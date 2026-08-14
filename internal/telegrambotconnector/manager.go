package telegrambotconnector

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/telegrambotadapter"
	"github.com/collectors-tech/cabinet/internal/telegramcapture"
)

const BotTokenSecretKey = "telegram_bot_token"

var (
	ErrWebhookConflict = errors.New("telegram webhook conflicts with long polling")
	ErrConnectorPaused = errors.New("telegram connector is paused")
)

type ConnectionTestResult struct {
	ProfileID          string  `json:"profile_id"`
	Status             string  `json:"status"`
	Code               string  `json:"code"`
	Message            string  `json:"message"`
	NextAction         string  `json:"next_action"`
	Bot                BotUser `json:"bot"`
	WebhookConflict    bool    `json:"webhook_conflict"`
	PendingUpdateCount int     `json:"pending_update_count"`
	BotTokenPresent    bool    `json:"bot_token_present"`
	CredentialReturned bool    `json:"credential_returned"`
	Transport          string  `json:"transport"`
	PublicListener     bool    `json:"public_listener"`
}

type ConnectionStatus struct {
	ProfileID          string `json:"profile_id"`
	Status             string `json:"status"`
	State              string `json:"state"`
	Code               string `json:"code"`
	Message            string `json:"message"`
	NextAction         string `json:"next_action"`
	Transport          string `json:"transport"`
	PublicListener     bool   `json:"public_listener"`
	BotTokenPresent    bool   `json:"bot_token_present"`
	BotValidated       bool   `json:"bot_validated"`
	CredentialReturned bool   `json:"credential_returned"`
	BotUsername        string `json:"bot_username,omitempty"`
	SenderID           string `json:"sender_id,omitempty"`
	ChatID             string `json:"chat_id,omitempty"`
	Paired             bool   `json:"paired"`
	Paused             bool   `json:"paused"`
	WebhookConflict    bool   `json:"webhook_conflict"`
	Offset             int64  `json:"offset"`
	LastSuccessAt      string `json:"last_success_at,omitempty"`
	LastUpdateID       string `json:"last_update_id,omitempty"`
	LastErrorCode      string `json:"last_error_code,omitempty"`
	RetryAfterSeconds  int    `json:"retry_after_seconds,omitempty"`
}

type Manager struct {
	db         *sql.DB
	profiles   *profile.Repository
	gateway    telegrambotadapter.CabinetGateway
	baseURL    string
	httpClient telegrambotadapter.HTTPDoer
	store      *SQLStore
	pairing    *PairingService
}

func NewManager(db *sql.DB, profiles *profile.Repository, gateway telegrambotadapter.CabinetGateway, baseURL string, httpClient telegrambotadapter.HTTPDoer) *Manager {
	store := NewSQLStore(db)
	return &Manager{
		db:         db,
		profiles:   profiles,
		gateway:    gateway,
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: httpClient,
		store:      store,
		pairing:    NewPairingService(store),
	}
}

