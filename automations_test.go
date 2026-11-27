package stackshift

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const automationSummaryJSON = `{"id":"run_1","automation_id":"auto_1","automation_name":"Redeploy","number":4,"test":true,` +
	`"status":"waiting","trigger":{"type":"manual","label":"Test run"},"started_by":{"id":"user_1","name":"Ada"},` +
	`"steps_done":1,"steps_total":2,"stopped_at_step":2,"approval":{"id":"card_1","node_id":"redeploy","step_name":"Redeploy",` +
	`"action":"projects.redeploy","tier":"ask","summary":"Redeploy a project","status":"pending","decided_by":null,` +
	`"expires_at":"2026-12-04T10:00:00Z"},"started_at":"2026-11-27T10:00:00Z","completed_at":null,"duration_ms":1200,` +
	`"created_at":"2026-11-27T10:00:00Z"}`

const automationJSON = `{"id":"auto_1","name":"Redeploy","description":"","team_id":"team_1","created_by":{"id":"user_1","name":"Ada"},` +
	`"enabled":true,"revision":3,"trigger":{"type":"event","event_types":["deploy.failed"],"has_webhook":false},` +
	`"definition":{"version":1,"nodes":[{"id":"trigger","type":"trigger.event","config":{"event_types":["deploy.failed"]}}],"edges":[]},` +
	`"access":{"actions":["projects.redeploy"],"can_widen":true,"can_change":true},"secrets":[],"last_run":null,` +
	`"runs_this_week":2,"created_at":"2026-11-27T10:00:00Z","updated_at":"2026-11-27T10:00:00Z"}`

