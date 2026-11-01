package stackshift

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type AgentLevel string

const (
	AgentLevelAskEveryTime AgentLevel = "ask_every_time"
	AgentLevelSafeChanges  AgentLevel = "safe_changes"
	AgentLevelWholeTasks   AgentLevel = "whole_tasks"
)

type AgentConnectionClientInfo struct {
	ID               string `json:"id"`
	CatalogID        string `json:"catalog_id"`
	Name             string `json:"name"`
	IdentityVerified bool   `json:"identity_verified"`
}

type AgentConnectionGrantedResource struct {
	Kind            string `json:"kind"`
	ID              string `json:"id"`
	IncludeChildren bool   `json:"include_children"`
	ApplicationID   string `json:"application_id,omitempty"`
	All             bool   `json:"all,omitempty"`
	FromKind        bool   `json:"from_kind,omitempty"`
	Name            string `json:"name,omitempty"`
}

type AgentConnectionCreation struct {
	DeploymentMode         string  `json:"deployment_mode,omitempty"`
	TargetNodeID           string  `json:"target_node_id,omitempty"`
	Kind                   string  `json:"kind"`
	IncludeCreatedChildren bool    `json:"include_created_children"`
	BoundApplicationID     *string `json:"bound_application_id"`
}

type AgentConnectionCreationAllowance struct {
	TargetNodeID     string  `json:"target_node_id,omitempty"`
	SourceDatabaseID string  `json:"source_database_id,omitempty"`
	Reserved         bool    `json:"reserved"`
	ResourceID       *string `json:"resource_id"`
	Kind             string  `json:"kind"`
	Maximum          int     `json:"maximum"`
	DeploymentMode   string  `json:"deployment_mode,omitempty"`
	Region           string  `json:"region,omitempty"`
	PlanID           string  `json:"plan_id,omitempty"`
}

type AgentConnectionBeforeEachChange struct {
	Level      AgentLevel `json:"level"`
	Label      string     `json:"label"`
	MaxLevel   AgentLevel `json:"max_level"`
	Exceptions int        `json:"exceptions"`
}

type AgentConnectionIdentity struct {
	OwnerID            string                             `json:"owner_id"`
	ConnectionID       string                             `json:"connection_id"`
	Client             AgentConnectionClientInfo          `json:"client"`
	Resources          []AgentConnectionGrantedResource   `json:"resources"`
	Operations         []string                           `json:"operations"`
	Creation           *AgentConnectionCreation           `json:"creation"`
	CreationAllowances []AgentConnectionCreationAllowance `json:"creation_allowances"`
	ExpiresAt          string                             `json:"expires_at"`
	BeforeEachChange   *AgentConnectionBeforeEachChange   `json:"before_each_change,omitempty"`
}

type AgentConnectionResourceDetail struct {
	Type    *string `json:"type"`
	Version *string `json:"version"`
	Plan    *string `json:"plan"`
	Region  *string `json:"region"`
	Host    *string `json:"host,omitempty"`
}

type AgentConnectionResource struct {
	Kind               string                         `json:"kind"`
	ID                 string                         `json:"id"`
	Name               string                         `json:"name"`
	ApplicationID      *string                        `json:"application_id"`
	CanIncludeChildren bool                           `json:"can_include_children"`
	Detail             *AgentConnectionResourceDetail `json:"detail,omitempty"`
}

type AgentConnectionResourcePage struct {
	Items      []AgentConnectionResource `json:"items"`
	NextCursor *string                   `json:"next_cursor"`
	Total      int                       `json:"total"`
}

type AgentConnectionResourceOptions struct {
	Kind   string
	Query  string
	Limit  int
	Cursor string
}

type AgentConnectionCapability struct {
	Action                string         `json:"action"`
	ExecutionMode         string         `json:"execution_mode"`
	Approval              string         `json:"approval"`
	SupportedEnvironments []string       `json:"supported_environments"`
	Funding               string         `json:"funding,omitempty"`
	ConfigurationSchema   map[string]any `json:"configuration_schema"`
	NativeAISchema        map[string]any `json:"native_ai_schema,omitempty"`
}

type AgentConnectionTarget struct {
	Kind            string  `json:"kind"`
	ID              string  `json:"id"`
	Name            *string `json:"name"`
	ApplicationID   *string `json:"application_id"`
	Role            string  `json:"role"`
	ResourceVersion *string `json:"resource_version"`
}