func (m *Manager) TestConnection(ctx context.Context, profileID, candidateToken string) (ConnectionTestResult, error) {
	result := baseConnectionTestResult(profileID)
	if err := m.validateProfile(ctx, profileID); err != nil {
		result.Code = "TELEGRAM_PROFILE_REQUIRED"
		result.Message = "Select a Cabinet profile before connecting Telegram."
		result.NextAction = "select_profile"
		return result, err
	}
	token := strings.TrimSpace(candidateToken)
	if token == "" {
		var err error
		token, err = m.profiles.GetSecret(ctx, profileID, BotTokenSecretKey)
		if err != nil || strings.TrimSpace(token) == "" {
			result.Code = "TELEGRAM_BOT_TOKEN_REQUIRED"
			result.Message = "Paste the BotFather token to test this Telegram bot."
			result.NextAction = "store_bot_token_secret"
			return result, errors.New("telegram bot token is required")
		}
	}
	client := NewClient(m.baseURL, token, m.httpClient)
	bot, err := client.GetMe(ctx)
	if err != nil {
		result.Status = "failed"
		result.Code = telegramConnectionErrorCode(err)
		result.Message = telegramConnectionErrorMessage(err)
		result.NextAction = "review_bot_token"
		return result, err
	}
	if !bot.IsBot {
		result.Status = "failed"
		result.Code = "TELEGRAM_BOT_IDENTITY_INVALID"
		result.Message = "Telegram accepted the credential, but it does not identify a bot."
		result.NextAction = "review_bot_token"
		return result, errors.New("telegram credential does not identify a bot")
	}
	existingSettings, settingsErr := m.profiles.GetSettings(ctx, profileID)
	if settingsErr != nil {
		return result, settingsErr
	}
	previousBotID := strings.TrimSpace(existingSettings["telegram.bot_id"])
	rotatedToDifferentBot := previousBotID != "" && previousBotID != strconv.FormatInt(bot.ID, 10)
	if strings.TrimSpace(candidateToken) != "" {
		if err := m.profiles.PutSecret(ctx, profileID, BotTokenSecretKey, token); err != nil {
			return result, err
		}
	}
	if rotatedToDifferentBot {
		if err := m.store.ResetProfile(ctx, profileID); err != nil {
			return result, err
		}
		if err := m.profiles.PutSettings(ctx, profileID, map[string]string{
			"telegram.catalog_capture.sender_id": "",
			"telegram.catalog_capture.chat_id":   "",
			"telegram.polling.paired_at":         "",
			"telegram.polling.last_success_at":   "",
			"telegram.polling.last_update_id":    "",
		}); err != nil {
			return result, err
		}
	}
	result.Bot = bot
	result.BotTokenPresent = true
	settings := map[string]string{
		"telegram.bot_username":             strings.TrimSpace(bot.Username),
		"telegram.bot_id":                   strconv.FormatInt(bot.ID, 10),
		"telegram.bot_token_secret_present": "true",
		"telegram.polling.transport":        "long_polling",
		"telegram.polling.public_listener":  "false",
		"telegram.polling.tested_at":        time.Now().UTC().Format(time.RFC3339),
	}
	webhook, err := client.GetWebhookInfo(ctx)
	if err != nil {
		result.Status = "failed"
		result.Code = telegramConnectionErrorCode(err)
		result.Message = telegramConnectionErrorMessage(err)
		result.NextAction = "retry_connection_test"
		_ = m.profiles.PutSettings(ctx, profileID, settings)
		return result, err
	}
	result.WebhookConflict = webhook.ConflictsWithPolling()
	result.PendingUpdateCount = webhook.PendingUpdateCount
	if result.WebhookConflict {
		settings["telegram.polling.webhook_conflict"] = "true"
		_ = m.profiles.PutSettings(ctx, profileID, settings)
		result.Status = "blocked"
		result.Code = "TELEGRAM_WEBHOOK_CONFLICT"
		result.Message = "This bot currently has a webhook. Remove it explicitly before Cabinet starts outbound-only polling."
		result.NextAction = "resolve_webhook_conflict"
		return result, ErrWebhookConflict
	}
	settings["telegram.polling.webhook_conflict"] = "false"
	settings["telegram.polling.enabled"] = "true"
	if err := m.profiles.PutSettings(ctx, profileID, settings); err != nil {
		return result, err
	}
	result.Status = "ready"
	result.Code = "TELEGRAM_BOT_VALIDATED"
	result.Message = "Telegram bot validated for outbound-only long polling. Pair a private chat to continue."
	result.NextAction = "create_pairing_code"
	return result, nil
}

func (m *Manager) ResolveWebhookConflict(ctx context.Context, profileID string) (ConnectionTestResult, error) {
	result := baseConnectionTestResult(profileID)
	token, err := m.token(ctx, profileID)
	if err != nil {
		return result, err
	}
	client := NewClient(m.baseURL, token, m.httpClient)
	if err := client.DeleteWebhook(ctx, false); err != nil {
		return result, err
	}
	result, err = m.TestConnection(ctx, profileID, "")
	if err == nil {
		result.Code = "TELEGRAM_WEBHOOK_REMOVED"
		result.Message = "Telegram webhook removed. Cabinet can now use outbound-only long polling."
		result.NextAction = "create_pairing_code"
	}
	return result, err
}

