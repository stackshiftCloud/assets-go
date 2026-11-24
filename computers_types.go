package stackshift

import (
	"encoding/json"
	"time"
)

type ComputerSize struct {
	VCPUs       int   `json:"vcpus"`
	MemoryBytes int64 `json:"memory_bytes"`
	DiskBytes   int64 `json:"disk_bytes"`
}

func NewComputerSize(vcpus int, memoryGB, diskGB float64) ComputerSize {
	const gib = 1 << 30
	return ComputerSize{VCPUs: vcpus, MemoryBytes: int64(memoryGB * gib), DiskBytes: int64(diskGB * gib)}
}

const (
	ComputerInternetOpen      = "open"
	ComputerInternetAllowlist = "allowlist"
	ComputerInternetNone      = "none"
)

const (
	ComputerErrorNotFound            = "computer_not_found"
	ComputerErrorCommandNotFound     = "command_not_found"
	ComputerErrorInvalidState        = "computer_invalid_state"
	ComputerErrorInvalidRequest      = "invalid_request"
	ComputerErrorTemplateUnavailable = "template_unavailable"
	ComputerErrorSizeUnavailable     = "size_unavailable"
	ComputerErrorCapacityUnavailable = "capacity_unavailable"
	ComputerErrorCapacityBusy        = "capacity_busy"
	ComputerErrorLimitExceeded       = "limit_exceeded"
	ComputerErrorSpendCapReached     = "spend_cap_reached"
	ComputerErrorInsufficientFunds   = "insufficient_funds"
	ComputerErrorInputLocked         = "input_locked"
	ComputerErrorApprovalsDisabled   = "approvals_disabled"
	ComputerErrorForbidden           = "forbidden"
	ComputerErrorIdempotencyConflict = "idempotency_conflict"
	ComputerErrorNodeUnavailable     = "node_unavailable"
	ComputerErrorUnauthorized        = "unauthorized"
	ComputerErrorInternal            = "internal_error"
	ComputerErrorNotAvailable        = "not_available"
)

type ComputerInternet struct {
	Mode           string   `json:"mode"`
	AllowedDomains []string `json:"allowed_domains,omitempty"`
}

type ComputerIdlePolicy struct {
	PauseAfterSeconds int    `json:"pause_after_seconds"`
	Mode              string `json:"mode"`
}

type ComputerAccess struct {
	Keys           *bool    `json:"keys,omitempty"`
	HostedAgentIDs []string `json:"hosted_agent_ids,omitempty"`
	Approvals      *bool    `json:"approvals,omitempty"`
	Recording      *bool    `json:"recording,omitempty"`
}

type ComputerPort struct {
	Port       int    `json:"port"`
	Visibility string `json:"visibility"`
	URL        string `json:"url"`
}

