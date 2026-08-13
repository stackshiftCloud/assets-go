package stackshift

import (
	"context"
	"net/http"
	"time"
)

type MailMessageStream struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Protected   bool      `json:"protected"`
	IPPoolID    string    `json:"ipPoolId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type MailMessageStreamCreateRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Type string `json:"type"`
}

type MailMessageStreamListResult struct {
	Data []MailMessageStream `json:"data"`
}

func (m *MailStreamsClient) Create(ctx context.Context, input MailMessageStreamCreateRequest) (*MailMessageStream, error) {
	var result MailMessageStream
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/streams", input, &result)
	return &result, err
}

func (m *MailStreamsClient) List(ctx context.Context) (*MailMessageStreamListResult, error) {
	var result MailMessageStreamListResult
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/streams", nil, &result)
	return &result, err
}
