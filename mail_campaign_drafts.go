package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type MailCampaignContent struct {
	Subject     string         `json:"subject"`
	PreviewText string         `json:"previewText"`
	HTML        string         `json:"html"`
	Text        string         `json:"text"`
	EditorMode  string         `json:"editorMode"`
	EditorState map[string]any `json:"editorState,omitempty"`
}
type MailCampaignUpdateRequest struct {
	MailCampaignCreateRequest
	Revision int `json:"revision"`
}
type MailCampaignPreviewRequest struct {
	Simulation *MailSimulationOptions `json:"simulation,omitempty"`
	Data       map[string]any         `json:"data"`
	Variant    string                 `json:"variant,omitempty"`
	To         string                 `json:"to,omitempty"`
}

func (m *MailCampaignsClient) Update(ctx context.Context, id string, input MailCampaignUpdateRequest) (*MailCampaign, error) {
	var out MailCampaign
	err := m.client.mailRequest(ctx, http.MethodPut, "/mail/campaigns/"+url.PathEscape(id), input, &out)
	return &out, err
}
func (m *MailCampaignsClient) Preview(ctx context.Context, id string, input MailCampaignPreviewRequest) (map[string]any, error) {
	var out map[string]any
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/campaigns/"+url.PathEscape(id)+"/preview", input, &out)
	return out, err
}
func (m *MailCampaignsClient) Test(ctx context.Context, id string, input MailCampaignPreviewRequest) (*MailSendResponse, error) {
	var out MailSendResponse
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/campaigns/"+url.PathEscape(id)+"/test", input, &out)
	return &out, err
}
func (m *MailCampaignsClient) Schedule(ctx context.Context, id string, sendAt time.Time, revision ...int) (*MailCampaign, error) {
	var out MailCampaign
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/campaigns/"+url.PathEscape(id)+"/schedule", map[string]any{"sendAt": sendAt, "revision": campaignRevision(revision)["revision"]}, &out)
	return &out, err
}

func campaignRevision(revision []int) map[string]int {
	if len(revision) == 0 {
		return nil
	}
	return map[string]int{"revision": revision[0]}
}
