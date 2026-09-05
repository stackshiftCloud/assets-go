package stackshift

import (
	"context"
	"net/http"
	"net/url"
)

type MailImportMode string

const (
	MailImportRecordedConsent MailImportMode = "recorded_consent"
	MailImportConfirmation    MailImportMode = "confirmation"
)

type MailSubscriberImportRequest struct {
	Mode    MailImportMode             `json:"mode"`
	Members []MailAudienceMemberImport `json:"members"`
}

type MailImportRow struct {
	Row     int    `json:"row"`
	Email   string `json:"email"`
	Outcome string `json:"outcome"`
	Code    string `json:"code"`
	Reason  string `json:"reason"`
	Status  string `json:"status,omitempty"`
}

type MailImportSummary struct {
	Imported int `json:"imported"`
	Updated  int `json:"updated"`
	Invalid  int `json:"invalid"`
	Skipped  int `json:"skipped"`
	Pending  int `json:"pending"`
}

type MailImportResult struct {
	Imported int               `json:"imported"`
	Skipped  int               `json:"skipped"`
	Summary  MailImportSummary `json:"summary"`
	Rows     []MailImportRow   `json:"rows"`
	Preview  bool              `json:"preview"`
}

func (m *MailAudiencesClient) PreviewImport(ctx context.Context, id string, input MailSubscriberImportRequest) (*MailImportResult, error) {
	var out MailImportResult
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/audiences/"+url.PathEscape(id)+"/imports/preview", input, &out)
	return &out, err
}

func (m *MailAudiencesClient) ImportDetailed(ctx context.Context, id string, input MailSubscriberImportRequest) (*MailImportResult, error) {
	var out MailImportResult
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/audiences/"+url.PathEscape(id)+"/imports", input, &out)
	return &out, err
}

func (m *MailAudiencesClient) SetEditableFields(ctx context.Context, id string, fields []string) error {
	if fields == nil {
		fields = []string{}
	}
	return m.client.mailRequest(ctx, http.MethodPut, "/mail/audiences/"+url.PathEscape(id)+"/preferences", map[string]any{"editableFields": fields}, nil)
}
