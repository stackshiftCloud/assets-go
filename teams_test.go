package stackshift

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type teamsMock struct {
	mu       sync.Mutex
	calls    []computersCall
	teams    map[string]map[string]any
	keys     map[string]map[string]any
	nextID   int
	messages int
	empty    int
}

func newTeamsMock() *teamsMock {
	return &teamsMock{teams: map[string]map[string]any{}, keys: map[string]map[string]any{}}
}

func (m *teamsMock) id() string {
	m.nextID++
	return fmt.Sprintf("00000000-0000-4000-9000-%012d", m.nextID)
}

func (m *teamsMock) last() computersCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls[len(m.calls)-1]
}

func (m *teamsMock) member(input map[string]any) map[string]any {
	name, _ := input["name"].(string)
	return map[string]any{
		"id": m.id(), "name": name, "agent": input["agent"], "role": input["role"], "computer_id": m.id(), "computer_state": "creating", "status": "starting",
		"hostname": name + ".team.internal", "address": "10.213.0.11", "size": map[string]any{"vcpus": 2, "memory_bytes": 4294967296, "disk_bytes": 21474836480},
		"template": "coding", "joined_at": "2026-11-09T00:00:00Z",
	}
}

func (m *teamsMock) message(seq int) map[string]any {
	return map[string]any{
		"seq": seq, "kind": "message", "from": map[string]any{"kind": "member", "member_id": "m1", "name": "claude-fixer"},
		"to": map[string]any{"kind": "everyone"}, "text": "message " + strconv.Itoa(seq), "created_at": "2026-11-09T00:00:00Z",
	}
}