func (m *Manager) CreatePairing(ctx context.Context, profileID string) (PairingCode, error) {
	if _, err := m.token(ctx, profileID); err != nil {
		return PairingCode{}, err
	}
	settings, err := m.profiles.GetSettings(ctx, profileID)
	if err != nil {
		return PairingCode{}, err
	}
	if strings.EqualFold(settings["telegram.polling.webhook_conflict"], "true") {
		return PairingCode{}, ErrWebhookConflict
	}
	if strings.TrimSpace(settings["telegram.bot_id"]) == "" || strings.TrimSpace(settings["telegram.polling.tested_at"]) == "" {
		return PairingCode{}, errors.New("telegram bot must pass getMe validation before pairing")
	}
	code, err := m.pairing.Create(ctx, profileID, 5*time.Minute)
	if err != nil {
		return PairingCode{}, err
	}
	if err := m.profiles.PutSettings(ctx, profileID, map[string]string{
		"telegram.polling.enabled":     "true",
		"telegram.polling.paused":      "false",
		"integration.telegram.enabled": "true",
	}); err != nil {
		return PairingCode{}, err
	}
	return code, nil
}

func (m *Manager) PollOnce(ctx context.Context, profileID string) (PollResult, error) {
	settings, err := m.profiles.GetSettings(ctx, profileID)
	if err != nil {
		return PollResult{}, err
	}
	if strings.EqualFold(settings["telegram.polling.paused"], "true") {
		return PollResult{}, ErrConnectorPaused
	}
	if strings.EqualFold(settings["telegram.polling.webhook_conflict"], "true") {
		return PollResult{}, ErrWebhookConflict
	}
	token, err := m.token(ctx, profileID)
	if err != nil {
		return PollResult{}, err
	}
	client := NewClient(m.baseURL, token, m.httpClient)
	poller := NewPoller(profileID, client, m.store, func(updateCtx context.Context, update telegrambotadapter.Update) error {
		return m.processUpdate(updateCtx, profileID, token, update)
	})
	result, err := poller.PollOnce(ctx)
	if err != nil {
		m.recordPollError(ctx, profileID, err)
		return result, err
	}
	values := map[string]string{
		"telegram.polling.state":               "running",
		"telegram.polling.last_success_at":     time.Now().UTC().Format(time.RFC3339),
		"telegram.polling.last_error_code":     "",
		"telegram.polling.retry_after_seconds": "0",
	}
	if result.Offset > 0 {
		values["telegram.polling.last_update_id"] = strconv.FormatInt(result.Offset-1, 10)
	}
	_ = m.profiles.PutSettings(ctx, profileID, values)
	return result, nil
}

func (m *Manager) processUpdate(ctx context.Context, profileID, token string, update telegrambotadapter.Update) error {
	if update.Message == nil && update.CallbackQuery == nil {
		return nil
	}
	if update.CallbackQuery != nil && update.CallbackQuery.Message == nil {
		return nil
	}
	if update.Message != nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(update.Message.Text)), "/start") {
		paired, err := m.pairing.Consume(ctx, update)
		if err != nil {
			return m.sendPairingReply(ctx, token, update.Message.Chat.ID, pairingFailureMessage(err))
		}
		if paired.ProfileID != profileID {
			return m.sendPairingReply(ctx, token, update.Message.Chat.ID, "This pairing code belongs to a different Cabinet profile.")
		}
		if err := m.profiles.PutSettings(ctx, profileID, map[string]string{
			"telegram.catalog_capture.sender_id": paired.SenderID,
			"telegram.catalog_capture.chat_id":   paired.ChatID,
			"telegram.polling.paired_at":         time.Now().UTC().Format(time.RFC3339),
			"telegram.polling.paused":            "false",
			"integration.telegram.enabled":       "true",
		}); err != nil {
			return err
		}
		return m.sendPairingReply(ctx, token, update.Message.Chat.ID, "Cabinet is connected to this private Telegram chat. You can now talk to Cabinet Agent here.")
	}
	if m.gateway == nil {
		return errors.New("Cabinet Telegram gateway is unavailable")
	}
	ctx = withInProcessProfile(ctx, profileID)
	result, err := telegrambotadapter.DispatchUpdate(ctx, m.gateway, update)
	if err != nil {
		return sanitizeError(err, token)
	}
	deliveryErr := telegrambotadapter.ExecuteBotAPICalls(ctx, &result, telegrambotadapter.BotAPIEndpoint{BaseURL: m.baseURL, Token: token}, m.httpClient)
	if strings.TrimSpace(result.CabinetError) != "" {
		return errors.New("Cabinet governed Telegram processing failed")
	}
	if deliveryErr != nil {
		return MarkProcessed(sanitizeError(deliveryErr, token))
	}
	return nil
}

