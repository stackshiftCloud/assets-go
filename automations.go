package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type AutomationsClient struct {
	client *Client
}

func (a *AutomationsClient) List(ctx context.Context, teamID string) (*AutomationList, error) {
	query := url.Values{}
	setQuery(query, "team_id", teamID)
	var result AutomationList
	err := a.client.request(ctx, http.MethodGet, withAutomationQuery("/automations", query), nil, &result)
	return &result, err
}

func (a *AutomationsClient) Create(ctx context.Context, input CreateAutomationInput) (*Automation, error) {
	var result Automation
	err := a.client.request(ctx, http.MethodPost, "/automations", input, &result)
	return &result, err
}

func (a *AutomationsClient) Get(ctx context.Context, automationID string) (*Automation, error) {
	var result Automation
	err := a.client.request(ctx, http.MethodGet, automationPath(automationID), nil, &result)
	return &result, err
}

func (a *AutomationsClient) Save(ctx context.Context, automationID string, input SaveAutomationInput) (*Automation, error) {
	var result Automation
	err := a.client.request(ctx, http.MethodPut, automationPath(automationID), input, &result)
	return &result, err
}

func (a *AutomationsClient) Delete(ctx context.Context, automationID string) error {
	return a.client.request(ctx, http.MethodDelete, automationPath(automationID), nil, nil)
}

func (a *AutomationsClient) Enable(ctx context.Context, automationID string) (*Automation, error) {
	var result Automation
	err := a.client.request(ctx, http.MethodPost, automationPath(automationID)+"/enable", nil, &result)
	return &result, err
}

func (a *AutomationsClient) Disable(ctx context.Context, automationID string) (*Automation, error) {
	var result Automation
	err := a.client.request(ctx, http.MethodPost, automationPath(automationID)+"/disable", nil, &result)
	return &result, err
}

func (a *AutomationsClient) SetAccess(ctx context.Context, automationID string, actions []string) (*Automation, error) {
	var result Automation
	err := a.client.request(ctx, http.MethodPut, automationPath(automationID)+"/access",
		map[string][]string{"actions": actions}, &result)
	return &result, err
}

func (a *AutomationsClient) StartRun(ctx context.Context, automationID string, input StartAutomationRunInput) (*AutomationRunSummary, error) {
	var result AutomationRunSummary
	err := a.client.request(ctx, http.MethodPost, automationPath(automationID)+"/runs", input, &result)
	return &result, err
}

func (a *AutomationsClient) RerunFailed(ctx context.Context, automationID string, input RerunFailedAutomationRunsInput) (int, error) {
	var result struct {
		Runs int `json:"runs"`
	}
	err := a.client.request(ctx, http.MethodPost, automationPath(automationID)+"/rerun-failed", input, &result)
	return result.Runs, err
}

func (a *AutomationsClient) PutSecret(ctx context.Context, automationID, name string, input AutomationSecretInput) (*AutomationSecret, error) {
	var result AutomationSecret
	err := a.client.request(ctx, http.MethodPut, automationPath(automationID)+"/secrets/"+url.PathEscape(name), input, &result)
	return &result, err
}

func (a *AutomationsClient) DeleteSecret(ctx context.Context, automationID, name string) error {
	return a.client.request(ctx, http.MethodDelete, automationPath(automationID)+"/secrets/"+url.PathEscape(name), nil, nil)
}

func (a *AutomationsClient) RotateWebhook(ctx context.Context, automationID string) (*AutomationWebhook, error) {
	var result AutomationWebhook
	err := a.client.request(ctx, http.MethodPost, automationPath(automationID)+"/webhook", nil, &result)
	return &result, err
}

func (a *AutomationsClient) Actions(ctx context.Context) (*AutomationCatalog, error) {
	var result AutomationCatalog
	err := a.client.request(ctx, http.MethodGet, "/automations/actions", nil, &result)
	return &result, err
}

func (a *AutomationsClient) Templates(ctx context.Context, category string) ([]AutomationTemplate, error) {
	query := url.Values{}
	setQuery(query, "category", category)
	var result []AutomationTemplate
	err := a.client.request(ctx, http.MethodGet, withAutomationQuery("/automations/templates", query), nil, &result)
	return result, err
}

func (a *AutomationsClient) PreviewTrigger(ctx context.Context, input AutomationTriggerPreviewInput) (int, error) {
	var result struct {
		Matched30d int `json:"matched_30d"`
	}
	err := a.client.request(ctx, http.MethodPost, "/automations/trigger-preview", input, &result)
	return result.Matched30d, err
}

func (a *AutomationsClient) ListRuns(ctx context.Context, options AutomationRunListOptions) (*AutomationRunPage, error) {
	query := url.Values{}
	setQuery(query, "team_id", options.TeamID)
	setQuery(query, "automation_id", options.AutomationID)
	setQuery(query, "status", options.Status)
	setQuery(query, "cursor", options.Cursor)
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	var result AutomationRunPage
	err := a.client.request(ctx, http.MethodGet, withAutomationQuery("/automations/runs", query), nil, &result)
	return &result, err
}

func (a *AutomationsClient) GetRun(ctx context.Context, runID string) (*AutomationRun, error) {
	var result AutomationRun
	err := a.client.request(ctx, http.MethodGet, automationRunPath(runID), nil, &result)
	return &result, err
}

func (a *AutomationsClient) CancelRun(ctx context.Context, runID string) (*AutomationRunSummary, error) {
	var result AutomationRunSummary
	err := a.client.request(ctx, http.MethodPost, automationRunPath(runID)+"/cancel", nil, &result)
	return &result, err
}

func (a *AutomationsClient) Rerun(ctx context.Context, runID string, input RerunAutomationRunInput) (*AutomationRunSummary, error) {
	var result AutomationRunSummary
	err := a.client.request(ctx, http.MethodPost, automationRunPath(runID)+"/rerun", input, &result)
	return &result, err
}

func (a *AutomationsClient) Decide(ctx context.Context, runID string, input DecideAutomationRunInput) (*AutomationRunSummary, error) {
	var result AutomationRunSummary
	err := a.client.request(ctx, http.MethodPost, automationRunPath(runID)+"/approval", input, &result)
	return &result, err
}

func automationPath(automationID string) string {
	return "/automations/" + url.PathEscape(automationID)
}
func automationRunPath(runID string) string { return "/automations/runs/" + url.PathEscape(runID) }

func withAutomationQuery(path string, query url.Values) string {
	if encoded := query.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}