func (m *teamsMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	raw, _ := io.ReadAll(r.Body)
	call := computersCall{Method: r.Method, Path: strings.TrimPrefix(r.URL.Path, "/api/v1"), Query: map[string]string{}, Headers: r.Header.Clone(), Raw: raw}
	for key, values := range r.URL.Query() {
		call.Query[key] = values[0]
	}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") && len(raw) > 0 {
		_ = json.Unmarshal(raw, &call.Body)
	}
	m.calls = append(m.calls, call)
	if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		writeFailure(w, 401, "unauthorized", "missing bearer token")
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(call.Path, "/computers/teams"), "/"), "/")
	teamID, sub := parts[0], strings.Join(parts[min(1, len(parts)):], "/")
	if !strings.HasPrefix(call.Path, "/computers/teams") {
		writeFailure(w, 404, "invalid_request", "no route")
		return
	}
	if teamID == "" {
		m.collection(w, call)
		return
	}
	team, ok := m.teams[teamID]
	if !ok {
		writeFailure(w, 404, "team_not_found", "team not found")
		return
	}
	switch {
	case sub == "" && r.Method == "GET":
		writeJSON(w, 200, team)
	case sub == "" && r.Method == "PATCH":
		for key, value := range call.Body {
			team[key] = value
		}
		writeJSON(w, 200, team)
	case sub == "" && r.Method == "DELETE":
		if call.Headers.Get("X-StackShift-Confirm-Delete") != teamID {
			writeFailure(w, 428, "confirmation_required", "confirm the delete")
			return
		}
		team["state"] = "deleting"
		writeJSON(w, 200, team)
	case sub == "pause":
		team["state"], team["pause_reason"] = "paused", "person"
		writeJSON(w, 200, team)
	case sub == "resume":
		team["state"] = "working"
		delete(team, "pause_reason")
		writeJSON(w, 200, team)
	case sub == "members" && r.Method == "POST":
		if call.Headers.Get("Idempotency-Key") == "" {
			writeFailure(w, 400, "invalid_request", "Idempotency-Key is required")
			return
		}
		member := m.member(call.Body)
		team["members"] = append(team["members"].([]any), member)
		writeJSON(w, 201, member)
	case strings.HasPrefix(sub, "members/") && r.Method == "DELETE":
		memberID := strings.TrimPrefix(sub, "members/")
		kept := []any{}
		found := false
		for _, item := range team["members"].([]any) {
			if item.(map[string]any)["id"] == memberID {
				found = true
				continue
			}
			kept = append(kept, item)
		}
		if !found {
			writeFailure(w, 404, "member_not_found", "member not found")
			return
		}
		team["members"] = kept
		w.WriteHeader(204)
	case sub == "messages" && r.Method == "POST":
		m.messages++
		message := m.message(m.messages)
		message["text"], message["to"] = call.Body["text"], map[string]any{"kind": "member", "name": call.Body["to"]}
		writeJSON(w, 201, message)
	case sub == "messages":
		after, _ := strconv.Atoi(call.Query["after_seq"])
		items := []any{}
		if call.Query["wait"] != "" && m.empty > 0 {
			m.empty--
			writeJSON(w, 200, map[string]any{"items": items, "next_seq": after, "has_more": false})
			return
		}
		for seq := after + 1; seq <= 3; seq++ {
			items = append(items, m.message(seq))
		}
		next := after
		if len(items) > 0 {
			next = 3
		}
		writeJSON(w, 200, map[string]any{"items": items, "next_seq": next, "has_more": false})
	case sub == "messages/stream":
		after, _ := strconv.Atoi(call.Query["after_seq"])
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, ": ping\n\n")
		for seq := 1; seq <= 3; seq++ {
			if seq > after {
				encoded, _ := json.Marshal(m.message(seq))
				fmt.Fprintf(w, "event: message\ndata: %s\n\n", encoded)
				fmt.Fprint(w, ": ping\n\n")
			}
		}
	case sub == "files":
		writeJSON(w, 200, map[string]any{"path": call.Query["path"], "used_bytes": 12, "capacity_bytes": 10737418240, "entries": []any{map[string]any{"name": "refund-fix.patch", "path": "/shared/refund-fix.patch", "type": "file", "size_bytes": 12, "modified_at": "2026-11-09T00:00:00Z", "changed_by": "claude-fixer"}}})
	case sub == "network":
		writeJSON(w, 200, map[string]any{"domain": "team.internal", "members": []any{map[string]any{"member_id": "m1", "name": "claude-fixer", "hostname": "claude-fixer.team.internal", "address": "10.213.0.11", "ports": []any{map[string]any{"port": 3000, "process": "node"}}}}})
	case sub == "activity":
		writeJSON(w, 200, map[string]any{"members": []any{map[string]any{"member_id": "m1", "name": "claude-fixer", "cwd": "/shared/repo", "command": "claude", "lines": []any{"a", "b", "c"}}}})
	default:
		writeFailure(w, 404, "invalid_request", "no route")
	}
}

func (m *teamsMock) collection(w http.ResponseWriter, call computersCall) {
	if call.Method == "GET" {
		items := []any{}
		for _, team := range m.teams {
			items = append(items, team)
		}
		writeJSON(w, 200, map[string]any{"items": items, "next_cursor": "", "has_more": false})
		return
	}
	key := call.Headers.Get("Idempotency-Key")
	if key == "" {
		writeFailure(w, 400, "invalid_request", "Idempotency-Key is required")
		return
	}
	if seen, ok := m.keys[key]; ok {
		writeJSON(w, 201, seen)
		return
	}
	members := []any{}
	for _, item := range call.Body["members"].([]any) {
		members = append(members, m.member(item.(map[string]any)))
	}
	team := map[string]any{
		"id": m.id(), "project_id": call.Body["project_id"], "name": call.Body["name"], "job": call.Body["job"], "state": "starting", "stop_rule": "quiet", "quiet_after_seconds": 1800,
		"spend_limit": call.Body["spend_limit"], "spent": map[string]any{"amount_minor": 0, "currency": "NGN"}, "shared": map[string]any{"enabled": true, "path": "/shared", "capacity_bytes": 10737418240},
		"network": map[string]any{"enabled": true, "domain": "team.internal"}, "ask_before_acting": false, "repository": nil, "members": members, "running_count": 0,
		"last_message_at": "2026-11-09T00:00:00Z", "created_at": "2026-11-09T00:00:00Z", "updated_at": "2026-11-09T00:00:00Z",
	}
	m.teams[team["id"].(string)] = team
	m.keys[key] = team
	writeJSON(w, 201, team)
}