func (m *Manager) sendPairingReply(ctx context.Context, token string, chatID int64, text string) error {
	call, err := telegrambotadapter.SendMessageFromReply(strconv.FormatInt(chatID, 10), telegramcapture.TelegramReply{Text: text, ConfirmationState: "informational"})
	if err != nil {
		return err
	}
	req, err := (telegrambotadapter.BotAPIEndpoint{BaseURL: m.baseURL, Token: token}).NewRequest(ctx, call)
	if err != nil {
		return err
	}
	client := m.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return sanitizeError(err, token)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram sendMessage returned status %d", resp.StatusCode)
	}
	return nil
}

func (m *Manager) Status(ctx context.Context, profileID string) (ConnectionStatus, error) {
	status := ConnectionStatus{ProfileID: profileID, Transport: "long_polling", PublicListener: false, CredentialReturned: false}
	if err := m.validateProfile(ctx, profileID); err != nil {
		return status, err
	}
	settings, err := m.profiles.GetSettings(ctx, profileID)
	if err != nil {
		return status, err
	}
	token, tokenErr := m.profiles.GetSecret(ctx, profileID, BotTokenSecretKey)
	status.BotTokenPresent = tokenErr == nil && strings.TrimSpace(token) != ""
	status.BotValidated = strings.TrimSpace(settings["telegram.bot_id"]) != "" && strings.TrimSpace(settings["telegram.polling.tested_at"]) != ""
	status.BotUsername = strings.TrimSpace(settings["telegram.bot_username"])
	status.SenderID = strings.TrimSpace(settings["telegram.catalog_capture.sender_id"])
	status.ChatID = strings.TrimSpace(settings["telegram.catalog_capture.chat_id"])
	status.Paired = status.SenderID != "" && status.ChatID != ""
	status.Paused = strings.EqualFold(settings["telegram.polling.paused"], "true")
	status.WebhookConflict = strings.EqualFold(settings["telegram.polling.webhook_conflict"], "true")
	status.LastSuccessAt = strings.TrimSpace(settings["telegram.polling.last_success_at"])
	status.LastUpdateID = strings.TrimSpace(settings["telegram.polling.last_update_id"])
	status.LastErrorCode = strings.TrimSpace(settings["telegram.polling.last_error_code"])
	status.RetryAfterSeconds, _ = strconv.Atoi(strings.TrimSpace(settings["telegram.polling.retry_after_seconds"]))
	status.Offset, _ = m.store.LoadOffset(ctx, profileID)
	switch {
	case !status.BotTokenPresent:
		status.Status, status.State, status.Code = "needs_config", "disabled", "TELEGRAM_BOT_TOKEN_REQUIRED"
		status.Message, status.NextAction = "Paste and validate a BotFather token.", "test_connection"
	case !status.BotValidated:
		status.Status, status.State, status.Code = "needs_config", "validation_required", "TELEGRAM_BOT_VALIDATION_REQUIRED"
		status.Message, status.NextAction = "Validate this BotFather token with Telegram before pairing.", "test_connection"
	case status.WebhookConflict:
		status.Status, status.State, status.Code = "blocked", "webhook_conflict", "TELEGRAM_WEBHOOK_CONFLICT"
		status.Message, status.NextAction = "Remove the existing Telegram webhook explicitly before polling.", "resolve_webhook_conflict"
	case !status.Paired:
		status.Status, status.State, status.Code = "needs_config", "pairing_required", "TELEGRAM_PAIRING_REQUIRED"
		status.Message, status.NextAction = "Create a short-lived pairing code and send it to the bot from a private chat.", "create_pairing_code"
	case status.Paused:
		status.Status, status.State, status.Code = "paused", "paused", "TELEGRAM_POLLING_PAUSED"
		status.Message, status.NextAction = "Telegram polling is paused.", "resume_polling"
	default:
		status.Status, status.State, status.Code = "ok", "connected", "TELEGRAM_POLLING_READY"
		status.Message, status.NextAction = "Telegram is paired for outbound-only long polling.", "talk_to_cabinet"
	}
	return status, nil
}