type ComputerTemplate struct {
	Slug             string         `json:"slug"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	Version          string         `json:"version"`
	Desktop          bool           `json:"desktop"`
	DefaultDiskBytes int64          `json:"default_disk_bytes"`
	MinDiskBytes     int64          `json:"min_disk_bytes"`
	MaxDiskBytes     int64          `json:"max_disk_bytes"`
	Sizes            []ComputerSize `json:"sizes"`
	OSName           string         `json:"os_name,omitempty"`
}

type Computer struct {
	ID              string             `json:"id"`
	ProjectID       string             `json:"project_id"`
	Name            string             `json:"name"`
	Region          string             `json:"region"`
	Template        string             `json:"template"`
	TemplateVersion string             `json:"template_version"`
	Size            ComputerSize       `json:"size"`
	State           string             `json:"state"`
	PauseMode       string             `json:"pause_mode,omitempty"`
	ParentID        string             `json:"parent_id,omitempty"`
	ForkMode        string             `json:"fork_mode,omitempty"`
	FromSnapshotID  string             `json:"from_snapshot_id,omitempty"`
	Internet        ComputerInternet   `json:"internet"`
	IdlePolicy      ComputerIdlePolicy `json:"idle_policy"`
	Access          ComputerAccess     `json:"access"`
	Logins          []string           `json:"logins,omitempty"`
	InputHolder     string             `json:"input_holder"`
	Ports           []ComputerPort     `json:"ports,omitempty"`
	SSHCommand      string             `json:"ssh_command"`
	NeedsYou        bool               `json:"needs_you"`
	AgentPaused     bool               `json:"agent_paused"`
	UsedBy          string             `json:"used_by,omitempty"`
	LastActiveAt    *time.Time         `json:"last_active_at,omitempty"`
	RegionName      string             `json:"region_name,omitempty"`
	Task            string             `json:"task,omitempty"`
	NextStep        string             `json:"next_step,omitempty"`
	TaskUpdatedAt   *time.Time         `json:"task_updated_at,omitempty"`
	ForkDurationMS  *int64             `json:"fork_duration_ms,omitempty"`
	LastOperationID string             `json:"last_operation_id,omitempty"`
	FailureCode     string             `json:"failure_code,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	StartedAt       *time.Time         `json:"started_at,omitempty"`
	PausedAt        *time.Time         `json:"paused_at,omitempty"`
}

type ComputerPage struct {
	Items      []Computer `json:"items"`
	NextCursor string     `json:"next_cursor,omitempty"`
	HasMore    bool       `json:"has_more"`
}

type ComputerResult struct {
	Computer    Computer `json:"computer"`
	OperationID string   `json:"operation_id"`
	Replayed    bool     `json:"replayed"`
}

type ComputerForkResult struct {
	Source      Computer   `json:"source"`
	Forks       []Computer `json:"forks"`
	OperationID string     `json:"operation_id"`
	Replayed    bool       `json:"replayed"`
}

type ComputerSource struct {
	Repository string `json:"repository"`
	Revision   string `json:"revision,omitempty"`
	Path       string `json:"path,omitempty"`
	Token      string `json:"token,omitempty"`
}

type ComputerSetupScript struct {
	Content string `json:"content"`
	CWD     string `json:"cwd,omitempty"`
}

type ComputerSetupDevcontainer struct {
	WorkspaceFolder string `json:"workspace_folder"`
	ConfigPath      string `json:"config_path,omitempty"`
}

type ComputerSetupSource struct {
	Repository string `json:"repository"`
	Revision   string `json:"revision,omitempty"`
	Path       string `json:"path,omitempty"`
	Directory  string `json:"directory,omitempty"`
	Token      string `json:"token,omitempty"`
}

type ComputerSetupInput struct {
	Script       *ComputerSetupScript       `json:"script,omitempty"`
	Devcontainer *ComputerSetupDevcontainer `json:"devcontainer,omitempty"`
	Source       *ComputerSetupSource       `json:"source,omitempty"`
}

func ComputerSeconds(value int64) *int64 { return &value }

type CreateComputerInput struct {
	ProjectID      string              `json:"project_id"`
	Name           string              `json:"name"`
	Region         string              `json:"region,omitempty"`
	Template       string              `json:"template"`
	FromSnapshotID string              `json:"from_snapshot_id,omitempty"`
	Source         *ComputerSource     `json:"source,omitempty"`
	Setup          *ComputerSetupInput `json:"setup,omitempty"`
	Size           ComputerSize        `json:"size"`
	Internet       *ComputerInternet   `json:"internet,omitempty"`
	IdlePolicy     *ComputerIdlePolicy `json:"idle_policy,omitempty"`
	Access         *ComputerAccess     `json:"access,omitempty"`
	Logins         []string            `json:"logins,omitempty"`

	MaxRunSeconds       *int64 `json:"max_run_seconds,omitempty"`
	ArchiveAfterSeconds *int64 `json:"archive_after_seconds,omitempty"`
	DeleteAfterSeconds  *int64 `json:"delete_after_seconds,omitempty"`
}

