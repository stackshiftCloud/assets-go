package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type MailAudience struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	MemberCount int       `json:"memberCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type MailAudienceMember struct {
	ID            string    `json:"id"`
	AudienceID    string    `json:"audienceId"`
	Email         string    `json:"email"`
	Name          string    `json:"name,omitempty"`
	Status        string    `json:"status"`
	ConsentSource string    `json:"consentSource"`
	ConsentAt     time.Time `json:"consentAt"`
	ConsentIP     string    `json:"consentIp,omitempty"`
}

type MailAudienceCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type MailAudienceMemberImport struct {
	Email         string    `json:"email"`
	Name          string    `json:"name,omitempty"`
	ConsentSource string    `json:"consentSource"`
	ConsentAt     time.Time `json:"consentAt"`
	ConsentIP     string    `json:"consentIp,omitempty"`
}

type MailCampaign struct {
	ID         string     `json:"id"`
	AudienceID string     `json:"audienceId"`
	BatchID    string     `json:"batchId,omitempty"`
	Name       string     `json:"name"`
	Template   string     `json:"template"`
	VersionID  string     `json:"versionId,omitempty"`
	From       string     `json:"from"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	QueuedAt   *time.Time `json:"queuedAt,omitempty"`
}

type MailCampaignCreateRequest struct {
	Name       string `json:"name"`
	AudienceID string `json:"audienceId"`
	Template   string `json:"template"`
	VersionID  string `json:"versionId,omitempty"`
	From       string `json:"from"`
}

type MailAudiencesClient struct{ client *Client }
type MailCampaignsClient struct{ client *Client }

func (m *MailAudiencesClient) Create(ctx context.Context, input MailAudienceCreateRequest) (*MailAudience, error) {
	var out MailAudience
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/audiences", input, &out)
	return &out, err
}

func (m *MailAudiencesClient) List(ctx context.Context) ([]MailAudience, error) {
	var out struct {
		Data []MailAudience `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/audiences", nil, &out)
	return out.Data, err
}

func (m *MailAudiencesClient) Get(ctx context.Context, id string) (*MailAudience, error) {
	var out MailAudience
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/audiences/"+url.PathEscape(id), nil, &out)
	return &out, err
}

func (m *MailAudiencesClient) Members(ctx context.Context, id string) ([]MailAudienceMember, error) {
	var out struct {
		Data []MailAudienceMember `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/audiences/"+url.PathEscape(id)+"/members", nil, &out)
	return out.Data, err
}

func (m *MailAudiencesClient) Import(ctx context.Context, id string, members []MailAudienceMemberImport) (int, int, error) {
	var out struct{ Imported, Skipped int }
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/audiences/"+url.PathEscape(id)+"/imports", map[string]any{"members": members}, &out)
	return out.Imported, out.Skipped, err
}

func (m *MailCampaignsClient) Create(ctx context.Context, input MailCampaignCreateRequest) (*MailCampaign, error) {
	var out MailCampaign
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/campaigns", input, &out)
	return &out, err
}

func (m *MailCampaignsClient) List(ctx context.Context) ([]MailCampaign, error) {
	var out struct {
		Data []MailCampaign `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/campaigns", nil, &out)
	return out.Data, err
}

func (m *MailCampaignsClient) Get(ctx context.Context, id string) (*MailCampaign, error) {
	var out MailCampaign
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/campaigns/"+url.PathEscape(id), nil, &out)
	return &out, err
}

func (m *MailCampaignsClient) Send(ctx context.Context, id string) (*MailCampaign, error) {
	var out MailCampaign
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/campaigns/"+url.PathEscape(id)+"/send", nil, &out)
	return &out, err
}
