package stackshift

import (
	"context"
	"net/http"
)

type MailAutomationInput struct {
	AudienceID   string                 `json:"audienceId"`
	Name         string                 `json:"name"`
	Template     string                 `json:"template"`
	VersionID    string                 `json:"versionId"`
	From         string                 `json:"from"`
	DelaySeconds int                    `json:"delaySeconds"`
	Simulation   *MailSimulationOptions `json:"simulation,omitempty"`
}
type MailAutomation struct {
	TriggerType string `json:"triggerType"`
	RunCount    int    `json:"runCount"`
	WorkspaceID string `json:"workspaceId"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	MailAutomationInput
	ID      string   `json:"id"`
	BrandID string   `json:"brandId"`
	Mode    MailMode `json:"mode"`
	Status  string   `json:"status"`
}
type MailAutomationsClient struct{ client *Client }

func (m *MailAutomationsClient) Create(ctx context.Context, input MailAutomationInput) (*MailAutomation, error) {
	var out MailAutomation
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/automations", input, &out)
	return &out, err
}
func (m *MailAutomationsClient) List(ctx context.Context) ([]MailAutomation, error) {
	var out struct {
		Data []MailAutomation `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/automations", nil, &out)
	return out.Data, err
}

type MailSignupFormInput struct {
	AudienceID     string `json:"audienceId"`
	Name           string `json:"name"`
	Heading        string `json:"heading"`
	Description    string `json:"description"`
	ButtonLabel    string `json:"buttonLabel"`
	SuccessMessage string `json:"successMessage"`
}
type MailSignupForm struct {
	PublicToken string     `json:"publicToken"`
	Brand       *MailBrand `json:"brand,omitempty"`
	WorkspaceID string     `json:"workspaceId"`
	CreatedAt   string     `json:"createdAt"`
	UpdatedAt   string     `json:"updatedAt"`
	MailSignupFormInput
	ID      string   `json:"id"`
	BrandID string   `json:"brandId"`
	Mode    MailMode `json:"mode"`
	Status  string   `json:"status"`
}
type MailFormsClient struct{ client *Client }

func (m *MailFormsClient) Create(ctx context.Context, input MailSignupFormInput) (*MailSignupForm, error) {
	var out MailSignupForm
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/forms", input, &out)
	return &out, err
}
func (m *MailFormsClient) List(ctx context.Context) ([]MailSignupForm, error) {
	var out struct {
		Data []MailSignupForm `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/forms", nil, &out)
	return out.Data, err
}

type MailSegmentCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value,omitempty"`
}

type MailSegmentInput struct {
	Match       string                 `json:"match,omitempty"`
	Conditions  []MailSegmentCondition `json:"conditions,omitempty"`
	AudienceID  string                 `json:"audienceId"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Field       string                 `json:"field"`
	Operator    string                 `json:"operator"`
	Value       string                 `json:"value"`
}
type MailSegment struct {
	MemberCount int    `json:"memberCount"`
	WorkspaceID string `json:"workspaceId"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	MailSegmentInput
	ID      string   `json:"id"`
	BrandID string   `json:"brandId"`
	Mode    MailMode `json:"mode"`
	Status  string   `json:"status"`
}
type MailSegmentsClient struct{ client *Client }

func (m *MailSegmentsClient) Create(ctx context.Context, input MailSegmentInput) (*MailSegment, error) {
	var out MailSegment
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/segments", input, &out)
	return &out, err
}
func (m *MailSegmentsClient) List(ctx context.Context) ([]MailSegment, error) {
	var out struct {
		Data []MailSegment `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/segments", nil, &out)
	return out.Data, err
}