type UpdateComputerInput struct {
	Name       *string             `json:"name,omitempty"`
	Size       *ComputerSize       `json:"size,omitempty"`
	Internet   *ComputerInternet   `json:"internet,omitempty"`
	IdlePolicy *ComputerIdlePolicy `json:"idle_policy,omitempty"`
	Access     *ComputerAccess     `json:"access,omitempty"`
	Logins     []string            `json:"logins,omitempty"`

	MaxRunSeconds       *int64 `json:"max_run_seconds,omitempty"`
	ArchiveAfterSeconds *int64 `json:"archive_after_seconds,omitempty"`
	DeleteAfterSeconds  *int64 `json:"delete_after_seconds,omitempty"`
}

type ComputerForkInput struct {
	Mode         string   `json:"mode,omitempty"`
	Count        int      `json:"count"`
	Names        []string `json:"names,omitempty"`
	Instructions []string `json:"instructions,omitempty"`
}

type ComputerSnapshot struct {
	ID               string       `json:"id"`
	ProjectID        string       `json:"project_id"`
	ComputerID       string       `json:"computer_id"`
	Name             string       `json:"name"`
	Mode             string       `json:"mode"`
	Status           string       `json:"status"`
	Template         string       `json:"template"`
	TemplateVersion  string       `json:"template_version"`
	Size             ComputerSize `json:"size"`
	StoredBytes      int64        `json:"stored_bytes"`
	ComputersStarted int          `json:"computers_started"`
	CreatedAt        time.Time    `json:"created_at"`
	ReadyAt          *time.Time   `json:"ready_at,omitempty"`
}

type ComputerSnapshotPage struct {
	Items      []ComputerSnapshot `json:"items"`
	NextCursor string             `json:"next_cursor,omitempty"`
	HasMore    bool               `json:"has_more"`
}

type ComputerExecInput struct {
	Command        []string          `json:"command,omitempty"`
	Shell          string            `json:"shell,omitempty"`
	CWD            string            `json:"cwd,omitempty"`
	Environment    map[string]string `json:"environment,omitempty"`
	User           string            `json:"user,omitempty"`
	TimeoutSeconds int               `json:"timeout_seconds,omitempty"`
	Background     bool              `json:"background,omitempty"`
	Stdin          string            `json:"stdin,omitempty"`
	StdinOpen      bool              `json:"stdin_open,omitempty"`
}

type ComputerExecResult struct {
	ProcessID  string `json:"process_id"`
	ExitCode   *int   `json:"exit_code,omitempty"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	Truncated  bool   `json:"truncated"`
	TimedOut   bool   `json:"timed_out"`
	Signal     string `json:"signal,omitempty"`
	DurationMS int64  `json:"duration_ms"`
}

type ComputerFileEntry struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Type      string `json:"type"`
	SizeBytes int64  `json:"size_bytes"`
}

type ComputerPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ComputerAction struct {
	Type            string          `json:"type"`
	X               *int            `json:"x,omitempty"`
	Y               *int            `json:"y,omitempty"`
	Button          string          `json:"button,omitempty"`
	Path            []ComputerPoint `json:"path,omitempty"`
	ScrollX         *int            `json:"scroll_x,omitempty"`
	ScrollY         *int            `json:"scroll_y,omitempty"`
	Text            string          `json:"text,omitempty"`
	Keys            []string        `json:"keys,omitempty"`
	DurationMS      *int64          `json:"duration_ms,omitempty"`
	WindowID        string          `json:"window_id,omitempty"`
	URL             string          `json:"url,omitempty"`
	CredentialToken string          `json:"credential_token,omitempty"`
	Login           string          `json:"login,omitempty"`
	Submit          bool            `json:"submit,omitempty"`
	Screenshot      bool            `json:"screenshot,omitempty"`
}

type ComputerWindow struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	App     string `json:"app"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Focused bool   `json:"focused"`
}

