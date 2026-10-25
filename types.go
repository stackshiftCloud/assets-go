package stackshift

import (
	"encoding/json"
	"io"
	"time"
)

type Options struct {
	APIKey      string
	BaseURL     string
	MailBaseURL string
	CDNBaseURL  string
	HTTPClient  HTTPDoer
}

type MailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type MailSendRequest struct {
	Simulation     *MailSimulationOptions `json:"simulation,omitempty"`
	Tracking       *MailTrackingOptions   `json:"tracking,omitempty"`
	From           any                    `json:"from"`
	To             any                    `json:"to"`
	Cc             any                    `json:"cc,omitempty"`
	Bcc            any                    `json:"bcc,omitempty"`
	ReplyTo        any                    `json:"replyTo,omitempty"`
	Subject        string                 `json:"subject"`
	HTML           string                 `json:"html,omitempty"`
	Text           string                 `json:"text,omitempty"`
	IdempotencyKey string                 `json:"idempotencyKey,omitempty"`
	Attachments    []MailAttachmentInput  `json:"attachments,omitempty"`
	Headers        map[string]string      `json:"headers,omitempty"`
	Tags           map[string]string      `json:"tags,omitempty"`
	Stream         string                 `json:"stream,omitempty"`
}

type MailSendResponse struct {
	Mode              MailMode `json:"mode"`
	ID                string   `json:"id"`
	Status            string   `json:"status"`
	IdempotencyStatus string   `json:"idempotencyStatus"`
}

type MailDNSRecord struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Status      string `json:"status"`
	Required    bool   `json:"required"`
	Explanation string `json:"explanation,omitempty"`
}

type MailDomain struct {
	Health           *MailSendingDomainHealth `json:"health,omitempty"`
	ID               string                   `json:"id"`
	Domain           string                   `json:"domain"`
	Status           string                   `json:"status"`
	SPFStatus        string                   `json:"spfStatus"`
	DKIMStatus       string                   `json:"dkimStatus"`
	DMARCStatus      string                   `json:"dmarcStatus"`
	ReturnPathStatus string                   `json:"returnPathStatus"`
	Records          []MailDNSRecord          `json:"records,omitempty"`
	CreatedAt        time.Time                `json:"createdAt"`
	VerifiedAt       *time.Time               `json:"verifiedAt,omitempty"`
	LastCheckedAt    *time.Time               `json:"lastCheckedAt,omitempty"`
}

type MailDomainListResult struct {
	Data       []MailDomain `json:"data"`
	NextCursor string       `json:"nextCursor,omitempty"`
}

type MailMessageListOptions struct {
	Status string
	Limit  int
	Cursor string
	To     string
	From   string
}

type MailMessageSummary struct {
	Mode          MailMode   `json:"mode"`
	ID            string     `json:"id"`
	From          string     `json:"from"`
	To            []string   `json:"to"`
	Subject       string     `json:"subject"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	QueuedAt      *time.Time `json:"queuedAt,omitempty"`
	MTAAcceptedAt *time.Time `json:"mtaAcceptedAt,omitempty"`
}

type MailMessageListResult struct {
	Data       []MailMessageSummary `json:"data"`
	NextCursor string               `json:"nextCursor,omitempty"`
}

type MailMessage struct {
	Mode                MailMode             `json:"mode"`
	ID                  string               `json:"id"`
	From                MailAddress          `json:"from"`
	To                  []string             `json:"to"`
	Cc                  []string             `json:"cc"`
	Bcc                 []string             `json:"bcc"`
	ReplyTo             *string              `json:"replyTo,omitempty"`
	Subject             string               `json:"subject"`
	HTML                *string              `json:"html,omitempty"`
	Text                *string              `json:"text,omitempty"`
	Status              string               `json:"status"`
	ErrorCode           *string              `json:"errorCode,omitempty"`
	ErrorMessage        *string              `json:"errorMessage,omitempty"`
	CreatedAt           time.Time            `json:"createdAt"`
	QueuedAt            *time.Time           `json:"queuedAt,omitempty"`
	SendingAt           *time.Time           `json:"sendingAt,omitempty"`
	MTAAcceptedAt       *time.Time           `json:"mtaAcceptedAt,omitempty"`
	FailedAt            *time.Time           `json:"failedAt,omitempty"`
	BouncedAt           *time.Time           `json:"bouncedAt,omitempty"`
	SuppressedAt        *time.Time           `json:"suppressedAt,omitempty"`
	BounceType          *string              `json:"bounceType,omitempty"`
	BounceReason        *string              `json:"bounceReason,omitempty"`
	FinalSMTPStatusCode *string              `json:"finalSmtpStatusCode,omitempty"`
	FinalSMTPResponse   *string              `json:"finalSmtpResponse,omitempty"`
	Recipients          []MailRecipient      `json:"recipients,omitempty"`
	RecipientSummary    MailRecipientSummary `json:"recipientSummary"`
	Headers             map[string]string    `json:"headers,omitempty"`
	Tags                map[string]string    `json:"tags,omitempty"`
	Stream              string               `json:"stream,omitempty"`
	StreamType          string               `json:"streamType,omitempty"`
}

type MailRecipient struct {
	ID                string     `json:"id"`
	MessageID         string     `json:"messageId"`
	Kind              string     `json:"kind"`
	Email             string     `json:"email"`
	Status            string     `json:"status"`
	RFCMessageID      *string    `json:"rfcMessageId,omitempty"`
	PostfixQueueID    *string    `json:"postfixQueueId,omitempty"`
	RemoteMX          *string    `json:"remoteMx,omitempty"`
	Provider          *string    `json:"provider,omitempty"`
	SMTPStatus        *string    `json:"smtpStatus,omitempty"`
	EnhancedStatus    *string    `json:"enhancedStatus,omitempty"`
	SMTPResponse      *string    `json:"smtpResponse,omitempty"`
	LatencyMS         *int64     `json:"latencyMs,omitempty"`
	Disposition       *string    `json:"disposition,omitempty"`
	BounceCategory    *string    `json:"bounceCategory,omitempty"`
	SuppressionAction *string    `json:"suppressionAction,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	MTAAcceptedAt     *time.Time `json:"mtaAcceptedAt,omitempty"`
	DeliveredAt       *time.Time `json:"deliveredAt,omitempty"`
	BouncedAt         *time.Time `json:"bouncedAt,omitempty"`
	ComplainedAt      *time.Time `json:"complainedAt,omitempty"`
}