type AgentConnectionOperationApproval struct {
	ID              string         `json:"id"`
	ActionHash      string         `json:"action_hash"`
	State           string         `json:"state"`
	ExpiresAt       string         `json:"expires_at"`
	RiskExplanation string         `json:"risk_explanation,omitempty"`
	Safety          map[string]any `json:"safety,omitempty"`
	DecidableVia    []string       `json:"decidable_via,omitempty"`
	ReviewURL       string         `json:"review_url,omitempty"`
}

type AgentConnectionControllerRef struct {
	Kind     string `json:"kind"`
	ID       string `json:"id"`
	Status   string `json:"status"`
	Resource struct {
		Kind string `json:"kind"`
		ID   string `json:"id"`
	} `json:"resource"`
}

type AgentConnectionOperationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type AgentConnectionOperation struct {
	ID                   string                            `json:"id"`
	ConnectionID         string                            `json:"connection_id"`
	Action               string                            `json:"action"`
	Reason               *string                           `json:"reason"`
	ResourceID           *string                           `json:"resource_id"`
	CreatedAt            string                            `json:"created_at"`
	State                string                            `json:"state"`
	DispatchState        string                            `json:"dispatch_state,omitempty"`
	ControllerStatus     string                            `json:"controller_status,omitempty"`
	ExecutionMode        string                            `json:"execution_mode"`
	Configuration        json.RawMessage                   `json:"configuration"`
	ResourceVersion      string                            `json:"resource_version"`
	AffectedTargets      []AgentConnectionTarget           `json:"affected_targets"`
	Approval             *AgentConnectionOperationApproval `json:"approval"`
	Result               json.RawMessage                   `json:"result"`
	ControllerOperations []AgentConnectionControllerRef    `json:"controller_operations"`
	Error                *AgentConnectionOperationError    `json:"error"`
	RecordedCost         map[string]any                    `json:"recorded_cost,omitempty"`
	Estimate             map[string]any                    `json:"estimate"`
}

type AgentConnectionOperationInput struct {
	Action                  string         `json:"action"`
	ResourceID              string         `json:"resource_id,omitempty"`
	Configuration           map[string]any `json:"configuration"`
	ExpectedResourceVersion string         `json:"expected_resource_version,omitempty"`
	Reason                  string         `json:"reason,omitempty"`
	IdempotencyKey          string         `json:"-"`
}

type AgentConnectionOperationListOptions struct {
	Limit  int
	Cursor string
}

type AgentConnectionOperationHistory struct {
	Items      []AgentConnectionOperation `json:"items"`
	NextCursor *string                    `json:"next_cursor"`
}

type AgentConnectionSourceStatus struct {
	Current string `json:"current"`
	History string `json:"history"`
}

type AgentConnectionSourceError struct {
	Source string `json:"source"`
	Code   string `json:"code"`
}

type AgentConnectionOperationPage struct {
	Current      []AgentConnectionOperation      `json:"current"`
	History      AgentConnectionOperationHistory `json:"history"`
	SourceStatus AgentConnectionSourceStatus     `json:"source_status"`
	Errors       []AgentConnectionSourceError    `json:"errors"`
}

type AgentConnectionOperationWait struct {
	Settled   bool                      `json:"settled"`
	Operation *AgentConnectionOperation `json:"operation"`
}

type AgentConnectionApprovalStatus struct {
	State              string  `json:"state"`
	ReviewURL          *string `json:"review_url"`
	ExpiresAt          *string `json:"expires_at"`
	NeedsOwnerSignedIn bool    `json:"needs_owner_signed_in"`
	DecidedAt          *string `json:"decided_at"`
}

type AgentConnectionTaskStep struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

type AgentConnectionTask struct {
	Title      string                    `json:"title"`
	Steps      []AgentConnectionTaskStep `json:"steps"`
	ReportedAt string                    `json:"reported_at"`
}

type AgentConnectionPlanStep struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type AgentConnectionPlan struct {
	ObjectiveID string                    `json:"objective_id"`
	Intent      string                    `json:"intent"`
	Status      string                    `json:"status"`
	Steps       []AgentConnectionPlanStep `json:"steps"`
}

type AgentConnectionPlans struct {
	Task  *AgentConnectionTask  `json:"task"`
	Plans []AgentConnectionPlan `json:"plans"`
}

type AgentConnectionTaskFact struct {
	Key        string          `json:"key"`
	Value      json.RawMessage `json:"value"`
	WrittenAt  string          `json:"written_at"`
	Influenced bool            `json:"influenced"`
}

type AgentConnectionTaskFacts struct {
	Facts []AgentConnectionTaskFact `json:"facts"`
}

