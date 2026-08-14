package telegrambotconnector

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/collectors-tech/cabinet/internal/telegrambotadapter"
)

const (
	defaultBotAPIBaseURL = "https://api.telegram.org"
	defaultPollTimeout   = 25
)

var requiredUpdateTypes = []string{"message", "callback_query"}

func RequiredUpdateTypes() []string {
	return append([]string(nil), requiredUpdateTypes...)
}

type BotUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

type WebhookInfo struct {
	URL                string `json:"url"`
	PendingUpdateCount int    `json:"pending_update_count"`
	LastErrorDate      int64  `json:"last_error_date,omitempty"`
	LastErrorMessage   string `json:"last_error_message,omitempty"`
	MaxConnections     int    `json:"max_connections,omitempty"`
	IP                 string `json:"ip_address,omitempty"`
}

func (w WebhookInfo) ConflictsWithPolling() bool {
	return strings.TrimSpace(w.URL) != ""
}

type APIError struct {
	Code        int
	Description string
	RetryAfter  time.Duration
}

func (e *APIError) Error() string {
	if e == nil {
		return "telegram bot api error"
	}
	description := strings.TrimSpace(e.Description)
	if description == "" {
		description = "request failed"
	}
	return fmt.Sprintf("telegram bot api error %d: %s", e.Code, description)
}

type Client struct {
	baseURL    string
	token      string
	httpClient telegrambotadapter.HTTPDoer
}

func NewClient(baseURL, token string, client telegrambotadapter.HTTPDoer) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultBotAPIBaseURL
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{baseURL: baseURL, token: strings.TrimSpace(token), httpClient: client}
}

func (c *Client) SafeEndpoint() string {
	if c == nil {
		return defaultBotAPIBaseURL
	}
	return c.baseURL
}

func (c *Client) GetMe(ctx context.Context) (BotUser, error) {
	var result BotUser
	err := c.call(ctx, "getMe", map[string]any{}, &result)
	return result, err
}

func (c *Client) GetWebhookInfo(ctx context.Context) (WebhookInfo, error) {
	var result WebhookInfo
	err := c.call(ctx, "getWebhookInfo", map[string]any{}, &result)
	return result, err
}

func (c *Client) DeleteWebhook(ctx context.Context, dropPendingUpdates bool) error {
	var deleted bool
	if err := c.call(ctx, "deleteWebhook", map[string]any{"drop_pending_updates": dropPendingUpdates}, &deleted); err != nil {
		return err
	}
	if !deleted {
		return errors.New("telegram did not confirm webhook removal")
	}
	return nil
}

func (c *Client) GetUpdates(ctx context.Context, offset int64, timeout int, allowed []string) ([]telegrambotadapter.Update, error) {
	if timeout <= 0 {
		timeout = defaultPollTimeout
	}
	if len(allowed) == 0 {
		allowed = RequiredUpdateTypes()
	}
	var result []telegrambotadapter.Update
	err := c.call(ctx, "getUpdates", map[string]any{
		"offset":          offset,
		"timeout":         timeout,
		"allowed_updates": allowed,
	}, &result)
	return result, err
}

func (c *Client) call(ctx context.Context, method string, body any, result any) error {
	if c == nil || strings.TrimSpace(c.token) == "" {
		return errors.New("telegram bot token is required")
	}
	endpoint := telegrambotadapter.BotAPIEndpoint{BaseURL: c.baseURL, Token: c.token}
	req, err := endpoint.NewRequest(ctx, telegrambotadapter.BotAPICall{Method: method, Body: body})
	if err != nil {
		return errors.New("could not prepare Telegram Bot API request")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram bot api unavailable: %w", sanitizeError(err, c.token))
	}
	defer resp.Body.Close()
	var envelope struct {
		OK          bool            `json:"ok"`
		Result      json.RawMessage `json:"result"`
		ErrorCode   int             `json:"error_code,omitempty"`
		Description string          `json:"description,omitempty"`
		Parameters  struct {
			RetryAfter int `json:"retry_after,omitempty"`
		} `json:"parameters,omitempty"`
	}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 4<<20))
	if err := decoder.Decode(&envelope); err != nil {
		return fmt.Errorf("telegram bot api returned an invalid response for %s", method)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || !envelope.OK {
		code := envelope.ErrorCode
		if code == 0 {
			code = resp.StatusCode
		}
		return &APIError{
			Code:        code,
			Description: sanitizeText(envelope.Description, c.token),
			RetryAfter:  time.Duration(envelope.Parameters.RetryAfter) * time.Second,
		}
	}
	if result == nil || len(envelope.Result) == 0 || string(envelope.Result) == "null" {
		return nil
	}
	if err := json.Unmarshal(envelope.Result, result); err != nil {
		return fmt.Errorf("telegram bot api returned an invalid %s result", method)
	}
	return nil
}

func sanitizeError(err error, token string) error {
	if err == nil {
		return nil
	}
	return errors.New(sanitizeText(err.Error(), token))
}