func (m *Manager) SetPaused(ctx context.Context, profileID string, paused bool) error {
	value := "false"
	state := "running"
	if paused {
		value, state = "true", "paused"
	}
	return m.profiles.PutSettings(ctx, profileID, map[string]string{"telegram.polling.paused": value, "telegram.polling.state": state})
}

func (m *Manager) Disconnect(ctx context.Context, profileID string) error {
	if err := m.profiles.DeleteSecret(ctx, profileID, BotTokenSecretKey); err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		return err
	}
	if err := m.store.ResetProfile(ctx, profileID); err != nil {
		return err
	}
	return m.profiles.PutSettings(ctx, profileID, map[string]string{
		"telegram.bot_token_secret_present":    "false",
		"telegram.bot_username":                "",
		"telegram.bot_id":                      "",
		"telegram.catalog_capture.sender_id":   "",
		"telegram.catalog_capture.chat_id":     "",
		"telegram.polling.enabled":             "false",
		"telegram.polling.paused":              "true",
		"telegram.polling.state":               "disconnected",
		"telegram.polling.tested_at":           "",
		"telegram.polling.paired_at":           "",
		"telegram.polling.webhook_conflict":    "false",
		"telegram.polling.last_success_at":     "",
		"telegram.polling.last_update_id":      "",
		"telegram.polling.last_error_code":     "",
		"telegram.polling.retry_after_seconds": "0",
		"integration.telegram.enabled":         "false",
	})
}

func (m *Manager) Run(ctx context.Context) error {
	backoff := time.Second
	const maxBackoff = 30 * time.Second
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		profileIDs, err := m.enabledProfileIDs(ctx)
		if err == nil && len(profileIDs) == 0 {
			err = errors.New("telegram bot token is not configured")
		}
		for _, profileID := range profileIDs {
			if _, pollErr := m.PollOnce(ctx, profileID); pollErr != nil {
				err = pollErr
				break
			}
			err = nil
		}
		if err != nil {
			if errors.Is(err, ErrConnectorPaused) || errors.Is(err, ErrWebhookConflict) || strings.Contains(err.Error(), "token") {
				backoff = 2 * time.Second
			} else if backoff < maxBackoff {
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
		} else {
			backoff = 100 * time.Millisecond
		}
		delay := managerRetryDelay(err, backoff, maxBackoff)
		if err := waitContext(ctx, delay); err != nil {
			return err
		}
	}
}

func (m *Manager) enabledProfileIDs(ctx context.Context) ([]string, error) {
	profiles, err := m.profiles.List(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(profiles))
	for _, candidate := range profiles {
		settings, settingsErr := m.profiles.GetSettings(ctx, candidate.ID)
		if settingsErr != nil {
			return nil, settingsErr
		}
		if strings.EqualFold(strings.TrimSpace(settings["telegram.polling.enabled"]), "true") {
			ids = append(ids, candidate.ID)
		}
	}
	return ids, nil
}

func managerRetryDelay(err error, fallback, maximum time.Duration) time.Duration {
	delay := fallback
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.RetryAfter > 0 {
		delay = apiErr.RetryAfter
	}
	if maximum > 0 && delay > maximum {
		return maximum
	}
	return delay
}