func teamsFixture(t *testing.T) (*Client, *teamsMock) {
	t.Helper()
	mock := newTeamsMock()
	server := httptest.NewServer(mock)
	t.Cleanup(server.Close)
	client, err := New(Options{APIKey: "csk_fixture", BaseURL: server.URL + "/api/v1"})
	if err != nil {
		t.Fatal(err)
	}
	return client, mock
}

func TestTeamsLifecycleMembersMessagesAndViews(t *testing.T) {
	client, mock := teamsFixture(t)
	ctx := context.Background()

	input := CreateTeamInput{
		ProjectID: computersProject, Name: "checkout-fix", Job: "Refunds fail on partial orders.",
		Repository: &TeamRepositoryInput{Repository: "github.com/northwind/api", Revision: "main"},
		Members: []TeamMemberInput{
			{Name: "claude-fixer", Agent: "claude-code", Role: "writes the fix"},
			{Name: "codex-reviewer", Agent: "codex", Size: &TeamMemberSizeInput{VCPUs: 2, MemoryBytes: 4 << 30}},
		},
		StopRule: TeamStopDone, SpendLimit: &TeamMoney{AmountMinor: 2000000, Currency: "NGN"},
	}
	team, err := client.Teams.Create(ctx, input, "team-key")
	if err != nil || team.State != TeamStateStarting || len(team.Members) != 2 || team.Members[0].Hostname != "claude-fixer.team.internal" || team.Shared.Path != "/shared" || team.SpendLimit == nil || team.SpendLimit.AmountMinor != 2000000 {
		t.Fatalf("create %+v %v", team, err)
	}
	call := mock.last()
	if call.Method != "POST" || call.Path != "/computers/teams" || call.Headers.Get("Idempotency-Key") != "team-key" || call.Headers.Get("Authorization") != "Bearer csk_fixture" || call.Body["stop_rule"] != "done" {
		t.Fatalf("create call %+v", call)
	}
	replay, err := client.Teams.Create(ctx, input, "team-key")
	if err != nil || replay.ID != team.ID {
		t.Fatalf("replay %+v %v", replay, err)
	}
	generated, err := client.Teams.Create(ctx, input, "")
	if err != nil || generated.ID == team.ID || len(mock.last().Headers.Get("Idempotency-Key")) != 36 {
		t.Fatalf("generated key %+v %v", generated, err)
	}

	page, err := client.Teams.List(ctx, computersProject, TeamListOptions{State: "starting", Limit: 10})
	if err != nil || len(page.Items) != 2 {
		t.Fatalf("list %+v %v", page, err)
	}
	if call := mock.last(); call.Query["project_id"] != computersProject || call.Query["state"] != "starting" || call.Query["limit"] != "10" {
		t.Fatalf("list query %+v", call.Query)
	}
	if all, err := client.Teams.ListAll(ctx, computersProject, ""); err != nil || len(all) != 2 {
		t.Fatalf("list all %v %v", all, err)
	}
	got, err := client.Teams.Get(ctx, team.ID)
	if err != nil || got.Name != "checkout-fix" {
		t.Fatalf("get %+v %v", got, err)
	}

	job, stop, ask := "Fix refunds and prove it.", TeamStopNever, true
	if _, err := client.Teams.Update(ctx, team.ID, UpdateTeamInput{Job: &job, StopRule: &stop, AskBeforeActing: &ask, SpendLimit: &TeamMoney{AmountMinor: 5, Currency: "USD"}}); err != nil {
		t.Fatal(err)
	}
	if call := mock.last(); call.Method != "PATCH" || call.Body["job"] != job || call.Body["stop_rule"] != "never" || call.Body["ask_before_acting"] != true || call.Body["spend_limit"].(map[string]any)["currency"] != "USD" {
		t.Fatalf("update call %+v", call)
	}
	if _, err := client.Teams.Update(ctx, team.ID, UpdateTeamInput{ClearSpendLimit: true}); err != nil {
		t.Fatal(err)
	}
	if call := mock.last(); string(call.Raw) != `{"spend_limit":null}` {
		t.Fatalf("clear body %s", call.Raw)
	}

	paused, err := client.Teams.Pause(ctx, team.ID)
	if err != nil || paused.State != TeamStatePaused || paused.PauseReason != "person" {
		t.Fatalf("pause %+v %v", paused, err)
	}
	resumed, err := client.Teams.Resume(ctx, team.ID)
	if err != nil || resumed.State != TeamStateWorking || resumed.PauseReason != "" {
		t.Fatalf("resume %+v %v", resumed, err)
	}

	member, err := client.Teams.AddMember(ctx, team.ID, TeamMemberInput{Name: "gemini-tests", Agent: "gemini", Role: "proves it", ModelSecretID: "secret-1"}, "member-key")
	if err != nil || member.Name != "gemini-tests" || member.ComputerID == "" {
		t.Fatalf("add member %+v %v", member, err)
	}
	if call := mock.last(); call.Path != "/computers/teams/"+team.ID+"/members" || call.Headers.Get("Idempotency-Key") != "member-key" || call.Body["model_secret_id"] != "secret-1" {
		t.Fatalf("add member call %+v", call)
	}
	if err := client.Teams.RemoveMember(ctx, team.ID, member.ID, true); err != nil {
		t.Fatal(err)
	}
	if call := mock.last(); call.Method != "DELETE" || call.Path != "/computers/teams/"+team.ID+"/members/"+member.ID || call.Query["keep_computer"] != "true" {
		t.Fatalf("remove member call %+v", call)
	}
	other := team.Members[1].ID
	if err := client.Teams.RemoveMember(ctx, team.ID, other, false); err != nil {
		t.Fatal(err)
	}
	if call := mock.last(); call.Query["keep_computer"] != "" {
		t.Fatalf("remove member query %+v", call.Query)
	}
	var missing *APIError
	if err := client.Teams.RemoveMember(ctx, team.ID, "nope", false); !errors.As(err, &missing) || missing.Code != ComputerErrorMemberNotFound || missing.StatusCode != 404 {
		t.Fatalf("missing member %v", err)
	}

	sent, err := client.Teams.Send(ctx, team.ID, SendTeamMessageInput{Text: "Patch is ready.", To: "codex-reviewer", File: "/shared/refund-fix.patch", AsMember: "claude-fixer"})
	if err != nil || sent.Seq != 1 || sent.To.Name != "codex-reviewer" {
		t.Fatalf("send %+v %v", sent, err)
	}
	if call := mock.last(); call.Body["file"] != "/shared/refund-fix.patch" || call.Body["as_member"] != "claude-fixer" || call.Body["to"] != "codex-reviewer" {
		t.Fatalf("send call %+v", call)
	}

	messages, err := client.Teams.ListMessages(ctx, team.ID, TeamMessageOptions{AfterSeq: 1, Limit: 50, Wait: 5})
	if err != nil || len(messages.Items) != 2 || messages.NextSeq != 3 || messages.Items[0].From.Name != "claude-fixer" {
		t.Fatalf("messages %+v %v", messages, err)
	}
	if call := mock.last(); call.Query["after_seq"] != "1" || call.Query["limit"] != "50" || call.Query["wait"] != "5" {
		t.Fatalf("messages query %+v", call.Query)
	}

	mock.mu.Lock()
	mock.empty = 1
	mock.mu.Unlock()
	waited, err := client.Teams.WaitForMessages(ctx, team.ID, 0, time.Minute)
	if err != nil || len(waited.Items) != 3 || waited.NextSeq != 3 {
		t.Fatalf("wait %+v %v", waited, err)
	}
	if call := mock.last(); call.Query["wait"] != "25" {
		t.Fatalf("wait query %+v", call.Query)
	}
	mock.mu.Lock()
	mock.empty = 1000
	mock.mu.Unlock()
	short, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	quiet, err := client.Teams.WaitForMessages(short, team.ID, 3, 0)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("wait quiet %v", err)
	}
	if err == nil && (len(quiet.Items) != 0 || quiet.NextSeq != 3) {
		t.Fatalf("wait quiet %+v", quiet)
	}
	timedOut, err := client.Teams.WaitForMessages(ctx, team.ID, 3, 500*time.Millisecond)
	if err != nil || len(timedOut.Items) != 0 || timedOut.NextSeq != 3 {
		t.Fatalf("wait timeout %+v %v", timedOut, err)
	}

	var seen []int64
	if err := client.Teams.StreamMessages(ctx, team.ID, 0, false, func(message TeamMessage) bool { seen = append(seen, message.Seq); return true }); err != nil || fmt.Sprint(seen) != "[1 2 3]" {
		t.Fatalf("stream %v %v", seen, err)
	}
	if call := mock.last(); call.Headers.Get("Accept") != "text/event-stream" {
		t.Fatalf("stream headers %+v", call.Headers)
	}
	seen = nil
	if err := client.Teams.StreamMessages(ctx, team.ID, 2, false, func(message TeamMessage) bool { seen = append(seen, message.Seq); return true }); err != nil || fmt.Sprint(seen) != "[3]" {
		t.Fatalf("resumed stream %v %v", seen, err)
	}
	if call := mock.last(); call.Headers.Get("Last-Event-ID") != "2" || call.Query["after_seq"] != "2" {
		t.Fatalf("resume call %+v", call)
	}
	seen = nil
	if err := client.Teams.StreamMessages(ctx, team.ID, 0, true, func(message TeamMessage) bool { seen = append(seen, message.Seq); return false }); err != nil || len(seen) != 1 {
		t.Fatalf("stopped stream %v %v", seen, err)
	}

	files, err := client.Teams.Files(ctx, team.ID, "/shared")
	if err != nil || files.UsedBytes != 12 || len(files.Entries) != 1 || files.Entries[0].ChangedBy != "claude-fixer" {
		t.Fatalf("files %+v %v", files, err)
	}
	if call := mock.last(); call.Query["path"] != "/shared" {
		t.Fatalf("files query %+v", call.Query)
	}
	network, err := client.Teams.Network(ctx, team.ID)
	if err != nil || network.Domain != "team.internal" || network.Members[0].Ports[0].Port != 3000 {
		t.Fatalf("network %+v %v", network, err)
	}
	activity, err := client.Teams.Activity(ctx, team.ID)
	if err != nil || len(activity.Members) != 1 || len(activity.Members[0].Lines) != 3 {
		t.Fatalf("activity %+v %v", activity, err)
	}

	deleted, err := client.Teams.Delete(ctx, team.ID)
	if err != nil || deleted.State != TeamStateDeleting {
		t.Fatalf("delete %+v %v", deleted, err)
	}
	if call := mock.last(); call.Headers.Get("X-StackShift-Confirm-Delete") != team.ID {
		t.Fatalf("delete headers %+v", call.Headers)
	}

	_, err = client.Teams.Get(ctx, "00000000-0000-4000-9000-999999999999")
	if !errors.As(err, &missing) || missing.StatusCode != 404 || missing.Code != ComputerErrorTeamNotFound {
		t.Fatalf("missing %v", err)
	}
}

func TestTeamsErrorsUseComputersShape(t *testing.T) {
	client, _ := teamsFixture(t)
	_, err := client.Teams.Create(context.Background(), CreateTeamInput{ProjectID: computersProject, Name: "x", Job: "j", Members: []TeamMemberInput{{Name: "a", Agent: "codex"}}}, "k")
	if err != nil {
		t.Fatal(err)
	}
	var apiErr *APIError
	_, err = client.Teams.Pause(context.Background(), "missing")
	if !errors.As(err, &apiErr) || apiErr.Code != "team_not_found" || apiErr.Message != "team not found" {
		t.Fatalf("pause missing %v", err)
	}
}
