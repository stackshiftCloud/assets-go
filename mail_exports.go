package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type MailExport struct {
	ID          string     `json:"id"`
	Kind        string     `json:"kind"`
	Format      string     `json:"format"`
	Status      string     `json:"status"`
	From        time.Time  `json:"from"`
	To          time.Time  `json:"to"`
	RecordCount int64      `json:"recordCount"`
	SizeBytes   int64      `json:"sizeBytes"`
	Checksum    string     `json:"checksum,omitempty"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	ExpiresAt   time.Time  `json:"expiresAt"`
}

type MailExportCreateRequest struct {
	Kind   string     `json:"kind,omitempty"`
	Format string     `json:"format,omitempty"`
	From   *time.Time `json:"from,omitempty"`
	To     *time.Time `json:"to,omitempty"`
}

type MailExportsClient struct{ client *Client }

func (m *MailExportsClient) Create(ctx context.Context, input MailExportCreateRequest) (*MailExport, error) {
	var out MailExport
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/exports", input, &out)
	return &out, err
}

func (m *MailExportsClient) List(ctx context.Context) ([]MailExport, error) {
	var out struct {
		Data []MailExport `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/exports", nil, &out)
	return out.Data, err
}

func (m *MailExportsClient) Get(ctx context.Context, id string) (*MailExport, error) {
	var out MailExport
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/exports/"+url.PathEscape(id), nil, &out)
	return &out, err
}

func (m *MailExportsClient) DownloadPath(id string) string {
	return "/mail/exports/" + url.PathEscape(id) + "/download"
}