type AgentConnectionHistoryView struct {
	What      string `json:"what"`
	Where     string `json:"where"`
	Why       string `json:"why"`
	AllowedBy string `json:"allowed_by"`
	Result    string `json:"result"`
	Tone      string `json:"tone"`
}

type AgentConnectionHistoryItem struct {
	ID        string                     `json:"id"`
	At        string                     `json:"at"`
	ActionKey string                     `json:"action_key"`
	View      AgentConnectionHistoryView `json:"view"`
}

type AgentConnectionHistoryPage struct {
	Items      []AgentConnectionHistoryItem `json:"items"`
	NextCursor *string                      `json:"next_cursor"`
}

type AgentConnectionHistoryOptions struct {
	Limit  int
	Cursor string
}

type AgentLevelException struct {
	Kind  string     `json:"kind"`
	ID    string     `json:"id"`
	Level AgentLevel `json:"level"`
	Name  string     `json:"name,omitempty"`
	Note  string     `json:"note,omitempty"`
}

type AgentLevelTrust struct {
	Tier   string `json:"tier"`
	Reason string `json:"reason,omitempty"`
}

type AgentLevelLimits struct {
	AITaskCents      *int  `json:"ai_task_cents,omitempty"`
	AIDayCents       *int  `json:"ai_day_cents,omitempty"`
	OnCallAIDayCents *int  `json:"on_call_ai_day_cents,omitempty"`
	InfraMonthlyKobo int64 `json:"infra_monthly_kobo"`
	ChangesPerMinute int   `json:"changes_per_minute"`
	ChangesPerHour   int   `json:"changes_per_hour"`
	ChangesPerDay    int   `json:"changes_per_day"`
	FailuresInARow   int   `json:"failures_in_a_row"`
}

type AgentLevelView struct {
	Level       AgentLevel            `json:"level"`
	Exceptions  []AgentLevelException `json:"exceptions"`
	Version     int64                 `json:"version"`
	UpdatedAt   string                `json:"updated_at,omitempty"`
	MaxLevel    AgentLevel            `json:"max_level"`
	Trust       AgentLevelTrust       `json:"trust"`
	Limits      AgentLevelLimits      `json:"limits"`
	SafeActions []string              `json:"safe_actions"`
}

type AgentConnectionLevelLowerInput struct {
	Level           AgentLevel `json:"level"`
	ExpectedVersion *int64     `json:"expected_version,omitempty"`
}

type AgentsClient struct {
	Connections *AgentConnectionsClient
}

type AgentConnectionsClient struct {
	client     *Client
	Operations *AgentConnectionOperationsClient
}

type AgentConnectionOperationsClient struct{ client *Client }

func decodeConnectionData(raw json.RawMessage, out any) error {
	if len(raw) == 0 {
		return nil
	}
	var wrapped struct {
		Data json.RawMessage `json:"data"`
	}
	if json.Unmarshal(raw, &wrapped) == nil && len(wrapped.Data) > 0 {
		return json.Unmarshal(wrapped.Data, out)
	}
	return json.Unmarshal(raw, out)
}

func (a *AgentConnectionsClient) Identity(ctx context.Context) (*AgentConnectionIdentity, error) {
	var raw json.RawMessage
	if err := a.client.request(ctx, http.MethodGet, "/agents/connections/identity", nil, &raw); err != nil {
		return nil, err
	}
	var out AgentConnectionIdentity
	return &out, decodeConnectionData(raw, &out)
}

func (a *AgentConnectionsClient) Resources(ctx context.Context, options AgentConnectionResourceOptions) (*AgentConnectionResourcePage, error) {
	query := url.Values{}
	setQuery(query, "kind", options.Kind)
	setQuery(query, "query", options.Query)
	setQuery(query, "cursor", options.Cursor)
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	var raw json.RawMessage
	if err := a.client.request(ctx, http.MethodGet, "/agents/connections/resources"+queryString(query), nil, &raw); err != nil {
		return nil, err
	}
	var out AgentConnectionResourcePage
	return &out, decodeConnectionData(raw, &out)
}

func (a *AgentConnectionsClient) Capabilities(ctx context.Context, family string) ([]AgentConnectionCapability, error) {
	query := url.Values{}
	setQuery(query, "family", family)
	var raw json.RawMessage
	if err := a.client.request(ctx, http.MethodGet, "/agents/connections/capabilities"+queryString(query), nil, &raw); err != nil {
		return nil, err
	}
	var out []AgentConnectionCapability
	return out, decodeConnectionData(raw, &out)
}

