package stackshift

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type agentCall struct {
	method string
	uri    string
	key    string
	auth   string
	body   map[string]any
}

type agentReply struct {
	status int
	body   string
}

func agentConnectionClient(t *testing.T, reply *agentReply) (*Client, *[]agentCall) {
	t.Helper()
	var calls []agentCall
	client, err := New(Options{APIKey: "ssc_fixture_connection_credential", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		call := agentCall{method: req.Method, uri: strings.TrimPrefix(req.URL.RequestURI(), "/api/v1"), key: req.Header.Get("Idempotency-Key"), auth: req.Header.Get("Authorization")}
		if req.Body != nil {
			if raw, _ := io.ReadAll(req.Body); len(raw) > 0 {
				_ = json.Unmarshal(raw, &call.body)
			}
		}
		calls = append(calls, call)
		return jsonResponse(reply.status, reply.body), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	return client, &calls
}

func agentOperationJSON(id, state string) string {
	return `{"id":"` + id + `","connection_id":"conn_1","action":"deployment.restart","reason":null,"resource_id":"p1","created_at":"x","state":"` + state + `","execution_mode":"external","configuration":{},"resource_version":"v1","affected_targets":[],"approval":null,"result":null,"controller_operations":[],"error":null,"estimate":{}}`
}

func TestAgentConnectionIdentityResourcesAndCapabilities(t *testing.T) {
	reply := &agentReply{200, `{"data":{"owner_id":"u1","connection_id":"conn_1","client":{"id":"claude","name":"Claude Code","identity_verified":false},"resources":[],"operations":["deployment.restart"],"creation":null,"creation_allowances":[],"expires_at":"x","before_each_change":{"level":"ask_every_time","label":"Ask me every time","max_level":"safe_changes","exceptions":0}}}`}
	client, calls := agentConnectionClient(t, reply)
	identity, err := client.Agents.Connections.Identity(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if identity.ConnectionID != "conn_1" || identity.BeforeEachChange == nil || identity.BeforeEachChange.Level != AgentLevelAskEveryTime {
		t.Fatalf("identity lost fields: %+v", identity)
	}
	if call := (*calls)[0]; call.method != "GET" || call.uri != "/agents/connections/identity" || call.auth != "Bearer ssc_fixture_connection_credential" {
		t.Fatalf("unexpected call %+v", call)
	}

	reply.body = `{"data":{"items":[{"kind":"project","id":"p1","name":"shop","application_id":null,"can_include_children":false}],"next_cursor":"1","total":2}}`
	page, err := client.Agents.Connections.Resources(context.Background(), AgentConnectionResourceOptions{Kind: "project", Query: "sh", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if page.Items[0].Name != "shop" || page.NextCursor == nil || *page.NextCursor != "1" || page.Total != 2 {
		t.Fatalf("resources lost fields: %+v", page)
	}
	if uri := (*calls)[1].uri; uri != "/agents/connections/resources?kind=project&limit=1&query=sh" {
		t.Fatalf("resources uri %s", uri)
	}

	reply.body = `{"data":[{"action":"deployment.restart","execution_mode":"external","approval":"ask","supported_environments":["production"],"configuration_schema":{}}]}`
	capabilities, err := client.Agents.Connections.Capabilities(context.Background(), "deployments")
	if err != nil {
		t.Fatal(err)
	}
	if len(capabilities) != 1 || capabilities[0].Action != "deployment.restart" || (*calls)[2].uri != "/agents/connections/capabilities?family=deployments" {
		t.Fatalf("capabilities wrong: %+v %s", capabilities, (*calls)[2].uri)
	}
}

func TestAgentConnectionOperations(t *testing.T) {
	reply := &agentReply{202, `{"data":` + strings.Replace(agentOperationJSON("op_1", "waiting_approval"), `"approval":null`, `"approval":{"id":"a1","action_hash":"h","state":"pending","expires_at":"x","review_url":"https://app/review"}`, 1) + `}`}
	client, calls := agentConnectionClient(t, reply)
	ctx := context.Background()
	operations := client.Agents.Connections.Operations

	if _, err := operations.Propose(ctx, AgentConnectionOperationInput{Action: "deployment.restart"}); err == nil || len(*calls) != 0 {
		t.Fatalf("a proposal without an idempotency key was sent: %v", err)
	}
	proposed, err := operations.Propose(ctx, AgentConnectionOperationInput{Action: "deployment.restart", ResourceID: "p1", ExpectedResourceVersion: "v1", Reason: "Restart after the config fix", IdempotencyKey: "retain-1"})
	if err != nil {
		t.Fatal(err)
	}
	if proposed.State != "waiting_approval" || proposed.Approval == nil || proposed.Approval.ReviewURL != "https://app/review" {
		t.Fatalf("proposal lost fields: %+v", proposed)
	}
	call := (*calls)[0]
	if call.method != "POST" || call.uri != "/agents/connections/operations" || call.key != "retain-1" {
		t.Fatalf("unexpected call %+v", call)
	}
	if call.body["action"] != "deployment.restart" || call.body["resource_id"] != "p1" || call.body["expected_resource_version"] != "v1" || call.body["reason"] != "Restart after the config fix" {
		t.Fatalf("body %+v", call.body)
	}
	if configuration, ok := call.body["configuration"].(map[string]any); !ok || len(configuration) != 0 {
		t.Fatalf("configuration should be an empty object: %+v", call.body["configuration"])
	}
	if _, leaked := call.body["IdempotencyKey"]; leaked {
		t.Fatal("idempotency key leaked into the body")
	}

	reply.status, reply.body = 200, `{"data":{"current":[`+agentOperationJSON("op_1", "running")+`],"history":{"items":[],"next_cursor":"n1"},"source_status":{"current":"available","history":"available"},"errors":[]}}`
	list, err := operations.List(ctx, AgentConnectionOperationListOptions{Limit: 5, Cursor: "c1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Current) != 1 || list.History.NextCursor == nil || *list.History.NextCursor != "n1" || list.SourceStatus.History != "available" {
		t.Fatalf("list lost fields: %+v", list)
	}
	if uri := (*calls)[1].uri; uri != "/agents/connections/operations?cursor=c1&limit=5" {
		t.Fatalf("list uri %s", uri)
	}

	reply.body = `{"data":` + agentOperationJSON("op 1", "running") + `}`
	got, err := operations.Get(ctx, "op 1")
	if err != nil || got.State != "running" || (*calls)[2].uri != "/agents/connections/operations/op%201" {
		t.Fatalf("get: %v %+v %s", err, got, (*calls)[2].uri)
	}

	reply.body = `{"data":{"settled":false,"operation":` + agentOperationJSON("op_1", "waiting_approval") + `}}`
	waited, err := operations.Wait(ctx, "op_1", 20)
	if err != nil || waited.Settled || waited.Operation == nil || waited.Operation.State != "waiting_approval" || (*calls)[3].uri != "/agents/connections/operations/op_1/wait?timeout_seconds=20" {
		t.Fatalf("wait: %v %+v %s", err, waited, (*calls)[3].uri)
	}
	if _, err := operations.Wait(ctx, "op_1", 0); err != nil || (*calls)[4].uri != "/agents/connections/operations/op_1/wait" {
		t.Fatalf("wait without timeout: %v %s", err, (*calls)[4].uri)
	}

	reply.body = `{"data":{"state":"pending","review_url":"https://app/review","expires_at":"x","needs_owner_signed_in":true,"decided_at":null}}`
	approval, err := operations.Approval(ctx, "op_1")
	if err != nil || approval.State != "pending" || !approval.NeedsOwnerSignedIn || approval.ReviewURL == nil || approval.DecidedAt != nil || (*calls)[5].uri != "/agents/connections/operations/op_1/approval" {
		t.Fatalf("approval: %v %+v %s", err, approval, (*calls)[5].uri)
	}
}

func TestAgentConnectionPlanTaskFactsAndHistory(t *testing.T) {
	reply := &agentReply{200, `{"data":{"task":{"title":"Fix checkout","steps":[{"name":"Read logs","state":"done"}],"reported_at":"x"},"plans":[{"objective_id":"obj_1","intent":"Fix","status":"running","steps":[{"id":"s1","title":"Read","status":"done"}]}]}}`}
	client, calls := agentConnectionClient(t, reply)
	ctx := context.Background()
	connections := client.Agents.Connections

	plan, err := connections.Plan(ctx)
	if err != nil || plan.Task == nil || plan.Task.Title != "Fix checkout" || plan.Plans[0].Steps[0].Title != "Read" || (*calls)[0].uri != "/agents/connections/plan" {
		t.Fatalf("plan: %v %+v", err, plan)
	}

	reply.body = `{"data":{"facts":[{"key":"region","value":"lagos","written_at":"x","influenced":false}]}}`
	facts, err := connections.TaskFacts(ctx, "obj_1")
	if err != nil || len(facts.Facts) != 1 || facts.Facts[0].Key != "region" || string(facts.Facts[0].Value) != `"lagos"` || (*calls)[1].uri != "/agents/connections/task-facts?objective_id=obj_1" {
		t.Fatalf("task facts: %v %+v %s", err, facts, (*calls)[1].uri)
	}

	reply.body = `{"data":{"items":[{"id":"h1","at":"x","action_key":"deployment.restart","view":{"what":"Restarted shop","where":"shop","why":"Fix","allowed_by":"You","result":"Done","tone":"ok"}}],"next_cursor":"next"}}`
	history, err := connections.History(ctx, AgentConnectionHistoryOptions{Limit: 10})
	if err != nil || history.Items[0].View.What != "Restarted shop" || history.NextCursor == nil || *history.NextCursor != "next" || (*calls)[2].uri != "/agents/connections/history?limit=10" {
		t.Fatalf("history: %v %+v %s", err, history, (*calls)[2].uri)
	}
}

func TestAgentConnectionLowerLevelAndRefusedRaise(t *testing.T) {
	reply := &agentReply{200, `{"success":true,"data":{"level":"ask_every_time","exceptions":[],"version":4,"max_level":"safe_changes","trust":{"tier":"new"},"limits":{"infra_monthly_kobo":0,"changes_per_minute":10,"changes_per_hour":120,"changes_per_day":300,"failures_in_a_row":5},"safe_actions":[]}}`}
	client, calls := agentConnectionClient(t, reply)
	ctx := context.Background()
	version := int64(3)
	lowered, err := client.Agents.Connections.LowerLevel(ctx, AgentConnectionLevelLowerInput{Level: AgentLevelAskEveryTime, ExpectedVersion: &version})
	if err != nil {
		t.Fatal(err)
	}
	if lowered.Version != 4 || lowered.Level != AgentLevelAskEveryTime || lowered.Limits.ChangesPerHour != 120 {
		t.Fatalf("level lost fields: %+v", lowered)
	}
	call := (*calls)[0]
	if call.method != "PUT" || call.uri != "/agents/connections/level" || call.body["level"] != "ask_every_time" || call.body["expected_version"] != float64(3) {
		t.Fatalf("unexpected call %+v", call)
	}

	reply.status, reply.body = 403, `{"success":false,"status_code":403,"message":"Only the owner can raise this.","error":{"code":"session_required"}}`
	_, err = client.Agents.Connections.LowerLevel(ctx, AgentConnectionLevelLowerInput{Level: AgentLevelWholeTasks})
	var apiError *APIError
	if !errors.As(err, &apiError) || apiError.StatusCode != 403 || apiError.Code != "session_required" || apiError.Message != "Only the owner can raise this." {
		t.Fatalf("a refused raise should be an APIError with its code: %v", err)
	}
}