func TestAutomationsRoutes(t *testing.T) {
	t.Parallel()
	type call struct{ method, path, body string }
	var calls []call
	routes := map[string]struct {
		status int
		data   string
	}{
		"GET /automations?team_id=team_1":           {http.StatusOK, `{"automations":[` + automationJSON + `],"happening":[` + automationSummaryJSON + `]}`},
		"POST /automations":                         {http.StatusCreated, automationJSON},
		"GET /automations/auto_1":                   {http.StatusOK, automationJSON},
		"PUT /automations/auto_1":                   {http.StatusOK, automationJSON},
		"DELETE /automations/auto_1":                {http.StatusNoContent, ""},
		"POST /automations/auto_1/enable":           {http.StatusOK, automationJSON},
		"POST /automations/auto_1/disable":          {http.StatusOK, automationJSON},
		"PUT /automations/auto_1/access":            {http.StatusOK, automationJSON},
		"POST /automations/auto_1/runs":             {http.StatusCreated, automationSummaryJSON},
		"POST /automations/auto_1/rerun-failed":     {http.StatusOK, `{"runs":3}`},
		"PUT /automations/auto_1/secrets/crm":       {http.StatusOK, `{"name":"crm","host":"api.crm.example","header":"Authorization","prefix":"Bearer ","set_by":{"id":"user_1","name":"Ada"},"updated_at":"2026-11-27T10:00:00Z","used_by":["notify"]}`},
		"DELETE /automations/auto_1/secrets/crm":    {http.StatusNoContent, ""},
		"POST /automations/auto_1/webhook":          {http.StatusOK, `{"url":"https://api.stackshift.cloud/api/v1/automations/hooks/tok","secret":"whsec_1"}`},
		"GET /automations/actions":                  {http.StatusOK, `{"actions":[{"key":"projects.redeploy","product":"hosting","summary":"Redeploy a project","tier":"write","available":true}],"steps":[{"type":"control.delay","name":"Wait","category":"Control","input_ports":[],"output_ports":[]}]}`},
		"GET /automations/templates?category=media": {http.StatusOK, `[{"key":"media-publish","name":"Publish media","description":"","category":"media","definition":{"version":1,"nodes":[],"edges":[]}}]`},
		"POST /automations/trigger-preview":         {http.StatusOK, `{"matched_30d":12}`},
		"GET /automations/runs?automation_id=auto_1&cursor=2026-11-27T10%3A00%3A00Z&limit=25&status=failed&team_id=team_1": {http.StatusOK, `{"runs":[` + automationSummaryJSON + `],"next_cursor":null,"counts":{"all":4,"running":0,"waiting":1,"failed":1,"done":2}}`},
		"GET /automations/runs/run_1": {http.StatusOK, strings.TrimSuffix(automationSummaryJSON, "}") +
			`,"steps":[{"node_id":"redeploy","name":"Redeploy","type":"platform.action","action":"projects.redeploy","tier":"ask","status":"waiting","input":null,"output":null,"error":null,"started_at":null,"completed_at":null,"duration_ms":null}],"failure":null}`},
		"POST /automations/runs/run_1/cancel":   {http.StatusOK, automationSummaryJSON},
		"POST /automations/runs/run_1/rerun":    {http.StatusCreated, automationSummaryJSON},
		"POST /automations/runs/run_1/approval": {http.StatusOK, automationSummaryJSON},
	}
	client, err := New(Options{APIKey: "sk_test", BaseURL: "https://api.stackshift.test/api/v1", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var body []byte
		if req.Body != nil {
			body, _ = io.ReadAll(req.Body)
		}
		path := strings.TrimPrefix(req.URL.Path, "/api/v1")
		if req.URL.RawQuery != "" {
			path += "?" + req.URL.RawQuery
		}
		calls = append(calls, call{req.Method, path, string(body)})
		if path == "/automations/auto_bad" {
			return jsonResponse(http.StatusUnprocessableEntity, `{"success":false,"message":"Step is not allowed.","error":{"code":"automation_invalid","details":{"issues":[{"code":"not_allowed","message":"Step is not allowed.","node_id":"redeploy"}]}}}`), nil
		}
		route, ok := routes[req.Method+" "+path]
		if !ok {
			return jsonResponse(http.StatusNotFound, `{"success":false,"message":"Not found"}`), nil
		}
		if route.status == http.StatusNoContent {
			return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}}, nil
		}
		return jsonResponse(route.status, `{"success":true,"data":`+route.data+`}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	automations := client.Automations
	lastBody := func() map[string]any {
		var body map[string]any
		if err := json.Unmarshal([]byte(calls[len(calls)-1].body), &body); err != nil {
			t.Fatalf("body %q: %v", calls[len(calls)-1].body, err)
		}
		return body
	}

	list, err := automations.List(ctx, "team_1")
	if err != nil || list.Automations[0].ID != "auto_1" || list.Happening[0].Approval.Tier != "ask" {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	if _, err := automations.Create(ctx, CreateAutomationInput{TeamID: "team_1", Name: "Redeploy", TemplateKey: "deploy-failed"}); err != nil {
		t.Fatal(err)
	}
	if body := lastBody(); body["template_key"] != "deploy-failed" || body["definition"] != nil {
		t.Fatalf("create body=%v", body)
	}
	got, err := automations.Get(ctx, "auto_1")
	if err != nil || got.Definition.Nodes[0].Type != "trigger.event" || *got.TeamID != "team_1" {
		t.Fatalf("get=%+v err=%v", got, err)
	}

	definition := AutomationDefinition{Version: 1, Nodes: []AutomationNode{
		{ID: "trigger", Type: "trigger.manual", Config: json.RawMessage(`{}`)},
		{ID: "redeploy", Type: "platform.action", Action: "projects.redeploy", Config: json.RawMessage(`{"project_id":"proj_1"}`)},
		{ID: "success", Type: "control.success"},
	}, Edges: []AutomationEdge{{ID: "a", Source: "trigger", Target: "redeploy"}, {ID: "b", Source: "redeploy", SourcePort: "success", Target: "success"}}}
	if _, err := automations.Save(ctx, "auto_1", SaveAutomationInput{Name: "Redeploy", Definition: definition, ExpectedRevision: 2}); err != nil {
		t.Fatal(err)
	}
	saved := calls[len(calls)-1].body
	if strings.Contains(saved, "action_version") || strings.Contains(saved, "retry") || !strings.Contains(saved, `"expected_revision":2`) {
		t.Fatalf("save body=%s", saved)
	}
	if err := automations.Delete(ctx, "auto_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := automations.Enable(ctx, "auto_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := automations.Disable(ctx, "auto_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := automations.SetAccess(ctx, "auto_1", []string{"projects.redeploy"}); err != nil || calls[len(calls)-1].body != `{"actions":["projects.redeploy"]}` {
		t.Fatalf("access body=%s err=%v", calls[len(calls)-1].body, err)
	}
	run, err := automations.StartRun(ctx, "auto_1", StartAutomationRunInput{Input: json.RawMessage(`{"project_id":"proj_1"}`), Test: true})
	if err != nil || !run.Test || *run.StoppedAtStep != 2 || lastBody()["test"] != true {
		t.Fatalf("run=%+v err=%v", run, err)
	}
	since := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	started, err := automations.RerunFailed(ctx, "auto_1", RerunFailedAutomationRunsInput{FromNodeID: "redeploy", Since: since})
	if err != nil || started != 3 || lastBody()["since"] != "2026-11-20T00:00:00Z" {
		t.Fatalf("started=%d err=%v", started, err)
	}
	secret, err := automations.PutSecret(ctx, "auto_1", "crm", AutomationSecretInput{Host: "api.crm.example", Header: "Authorization", Prefix: "Bearer ", Value: "sk_live"})
	if err != nil || secret.UsedBy[0] != "notify" {
		t.Fatalf("secret=%+v err=%v", secret, err)
	}
	if err := automations.DeleteSecret(ctx, "auto_1", "crm"); err != nil {
		t.Fatal(err)
	}
	if hook, err := automations.RotateWebhook(ctx, "auto_1"); err != nil || hook.Secret != "whsec_1" {
		t.Fatalf("hook=%+v err=%v", hook, err)
	}
	if catalog, err := automations.Actions(ctx); err != nil || catalog.Actions[0].Key != "projects.redeploy" || catalog.Steps[0].Type != "control.delay" {
		t.Fatalf("catalog=%+v err=%v", catalog, err)
	}
	if templates, err := automations.Templates(ctx, "media"); err != nil || templates[0].Key != "media-publish" {
		t.Fatalf("templates=%+v err=%v", templates, err)
	}
	if matched, err := automations.PreviewTrigger(ctx, AutomationTriggerPreviewInput{EventTypes: []string{"deploy.failed"}}); err != nil || matched != 12 {
		t.Fatalf("matched=%d err=%v", matched, err)
	}
	page, err := automations.ListRuns(ctx, AutomationRunListOptions{TeamID: "team_1", AutomationID: "auto_1", Status: "failed", Cursor: "2026-11-27T10:00:00Z", Limit: 25})
	if err != nil || page.Counts.Done != 2 || page.NextCursor != nil {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	detail, err := automations.GetRun(ctx, "run_1")
	if err != nil || detail.ID != "run_1" || detail.Steps[0].Tier != "ask" || detail.Failure != nil {
		t.Fatalf("detail=%+v err=%v", detail, err)
	}
	if _, err := automations.CancelRun(ctx, "run_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := automations.Rerun(ctx, "run_1", RerunAutomationRunInput{FromNodeID: "redeploy"}); err != nil || lastBody()["from_node_id"] != "redeploy" {
		t.Fatalf("rerun err=%v", err)
	}
	if _, err := automations.Decide(ctx, "run_1", DecideAutomationRunInput{Decision: "approved", ConfirmName: "Redeploy"}); err != nil || lastBody()["decision"] != "approved" {
		t.Fatalf("decide err=%v", err)
	}

	_, err = automations.Save(ctx, "auto_bad", SaveAutomationInput{Name: "Bad", Definition: definition, ExpectedRevision: 1})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusUnprocessableEntity || apiErr.Code != "automation_invalid" || apiErr.Details["issues"] == nil {
		t.Fatalf("err=%v", err)
	}
	for _, call := range calls {
		if strings.Contains(call.path, "/assets/workflows") {
			t.Fatalf("dead route called: %s", call.path)
		}
	}
}