type ComputerScreenshot struct {
	MediaType string    `json:"media_type"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Scale     float64   `json:"scale"`
	Data      string    `json:"data"`
	TakenAt   time.Time `json:"taken_at"`
}

type ComputerActionError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type ComputerActionResult struct {
	Type       string               `json:"type"`
	OK         bool                 `json:"ok"`
	Cursor     *ComputerPoint       `json:"cursor,omitempty"`
	Text       string               `json:"text,omitempty"`
	Windows    []ComputerWindow     `json:"windows,omitempty"`
	Screenshot *ComputerScreenshot  `json:"screenshot,omitempty"`
	Error      *ComputerActionError `json:"error,omitempty"`
}

type ComputerActionsRequest struct {
	Actor    string           `json:"actor"`
	StepID   string           `json:"step_id"`
	Actions  []ComputerAction `json:"actions"`
	Task     string           `json:"task,omitempty"`
	NextStep string           `json:"next_step,omitempty"`
}

type ComputerActOptions struct {
	Actor    string
	StepID   string
	Task     string
	NextStep string
}

type ComputerActionsResponse struct {
	StepID  string                 `json:"step_id"`
	Results []ComputerActionResult `json:"results"`
}

type ComputerInputLock struct {
	Holder    string     `json:"holder"`
	SessionID string     `json:"session_id,omitempty"`
	Since     *time.Time `json:"since,omitempty"`
}

type ComputerStep struct {
	Seq        int64          `json:"seq"`
	StepID     string         `json:"step_id"`
	Actor      string         `json:"actor"`
	Action     ComputerAction `json:"action"`
	OK         bool           `json:"ok"`
	ErrorCode  string         `json:"error_code,omitempty"`
	StartedAt  time.Time      `json:"started_at"`
	DurationMS int64          `json:"duration_ms"`
}

type ComputerStepPage struct {
	Steps   []ComputerStep `json:"steps"`
	NextSeq int64          `json:"next_seq"`
}

type ComputerRecording struct {
	RecordingID string     `json:"recording_id"`
	Path        string     `json:"path"`
	StartedAt   time.Time  `json:"started_at"`
	StoppedAt   *time.Time `json:"stopped_at,omitempty"`
	SizeBytes   int64      `json:"size_bytes"`
}

type ComputerApproval struct {
	ID               string     `json:"id"`
	ComputerID       string     `json:"computer_id"`
	Kind             string     `json:"kind"`
	Question         string     `json:"question"`
	Details          string     `json:"details,omitempty"`
	PageURL          string     `json:"page_url,omitempty"`
	AskedBy          string     `json:"asked_by,omitempty"`
	Status           string     `json:"status"`
	NextStep         string     `json:"next_step,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	DecidedAt        *time.Time `json:"decided_at,omitempty"`
	NotifiedAt       *time.Time `json:"notified_at,omitempty"`
	NotifiedChannels []string   `json:"notified_channels,omitempty"`
}

type ComputerSession struct {
	ID         string    `json:"id"`
	ComputerID string    `json:"computer_id"`
	Kind       string    `json:"kind"`
	URL        string    `json:"url"`
	Token      string    `json:"token"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type ComputerEvent struct {
	Seq        int64              `json:"seq"`
	Type       string             `json:"type"`
	ComputerID string             `json:"computer_id"`
	State      string             `json:"state,omitempty"`
	Step       *ComputerStep      `json:"step,omitempty"`
	InputLock  *ComputerInputLock `json:"input_lock,omitempty"`
	Approval   *ComputerApproval  `json:"approval,omitempty"`
	Snapshot   *ComputerSnapshot  `json:"snapshot,omitempty"`
	ForkIDs    []string           `json:"fork_ids,omitempty"`
	Port       *ComputerPort      `json:"port,omitempty"`
	Time       time.Time          `json:"time"`
}

type ComputerKeyLimits struct {
	MaxRunning    int           `json:"max_running,omitempty"`
	MaxSize       *ComputerSize `json:"max_size,omitempty"`
	Templates     []string      `json:"templates,omitempty"`
	SpendCapMinor int64         `json:"spend_cap_minor,omitempty"`
}

type ComputerKey struct {
	ID         string             `json:"id"`
	ProjectID  string             `json:"project_id"`
	Name       string             `json:"name"`
	Prefix     string             `json:"prefix"`
	Limits     *ComputerKeyLimits `json:"limits,omitempty"`
	Scopes     []ComputerKeyScope `json:"scopes,omitempty"`
	LastUsedAt *time.Time         `json:"last_used_at,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
	RevokedAt  *time.Time         `json:"revoked_at,omitempty"`
}