type MailRecipientSummary struct {
	Queued      int `json:"queued"`
	Sending     int `json:"sending"`
	MTAAccepted int `json:"mtaAccepted"`
	Delayed     int `json:"delayed"`
	Delivered   int `json:"delivered"`
	Bounced     int `json:"bounced"`
	Failed      int `json:"failed"`
	Suppressed  int `json:"suppressed"`
	Complained  int `json:"complained"`
}

type MailSendAttempt struct {
	ID             string     `json:"id"`
	AttemptNumber  int        `json:"attemptNumber"`
	Status         string     `json:"status"`
	MTAQueueID     *string    `json:"mtaQueueId,omitempty"`
	SMTPStatusCode *string    `json:"smtpStatusCode,omitempty"`
	SMTPResponse   *string    `json:"smtpResponse,omitempty"`
	StartedAt      time.Time  `json:"startedAt"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
}

type MailLog struct {
	Level     string         `json:"level"`
	Event     string         `json:"event"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

type MailSuppressionListOptions struct {
	Email   string
	Reason  string
	BrandID string
	Limit   int
	Cursor  string
}

type MailSuppressionCreateRequest struct {
	Email    string         `json:"email"`
	Reason   string         `json:"reason,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type MailSuppression struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Reason    string     `json:"reason"`
	Source    string     `json:"source"`
	MessageID *string    `json:"messageId,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type MailSuppressionListResult struct {
	Data       []MailSuppression `json:"data"`
	NextCursor string            `json:"nextCursor,omitempty"`
}

type MailBounceListOptions struct {
	Email     string
	Type      string
	MessageID string
	Limit     int
	Cursor    string
}

type MailBounce struct {
	ID             string    `json:"id"`
	MessageID      *string   `json:"messageId,omitempty"`
	RecipientEmail string    `json:"recipientEmail"`
	BounceType     string    `json:"bounceType"`
	Reason         string    `json:"reason"`
	SMTPStatusCode *string   `json:"smtpStatusCode,omitempty"`
	DiagnosticCode *string   `json:"diagnosticCode,omitempty"`
	ReceivedAt     time.Time `json:"receivedAt"`
}

type MailBounceListResult struct {
	Data       []MailBounce `json:"data"`
	NextCursor string       `json:"nextCursor,omitempty"`
}

type Asset struct {
	ID                string          `json:"id"`
	Bucket            string          `json:"bucket"`
	Key               string          `json:"key"`
	OriginalName      string          `json:"original_name"`
	MimeType          string          `json:"mime_type"`
	Size              int64           `json:"size"`
	ChecksumSHA256    string          `json:"checksum_sha256"`
	Visibility        string          `json:"visibility"`
	Status            string          `json:"status"`
	ReplicationStatus string          `json:"replication_status"`
	ScanStatus        string          `json:"scan_status,omitempty"`
	ScanResult        string          `json:"scan_result,omitempty"`
	QuarantineStatus  string          `json:"quarantine_status,omitempty"`
	VideoStatus       string          `json:"video_status,omitempty"`
	VideoProgress     *int            `json:"video_progress_percent,omitempty"`
	AIStatus          string          `json:"ai_status,omitempty"`
	ModerationStatus  string          `json:"moderation_status,omitempty"`
	ModerationReason  string          `json:"moderation_reason,omitempty"`
	CurrentVersionID  string          `json:"current_version_id,omitempty"`
	DefaultBranch     string          `json:"default_branch,omitempty"`
	Metadata          map[string]any  `json:"metadata,omitempty"`
	CacheControl      string          `json:"cache_control,omitempty"`
	Tags              []string        `json:"tags,omitempty"`
	Video             *AssetVideoInfo `json:"video,omitempty"`
	AI                *AssetAIInfo    `json:"ai,omitempty"`
	URL               string          `json:"url,omitempty"`
	Generation        int64           `json:"generation"`
	Revision          int64           `json:"revision"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type AssetAIInfo struct {
	Annotation *AssetAIAnnotation     `json:"annotation,omitempty"`
	Moderation *AssetModerationResult `json:"moderation,omitempty"`
	Transcript *AssetTranscript       `json:"transcript,omitempty"`
}

type AssetAIAnnotation struct {
	Provider string         `json:"provider"`
	Model    string         `json:"model"`
	Status   string         `json:"status"`
	Caption  string         `json:"caption,omitempty"`
	AltText  string         `json:"alt_text,omitempty"`
	Tags     []string       `json:"tags,omitempty"`
	Objects  []string       `json:"objects,omitempty"`
	Colors   []string       `json:"colors,omitempty"`
	Raw      map[string]any `json:"raw,omitempty"`
}

type AssetModerationResult struct {
	ID             string             `json:"id"`
	Provider       string             `json:"provider"`
	Model          string             `json:"model"`
	Status         string             `json:"status"`
	Flagged        bool               `json:"flagged"`
	Categories     map[string]bool    `json:"categories"`
	CategoryScores map[string]float64 `json:"category_scores"`
	Reason         string             `json:"reason,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
}

type AssetTranscript struct {
	Provider string           `json:"provider"`
	Model    string           `json:"model"`
	Status   string           `json:"status"`
	Language string           `json:"language,omitempty"`
	Text     string           `json:"text,omitempty"`
	Segments []map[string]any `json:"segments,omitempty"`
}

type AssetVersion struct {
	ID              string         `json:"id"`
	AssetID         string         `json:"asset_id"`
	Branch          string         `json:"branch"`
	VersionNumber   int            `json:"version_number"`
	ParentVersionID string         `json:"parent_version_id,omitempty"`
	OriginalName    string         `json:"original_name"`
	MimeType        string         `json:"mime_type"`
	Size            int64          `json:"size"`
	ChecksumSHA256  string         `json:"checksum_sha256"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	CacheControl    string         `json:"cache_control,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}

type AssetCollection struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	AssetIDs    []string  `json:"asset_ids,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AssetSavedSearch struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Filters   map[string]any `json:"filters"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type AssetVideoInfo struct {
	Metadata  map[string]any   `json:"metadata,omitempty"`
	Outputs   []map[string]any `json:"outputs,omitempty"`
	HLSURL    string           `json:"hls_url,omitempty"`
	PosterURL string           `json:"poster_url,omitempty"`
}

type AssetContentPolicy struct {
	AllowedMimeTypes []string `json:"allowed_mime_types"`
	MaxImageBytes    int64    `json:"max_image_bytes"`
	MaxVideoBytes    int64    `json:"max_video_bytes"`
	MaxOtherBytes    int64    `json:"max_other_bytes"`
	RequireScan      bool     `json:"require_scan"`
}

type AssetUpload struct {
	File         io.Reader
	FileName     string
	Bucket       string
	Key          string
	Folder       string
	Visibility   string
	CacheControl string
	Metadata     map[string]any
}

type AssetListOptions struct {
	Bucket            string
	Prefix            string
	Visibility        string
	Query             string
	Tag               string
	MimeType          string
	Type              string
	Folder            string
	Checksum          string
	Status            string
	ScanStatus        string
	ReplicationStatus string
	ModerationStatus  string
	QuarantineStatus  string
	Sort              string
	Direction         string
	Limit             int
	Cursor            string
}

type AssetListResult struct {
	Assets     []Asset `json:"assets"`
	Total      int     `json:"total"`
	NextCursor string  `json:"next_cursor,omitempty"`
}

type SignedURLOptions struct {
	ExpiresIn    string
	MaxDownloads *int
}

type SignedURL struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SignedUploadURLOptions struct {
	Mode           string
	FileName       string
	FileSize       int64
	MimeType       string
	ChecksumSHA256 string
	PartSize       int64
	Bucket         string
	Key            string
	Folder         string
	Visibility     string
	CacheControl   string
	Metadata       map[string]any
	ExpiresIn      string
	MaxBytes       int64
	IdempotencyKey string
}

type AssetBucket struct {
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	DefaultVisibility string         `json:"default_visibility"`
	CacheControl      string         `json:"cache_control,omitempty"`
	Versioning        bool           `json:"versioning"`
	RetentionDays     int            `json:"retention_days"`
	MaxObjectBytes    *int64         `json:"max_object_bytes,omitempty"`
	AllowedMimeTypes  []string       `json:"allowed_mime_types"`
	HomeRegion        string         `json:"home_region"`
	ReplicationPolicy string         `json:"replication_policy"`
	CORSOrigins       []string       `json:"cors_origins"`
	AllowedOrigins    []string       `json:"allowed_origins"`
	LifecyclePolicy   map[string]any `json:"lifecycle_policy"`
	CustomDomainID    string         `json:"custom_domain_id,omitempty"`
	Revision          int64          `json:"revision"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type AssetBucketInput struct {
	Name              string         `json:"name,omitempty"`
	DefaultVisibility string         `json:"default_visibility,omitempty"`
	CacheControl      string         `json:"cache_control,omitempty"`
	Versioning        *bool          `json:"versioning,omitempty"`
	RetentionDays     int            `json:"retention_days,omitempty"`
	MaxObjectBytes    *int64         `json:"max_object_bytes,omitempty"`
	AllowedMimeTypes  []string       `json:"allowed_mime_types,omitempty"`
	HomeRegion        string         `json:"home_region,omitempty"`
	ReplicationPolicy string         `json:"replication_policy,omitempty"`
	CORSOrigins       []string       `json:"cors_origins,omitempty"`
	AllowedOrigins    []string       `json:"allowed_origins,omitempty"`
	LifecyclePolicy   map[string]any `json:"lifecycle_policy,omitempty"`
	CustomDomainID    string         `json:"custom_domain_id,omitempty"`
}

type AssetPatchInput struct {
	Bucket            *string `json:"bucket,omitempty"`
	Key               *string `json:"key,omitempty"`
	OriginalName      *string `json:"original_name,omitempty"`
	CacheControl      *string `json:"cache_control,omitempty"`
	ClearCacheControl bool    `json:"clear_cache_control,omitempty"`
	Visibility        *string `json:"visibility,omitempty"`
}

type AssetJobSubmission struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	StatusURL string `json:"status_url"`
}

type AssetJob struct {
	ID              string         `json:"id"`
	JobName         string         `json:"job_name"`
	QueueName       string         `json:"queue_name"`
	Status          string         `json:"status"`
	ProgressPercent int            `json:"progress_percent"`
	Phase           string         `json:"phase"`
	Payload         map[string]any `json:"payload"`
	Result          map[string]any `json:"result,omitempty"`
	Error           map[string]any `json:"error,omitempty"`
	Attempts        int            `json:"attempts"`
	MaxAttempts     int            `json:"max_attempts"`
	ScheduledAt     time.Time      `json:"scheduled_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	StartedAt       *time.Time     `json:"started_at,omitempty"`
	CompletedAt     *time.Time     `json:"completed_at,omitempty"`
}

type AssetAIPolicy struct {
	EnabledActions        []string `json:"enabled_actions"`
	MonthlySpendCapMicros int64    `json:"monthly_spend_cap_micros"`
	Revision              int64    `json:"revision"`
}

type AssetAIAccountConfig struct {
	Provider               string        `json:"provider"`
	Configured             bool          `json:"configured"`
	Actions                []string      `json:"actions"`
	Policy                 AssetAIPolicy `json:"policy"`
	MonthlySpendUsedMicros int64         `json:"monthly_spend_used_micros"`
}

type AssetLifecycleRule struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Prefix    string     `json:"prefix,omitempty"`
	Action    string     `json:"action"`
	AgeDays   int        `json:"age_days"`
	Enabled   bool       `json:"enabled"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastRunAt *time.Time `json:"last_run_at,omitempty"`
	LastError *string    `json:"last_error,omitempty"`
}

type AssetCustomDomain struct {
	ID                string     `json:"id"`
	Domain            string     `json:"domain"`
	Status            string     `json:"status"`
	VerificationName  string     `json:"verification_name"`
	VerificationValue string     `json:"verification_value"`
	LastError         *string    `json:"last_error,omitempty"`
	ProviderDomainID  *string    `json:"provider_domain_id,omitempty"`
	CertificateID     *string    `json:"certificate_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
}

type AssetEvent struct {
	ID        string         `json:"id"`
	AssetID   *string        `json:"asset_id,omitempty"`
	Type      string         `json:"type"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
}

type AssetWebhook struct {
	ID            string   `json:"id"`
	URL           string   `json:"url"`
	EventTypes    []string `json:"event_types"`
	Status        string   `json:"status"`
	FailureCount  int      `json:"failure_count"`
	LastSuccessAt string   `json:"last_success_at,omitempty"`
	LastFailureAt string   `json:"last_failure_at,omitempty"`
	DisabledAt    string   `json:"disabled_at,omitempty"`
	CreatedBy     string   `json:"created_by"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type AssetWebhookDelivery struct {
	ID              string         `json:"id"`
	WebhookID       string         `json:"webhook_id"`
	EventID         string         `json:"event_id"`
	EventType       string         `json:"event_type"`
	Payload         map[string]any `json:"payload"`
	Status          string         `json:"status"`
	AttemptCount    int            `json:"attempt_count"`
	NextAttemptAt   string         `json:"next_attempt_at"`
	StatusCode      int            `json:"status_code,omitempty"`
	ResponseSnippet string         `json:"response_snippet,omitempty"`
	ErrorMessage    string         `json:"error_message,omitempty"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
	DeliveredAt     string         `json:"delivered_at,omitempty"`
}

type AssetUsageSummary struct {
	TotalAssets             int              `json:"total_assets"`
	TotalBytes              int64            `json:"total_bytes"`
	PeriodIngressBytes      int64            `json:"period_ingress_bytes"`
	PeriodEgressBytes       int64            `json:"period_egress_bytes"`
	PeriodOriginBytes       int64            `json:"period_origin_bytes"`
	PeriodVerificationBytes int64            `json:"period_verification_bytes"`
	PeriodTransformCount    int64            `json:"period_transform_count"`
	PeriodAIRequests        int64            `json:"period_ai_requests"`
	PeriodLogicalByteHours  int64            `json:"period_logical_storage_byte_hours"`
	PeriodPhysicalByteHours int64            `json:"period_physical_storage_byte_hours"`
	PeriodDerivedByteHours  int64            `json:"period_derived_storage_byte_hours"`
	PeriodTransformMillis   int64            `json:"period_transform_compute_ms"`
	PeriodVideoInputSeconds int64            `json:"period_video_input_seconds"`
	PeriodAICostMicros      int64            `json:"period_ai_cost_micros"`
	ByBucket                map[string]int64 `json:"by_bucket"`
	ByMimeFamily            map[string]int64 `json:"by_mime_family"`
	ByVisibility            map[string]int64 `json:"by_visibility"`
}

type SignedUploadURL struct {
	UploadSessionID string                 `json:"upload_session_id"`
	SessionID       string                 `json:"session_id,omitempty"`
	Mode            string                 `json:"mode,omitempty"`
	URL             string                 `json:"url"`
	Method          string                 `json:"method"`
	ExpiresAt       time.Time              `json:"expires_at"`
	MaxBytes        int64                  `json:"max_bytes"`
	PartSize        int64                  `json:"part_size,omitempty"`
	Parts           []AssetChunkUploadPart `json:"parts,omitempty"`
}

type AssetChunkUploadPart struct {
	PartNumber     int       `json:"part_number"`
	Size           int64     `json:"size"`
	ChecksumSHA256 string    `json:"checksum_sha256"`
	CreatedAt      time.Time `json:"created_at"`
}

type AssetTransformOptions struct {
	Width     int
	Height    int
	Crop      string
	Format    string
	Quality   string
	ExpiresIn string
	Gravity   string
	Animation string
}

type AssetTransformPreset struct {
	ID          string
	Name        string
	Category    string
	Description string
	AssetTransformOptions
}

type AssetTransformation struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Spec              string          `json:"spec"`
	NormalizedSpec    string          `json:"normalized_spec"`
	DefinitionVersion int             `json:"definition_version"`
	MediaKind         string          `json:"media_kind"`
	Definition        json.RawMessage `json:"definition,omitempty"`
	Eager             bool            `json:"eager"`
	Revision          int64           `json:"revision"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}
