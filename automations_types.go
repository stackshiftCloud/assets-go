package stackshift

import (
	"encoding/json"
	"time"
)

type AutomationDefinition struct {
	Version int              `json:"version"`
	Nodes   []AutomationNode `json:"nodes"`
	Edges   []AutomationEdge `json:"edges"`
}

type AutomationNode struct {
	ID             string           `json:"id"`
	Type           string           `json:"type"`
	Name           string           `json:"name,omitempty"`
	Action         string           `json:"action,omitempty"`
	Config         json.RawMessage  `json:"config,omitempty"`
	Retry          *AutomationRetry `json:"retry,omitempty"`
	TimeoutSeconds int              `json:"timeout_seconds,omitempty"`
	FailurePolicy  string           `json:"failure_policy,omitempty"`
}

type AutomationRetry struct {
	MaxAttempts    int `json:"max_attempts"`
	BackoffSeconds int `json:"backoff_seconds"`
}

type AutomationEdge struct {
	ID         string `json:"id"`
	Source     string `json:"source"`
	SourcePort string `json:"source_port,omitempty"`
	Target     string `json:"target"`
	TargetPort string `json:"target_port,omitempty"`
}

type AutomationPerson struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AutomationTrigger struct {
	Type       string        `json:"type"`
	EventTypes []string      `json:"event_types,omitempty"`
	Filters    *EventFilters `json:"filters,omitempty"`
	Schedule   string        `json:"schedule,omitempty"`
	Timezone   string        `json:"timezone,omitempty"`
	HasWebhook bool          `json:"has_webhook"`
}

type AutomationAccess struct {
	Actions   []string `json:"actions"`
	CanWiden  bool     `json:"can_widen"`
	CanChange bool     `json:"can_change"`
}

type AutomationSecret struct {
	Name      string           `json:"name"`
	Host      string           `json:"host"`
	Header    string           `json:"header"`
	Prefix    string           `json:"prefix"`
	SetBy     AutomationPerson `json:"set_by"`
	UpdatedAt time.Time        `json:"updated_at"`
	UsedBy    []string         `json:"used_by"`
}