func sanitizeText(value, token string) string {
	value = strings.TrimSpace(value)
	if strings.TrimSpace(token) != "" {
		value = strings.ReplaceAll(value, token, "[redacted]")
	}
	return value
}

type BotUpdateClient interface {
	GetUpdates(ctx context.Context, offset int64, timeout int, allowed []string) ([]telegrambotadapter.Update, error)
}

type OffsetStore interface {
	LoadOffset(ctx context.Context, profileID string) (int64, error)
	SaveOffset(ctx context.Context, profileID string, offset int64) error
}

type ProcessUpdate func(ctx context.Context, update telegrambotadapter.Update) error

type processedUpdateError struct{ err error }

func (e *processedUpdateError) Error() string { return e.err.Error() }
func (e *processedUpdateError) Unwrap() error { return e.err }

// MarkProcessed reports a post-processing failure, such as reply delivery,
// while preserving the exactly-once governed handoff boundary for the update.
func MarkProcessed(err error) error {
	if err == nil {
		return nil
	}
	return &processedUpdateError{err: err}
}

type PollResult struct {
	Received         int
	Processed        int
	SkippedDuplicate int
	Offset           int64
}

type Poller struct {
	profileID  string
	client     BotUpdateClient
	store      OffsetStore
	process    ProcessUpdate
	wait       func(context.Context, time.Duration) error
	minBackoff time.Duration
	maxBackoff time.Duration
	idleDelay  time.Duration
	timeout    int
}

func NewPoller(profileID string, client BotUpdateClient, store OffsetStore, process ProcessUpdate) *Poller {
	return &Poller{
		profileID:  strings.TrimSpace(profileID),
		client:     client,
		store:      store,
		process:    process,
		wait:       waitContext,
		minBackoff: 500 * time.Millisecond,
		maxBackoff: 30 * time.Second,
		idleDelay:  100 * time.Millisecond,
		timeout:    defaultPollTimeout,
	}
}

func (p *Poller) SetBackoff(minimum, maximum time.Duration) {
	if minimum > 0 {
		p.minBackoff = minimum
	}
	if maximum >= p.minBackoff {
		p.maxBackoff = maximum
	}
}

func (p *Poller) SetWait(wait func(context.Context, time.Duration) error) {
	if wait != nil {
		p.wait = wait
	}
}

func (p *Poller) PollOnce(ctx context.Context) (PollResult, error) {
	if p == nil || p.client == nil || p.store == nil || p.process == nil || p.profileID == "" {
		return PollResult{}, errors.New("telegram poller is not configured")
	}
	offset, err := p.store.LoadOffset(ctx, p.profileID)
	if err != nil {
		return PollResult{}, fmt.Errorf("load telegram update offset: %w", err)
	}
	updates, err := p.client.GetUpdates(ctx, offset, p.timeout, RequiredUpdateTypes())
	if err != nil {
		return PollResult{Offset: offset}, err
	}
	for _, update := range updates {
		if update.UpdateID <= 0 {
			return PollResult{Received: len(updates), Offset: offset}, errors.New("telegram returned a malformed update without a positive update id")
		}
	}
	sort.SliceStable(updates, func(i, j int) bool { return updates[i].UpdateID < updates[j].UpdateID })
	result := PollResult{Received: len(updates), Offset: offset}
	for _, update := range updates {
		if update.UpdateID < offset {
			result.SkippedDuplicate++
			continue
		}
		processErr := p.process(ctx, update)
		var processedErr *processedUpdateError
		if processErr != nil && !errors.As(processErr, &processedErr) {
			return result, fmt.Errorf("process telegram update %d: %w", update.UpdateID, processErr)
		}
		next := update.UpdateID + 1
		if err := p.store.SaveOffset(ctx, p.profileID, next); err != nil {
			return result, fmt.Errorf("persist telegram update offset %d: %w", next, err)
		}
		offset = next
		result.Offset = next
		result.Processed++
		if processErr != nil {
			return result, fmt.Errorf("telegram update %d processed but follow-up failed: %w", update.UpdateID, processErr)
		}
	}
	return result, nil
}

func (p *Poller) Run(ctx context.Context) error {
	failures := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		result, err := p.PollOnce(ctx)
		if err != nil {
			failures++
			delay := p.backoff(failures)
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.RetryAfter > 0 {
				delay = apiErr.RetryAfter
				if delay > p.maxBackoff {
					delay = p.maxBackoff
				}
			}
			if waitErr := p.wait(ctx, delay); waitErr != nil {
				return waitErr
			}
			continue
		}
		failures = 0
		if result.Received == 0 {
			if err := p.wait(ctx, p.idleDelay); err != nil {
				return err
			}
		}
	}
}

func (p *Poller) backoff(failures int) time.Duration {
	if failures < 1 {
		failures = 1
	}
	delay := p.minBackoff
	for i := 1; i < failures; i++ {
		if delay >= p.maxBackoff/2 {
			return p.maxBackoff
		}
		delay *= 2
	}
	if delay > p.maxBackoff {
		return p.maxBackoff
	}
	return delay
}

func waitContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

