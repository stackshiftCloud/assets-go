package stackshift

import (
	"encoding/json"
	"time"
)

type AssetWorkflowDefinition struct {
	Version int                 `json:"version"`
	Nodes   []AssetWorkflowNode `json:"nodes"`
	Edges   []AssetWorkflowEdge `json:"edges"`
}

type AssetWorkflowNode struct {
	ID             string          `json:"id"`
	Type           string          `json:"type"`
	Action         string          `json:"action,omitempty"`
	ActionVersion  string          `json:"action_version,omitempty"`
	Config         json.RawMessage `json:"config,omitempty"`
	Retry          WorkflowRetry   `json:"retry,omitempty"`
	TimeoutSeconds int             `json:"timeout_seconds,omitempty"`
	FailurePolicy  string          `json:"failure_policy,omitempty"`
}

type AssetWorkflowEdge struct {
	ID         string `json:"id"`
	Source     string `json:"source"`
	SourcePort string `json:"source_port,omitempty"`
	Target     string `json:"target"`
	TargetPort string `json:"target_port,omitempty"`
}

type WorkflowRetry struct {
	MaxAttempts    int `json:"max_attempts,omitempty"`
	BackoffSeconds int `json:"backoff_seconds,omitempty"`
}

type AssetWorkflowLayout struct {
	Nodes    []AssetWorkflowNodeLayout `json:"nodes"`
	Viewport AssetWorkflowViewport     `json:"viewport"`
	Groups   []AssetWorkflowGroup      `json:"groups,omitempty"`
}

