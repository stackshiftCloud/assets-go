// Package supportactions verifies scoped Support action requests and signs receipts.
package supportactions

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

const RequestContext = "stackshift.support.action/v1"
const ReceiptContext = "stackshift.support.action-receipt/v1"
const RequestLimit = 256 * 1024
const ReceiptLimit = 64 * 1024

// ProtocolError contains a stable code, never payloads or credentials.
type ProtocolError string

func (e ProtocolError) Error() string { return string(e) }

var ErrFenceLost = errors.New("fence_lost")

type Field struct {
	Type      string   `json:"type"`
	Required  bool     `json:"required"`
	MaxLength *int     `json:"max_length,omitempty"`
	Minimum   *float64 `json:"minimum,omitempty"`
	Maximum   *float64 `json:"maximum,omitempty"`
	Enum      []string `json:"enum,omitempty"`
}
type Action struct {
	Name                   string           `json:"name"`
	Version                string           `json:"version"`
	Endpoint               string           `json:"endpoint"`
	ReconciliationEndpoint string           `json:"reconciliation_endpoint,omitempty"`
	Input                  map[string]Field `json:"input"`
	Output                 map[string]Field `json:"output"`
	Risk                   string           `json:"risk"`
	IdentityRequired       bool             `json:"identity_required"`
	IdempotencySupported   bool             `json:"idempotency_supported"`
	SafeRetry              bool             `json:"safe_retry"`
	TimeoutSeconds         int              `json:"timeout_seconds"`
	VerificationOperation  string           `json:"verification_operation,omitempty"`
}
type Manifest struct {
	Schema      string   `json:"schema"`
	Version     string   `json:"version"`
	Permissions []string `json:"permissions"`
	EventTypes  []string `json:"event_types"`
	Placements  []string `json:"placements"`
	WebhookURL  string   `json:"webhook_url"`
	Description string   `json:"description"`
	Actions     []Action `json:"actions"`
}
type ActionRef struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type Scope struct {
	WorkspaceID    string `json:"workspace_id"`
	EnvironmentID  string `json:"environment_id"`
	InboxID        string `json:"inbox_id"`
	ConversationID string `json:"conversation_id"`
	ContactID      string `json:"contact_id"`
}
type Customer struct {
	ExternalID string `json:"external_id"`
}
type Authorization struct {
	GrantID  string `json:"grant_id,omitempty"`
	Revision int64  `json:"revision"`
	Mode     string `json:"mode"`
}
type Request struct {
	Schema               string         `json:"schema"`
	RequestID            string         `json:"request_id"`
	InstallationID       string         `json:"installation_id"`
	InstallationRevision int64          `json:"installation_revision"`
	AppVersionID         string         `json:"app_version_id"`
	Action               ActionRef      `json:"action"`
	Scope                Scope          `json:"scope"`
	Customer             Customer       `json:"customer"`
	Arguments            map[string]any `json:"arguments"`
	Authorization        Authorization  `json:"authorization"`
	IssuedAt             int64          `json:"issued_at"`
	ExpiresAt            int64          `json:"expires_at"`
	Mode                 string         `json:"mode"`
	TargetRequestID      string         `json:"target_request_id,omitempty"`
}
type Receipt struct {
	Schema    string         `json:"schema"`
	RequestID string         `json:"request_id"`
	State     string         `json:"state"`
	Result    map[string]any `json:"result"`
	ErrorCode string         `json:"error_code,omitempty"`
}
type Verification struct {
	KeyID, Secret, InstallationID, ExpectedCustomerExternalID string
	Actions                                                   []Action
	Now                                                       time.Time
	Tolerance                                                 time.Duration
}
type Claim struct {
	State, Fence string
	Response     []byte
}

// DurableStore must commit claims before dispatch and never automatically reclaim
// running or unknown work. Complete is a fenced compare-and-set from running.
type DurableStore interface {
	Claim(context.Context, string, string, string) (Claim, error)
	Complete(context.Context, string, string, string, []byte, bool) error
}
type Handler func(context.Context, Request) (Receipt, error)
type SignedResponse struct {
	Body    json.RawMessage
	Headers map[string]string
}
