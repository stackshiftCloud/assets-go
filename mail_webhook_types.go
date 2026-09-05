package stackshift

import "time"

type MailEvent struct {
	Mode         MailMode       `json:"mode"`
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	ResourceType *string        `json:"resourceType,omitempty"`
	ResourceID   *string        `json:"resourceId,omitempty"`
	MessageID    *string        `json:"messageId,omitempty"`
	Payload      map[string]any `json:"payload"`
	OccurredAt   time.Time      `json:"occurredAt"`
	CreatedAt    time.Time      `json:"createdAt"`
}

type MailEventListOptions struct {
	Type         string
	ResourceType string
	ResourceID   string
	MessageID    string
	Limit        int
	Cursor       string
	From         string
	To           string
}

type MailEventListResult struct {
	Data       []MailEvent `json:"data"`
	NextCursor string      `json:"nextCursor,omitempty"`
}

type MailTimelineItem struct {
	Type      string         `json:"type"`
	Label     string         `json:"label"`
	CreatedAt time.Time      `json:"createdAt"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type MailWebhookReplayRequest struct {
	EventIDs       []string `json:"eventIds"`
	IdempotencyKey string   `json:"-"`
}

type MailWebhookReplay struct {
	ID         string                `json:"id"`
	WebhookID  string                `json:"webhookId"`
	EventIDs   []string              `json:"eventIds"`
	Deliveries []MailWebhookDelivery `json:"deliveries"`
	CreatedAt  time.Time             `json:"createdAt"`
}

type MailWebhook struct {
	ID            string     `json:"id"`
	URL           string     `json:"url"`
	Description   *string    `json:"description,omitempty"`
	Status        string     `json:"status"`
	EventTypes    []string   `json:"eventTypes"`
	FailureCount  int        `json:"failureCount,omitempty"`
	LastSuccessAt *time.Time `json:"lastSuccessAt,omitempty"`
	LastFailureAt *time.Time `json:"lastFailureAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type MailWebhookCreateRequest struct {
	URL         string   `json:"url"`
	Description string   `json:"description,omitempty"`
	EventTypes  []string `json:"eventTypes"`
}

type MailWebhookCreateResponse struct {
	MailWebhook
	Secret string `json:"secret"`
}

type MailWebhookUpdateRequest struct {
	URL         *string  `json:"url,omitempty"`
	Description *string  `json:"description,omitempty"`
	EventTypes  []string `json:"eventTypes,omitempty"`
	Status      *string  `json:"status,omitempty"`
}

type MailWebhookListOptions struct {
	Limit  int
	Cursor string
}

type MailWebhookListResult struct {
	Data       []MailWebhook `json:"data"`
	NextCursor string        `json:"nextCursor,omitempty"`
}

type MailWebhookDeliveryListOptions struct {
	Status string
	Limit  int
	Cursor string
}

type MailWebhookDeliveryListResult struct {
	Data       []MailWebhookDelivery `json:"data"`
	NextCursor string                `json:"nextCursor,omitempty"`
}

type MailWebhookDelivery struct {
	Mode                MailMode   `json:"mode"`
	ID                  string     `json:"id"`
	EndpointID          string     `json:"endpointId"`
	EventID             string     `json:"eventId"`
	Status              string     `json:"status"`
	AttemptCount        int        `json:"attemptCount"`
	NextAttemptAt       *time.Time `json:"nextAttemptAt,omitempty"`
	LastAttemptAt       *time.Time `json:"lastAttemptAt,omitempty"`
	ResponseStatus      *int       `json:"responseStatus,omitempty"`
	ResponseBodySnippet *string    `json:"responseBodySnippet,omitempty"`
	ErrorMessage        *string    `json:"errorMessage,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	CompletedAt         *time.Time `json:"completedAt,omitempty"`
}

type MailWebhookDeliveryAttempt struct {
	ID                  string    `json:"id"`
	AttemptNumber       int       `json:"attemptNumber"`
	ResponseStatus      *int      `json:"responseStatus,omitempty"`
	ResponseBodySnippet *string   `json:"responseBodySnippet,omitempty"`
	ErrorMessage        *string   `json:"errorMessage,omitempty"`
	DurationMs          *int      `json:"durationMs,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
}

type MailWebhookDeliveryDetail struct {
	MailWebhookDelivery
	Attempts []MailWebhookDeliveryAttempt `json:"attempts"`
}

type MailWebhookSignatureInput struct {
	RawBody         []byte
	SignatureHeader string
	TimestampHeader string
	Secret          string
	Tolerance       time.Duration
	Now             time.Time
}