func (m *Manager) validateProfile(ctx context.Context, profileID string) error {
	if m == nil || m.profiles == nil || strings.TrimSpace(profileID) == "" {
		return errors.New("Cabinet profile is required")
	}
	_, err := m.profiles.GetByID(ctx, profileID)
	return err
}

func (m *Manager) token(ctx context.Context, profileID string) (string, error) {
	if err := m.validateProfile(ctx, profileID); err != nil {
		return "", err
	}
	token, err := m.profiles.GetSecret(ctx, profileID, BotTokenSecretKey)
	if err != nil || strings.TrimSpace(token) == "" {
		return "", errors.New("telegram bot token is required")
	}
	return strings.TrimSpace(token), nil
}

func (m *Manager) recordPollError(ctx context.Context, profileID string, err error) {
	code := telegramConnectionErrorCode(err)
	retry := 0
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		retry = int(apiErr.RetryAfter.Seconds())
	}
	_ = m.profiles.PutSettings(ctx, profileID, map[string]string{
		"telegram.polling.state":               "degraded",
		"telegram.polling.last_error_code":     code,
		"telegram.polling.retry_after_seconds": strconv.Itoa(retry),
	})
}

func baseConnectionTestResult(profileID string) ConnectionTestResult {
	return ConnectionTestResult{ProfileID: strings.TrimSpace(profileID), Status: "needs_config", CredentialReturned: false, Transport: "long_polling", PublicListener: false}
}

func telegramConnectionErrorCode(err error) string {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case http.StatusUnauthorized:
			return "TELEGRAM_BOT_TOKEN_INVALID"
		case http.StatusTooManyRequests:
			return "TELEGRAM_RATE_LIMITED"
		}
	}
	return "TELEGRAM_PROVIDER_UNAVAILABLE"
}

func telegramConnectionErrorMessage(err error) string {
	switch telegramConnectionErrorCode(err) {
	case "TELEGRAM_BOT_TOKEN_INVALID":
		return "Telegram rejected this bot token. Replace it with a current BotFather token."
	case "TELEGRAM_RATE_LIMITED":
		return "Telegram rate-limited the connector. Cabinet will retry with bounded backoff."
	default:
		return "Cabinet could not reach Telegram. Check connectivity and retry."
	}
}

func pairingFailureMessage(err error) string {
	switch {
	case errors.Is(err, ErrPairingCodeExpired):
		return "This Cabinet pairing code expired. Create a new code in Cabinet."
	case errors.Is(err, ErrPairingCodeUsed):
		return "This Cabinet pairing code was already used. Create a new code to re-pair."
	case errors.Is(err, ErrPrivateChatRequired):
		return "Cabinet pairing only works in a private chat with this bot."
	default:
		return "This Cabinet pairing code is invalid. Create a new code in Cabinet."
	}
}

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(db *sql.DB) *SQLStore { return &SQLStore{db: db} }

func (s *SQLStore) LoadOffset(ctx context.Context, profileID string) (int64, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("telegram state store is unavailable")
	}
	var offset int64
	err := s.db.QueryRowContext(ctx, `SELECT update_offset FROM telegram_connector_state WHERE profile_id = ?`, profileID).Scan(&offset)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return offset, err
}