type AssetWorkflowNodeLayout struct {
	ID     string  `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}
type AssetWorkflowViewport struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom float64 `json:"zoom"`
}
type AssetWorkflowGroup struct {
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	NodeIDs []string `json:"node_ids"`
}
type WorkflowPortDefinition struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	Required bool   `json:"required,omitempty"`
}
type WorkflowNodeDefinition struct {
	Type             string                   `json:"type"`
	Version          string                   `json:"version"`
	Name             string                   `json:"name"`
	Category         string                   `json:"category"`
	Description      string                   `json:"description"`
	InputPorts       []WorkflowPortDefinition `json:"input_ports"`
	OutputPorts      []WorkflowPortDefinition `json:"output_ports"`
	ConfigSchema     json.RawMessage          `json:"config_schema"`
	Risk             string                   `json:"risk"`
	RequiresApproval bool                     `json:"requires_approval"`
	ApprovalMode     string                   `json:"approval_mode"`
}

type AssetWorkflow struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description,omitempty"`
	Status               string          `json:"status"`
	Revision             int64           `json:"revision"`
	ExecutionPrincipalID string          `json:"execution_principal_id"`
	DraftDefinition      json.RawMessage `json:"draft_definition"`
	DraftLayout          json.RawMessage `json:"draft_layout"`
	CurrentVersionID     string          `json:"current_version_id,omitempty"`
	ActiveVersionID      string          `json:"active_version_id,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}
type AssetWorkflowVersion struct {
	ID             string          `json:"id"`
	WorkflowID     string          `json:"workflow_id"`
	Version        int             `json:"version"`
	SemanticDigest string          `json:"semantic_digest"`
	Definition     json.RawMessage `json:"definition"`
	Layout         json.RawMessage `json:"layout"`
	CreatedBy      string          `json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
}
type AssetWorkflowRun struct {
	ID           string          `json:"id"`
	WorkflowID   string          `json:"workflow_id"`
	VersionID    string          `json:"version_id"`
	DurableRunID string          `json:"durable_run_id,omitempty"`
	TriggerType  string          `json:"trigger_type"`
	TriggerKey   string          `json:"trigger_key,omitempty"`
	Status       string          `json:"status"`
	Input        json.RawMessage `json:"input"`
	Output       json.RawMessage `json:"output,omitempty"`
	Error        json.RawMessage `json:"error,omitempty"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
type AssetWorkflowNodeExecution struct {
	ID              string                         `json:"id"`
	RunID           string                         `json:"run_id"`
	NodeID          string                         `json:"node_id"`
	NodeType        string                         `json:"node_type"`
	Status          string                         `json:"status"`
	Attempt         int                            `json:"attempt"`
	Input           json.RawMessage                `json:"input,omitempty"`
	Output          json.RawMessage                `json:"output,omitempty"`
	Error           json.RawMessage                `json:"error,omitempty"`
	SelectedPorts   []string                       `json:"selected_ports"`
	WaitEvent       string                         `json:"wait_event,omitempty"`
	WaitCorrelation string                         `json:"wait_correlation,omitempty"`
	WaitUntil       *time.Time                     `json:"wait_until,omitempty"`
	StartedAt       *time.Time                     `json:"started_at,omitempty"`
	CompletedAt     *time.Time                     `json:"completed_at,omitempty"`
	StackieActivity []AssetWorkflowStackieActivity `json:"stackie_activity,omitempty"`
}
type AssetWorkflowStackieActivity struct {
	ID           string          `json:"id"`
	RunID        string          `json:"run_id"`
	NodeID       string          `json:"node_id"`
	InvocationID string          `json:"invocation_id"`
	Sequence     int             `json:"sequence"`
	ActivityType string          `json:"activity_type"`
	ToolName     string          `json:"tool_name,omitempty"`
	Status       string          `json:"status"`
	Input        json.RawMessage `json:"input,omitempty"`
	Output       json.RawMessage `json:"output,omitempty"`
	InputTokens  int64           `json:"input_tokens"`
	OutputTokens int64           `json:"output_tokens"`
	CostMicros   int64           `json:"cost_micros"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
type AssetWorkflowApproval struct {
	ID             string          `json:"id"`
	WorkflowID     string          `json:"workflow_id"`
	VersionID      string          `json:"version_id"`
	RunID          string          `json:"run_id,omitempty"`
	NodeID         string          `json:"node_id,omitempty"`
	ApprovalMode   string          `json:"approval_mode"`
	SemanticDigest string          `json:"semantic_digest"`
	Status         string          `json:"status"`
	RiskSummary    json.RawMessage `json:"risk_summary"`
	Constraints    json.RawMessage `json:"constraints"`
	RequestedBy    string          `json:"requested_by"`
	DecidedBy      string          `json:"decided_by,omitempty"`
	DecisionReason string          `json:"decision_reason,omitempty"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}
type AssetWorkflowHTTPConnection struct {
	ID                  string                            `json:"id"`
	Name                string                            `json:"name"`
	Host                string                            `json:"host"`
	AllowedPathPrefixes []string                          `json:"allowed_path_prefixes"`
	CredentialReference string                            `json:"credential_reference,omitempty"`
	Credential          *AssetWorkflowHTTPCredentialInput `json:"credential,omitempty"`
	DefaultHeaders      json.RawMessage                   `json:"default_headers"`
	Revision            int64                             `json:"revision"`
	CreatedAt           time.Time                         `json:"created_at"`
	UpdatedAt           time.Time                         `json:"updated_at"`
}
type CreateAssetWorkflowHTTPConnectionInput struct {
	Name                string                            `json:"name"`
	Host                string                            `json:"host"`
	AllowedPathPrefixes []string                          `json:"allowed_path_prefixes"`
	Credential          *AssetWorkflowHTTPCredentialInput `json:"credential,omitempty"`
	DefaultHeaders      json.RawMessage                   `json:"default_headers,omitempty"`
}
type AssetWorkflowHTTPCredentialInput struct {
	Name   string `json:"name"`
	Header string `json:"header"`
	Prefix string `json:"prefix,omitempty"`
	Value  string `json:"value"`
}
type AssetWorkflowValidationIssue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	NodeID  string `json:"node_id,omitempty"`
	EdgeID  string `json:"edge_id,omitempty"`
}
type AssetWorkflowValidation struct {
	Valid          bool                           `json:"valid"`
	SemanticDigest string                         `json:"semantic_digest,omitempty"`
	Definition     AssetWorkflowDefinition        `json:"definition"`
	Issues         []AssetWorkflowValidationIssue `json:"issues"`
}
type AssetWorkflowTemplate struct {
	Key         string                         `json:"key"`
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	Definition  AssetWorkflowDefinition        `json:"definition"`
	Layout      AssetWorkflowLayout            `json:"layout"`
	Valid       bool                           `json:"valid"`
	Issues      []AssetWorkflowValidationIssue `json:"issues"`
}
type AssetWorkflowCatalog struct {
	Nodes     []WorkflowNodeDefinition `json:"nodes"`
	Templates []AssetWorkflowTemplate  `json:"templates"`
}
type AssetWorkflowRunDetail struct {
	Run   AssetWorkflowRun             `json:"run"`
	Nodes []AssetWorkflowNodeExecution `json:"nodes"`
}
type AssetWorkflowHook struct {
	Hook struct {
		ID            string `json:"id"`
		WorkflowID    string `json:"workflow_id"`
		SecretVersion int    `json:"secret_version"`
		Enabled       bool   `json:"enabled"`
	} `json:"hook"`
	Token  string `json:"token"`
	Secret string `json:"secret"`
}
type AssetAutomationTool struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	RequiredScope  string   `json:"required_scope"`
	RequiredScopes []string `json:"required_scopes,omitempty"`
	Risk           string   `json:"risk"`
}
type AssetAutomationAuditItem struct {
	ID        string `json:"id"`
	Tool      string `json:"tool"`
	Transport string `json:"transport"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
type AssetAutomationOverview struct {
	MCPEndpoint      string                     `json:"mcp_endpoint"`
	AuthMethod       string                     `json:"auth_method"`
	Tools            []AssetAutomationTool      `json:"tools"`
	RecentExecutions []AssetAutomationAuditItem `json:"recent_executions"`
}

type CreateAssetWorkflowInput struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	TemplateKey string                   `json:"template_key,omitempty"`
	Definition  *AssetWorkflowDefinition `json:"definition,omitempty"`
	Layout      *AssetWorkflowLayout     `json:"layout,omitempty"`
}
type UpdateAssetWorkflowDraftInput struct {
	Name             string                  `json:"name"`
	Description      string                  `json:"description,omitempty"`
	Definition       AssetWorkflowDefinition `json:"definition"`
	Layout           AssetWorkflowLayout     `json:"layout"`
	ExpectedRevision int64                   `json:"expected_revision"`
}
type StartAssetWorkflowInput struct {
	VersionID      string          `json:"version_id,omitempty"`
	Input          json.RawMessage `json:"input,omitempty"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
}
type DecideAssetWorkflowApprovalInput struct {
	Reason    string     `json:"reason,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