type Automation struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	TeamID       *string               `json:"team_id"`
	CreatedBy    AutomationPerson      `json:"created_by"`
	Enabled      bool                  `json:"enabled"`
	Revision     int64                 `json:"revision"`
	Trigger      AutomationTrigger     `json:"trigger"`
	Definition   *AutomationDefinition `json:"definition,omitempty"`
	Access       AutomationAccess      `json:"access"`
	Secrets      []AutomationSecret    `json:"secrets"`
	LastRun      *AutomationRunSummary `json:"last_run"`
	RunsThisWeek int                   `json:"runs_this_week"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

type AutomationList struct {
	Automations []Automation           `json:"automations"`
	Happening   []AutomationRunSummary `json:"happening"`
}

type AutomationRunTrigger struct {
	Type      string `json:"type"`
	Label     string `json:"label"`
	EventType string `json:"event_type,omitempty"`
	EventID   string `json:"event_id,omitempty"`
}

type AutomationRunApproval struct {
	ID        string            `json:"id"`
	NodeID    string            `json:"node_id"`
	StepName  string            `json:"step_name"`
	Action    string            `json:"action"`
	Tier      string            `json:"tier"`
	Summary   string            `json:"summary"`
	Status    string            `json:"status"`
	DecidedBy *AutomationPerson `json:"decided_by"`
	ExpiresAt *time.Time        `json:"expires_at"`
}

type AutomationRunSummary struct {
	ID             string                 `json:"id"`
	AutomationID   string                 `json:"automation_id"`
	AutomationName string                 `json:"automation_name"`
	Number         int64                  `json:"number"`
	Test           bool                   `json:"test"`
	Status         string                 `json:"status"`
	Trigger        AutomationRunTrigger   `json:"trigger"`
	StartedBy      *AutomationPerson      `json:"started_by"`
	StepsDone      int                    `json:"steps_done"`
	StepsTotal     int                    `json:"steps_total"`
	StoppedAtStep  *int                   `json:"stopped_at_step"`
	Approval       *AutomationRunApproval `json:"approval"`
	StartedAt      *time.Time             `json:"started_at"`
	CompletedAt    *time.Time             `json:"completed_at"`
	DurationMs     *int64                 `json:"duration_ms"`
	CreatedAt      time.Time              `json:"created_at"`
}

type AutomationRunStep struct {
	NodeID      string          `json:"node_id"`
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Action      string          `json:"action,omitempty"`
	Tier        string          `json:"tier,omitempty"`
	Status      string          `json:"status"`
	Input       json.RawMessage `json:"input"`
	Output      json.RawMessage `json:"output"`
	Error       json.RawMessage `json:"error"`
	StartedAt   *time.Time      `json:"started_at"`
	CompletedAt *time.Time      `json:"completed_at"`
	DurationMs  *int64          `json:"duration_ms"`
}

type AutomationRunFailure struct {
	Code            string    `json:"code"`
	NodeID          string    `json:"node_id"`
	Step            int       `json:"step"`
	Action          string    `json:"action"`
	Resource        string    `json:"resource"`
	Needs           string    `json:"needs"`
	CanDoNow        []string  `json:"can_do_now"`
	FailedRunsSince int       `json:"failed_runs_since"`
	Since           time.Time `json:"since"`
}

type AutomationRun struct {
	AutomationRunSummary
	Steps   []AutomationRunStep   `json:"steps"`
	Failure *AutomationRunFailure `json:"failure"`
}

type AutomationRunCounts struct {
	All     int `json:"all"`
	Running int `json:"running"`
	Waiting int `json:"waiting"`
	Failed  int `json:"failed"`
	Done    int `json:"done"`
}

type AutomationRunPage struct {
	Runs       []AutomationRunSummary `json:"runs"`
	NextCursor *string                `json:"next_cursor"`
	Counts     AutomationRunCounts    `json:"counts"`
}

type AutomationRunListOptions struct {
	TeamID       string
	AutomationID string
	Status       string
	Cursor       string
	Limit        int
}

type AutomationAction struct {
	Key       string          `json:"key"`
	Product   string          `json:"product"`
	Summary   string          `json:"summary"`
	Tier      string          `json:"tier"`
	Input     json.RawMessage `json:"input,omitempty"`
	Available bool            `json:"available"`
	Reason    string          `json:"reason,omitempty"`
}

type AutomationPort struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}

type AutomationStepType struct {
	Type        string           `json:"type"`
	Name        string           `json:"name"`
	Category    string           `json:"category"`
	InputPorts  []AutomationPort `json:"input_ports"`
	OutputPorts []AutomationPort `json:"output_ports"`
}

type AutomationCatalog struct {
	Actions []AutomationAction   `json:"actions"`
	Steps   []AutomationStepType `json:"steps"`
}

type AutomationTemplate struct {
	Key         string               `json:"key"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Category    string               `json:"category"`
	Definition  AutomationDefinition `json:"definition"`
}

type AutomationIssue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	NodeID  string `json:"node_id,omitempty"`
	EdgeID  string `json:"edge_id,omitempty"`
}

type AutomationWebhook struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

type CreateAutomationInput struct {
	TeamID      string                `json:"team_id,omitempty"`
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	TemplateKey string                `json:"template_key,omitempty"`
	Definition  *AutomationDefinition `json:"definition,omitempty"`
}

type SaveAutomationInput struct {
	Name             string               `json:"name"`
	Description      string               `json:"description,omitempty"`
	Definition       AutomationDefinition `json:"definition"`
	ExpectedRevision int64                `json:"expected_revision"`
}

type StartAutomationRunInput struct {
	Input json.RawMessage `json:"input,omitempty"`
	Test  bool            `json:"test,omitempty"`
}

type RerunAutomationRunInput struct {
	FromNodeID string `json:"from_node_id,omitempty"`
}

type RerunFailedAutomationRunsInput struct {
	FromNodeID string    `json:"from_node_id"`
	Since      time.Time `json:"since"`
}

type DecideAutomationRunInput struct {
	Decision    string `json:"decision"`
	Note        string `json:"note,omitempty"`
	ConfirmName string `json:"confirm_name,omitempty"`
}

type AutomationSecretInput struct {
	Host   string `json:"host"`
	Header string `json:"header"`
	Prefix string `json:"prefix,omitempty"`
	Value  string `json:"value"`
}

type AutomationTriggerPreviewInput struct {
	TeamID     string        `json:"team_id,omitempty"`
	EventTypes []string      `json:"event_types"`
	Filters    *EventFilters `json:"filters,omitempty"`
}
