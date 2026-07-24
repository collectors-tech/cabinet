package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ReceiptSink interface {
	RecordMCPReceipt(ctx context.Context, receipt OperationReceipt)
}

type ReceiptSinkFunc func(ctx context.Context, receipt OperationReceipt)

func (f ReceiptSinkFunc) RecordMCPReceipt(ctx context.Context, receipt OperationReceipt) {
	f(ctx, receipt)
}

type OperationReceipt struct {
	OperationID   string `json:"operation_id"`
	SessionID     string `json:"session_id"`
	ProfileID     string `json:"profile_id"`
	ProfileLabel  string `json:"profile_label,omitempty"`
	ClientName    string `json:"client_name,omitempty"`
	ClientVersion string `json:"client_version,omitempty"`
	Capability    string `json:"capability"`
	Method        string `json:"method"`
	Version       string `json:"cabinet_version"`
	VersionDigest string `json:"cabinet_version_digest,omitempty"`
	InputClass    string `json:"input_class"`
	Outcome       string `json:"outcome"`
	ErrorClass    string `json:"error_class,omitempty"`
}

type receiptClientInfo struct {
	name    string
	version string
}

func buildReceipt(req mcp.Request, method string, err error, profileID, profileLabel, version, versionDigest, seed string, sequence uint64, clients *sync.Map) (OperationReceipt, bool) {
	capability, inputClass, material := classifyReceiptOperation(method, req.GetParams())
	if !material {
		return OperationReceipt{}, false
	}
	sessionID := sessionIDForReceipt(req.GetSession(), seed)
	client := clientInfoForReceipt(sessionID, method, req.GetParams(), clients)
	receipt := OperationReceipt{
		OperationID:   fmt.Sprintf("%s:%06d", operationIDSeed(sessionID, seed), sequence),
		SessionID:     sessionID,
		ProfileID:     profileID,
		ProfileLabel:  profileLabel,
		ClientName:    client.name,
		ClientVersion: client.version,
		Capability:    capability,
		Method:        method,
		Version:       version,
		VersionDigest: versionDigest,
		InputClass:    inputClass,
		Outcome:       "ok",
	}
	if err != nil {
		receipt.Outcome = "error"
		receipt.ErrorClass = errorClassForReceipt(err)
	}
	return receipt, true
}

func classifyReceiptOperation(method string, params mcp.Params) (capability string, inputClass string, material bool) {
	switch method {
	case "initialize":
		return "session.initialize", "client_metadata", true
	case "tools/call":
		name := "unknown"
		switch p := params.(type) {
		case *mcp.CallToolParams:
			name = strings.TrimSpace(p.Name)
		case *mcp.CallToolParamsRaw:
			name = strings.TrimSpace(p.Name)
		}
		if name == "" {
			name = "unknown"
		}
		return "tool:" + name, "tool_arguments", true
	case "resources/read":
		return "resource.read", "resource_locator", true
	case "prompts/get":
		return "prompt.get", "prompt_arguments", true
	default:
		return "", "", false
	}
}

func sessionIDForReceipt(session mcp.Session, seed string) string {
	if session != nil {
		if id := strings.TrimSpace(session.ID()); id != "" {
			return id
		}
	}
	if seed = strings.TrimSpace(seed); seed != "" {
		return seed
	}
	return "local-session"
}

func operationIDSeed(sessionID, seed string) string {
	if sessionID = strings.TrimSpace(sessionID); sessionID != "" {
		return sessionID
	}
	if seed = strings.TrimSpace(seed); seed != "" {
		return seed
	}
	return "mcp-operation"
}

func clientInfoForReceipt(sessionID, method string, params mcp.Params, clients *sync.Map) receiptClientInfo {
	if method == "initialize" {
		if p, ok := params.(*mcp.InitializeParams); ok && p.ClientInfo != nil {
			client := receiptClientInfo{
				name:    strings.TrimSpace(p.ClientInfo.Name),
				version: strings.TrimSpace(p.ClientInfo.Version),
			}
			clients.Store(sessionID, client)
			return client
		}
	}
	if v, ok := clients.Load(sessionID); ok {
		if client, ok := v.(receiptClientInfo); ok {
			return client
		}
	}
	return receiptClientInfo{}
}

func errorClassForReceipt(err error) string {
	switch {
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	default:
		return "protocol_error"
	}
}
