package stackshift

import "time"

type ComputerKeyScope string

const (
	ComputerScopeRead   ComputerKeyScope = "computers:read"
	ComputerScopeWrite  ComputerKeyScope = "computers:write"
	ComputerScopeUse    ComputerKeyScope = "computers:use"
	ComputerScopeDelete ComputerKeyScope = "computers:delete"
)

type ComputerExecExit struct {
	ProcessID  string `json:"process_id"`
	ExitCode   *int   `json:"exit_code,omitempty"`
	TimedOut   bool   `json:"timed_out"`
	Signal     string `json:"signal,omitempty"`
	DurationMS int64  `json:"duration_ms"`
}

type ComputerStreamError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ComputerExecEvent struct {
	Event     string
	ProcessID string
	Data      string
	Exit      *ComputerExecExit
	Error     *ComputerStreamError
}

type ComputerCommandRecord struct {
	ID              string     `json:"id"`
	Command         []string   `json:"command,omitempty"`
	Shell           string     `json:"shell,omitempty"`
	CWD             string     `json:"cwd"`
	Background      bool       `json:"background"`
	State           string     `json:"state"`
	ExitCode        *int       `json:"exit_code,omitempty"`
	Signal          string     `json:"signal,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationMS      int64      `json:"duration_ms"`
	StdoutBytes     int64      `json:"stdout_bytes"`
	StderrBytes     int64      `json:"stderr_bytes"`
	Actor           string     `json:"actor,omitempty"`
	Stdout          string     `json:"stdout,omitempty"`
	Stderr          string     `json:"stderr,omitempty"`
	StdoutTruncated bool       `json:"stdout_truncated,omitempty"`
	StderrTruncated bool       `json:"stderr_truncated,omitempty"`
}

type ComputerCommandPage struct {
	Items      []ComputerCommandRecord `json:"items"`
	NextCursor string                  `json:"next_cursor,omitempty"`
}

type ComputerCommandList struct {
	Items []ComputerCommandRecord `json:"items"`
}

type ComputerJobLogChunk struct {
	Stream string `json:"stream"`
	Data   string `json:"data"`
}

type ComputerJobLogs struct {
	Chunks     []ComputerJobLogChunk `json:"chunks"`
	NextOffset int64                 `json:"next_offset"`
	Done       bool                  `json:"done"`
	Job        ComputerCommandRecord `json:"job"`
}

type ComputerJobLogEvent struct {
	Event  string
	Offset int64
	Data   string
	Exit   *ComputerCommandRecord
}

type ComputerActivityLease struct {
	ComputerID string    `json:"computer_id"`
	Task       string    `json:"task"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ComputerActivityLeaseList struct {
	Items []ComputerActivityLease `json:"items"`
}

type ComputerActivityLeaseResult struct {
	ComputerActivityLease
	Woke bool `json:"woke"`
}

type ComputerLifecycleLimits struct {
	ComputerID    string     `json:"computer_id"`
	MaxRunSeconds int64      `json:"max_run_seconds"`
	ArchiveAt     *time.Time `json:"archive_at,omitempty"`
	DeleteAt      *time.Time `json:"delete_at,omitempty"`
	RunningSince  *time.Time `json:"running_since,omitempty"`
}

type ComputerMetricsSample struct {
	Time             time.Time `json:"time"`
	CPUPercent       float64   `json:"cpu_percent"`
	MemoryUsedBytes  int64     `json:"memory_used_bytes"`
	MemoryTotalBytes int64     `json:"memory_total_bytes"`
	DiskUsedBytes    int64     `json:"disk_used_bytes"`
	DiskTotalBytes   int64     `json:"disk_total_bytes"`
	NetRxBytes       int64     `json:"net_rx_bytes"`
	NetTxBytes       int64     `json:"net_tx_bytes"`
	OOMKills         int64     `json:"oom_kills"`
	VMMOOMKills      int64     `json:"vmm_oom_kills"`
}

type ComputerMetrics struct {
	ComputerID      string                  `json:"computer_id"`
	WorkloadID      string                  `json:"workload_id,omitempty"`
	IntervalSeconds int                     `json:"interval_seconds"`
	Samples         []ComputerMetricsSample `json:"samples"`
	OOMKills        int64                   `json:"oom_kills"`
	VMMOOMKills     int64                   `json:"vmm_oom_kills"`
}

type ComputerSSHToken struct {
	ID         string    `json:"id"`
	Token      string    `json:"token"`
	SSHCommand string    `json:"ssh_command"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type ComputerSSHCertificate struct {
	Certificate string    `json:"certificate"`
	Login       string    `json:"login"`
	Host        string    `json:"host"`
	SSHCommand  string    `json:"ssh_command"`
	ExpiresAt   time.Time `json:"expires_at"`
}

const (
	ComputerAgentClaude   = "claude"
	ComputerAgentCodex    = "codex"
	ComputerAgentGemini   = "gemini"
	ComputerAgentOpenCode = "opencode"
	ComputerAgentCursor   = "cursor"
	ComputerAgentCopilot  = "copilot"
	ComputerAgentAmp      = "amp"
	ComputerAgentQwen     = "qwen"
)

var ComputerAgentProviders = []string{ComputerAgentClaude, ComputerAgentCodex, ComputerAgentGemini, ComputerAgentOpenCode, ComputerAgentCursor, ComputerAgentCopilot, ComputerAgentAmp, ComputerAgentQwen}

const (
	ComputerAgentRunning     = "running"
	ComputerAgentIdle        = "idle"
	ComputerAgentFailed      = "failed"
	ComputerAgentInterrupted = "interrupted"
	ComputerAgentNeedsInput  = "needs_input"
	ComputerAgentLost        = "lost"
	ComputerAgentCompleted   = "completed"
	ComputerAgentStarting    = "starting"
)

type ComputerAgentPrompt struct {
	Prompt           string            `json:"prompt"`
	ConversationID   string            `json:"conversation_id,omitempty"`
	Provider         string            `json:"provider,omitempty"`
	Model            string            `json:"model,omitempty"`
	CWD              string            `json:"cwd,omitempty"`
	Env              map[string]string `json:"env,omitempty"`
	SecretBindingIDs []string          `json:"secret_binding_ids,omitempty"`
	Attachments      []string          `json:"attachments,omitempty"`
}

type ComputerAgentFanoutPrompt struct {
	Prompt           string            `json:"prompt,omitempty"`
	Prompts          []string          `json:"prompts,omitempty"`
	Forks            int               `json:"forks,omitempty"`
	Provider         string            `json:"provider,omitempty"`
	Model            string            `json:"model,omitempty"`
	CWD              string            `json:"cwd,omitempty"`
	Env              map[string]string `json:"env,omitempty"`
	SecretBindingIDs []string          `json:"secret_binding_ids,omitempty"`
	Attachments      []string          `json:"attachments,omitempty"`
}

type ComputerAgentChangedFiles struct {
	Files      int    `json:"files"`
	Insertions int    `json:"insertions"`
	Deletions  int    `json:"deletions"`
	Stat       string `json:"stat"`
}

type ComputerAgentTurnResult struct {
	DurationMS   int64                      `json:"duration_ms"`
	ExitCode     *int                       `json:"exit_code,omitempty"`
	LastMessage  string                     `json:"last_message,omitempty"`
	ChangedFiles *ComputerAgentChangedFiles `json:"changed_files,omitempty"`
}

type ComputerAgentConversation struct {
	ID         string                   `json:"id"`
	Provider   string                   `json:"provider"`
	Model      string                   `json:"model,omitempty"`
	CWD        string                   `json:"cwd"`
	SessionID  string                   `json:"session_id,omitempty"`
	Status     string                   `json:"status"`
	ExitCode   *int                     `json:"exit_code,omitempty"`
	Turns      int                      `json:"turns"`
	LastPrompt string                   `json:"last_prompt,omitempty"`
	Result     *ComputerAgentTurnResult `json:"result,omitempty"`
	CreatedAt  time.Time                `json:"created_at"`
	UpdatedAt  time.Time                `json:"updated_at"`
}

type ComputerAgentEvent struct {
	Seq      int64                    `json:"seq"`
	Turn     int                      `json:"turn"`
	Time     time.Time                `json:"time"`
	Kind     string                   `json:"kind"`
	Data     map[string]any           `json:"data,omitempty"`
	Text     string                   `json:"text,omitempty"`
	Tool     string                   `json:"tool,omitempty"`
	Command  string                   `json:"command,omitempty"`
	Path     string                   `json:"path,omitempty"`
	Prompt   string                   `json:"prompt,omitempty"`
	Status   string                   `json:"status,omitempty"`
	ExitCode *int                     `json:"exit_code,omitempty"`
	Result   *ComputerAgentTurnResult `json:"result,omitempty"`
}

type ComputerAgentProviderInfo struct {
	Provider     string   `json:"provider"`
	Name         string   `json:"name"`
	Installed    bool     `json:"installed"`
	Version      string   `json:"version,omitempty"`
	SignedIn     bool     `json:"signed_in"`
	Method       string   `json:"method,omitempty"`
	LoginCommand []string `json:"login_command"`
	ModelKeys    []string `json:"model_keys"`
}

type ComputerAgentResult struct {
	ConversationID string                   `json:"conversation_id"`
	Provider       string                   `json:"provider"`
	Turn           int                      `json:"turn"`
	Status         string                   `json:"status"`
	Result         *ComputerAgentTurnResult `json:"result,omitempty"`
}

type ComputerAgentTurn struct {
	ID             string                   `json:"id"`
	ConversationID string                   `json:"conversation_id"`
	Turn           int                      `json:"turn"`
	Provider       string                   `json:"provider"`
	Model          string                   `json:"model"`
	Prompt         string                   `json:"prompt"`
	Status         string                   `json:"status"`
	Error          string                   `json:"error"`
	Result         *ComputerAgentTurnResult `json:"result"`
	FanoutID       *string                  `json:"fanout_id"`
	CreatedAt      time.Time                `json:"created_at"`
	EndedAt        *time.Time               `json:"ended_at"`
}

type ComputerAgentFork struct {
	ComputerID     string                   `json:"computer_id"`
	Name           string                   `json:"name"`
	State          string                   `json:"state"`
	Prompt         string                   `json:"prompt"`
	ConversationID string                   `json:"conversation_id,omitempty"`
	Status         string                   `json:"status"`
	Error          string                   `json:"error,omitempty"`
	Result         *ComputerAgentTurnResult `json:"result,omitempty"`
}

type ComputerAgentFanout struct {
	ID               string              `json:"id"`
	SourceComputerID string              `json:"source_computer_id"`
	Provider         string              `json:"provider"`
	Model            string              `json:"model,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	Forks            []ComputerAgentFork `json:"forks"`
}

type ComputerAgentEventPage struct {
	ConversationID string               `json:"conversation_id"`
	Events         []ComputerAgentEvent `json:"events"`
	NextSeq        int64                `json:"next_seq"`
	Status         string               `json:"status"`
	ExitCode       *int                 `json:"exit_code,omitempty"`
}

type ComputerFileTransfer struct {
	ID             string     `json:"id"`
	Direction      string     `json:"direction"`
	Path           string     `json:"path"`
	SizeBytes      int64      `json:"size_bytes,omitempty"`
	ChecksumSHA256 string     `json:"checksum_sha256,omitempty"`
	State          string     `json:"state"`
	FailureCode    string     `json:"failure_code,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	ChunkBytes     int64      `json:"chunk_bytes,omitempty"`
	ChunkSHA256    []string   `json:"chunk_sha256,omitempty"`
}

type ComputerSignedRequest struct {
	URL       string              `json:"url"`
	Method    string              `json:"method"`
	Headers   map[string][]string `json:"headers"`
	ExpiresAt time.Time           `json:"expires_at"`
}

type ComputerTransfer struct {
	Transfer ComputerFileTransfer   `json:"transfer"`
	Download *ComputerSignedRequest `json:"download,omitempty"`
}

type ComputerUploadInput struct {
	Path      string   `json:"path"`
	SizeBytes int64    `json:"size_bytes"`
	SHA256    string   `json:"sha256"`
	Mode      *int     `json:"mode,omitempty"`
	PartMD5   []string `json:"part_md5,omitempty"`
}

type ComputerUploadPart struct {
	Number      int                 `json:"number"`
	OffsetBytes int64               `json:"offset_bytes"`
	SizeBytes   int64               `json:"size_bytes"`
	Done        bool                `json:"done"`
	ETag        string              `json:"etag,omitempty"`
	URL         string              `json:"url,omitempty"`
	Method      string              `json:"method,omitempty"`
	Headers     map[string][]string `json:"headers,omitempty"`
}

type ComputerUploadPlan struct {
	Transfer      ComputerFileTransfer `json:"transfer"`
	UploadID      string               `json:"upload_id"`
	PartSizeBytes int64                `json:"part_size_bytes"`
	Parts         []ComputerUploadPart `json:"parts"`
	ExpiresAt     time.Time            `json:"expires_at"`
}

type ComputerCompletedPart struct {
	Number int    `json:"number"`
	ETag   string `json:"etag"`
}

type ComputerCompleteUploadInput struct {
	UploadID string                  `json:"upload_id"`
	Parts    []ComputerCompletedPart `json:"parts"`
}

type ComputerResumeUploadInput struct {
	UploadID string   `json:"upload_id"`
	PartMD5  []string `json:"part_md5"`
}

type ComputerAgentRegistrationInput struct {
	Name   string             `json:"name,omitempty"`
	Scopes []ComputerKeyScope `json:"scopes,omitempty"`
}

type ComputerAgentRegistration struct {
	ClaimToken              string             `json:"claim_token"`
	UserCode                string             `json:"user_code"`
	VerificationURI         string             `json:"verification_uri"`
	VerificationURIComplete string             `json:"verification_uri_complete"`
	ExpiresIn               int                `json:"expires_in"`
	Interval                int                `json:"interval"`
	Scopes                  []ComputerKeyScope `json:"scopes"`
}

type ComputerAgentIssuedKey struct {
	Key       string             `json:"key"`
	KeyID     string             `json:"key_id"`
	Prefix    string             `json:"prefix"`
	Name      string             `json:"name"`
	ProjectID string             `json:"project_id"`
	Scopes    []ComputerKeyScope `json:"scopes"`
}

type ComputerAgentClaim struct {
	UserCode           string             `json:"user_code"`
	Name               string             `json:"name"`
	Status             string             `json:"status"`
	RequestedScopes    []ComputerKeyScope `json:"requested_scopes"`
	ApprovedScopes     []ComputerKeyScope `json:"approved_scopes"`
	RequesterIP        string             `json:"requester_ip"`
	RequesterUserAgent string             `json:"requester_user_agent"`
	ProjectID          string             `json:"project_id,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	ExpiresAt          time.Time          `json:"expires_at"`
	DecidedAt          *time.Time         `json:"decided_at,omitempty"`
}

type ComputerAgentClaimDecision struct {
	Approve   bool               `json:"approve"`
	Scopes    []ComputerKeyScope `json:"scopes,omitempty"`
	ProjectID string             `json:"project_id,omitempty"`
}