var (
	ErrPairingCodeInvalid  = errors.New("telegram pairing code is invalid")
	ErrPairingCodeExpired  = errors.New("telegram pairing code expired")
	ErrPairingCodeUsed     = errors.New("telegram pairing code already used")
	ErrPrivateChatRequired = errors.New("telegram pairing requires the same sender in a private chat")
)

type PairingRequest struct {
	ProfileID  string
	CodeHash   string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	ConsumedAt time.Time
}

type PairingCode struct {
	ProfileID string    `json:"profile_id"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PairingResult struct {
	ProfileID string `json:"profile_id"`
	SenderID  string `json:"sender_id"`
	ChatID    string `json:"chat_id"`
}

type PairingStore interface {
	CreatePairing(ctx context.Context, request PairingRequest) error
	ConsumePairing(ctx context.Context, codeHash string, at time.Time) (PairingRequest, error)
}

type PairingService struct {
	store    PairingStore
	now      func() time.Time
	generate func() (string, error)
}

func NewPairingService(store PairingStore) *PairingService {
	return &PairingService{store: store, now: time.Now, generate: generatePairingCode}
}

func (s *PairingService) SetClock(now func() time.Time) {
	if now != nil {
		s.now = now
	}
}

func (s *PairingService) SetCodeGenerator(generate func() (string, error)) {
	if generate != nil {
		s.generate = generate
	}
}

func (s *PairingService) Create(ctx context.Context, profileID string, ttl time.Duration) (PairingCode, error) {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" || s == nil || s.store == nil {
		return PairingCode{}, errors.New("telegram pairing profile is required")
	}
	if ttl <= 0 || ttl > 15*time.Minute {
		ttl = 5 * time.Minute
	}
	code, err := s.generate()
	if err != nil {
		return PairingCode{}, err
	}
	code = normalizePairingCode(code)
	createdAt := s.now().UTC()
	request := PairingRequest{ProfileID: profileID, CodeHash: hashPairingCode(code), CreatedAt: createdAt, ExpiresAt: createdAt.Add(ttl)}
	if err := s.store.CreatePairing(ctx, request); err != nil {
		return PairingCode{}, err
	}
	return PairingCode{ProfileID: profileID, Code: code, ExpiresAt: request.ExpiresAt}, nil
}

func (s *PairingService) Consume(ctx context.Context, update telegrambotadapter.Update) (PairingResult, error) {
	if s == nil || s.store == nil || update.Message == nil {
		return PairingResult{}, ErrPairingCodeInvalid
	}
	message := update.Message
	if !strings.EqualFold(strings.TrimSpace(message.Chat.Type), "private") || message.From.ID == 0 || message.Chat.ID == 0 || message.From.ID != message.Chat.ID {
		return PairingResult{}, ErrPrivateChatRequired
	}
	code, ok := pairingCodeFromStart(message.Text)
	if !ok {
		return PairingResult{}, ErrPairingCodeInvalid
	}
	request, err := s.store.ConsumePairing(ctx, hashPairingCode(code), s.now().UTC())
	if err != nil {
		return PairingResult{}, err
	}
	return PairingResult{ProfileID: request.ProfileID, SenderID: fmt.Sprintf("%d", message.From.ID), ChatID: fmt.Sprintf("%d", message.Chat.ID)}, nil
}

func pairingCodeFromStart(text string) (string, bool) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) != 2 || !strings.EqualFold(fields[0], "/start") {
		return "", false
	}
	code := normalizePairingCode(fields[1])
	return code, strings.HasPrefix(code, "CAB-")
}

func normalizePairingCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func hashPairingCode(code string) string {
	sum := sha256.Sum256([]byte(normalizePairingCode(code)))
	return hex.EncodeToString(sum[:])
}

func generatePairingCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	for i := range raw {
		raw[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return "CAB-" + string(raw[:4]) + "-" + string(raw[4:]), nil
}

type MemoryPairingStore struct {
	mu       sync.Mutex
	requests map[string]PairingRequest
}

func NewMemoryPairingStore() *MemoryPairingStore {
	return &MemoryPairingStore{requests: map[string]PairingRequest{}}
}

func (s *MemoryPairingStore) CreatePairing(_ context.Context, request PairingRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.requests[request.CodeHash]; exists {
		return errors.New("telegram pairing code collision")
	}
	s.requests[request.CodeHash] = request
	return nil
}

func (s *MemoryPairingStore) ConsumePairing(_ context.Context, codeHash string, at time.Time) (PairingRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	request, ok := s.requests[codeHash]
	if !ok {
		return PairingRequest{}, ErrPairingCodeInvalid
	}
	if !request.ConsumedAt.IsZero() {
		return PairingRequest{}, ErrPairingCodeUsed
	}
	if !at.Before(request.ExpiresAt) {
		return PairingRequest{}, ErrPairingCodeExpired
	}
	request.ConsumedAt = at
	s.requests[codeHash] = request
	return request, nil
}

func (s *MemoryPairingStore) ContainsPlaintext(value string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, request := range s.requests {
		if strings.Contains(key, value) || strings.Contains(request.CodeHash, value) {
			return true
		}
	}
	return false
}