type ComputerCreatedKey struct {
	ComputerKey
	Secret string `json:"secret"`
}

type CreateComputerKeyInput struct {
	ProjectID string             `json:"project_id"`
	Name      string             `json:"name"`
	Limits    *ComputerKeyLimits `json:"limits,omitempty"`
	Scopes    []ComputerKeyScope `json:"scopes,omitempty"`
}

type ComputerAskInput struct {
	Kind     string `json:"kind,omitempty"`
	Question string `json:"question"`
	Details  string `json:"details,omitempty"`
	PageURL  string `json:"page_url,omitempty"`
	NextStep string `json:"next_step,omitempty"`
}

type ComputerStepOptions struct {
	AfterSeq  int64
	BeforeSeq int64
	Order     string
	Limit     int
}

type ComputerUsageLine struct {
	Kind        string  `json:"kind"`
	Detail      string  `json:"detail"`
	Amount      float64 `json:"amount"`
	VCPUs       int     `json:"vcpus,omitempty"`
	MemoryBytes int64   `json:"memory_bytes,omitempty"`
	Seconds     int64   `json:"seconds,omitempty"`
	GBMonths    float64 `json:"gb_months,omitempty"`
	GB          float64 `json:"gb,omitempty"`
}

type ComputerUsageItem struct {
	ComputerID     string              `json:"computer_id"`
	RunningSeconds int64               `json:"running_seconds"`
	Total          float64             `json:"total"`
	Lines          []ComputerUsageLine `json:"lines"`
}

type ComputerUsage struct {
	Month    string              `json:"month"`
	Currency string              `json:"currency"`
	Total    float64             `json:"total"`
	Items    []ComputerUsageItem `json:"items"`
}

type ComputerUsageOptions struct {
	Month      string
	ComputerID string
	Currency   string
}

type ComputerPricing struct {
	Currency        string    `json:"currency"`
	VCPUHour        float64   `json:"vcpu_hour"`
	MemoryGiBHour   float64   `json:"memory_gib_hour"`
	DiskGBMonth     float64   `json:"disk_gb_month"`
	SnapshotGBMonth float64   `json:"snapshot_gb_month"`
	EgressGB        float64   `json:"egress_gb"`
	EffectiveAt     time.Time `json:"effective_at"`
}

