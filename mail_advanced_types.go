package stackshift

import "time"

type MailAttachmentInput struct {
	AssetID     string `json:"assetId,omitempty"`
	UploadID    string `json:"uploadId,omitempty"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType,omitempty"`
}

type MailScheduledCreateRequest struct {
	SendAt         time.Time             `json:"sendAt"`
	Message        *MailSendRequest      `json:"message,omitempty"`
	Template       string                `json:"template,omitempty"`
	VersionID      string                `json:"versionId,omitempty"`
	To             any                   `json:"to,omitempty"`
	From           any                   `json:"from,omitempty"`
	Data           map[string]any        `json:"data,omitempty"`
	IdempotencyKey string                `json:"idempotencyKey,omitempty"`
	Attachments    []MailAttachmentInput `json:"attachments,omitempty"`
}

type MailScheduledMessage struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"`
	SendAt     time.Time  `json:"sendAt"`
	MessageID  *string    `json:"messageId,omitempty"`
	CreatedAt  *time.Time `json:"createdAt,omitempty"`
	CanceledAt *time.Time `json:"canceledAt,omitempty"`
	SentAt     *time.Time `json:"sentAt,omitempty"`
}

type MailScheduledListResult struct {
	Data []MailScheduledMessage `json:"data"`
}

type MailBatchSendRequest struct {
	From     any               `json:"from"`
	Messages []MailSendRequest `json:"messages"`
}

type MailBatchTemplateRecipient struct {
	To             any            `json:"to"`
	Data           map[string]any `json:"data,omitempty"`
	IdempotencyKey string         `json:"idempotencyKey,omitempty"`
}

type MailBatchTemplateRequest struct {
	Template   string                       `json:"template"`
	VersionID  string                       `json:"versionId,omitempty"`
	From       any                          `json:"from"`
	Recipients []MailBatchTemplateRecipient `json:"recipients"`
}

type MailBatch struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	MessageCount int    `json:"messageCount"`
}

type MailInboundDomain struct {
	ID      string          `json:"id"`
	Domain  string          `json:"domain"`
	Status  string          `json:"status"`
	Records []MailDNSRecord `json:"records,omitempty"`
}

type MailInboundDomainListResult struct {
	Data []MailInboundDomain `json:"data"`
}

type MailInboundMessage struct {
	ID         string     `json:"id"`
	FromEmail  string     `json:"fromEmail"`
	ToEmails   []string   `json:"toEmails"`
	CcEmails   []string   `json:"ccEmails,omitempty"`
	Subject    *string    `json:"subject,omitempty"`
	Text       *string    `json:"text,omitempty"`
	HTML       *string    `json:"html,omitempty"`
	RawRef     *string    `json:"rawRef,omitempty"`
	ReceivedAt *time.Time `json:"receivedAt,omitempty"`
}

type MailInboundMessageListResult struct {
	Data []MailInboundMessage `json:"data"`
}

type MailAnalyticsOptions struct {
	Range    string
	From     time.Time
	To       time.Time
	Interval string
	Domain   string
	Template string
	Event    string
}

type MailAnalytics struct {
	Summary map[string]int   `json:"summary"`
	Series  []map[string]any `json:"series"`
}
