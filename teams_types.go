package stackshift

import "time"

const (
	ComputerErrorTeamNotFound     = "team_not_found"
	ComputerErrorMemberNotFound   = "member_not_found"
	ComputerErrorTeamInvalidState = "team_invalid_state"
)

const (
	TeamStateStarting = "starting"
	TeamStateWorking  = "working"
	TeamStatePaused   = "paused"
	TeamStateDeleting = "deleting"
	TeamStateDeleted  = "deleted"
)

const (
	TeamStopQuiet = "quiet"
	TeamStopDone  = "done"
	TeamStopNever = "never"
)

type TeamMoney struct {
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
}

type TeamShared struct {
	Enabled       bool   `json:"enabled"`
	Path          string `json:"path,omitempty"`
	CapacityBytes int64  `json:"capacity_bytes,omitempty"`
}

type TeamNetwork struct {
	Enabled bool   `json:"enabled"`
	Domain  string `json:"domain,omitempty"`
}

type TeamRepository struct {
	Repository string `json:"repository"`
	Revision   string `json:"revision,omitempty"`
	Status     string `json:"status,omitempty"`
	Commit     string `json:"commit,omitempty"`
	Error      string `json:"error,omitempty"`
}

type TeamMember struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Agent         string       `json:"agent"`
	Role          string       `json:"role"`
	ComputerID    string       `json:"computer_id"`
	ComputerState string       `json:"computer_state"`
	Status        string       `json:"status"`
	Hostname      string       `json:"hostname"`
	Address       string       `json:"address"`
	Size          ComputerSize `json:"size"`
	Template      string       `json:"template"`
	LastActiveAt  *time.Time   `json:"last_active_at,omitempty"`
	JoinedAt      time.Time    `json:"joined_at"`
}

type Team struct {
	ID                string          `json:"id"`
	ProjectID         string          `json:"project_id"`
	Name              string          `json:"name"`
	Job               string          `json:"job"`
	State             string          `json:"state"`
	PauseReason       string          `json:"pause_reason,omitempty"`
	StopRule          string          `json:"stop_rule"`
	QuietAfterSeconds int             `json:"quiet_after_seconds"`
	SpendLimit        *TeamMoney      `json:"spend_limit"`
	Spent             TeamMoney       `json:"spent"`
	Shared            TeamShared      `json:"shared"`
	Network           TeamNetwork     `json:"network"`
	AskBeforeActing   bool            `json:"ask_before_acting"`
	Repository        *TeamRepository `json:"repository"`
	Members           []TeamMember    `json:"members"`
	RunningCount      int             `json:"running_count"`
	DoneSummary       string          `json:"done_summary,omitempty"`
	LastMessageAt     *time.Time      `json:"last_message_at,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	PausedAt          *time.Time      `json:"paused_at,omitempty"`
}

type TeamPage struct {
	Items      []Team `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

type TeamMessageParty struct {
	Kind     string `json:"kind"`
	MemberID string `json:"member_id,omitempty"`
	Name     string `json:"name,omitempty"`
}

type TeamMessage struct {
	Seq       int64            `json:"seq"`
	Kind      string           `json:"kind"`
	From      TeamMessageParty `json:"from"`
	To        TeamMessageParty `json:"to"`
	Text      string           `json:"text"`
	File      string           `json:"file,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

type TeamMessagePage struct {
	Items   []TeamMessage `json:"items"`
	NextSeq int64         `json:"next_seq"`
	HasMore bool          `json:"has_more"`
}

type TeamFileEntry struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Type       string    `json:"type"`
	SizeBytes  int64     `json:"size_bytes"`
	ModifiedAt time.Time `json:"modified_at"`
	ChangedBy  string    `json:"changed_by,omitempty"`
}

type TeamFiles struct {
	Path          string          `json:"path"`
	UsedBytes     int64           `json:"used_bytes"`
	CapacityBytes int64           `json:"capacity_bytes"`
	Entries       []TeamFileEntry `json:"entries"`
}

type TeamNetworkPort struct {
	Port    int    `json:"port"`
	Process string `json:"process"`
}

type TeamNetworkMember struct {
	MemberID string            `json:"member_id"`
	Name     string            `json:"name"`
	Hostname string            `json:"hostname"`
	Address  string            `json:"address"`
	Ports    []TeamNetworkPort `json:"ports"`
}

type TeamNetworkMap struct {
	Domain  string              `json:"domain"`
	Members []TeamNetworkMember `json:"members"`
}

type TeamMemberActivity struct {
	MemberID string   `json:"member_id"`
	Name     string   `json:"name"`
	CWD      string   `json:"cwd"`
	Command  string   `json:"command"`
	Lines    []string `json:"lines"`
}

type TeamActivity struct {
	Members []TeamMemberActivity `json:"members"`
}

type TeamMemberSizeInput struct {
	VCPUs       int   `json:"vcpus,omitempty"`
	MemoryBytes int64 `json:"memory_bytes,omitempty"`
}

type TeamMemberInput struct {
	Name          string               `json:"name"`
	Agent         string               `json:"agent"`
	Role          string               `json:"role,omitempty"`
	Size          *TeamMemberSizeInput `json:"size,omitempty"`
	Template      string               `json:"template,omitempty"`
	ModelSecretID string               `json:"model_secret_id,omitempty"`
}

type TeamRepositoryInput struct {
	Repository string `json:"repository"`
	Revision   string `json:"revision,omitempty"`
	Token      string `json:"token,omitempty"`
}

type TeamSharedInput struct {
	Enabled       bool  `json:"enabled"`
	CapacityBytes int64 `json:"capacity_bytes,omitempty"`
}

type TeamNetworkInput struct {
	Enabled bool `json:"enabled"`
}

type CreateTeamInput struct {
	ProjectID       string               `json:"project_id"`
	Name            string               `json:"name"`
	Job             string               `json:"job"`
	Repository      *TeamRepositoryInput `json:"repository,omitempty"`
	Members         []TeamMemberInput    `json:"members"`
	Shared          *TeamSharedInput     `json:"shared,omitempty"`
	Network         *TeamNetworkInput    `json:"network,omitempty"`
	AskBeforeActing bool                 `json:"ask_before_acting,omitempty"`
	StopRule        string               `json:"stop_rule,omitempty"`
	SpendLimit      *TeamMoney           `json:"spend_limit,omitempty"`
	Region          string               `json:"region,omitempty"`
}

type UpdateTeamInput struct {
	Job             *string
	StopRule        *string
	SpendLimit      *TeamMoney
	ClearSpendLimit bool
	AskBeforeActing *bool
}

type SendTeamMessageInput struct {
	Text     string `json:"text"`
	To       string `json:"to,omitempty"`
	File     string `json:"file,omitempty"`
	AsMember string `json:"as_member,omitempty"`
}