type ComputerProcess struct {
	PID         int     `json:"pid"`
	Name        string  `json:"name"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryBytes int64   `json:"memory_bytes"`
}

type ComputerProcessList struct {
	Processes        []ComputerProcess `json:"processes"`
	MemoryUsedBytes  int64             `json:"memory_used_bytes"`
	MemoryTotalBytes int64             `json:"memory_total_bytes"`
	DiskUsedBytes    int64             `json:"disk_used_bytes"`
	DiskTotalBytes   int64             `json:"disk_total_bytes"`
}

type ComputerLogEntry struct {
	Time   time.Time `json:"time"`
	Source string    `json:"source"`
	Line   string    `json:"line"`
}

type ComputerLogPage struct {
	Entries    []ComputerLogEntry `json:"entries"`
	NextCursor string             `json:"next_cursor,omitempty"`
}

type ComputerLogOptions struct {
	Source string
	Cursor string
	Limit  int
}

type ComputerPublicPricing struct {
	NGN           ComputerPricing            `json:"ngn"`
	USD           ComputerPricing            `json:"usd"`
	FreePlan      ComputerFreePlan           `json:"free_plan"`
	RunningLimits []ComputerPlanRunningLimit `json:"running_limits"`
}

type ComputerFreePlan struct {
	Computers            int   `json:"computers"`
	VCPUs                int   `json:"vcpus"`
	MemoryBytes          int64 `json:"memory_bytes"`
	RunningHoursPerMonth int   `json:"running_hours_per_month"`
}

type ComputerPlanRunningLimit struct {
	Plan      string `json:"plan"`
	Computers int    `json:"computers"`
}

type ComputerRegion struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

type ComputerRegionList struct {
	Items []ComputerRegion `json:"items"`
}

type ComputerSystemService struct {
	Unit        string `json:"unit"`
	Description string `json:"description"`
	Active      string `json:"active"`
	Sub         string `json:"sub"`
	Enabled     bool   `json:"enabled"`
}

type ComputerListeningPort struct {
	Port      int    `json:"port"`
	Address   string `json:"address"`
	Protocol  string `json:"protocol"`
	Process   string `json:"process,omitempty"`
	Published bool   `json:"published"`
	URL       string `json:"url,omitempty"`
}

type ComputerSystem struct {
	ComputerID string                  `json:"computer_id"`
	Live       bool                    `json:"live"`
	Region     string                  `json:"region"`
	RegionName string                  `json:"region_name"`
	OSName     string                  `json:"os_name"`
	OSVersion  string                  `json:"os_version"`
	Kernel     string                  `json:"kernel"`
	ReadAt     *time.Time              `json:"read_at,omitempty"`
	Services   []ComputerSystemService `json:"services"`
	Listening  []ComputerListeningPort `json:"listening"`
	Idle       bool                    `json:"idle"`
	IdleSince  *time.Time              `json:"idle_since,omitempty"`
	PausesAt   *time.Time              `json:"pauses_at,omitempty"`
}

type ComputerWindowList struct {
	Windows []ComputerWindow `json:"windows"`
}

type ComputerTask struct {
	Task      string    `json:"task"`
	NextStep  string    `json:"next_step"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ComputerTaskInput struct {
	Task     *string `json:"task,omitempty"`
	NextStep *string `json:"next_step,omitempty"`
}

type ComputerMoney struct {
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
}

type ComputerNotificationChoices struct {
	Email           bool   `json:"email"`
	Push            bool   `json:"push"`
	Slack           bool   `json:"slack"`
	EmailAddress    string `json:"email_address"`
	SlackConfigured bool   `json:"slack_configured"`
	PushAvailable   bool   `json:"push_available"`
	PushDevices     int    `json:"push_devices"`
}

type ComputerAccountSettings struct {
	Plan              string                      `json:"plan"`
	CanEdit           bool                        `json:"can_edit"`
	RunningNow        int                         `json:"running_now"`
	SpendCap          *ComputerMoney              `json:"spend_cap"`
	SpentThisMonth    ComputerMoney               `json:"spent_this_month"`
	DefaultIdlePolicy *ComputerIdlePolicy         `json:"default_idle_policy"`
	DefaultRecording  bool                        `json:"default_recording"`
	DefaultApprovals  bool                        `json:"default_approvals"`
	Notifications     ComputerNotificationChoices `json:"notifications"`
}

type ComputerNotificationUpdate struct {
	Email *bool `json:"email,omitempty"`
	Push  *bool `json:"push,omitempty"`
	Slack *bool `json:"slack,omitempty"`
}

type SaveComputerSettingsInput struct {
	SpendCap               *ComputerMoney
	ClearSpendCap          bool
	DefaultIdlePolicy      *ComputerIdlePolicy
	ClearDefaultIdlePolicy bool
	DefaultRecording       *bool
	DefaultApprovals       *bool
	Notifications          *ComputerNotificationUpdate
}

func (i SaveComputerSettingsInput) MarshalJSON() ([]byte, error) {
	body := map[string]any{}
	if i.ClearSpendCap {
		body["spend_cap"] = nil
	} else if i.SpendCap != nil {
		body["spend_cap"] = i.SpendCap
	}
	if i.ClearDefaultIdlePolicy {
		body["default_idle_policy"] = nil
	} else if i.DefaultIdlePolicy != nil {
		body["default_idle_policy"] = i.DefaultIdlePolicy
	}
	if i.DefaultRecording != nil {
		body["default_recording"] = *i.DefaultRecording
	}
	if i.DefaultApprovals != nil {
		body["default_approvals"] = *i.DefaultApprovals
	}
	if i.Notifications != nil {
		body["notifications"] = i.Notifications
	}
	return json.Marshal(body)
}

type ComputerSetupStep struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	ExitCode *int   `json:"exit_code,omitempty"`
}

