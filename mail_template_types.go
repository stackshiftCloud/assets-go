package stackshift

import "time"

type MailTemplateCreateRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	Subject     string `json:"subject"`
	HTML        string `json:"html,omitempty"`
	Text        string `json:"text,omitempty"`
}

type MailTemplateUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Subject     string `json:"subject,omitempty"`
	HTML        string `json:"html,omitempty"`
	Text        string `json:"text,omitempty"`
}

type MailTemplateListOptions struct {
	Status string
	Limit  int
	Cursor string
	Search string
}

type MailTemplate struct {
	Mode            MailMode  `json:"mode"`
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     *string   `json:"description,omitempty"`
	Status          string    `json:"status"`
	ActiveVersionID *string   `json:"activeVersionId,omitempty"`
	Variables       []string  `json:"variables"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type MailTemplateVersion struct {
	ID            string    `json:"id"`
	TemplateID    string    `json:"templateId"`
	VersionNumber int       `json:"versionNumber"`
	Subject       string    `json:"subject"`
	HTML          *string   `json:"html,omitempty"`
	Text          *string   `json:"text,omitempty"`
	Variables     []string  `json:"variables"`
	CreatedAt     time.Time `json:"createdAt"`
}

type MailTemplateDetail struct {
	MailTemplate
	ActiveVersion *MailTemplateVersion  `json:"activeVersion,omitempty"`
	Versions      []MailTemplateVersion `json:"versions,omitempty"`
}

type MailTemplateListResult struct {
	Data       []MailTemplate `json:"data"`
	NextCursor string         `json:"nextCursor,omitempty"`
}

type MailTemplatePreviewRequest struct {
	Data      map[string]any `json:"data"`
	VersionID string         `json:"versionId,omitempty"`
}

type MailTemplatePreview struct {
	Subject          string   `json:"subject"`
	HTML             *string  `json:"html,omitempty"`
	Text             *string  `json:"text,omitempty"`
	MissingVariables []string `json:"missingVariables"`
}

type MailTemplateTestSendRequest struct {
	Simulation *MailSimulationOptions `json:"simulation,omitempty"`
	To         any                    `json:"to"`
	From       any                    `json:"from,omitempty"`
	Data       map[string]any         `json:"data"`
	VersionID  string                 `json:"versionId,omitempty"`
}

type MailTemplateTestSendResponse struct {
	Mode      MailMode `json:"mode"`
	MessageID string   `json:"messageId"`
	Status    string   `json:"status"`
}

type MailSendTemplateRequest struct {
	Simulation     *MailSimulationOptions `json:"simulation,omitempty"`
	Tracking       *MailTrackingOptions   `json:"tracking,omitempty"`
	Template       string                 `json:"template"`
	VersionID      string                 `json:"versionId,omitempty"`
	To             any                    `json:"to"`
	From           any                    `json:"from,omitempty"`
	Cc             any                    `json:"cc,omitempty"`
	Bcc            any                    `json:"bcc,omitempty"`
	ReplyTo        any                    `json:"replyTo,omitempty"`
	Data           map[string]any         `json:"data"`
	IdempotencyKey string                 `json:"idempotencyKey,omitempty"`
	Attachments    []MailAttachmentInput  `json:"attachments,omitempty"`
}

type MailSendTemplateResponse struct {
	Mode              MailMode `json:"mode"`
	ID                string   `json:"id"`
	Status            string   `json:"status"`
	TemplateID        string   `json:"templateId"`
	TemplateVersionID string   `json:"templateVersionId"`
	IdempotencyStatus string   `json:"idempotencyStatus"`
}