func (a *AgentConnectionsClient) Plan(ctx context.Context) (*AgentConnectionPlans, error) {
	var raw json.RawMessage
	if err := a.client.request(ctx, http.MethodGet, "/agents/connections/plan", nil, &raw); err != nil {
		return nil, err
	}
	var out AgentConnectionPlans
	return &out, decodeConnectionData(raw, &out)
}

func (a *AgentConnectionsClient) TaskFacts(ctx context.Context, objectiveID string) (*AgentConnectionTaskFacts, error) {
	query := url.Values{}
	query.Set("objective_id", objectiveID)
	var raw json.RawMessage
	if err := a.client.request(ctx, http.MethodGet, "/agents/connections/task-facts"+queryString(query), nil, &raw); err != nil {
		return nil, err
	}
	var out AgentConnectionTaskFacts
	return &out, decodeConnectionData(raw, &out)
}

func (a *AgentConnectionsClient) History(ctx context.Context, options AgentConnectionHistoryOptions) (*AgentConnectionHistoryPage, error) {
	query := url.Values{}
	setQuery(query, "cursor", options.Cursor)
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	var raw json.RawMessage
	if err := a.client.request(ctx, http.MethodGet, "/agents/connections/history"+queryString(query), nil, &raw); err != nil {
		return nil, err
	}
	var out AgentConnectionHistoryPage
	return &out, decodeConnectionData(raw, &out)
}

func (a *AgentConnectionsClient) LowerLevel(ctx context.Context, input AgentConnectionLevelLowerInput) (*AgentLevelView, error) {
	var raw json.RawMessage
	if err := a.client.request(ctx, http.MethodPut, "/agents/connections/level", input, &raw); err != nil {
		return nil, err
	}
	var out AgentLevelView
	return &out, decodeConnectionData(raw, &out)
}

func (o *AgentConnectionOperationsClient) List(ctx context.Context, options AgentConnectionOperationListOptions) (*AgentConnectionOperationPage, error) {
	query := url.Values{}
	setQuery(query, "cursor", options.Cursor)
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	var raw json.RawMessage
	if err := o.client.request(ctx, http.MethodGet, "/agents/connections/operations"+queryString(query), nil, &raw); err != nil {
		return nil, err
	}
	var out AgentConnectionOperationPage
	return &out, decodeConnectionData(raw, &out)
}

func (o *AgentConnectionOperationsClient) Propose(ctx context.Context, input AgentConnectionOperationInput) (*AgentConnectionOperation, error) {
	if input.IdempotencyKey == "" {
		return nil, errors.New("stackshift: proposing an operation needs an IdempotencyKey; keep it and send the same one when you retry")
	}
	if input.Configuration == nil {
		input.Configuration = map[string]any{}
	}
	var raw json.RawMessage
	if err := o.client.requestWithHeaders(ctx, http.MethodPost, "/agents/connections/operations", input, idempotencyHeader(input.IdempotencyKey), &raw); err != nil {
		return nil, err
	}
	var out AgentConnectionOperation
	return &out, decodeConnectionData(raw, &out)
}

func (o *AgentConnectionOperationsClient) Get(ctx context.Context, operationID string) (*AgentConnectionOperation, error) {
	var raw json.RawMessage
	if err := o.client.request(ctx, http.MethodGet, "/agents/connections/operations/"+url.PathEscape(operationID), nil, &raw); err != nil {
		return nil, err
	}
	var out AgentConnectionOperation
	return &out, decodeConnectionData(raw, &out)
}

func (o *AgentConnectionOperationsClient) Wait(ctx context.Context, operationID string, timeoutSeconds int) (*AgentConnectionOperationWait, error) {
	query := url.Values{}
	if timeoutSeconds > 0 {
		query.Set("timeout_seconds", strconv.Itoa(timeoutSeconds))
	}
	var raw json.RawMessage
	if err := o.client.request(ctx, http.MethodGet, "/agents/connections/operations/"+url.PathEscape(operationID)+"/wait"+queryString(query), nil, &raw); err != nil {
		return nil, err
	}
	var out AgentConnectionOperationWait
	return &out, decodeConnectionData(raw, &out)
}

func (o *AgentConnectionOperationsClient) Approval(ctx context.Context, operationID string) (*AgentConnectionApprovalStatus, error) {
	var raw json.RawMessage
	if err := o.client.request(ctx, http.MethodGet, "/agents/connections/operations/"+url.PathEscape(operationID)+"/approval", nil, &raw); err != nil {
		return nil, err
	}
	var out AgentConnectionApprovalStatus
	return &out, decodeConnectionData(raw, &out)
}