type ComputerSetupImage struct {
	Reference string `json:"reference,omitempty"`
	Digest    string `json:"digest,omitempty"`
	Service   bool   `json:"service,omitempty"`
}

type ComputerSetupImport struct {
	Name        string `json:"name,omitempty"`
	State       string `json:"state,omitempty"`
	SnapshotID  string `json:"snapshot_id,omitempty"`
	KeepBuilder bool   `json:"keep_builder,omitempty"`
}

type ComputerSetup struct {
	ID         string               `json:"id"`
	ComputerID string               `json:"computer_id"`
	ProjectID  string               `json:"project_id"`
	Kind       string               `json:"kind"`
	Spec       map[string]any       `json:"spec"`
	State      string               `json:"state"`
	ExitCode   *int                 `json:"exit_code,omitempty"`
	Error      string               `json:"error,omitempty"`
	Steps      []ComputerSetupStep  `json:"steps"`
	Image      *ComputerSetupImage  `json:"image,omitempty"`
	Import     *ComputerSetupImport `json:"import,omitempty"`
	Log        string               `json:"log"`
	LogBytes   int64                `json:"log_bytes"`
	CreatedAt  time.Time            `json:"created_at"`
	StartedAt  *time.Time           `json:"started_at,omitempty"`
	FinishedAt *time.Time           `json:"finished_at,omitempty"`
	UpdatedAt  time.Time            `json:"updated_at"`
}

type ComputerRegistryLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ImportComputerImageInput struct {
	ProjectID           string                 `json:"project_id"`
	Image               string                 `json:"image"`
	Name                string                 `json:"name"`
	Template            string                 `json:"template,omitempty"`
	Region              string                 `json:"region,omitempty"`
	Size                *ComputerSize          `json:"size,omitempty"`
	Registry            *ComputerRegistryLogin `json:"registry,omitempty"`
	NoEntrypoint        bool                   `json:"no_entrypoint,omitempty"`
	NoVerify            bool                   `json:"no_verify,omitempty"`
	KeepBuilder         bool                   `json:"keep_builder,omitempty"`
	ComputerName        string                 `json:"computer_name,omitempty"`
	Internet            *ComputerInternet      `json:"internet,omitempty"`
	IdlePolicy          *ComputerIdlePolicy    `json:"idle_policy,omitempty"`
	Access              *ComputerAccess        `json:"access,omitempty"`
	Logins              []string               `json:"logins,omitempty"`
	MaxRunSeconds       int64                  `json:"max_run_seconds,omitempty"`
	ArchiveAfterSeconds int64                  `json:"archive_after_seconds,omitempty"`
	DeleteAfterSeconds  int64                  `json:"delete_after_seconds,omitempty"`
}

