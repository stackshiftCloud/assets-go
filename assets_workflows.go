package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// WorkflowCatalog returns the live typed node registry and editable templates.
func (a *AssetsClient) WorkflowCatalog(ctx context.Context) (*AssetWorkflowCatalog, error) {
	var result AssetWorkflowCatalog
	err := a.client.request(ctx, http.MethodGet, "/assets/workflows/catalog", nil, &result)
	return &result, err
}

// CreateWorkflow creates a blank or template-backed workflow draft.
func (a *AssetsClient) CreateWorkflow(ctx context.Context, input CreateAssetWorkflowInput) (*AssetWorkflow, error) {
	var workflow AssetWorkflow
	err := a.client.request(ctx, http.MethodPost, "/assets/workflows", input, &workflow)
	return &workflow, err
}

// ListWorkflows lists workflow drafts and active workflows.
func (a *AssetsClient) ListWorkflows(ctx context.Context, page, perPage int) ([]AssetWorkflow, error) {
	path := "/assets/workflows?page=" + strconv.Itoa(page) + "&per_page=" + strconv.Itoa(perPage)
	var workflows []AssetWorkflow
	err := a.client.request(ctx, http.MethodGet, path, nil, &workflows)
	return workflows, err
}

// GetWorkflow returns one workflow and its current draft documents.
func (a *AssetsClient) GetWorkflow(ctx context.Context, workflowID string) (*AssetWorkflow, error) {
	var workflow AssetWorkflow
	err := a.client.request(ctx, http.MethodGet, workflowPath(workflowID), nil, &workflow)
	return &workflow, err
}

// UpdateWorkflowDraft replaces semantic and layout drafts with an optimistic revision precondition.
func (a *AssetsClient) UpdateWorkflowDraft(ctx context.Context, workflowID string, input UpdateAssetWorkflowDraftInput) (*AssetWorkflow, error) {
	var workflow AssetWorkflow
	err := a.client.request(ctx, http.MethodPut, workflowPath(workflowID)+"/draft", input, &workflow)
	return &workflow, err
}

// ValidateWorkflow validates the current draft without publishing it.
func (a *AssetsClient) ValidateWorkflow(ctx context.Context, workflowID string) (*AssetWorkflowValidation, error) {
	var validation AssetWorkflowValidation
	err := a.client.request(ctx, http.MethodPost, workflowPath(workflowID)+"/validate", nil, &validation)
	return &validation, err
}

// PublishWorkflow creates an immutable version and any required approval request.
func (a *AssetsClient) PublishWorkflow(ctx context.Context, workflowID string, expectedRevision int64) (*AssetWorkflowVersion, error) {
	var version AssetWorkflowVersion
	err := a.client.request(ctx, http.MethodPost, workflowPath(workflowID)+"/publish",
		map[string]int64{"expected_revision": expectedRevision}, &version)
	return &version, err
}

// ActivateWorkflow activates an approved immutable version.
func (a *AssetsClient) ActivateWorkflow(ctx context.Context, workflowID, versionID string, expectedRevision int64) error {
	return a.client.request(ctx, http.MethodPost, workflowPath(workflowID)+"/activate", map[string]any{
		"version_id": versionID, "expected_revision": expectedRevision,
	}, nil)
}

// ListWorkflowVersions returns immutable versions newest first.
func (a *AssetsClient) ListWorkflowVersions(ctx context.Context, workflowID string) ([]AssetWorkflowVersion, error) {
	var versions []AssetWorkflowVersion
	err := a.client.request(ctx, http.MethodGet, workflowPath(workflowID)+"/versions", nil, &versions)
	return versions, err
}

// StartWorkflowRun starts an active version manually.
func (a *AssetsClient) StartWorkflowRun(ctx context.Context, workflowID string, input StartAssetWorkflowInput) (*AssetWorkflowRun, error) {
	var run AssetWorkflowRun
	headers := map[string]string{}
	if input.IdempotencyKey != "" {
		headers["Idempotency-Key"] = input.IdempotencyKey
	}
	err := a.client.requestWithHeaders(ctx, http.MethodPost, workflowPath(workflowID)+"/run", input, headers, &run)
	return &run, err
}

// ListWorkflowRuns returns persisted runs for one workflow.
func (a *AssetsClient) ListWorkflowRuns(ctx context.Context, workflowID string, page, perPage int) ([]AssetWorkflowRun, error) {
	path := workflowPath(workflowID) + "/runs?page=" + strconv.Itoa(page) + "&per_page=" + strconv.Itoa(perPage)
	var runs []AssetWorkflowRun
	err := a.client.request(ctx, http.MethodGet, path, nil, &runs)
	return runs, err
}