func (s *SQLStore) SaveOffset(ctx context.Context, profileID string, offset int64) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO telegram_connector_state(profile_id, update_offset, updated_at)
		VALUES(?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(profile_id) DO UPDATE SET update_offset=excluded.update_offset, updated_at=CURRENT_TIMESTAMP
	`, profileID, offset)
	return err
}

func (s *SQLStore) ResetProfile(ctx context.Context, profileID string) error {
	if s == nil || s.db == nil {
		return errors.New("telegram state store is unavailable")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM telegram_connector_state WHERE profile_id = ?`, profileID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE telegram_pairing_requests SET consumed_at = ? WHERE profile_id = ? AND consumed_at = ''`, time.Now().UTC().Format(time.RFC3339Nano), profileID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLStore) CreatePairing(ctx context.Context, request PairingRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `UPDATE telegram_pairing_requests SET consumed_at = ? WHERE profile_id = ? AND consumed_at = ''`, request.CreatedAt.Format(time.RFC3339Nano), request.ProfileID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO telegram_pairing_requests(code_hash, profile_id, expires_at, created_at, consumed_at) VALUES(?, ?, ?, ?, '')`, request.CodeHash, request.ProfileID, request.ExpiresAt.Format(time.RFC3339Nano), request.CreatedAt.Format(time.RFC3339Nano)); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLStore) ConsumePairing(ctx context.Context, codeHash string, at time.Time) (PairingRequest, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PairingRequest{}, err
	}
	defer tx.Rollback()
	var request PairingRequest
	var expiresAt, createdAt, consumedAt string
	err = tx.QueryRowContext(ctx, `SELECT profile_id, code_hash, expires_at, created_at, consumed_at FROM telegram_pairing_requests WHERE code_hash = ?`, codeHash).Scan(&request.ProfileID, &request.CodeHash, &expiresAt, &createdAt, &consumedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return PairingRequest{}, ErrPairingCodeInvalid
	}
	if err != nil {
		return PairingRequest{}, err
	}
	request.ExpiresAt, _ = time.Parse(time.RFC3339Nano, expiresAt)
	request.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	if consumedAt != "" {
		return PairingRequest{}, ErrPairingCodeUsed
	}
	if !at.Before(request.ExpiresAt) {
		return PairingRequest{}, ErrPairingCodeExpired
	}
	result, err := tx.ExecContext(ctx, `UPDATE telegram_pairing_requests SET consumed_at = ? WHERE code_hash = ? AND consumed_at = ''`, at.Format(time.RFC3339Nano), codeHash)
	if err != nil {
		return PairingRequest{}, err
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return PairingRequest{}, ErrPairingCodeUsed
	}
	request.ConsumedAt = at
	if err := tx.Commit(); err != nil {
		return PairingRequest{}, err
	}
	return request, nil
}

type InProcessGateway struct {
	Handler http.Handler
}

type inProcessRequestContextKey struct{}
type inProcessProfileContextKey struct{}

// IsInProcessRequest reports whether a request was constructed by Cabinet's
// outbound Telegram connector. The marker lives only in process context and
// cannot be forged by an HTTP header or request body.
func IsInProcessRequest(ctx context.Context) bool {
	value, _ := ctx.Value(inProcessRequestContextKey{}).(bool)
	return value
}

// WithInProcessRequest is intended for Cabinet's connector and white-box tests.
func WithInProcessRequest(ctx context.Context) context.Context {
	return context.WithValue(ctx, inProcessRequestContextKey{}, true)
}

func withInProcessProfile(ctx context.Context, profileID string) context.Context {
	return context.WithValue(ctx, inProcessProfileContextKey{}, strings.TrimSpace(profileID))
}

func InProcessProfile(ctx context.Context) (string, bool) {
	profileID, _ := ctx.Value(inProcessProfileContextKey{}).(string)
	profileID = strings.TrimSpace(profileID)
	return profileID, profileID != ""
}

func (g InProcessGateway) PostJSON(ctx context.Context, path string, body any, response any) (int, error) {
	if g.Handler == nil {
		return 0, errors.New("Cabinet handler is unavailable")
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(WithInProcessRequest(ctx), http.MethodPost, path, bytes.NewReader(raw))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	recorder := &gatewayResponse{header: make(http.Header), status: http.StatusOK}
	g.Handler.ServeHTTP(recorder, req)
	if len(recorder.body) > 0 && response != nil {
		_ = json.Unmarshal(recorder.body, response)
	}
	if recorder.status < 200 || recorder.status >= 300 {
		return recorder.status, fmt.Errorf("Cabinet Telegram dispatch returned status %d", recorder.status)
	}
	return recorder.status, nil
}

type gatewayResponse struct {
	header http.Header
	body   []byte
	status int
}

func (w *gatewayResponse) Header() http.Header    { return w.header }
func (w *gatewayResponse) WriteHeader(status int) { w.status = status }
func (w *gatewayResponse) Write(body []byte) (int, error) {
	w.body = append(w.body, body...)
	return len(body), nil
}