type ComputerImageImportResult struct {
	Computer Computer      `json:"computer"`
	Setup    ComputerSetup `json:"setup"`
}

type ComputerCIConnection struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	UserID         string    `json:"user_id,omitempty"`
	InstallationID int64     `json:"installation_id"`
	AccountLogin   string    `json:"account_login"`
	AccountType    string    `json:"account_type"`
	CreatedAt      time.Time `json:"created_at"`
}

type ComputerCIRepository struct {
	FullName string `json:"full_name,omitempty"`
	Private  bool   `json:"private,omitempty"`
	HTMLURL  string `json:"html_url,omitempty"`
}

type ComputerCIOverviewConnection struct {
	ComputerCIConnection
	Repositories      []ComputerCIRepository `json:"repositories,omitempty"`
	RepositoriesError string                 `json:"repositories_error,omitempty"`
}

type ComputerCIInstallation struct {
	InstallationID  int64  `json:"installation_id,omitempty"`
	AccountLogin    string `json:"account_login,omitempty"`
	AccountType     string `json:"account_type,omitempty"`
	LinkedProjectID string `json:"linked_project_id,omitempty"`
}

type ComputerCIJob struct {
	Provider          string     `json:"provider"`
	ResourceClass     string     `json:"resource_class,omitempty"`
	ID                string     `json:"id"`
	ProjectID         string     `json:"project_id"`
	GitHubJobID       int64      `json:"github_job_id"`
	GitHubRunID       int64      `json:"github_run_id"`
	Repository        string     `json:"repository"`
	RepositoryPrivate bool       `json:"repository_private,omitempty"`
	WorkflowName      string     `json:"workflow_name"`
	JobName           string     `json:"job_name"`
	HeadBranch        string     `json:"head_branch,omitempty"`
	HeadSHA           string     `json:"head_sha,omitempty"`
	Labels            []string   `json:"labels"`
	Size              string     `json:"size"`
	SnapshotName      string     `json:"snapshot_name,omitempty"`
	SnapshotID        string     `json:"snapshot_id,omitempty"`
	State             string     `json:"state"`
	Conclusion        string     `json:"conclusion,omitempty"`
	Reason            string     `json:"reason,omitempty"`
	ComputerID        string     `json:"computer_id,omitempty"`
	RunnerName        string     `json:"runner_name,omitempty"`
	QueuedAt          time.Time  `json:"queued_at"`
	RunnerStartedAt   *time.Time `json:"runner_started_at,omitempty"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
}

type ComputerCIOverview struct {
	Connections         []ComputerCIOverviewConnection `json:"connections"`
	Installations       []ComputerCIInstallation       `json:"installations"`
	CircleCIConnections []ComputerCircleCIConnection   `json:"circleci_connections"`
	Sizes               []ComputerCISizeInfo           `json:"sizes"`
	Jobs                []ComputerCIJob                `json:"jobs"`
	Labels              []string                       `json:"labels"`
}

type ComputerCISizeInfo struct {
	Name        string `json:"name"`
	VCPUs       int    `json:"vcpus"`
	MemoryBytes int64  `json:"memory_bytes"`
}

type ComputerCircleCIConnection struct {
	ID                string    `json:"id"`
	ProjectID         string    `json:"project_id"`
	ResourceClass     string    `json:"resource_class"`
	Size              string    `json:"size"`
	SnapshotName      string    `json:"snapshot_name,omitempty"`
	MaxConcurrentJobs int       `json:"max_concurrent_jobs"`
	StatusReason      string    `json:"status_reason,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type ComputerCircleCIConnectInput struct {
	ProjectID         string `json:"project_id"`
	ResourceClass     string `json:"resource_class"`
	Size              string `json:"size"`
	SnapshotName      string `json:"snapshot_name,omitempty"`
	MaxConcurrentJobs int    `json:"max_concurrent_jobs,omitempty"`
	APIToken          string `json:"api_token"`
}