// GetWorkflowRun returns a run plus its persisted node executions.
func (a *AssetsClient) GetWorkflowRun(ctx context.Context, runID string) (*AssetWorkflowRunDetail, error) {
	var detail AssetWorkflowRunDetail
	err := a.client.request(ctx, http.MethodGet, workflowRunPath(runID), nil, &detail)
	return &detail, err
}

// CancelWorkflowRun requests cooperative cancellation.
func (a *AssetsClient) CancelWorkflowRun(ctx context.Context, runID string) (*AssetWorkflowRun, error) {
	return a.mutateWorkflowRun(ctx, runID, "cancel")
}

// RetryWorkflowRun resets only the failed node and its descendants.
func (a *AssetsClient) RetryWorkflowRun(ctx context.Context, runID string) (*AssetWorkflowRun, error) {
	return a.mutateWorkflowRun(ctx, runID, "retry")
}

// ListWorkflowApprovals lists approvals, optionally filtered by status.
func (a *AssetsClient) ListWorkflowApprovals(ctx context.Context, status string) ([]AssetWorkflowApproval, error) {
	path := "/assets/workflows/approvals"
	if status != "" {
		path += "?status=" + url.QueryEscape(status)
	}
	var approvals []AssetWorkflowApproval
	err := a.client.request(ctx, http.MethodGet, path, nil, &approvals)
	return approvals, err
}

// DecideWorkflowApproval approves or rejects a pending workflow approval.
func (a *AssetsClient) DecideWorkflowApproval(ctx context.Context, approvalID string, approved bool, input DecideAssetWorkflowApprovalInput) (*AssetWorkflowApproval, error) {
	decision := "reject"
	if approved {
		decision = "approve"
	}
	var approval AssetWorkflowApproval
	err := a.client.request(ctx, http.MethodPost,
		"/assets/workflows/approvals/"+url.PathEscape(approvalID)+"/"+decision, input, &approval)
	return &approval, err
}

// ListWorkflowHTTPConnections lists reusable restricted external HTTP connections.
func (a *AssetsClient) ListWorkflowHTTPConnections(ctx context.Context) ([]AssetWorkflowHTTPConnection, error) {
	var connections []AssetWorkflowHTTPConnection
	err := a.client.request(ctx, http.MethodGet, "/assets/workflows/connections", nil, &connections)
	return connections, err
}

// CreateWorkflowHTTPConnection creates a host and path-prefix restricted HTTPS connection.
func (a *AssetsClient) CreateWorkflowHTTPConnection(ctx context.Context, input CreateAssetWorkflowHTTPConnectionInput) (*AssetWorkflowHTTPConnection, error) {
	var connection AssetWorkflowHTTPConnection
	err := a.client.request(ctx, http.MethodPost, "/assets/workflows/connections", input, &connection)
	return &connection, err
}

// RotateWorkflowHTTPConnectionCredential atomically replaces a connection credential.
func (a *AssetsClient) RotateWorkflowHTTPConnectionCredential(ctx context.Context, connectionID string, input AssetWorkflowHTTPCredentialInput) (map[string]any, error) {
	var credential map[string]any
	err := a.client.request(ctx, http.MethodPost, "/assets/workflows/connections/"+
		url.PathEscape(connectionID)+"/credential/rotate", input, &credential)
	return credential, err
}

// RotateWorkflowInboundHook creates a new one-time-disclosed token and signing secret.
func (a *AssetsClient) RotateWorkflowInboundHook(ctx context.Context, workflowID string) (*AssetWorkflowHook, error) {
	var hook AssetWorkflowHook
	err := a.client.request(ctx, http.MethodPost, workflowPath(workflowID)+"/inbound-hook/rotate", nil, &hook)
	return &hook, err
}

// AutomationOverview returns the governed MCP tools and recent audited executions.
func (a *AssetsClient) AutomationOverview(ctx context.Context) (*AssetAutomationOverview, error) {
	var overview AssetAutomationOverview
	err := a.client.request(ctx, http.MethodGet, "/assets/automation", nil, &overview)
	return &overview, err
}

func (a *AssetsClient) mutateWorkflowRun(ctx context.Context, runID, operation string) (*AssetWorkflowRun, error) {
	var run AssetWorkflowRun
	err := a.client.request(ctx, http.MethodPost, workflowRunPath(runID)+"/"+operation, nil, &run)
	return &run, err
}

func workflowPath(workflowID string) string { return "/assets/workflows/" + url.PathEscape(workflowID) }
func workflowRunPath(runID string) string   { return "/assets/workflows/runs/" + url.PathEscape(runID) }
