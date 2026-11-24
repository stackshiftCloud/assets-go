package stackshift

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

type controlCall struct {
	Method  string
	Path    string
	Query   map[string]string
	Headers http.Header
	Raw     []byte
}

func (c controlCall) body(t *testing.T) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(c.Raw, &out); err != nil {
		t.Fatalf("body %q: %v", c.Raw, err)
	}
	return out
}

type controlServer struct {
	mu      sync.Mutex
	calls   []controlCall
	handler func(w http.ResponseWriter, r *http.Request, call controlCall)
}

func newControlServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request, call controlCall)) (*Client, *controlServer) {
	t.Helper()
	fake := &controlServer{handler: handler}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		call := controlCall{Method: r.Method, Path: strings.TrimPrefix(r.URL.Path, "/api/v1"), Query: map[string]string{}, Headers: r.Header.Clone(), Raw: raw}
		for key, values := range r.URL.Query() {
			call.Query[key] = values[0]
		}
		fake.mu.Lock()
		fake.calls = append(fake.calls, call)
		fake.mu.Unlock()
		fake.handler(w, r, call)
	}))
	t.Cleanup(server.Close)
	client, err := New(Options{APIKey: "csk_fixture", BaseURL: server.URL + "/api/v1"})
	if err != nil {
		t.Fatal(err)
	}
	return client, fake
}

func (s *controlServer) last() controlCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls[len(s.calls)-1]
}

func (s *controlServer) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

func sse(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "text/event-stream")
	_, _ = io.WriteString(w, body)
}

const commandRecordJSON = `{"id":"m1","shell":"sleep 5","cwd":"/workspace","background":true,"state":"exited","exit_code":0,"signal":"","started_at":"2026-11-10T00:00:00Z","ended_at":"2026-11-10T00:00:05Z","duration_ms":5000,"stdout_bytes":3,"stderr_bytes":0,"stdout":"hi\n"}`

func TestComputerExecStdinStreamAndCommandHistory(t *testing.T) {
	client, fake := newControlServer(t, func(w http.ResponseWriter, r *http.Request, call controlCall) {
		switch {
		case r.Method == "POST" && strings.HasSuffix(call.Path, "/exec") && strings.Contains(r.Header.Get("Accept"), "text/event-stream"):
			if call.body(t)["command"] == nil {
				writeFailure(w, 422, "invalid_request", "bad exec")
				return
			}
			sse(w, "event: start\ndata: {\"id\":\"p9\"}\n\n: ping\n\nevent: stdout\ndata: {\"data\":\"hel\"}\n\nevent: stderr\ndata: {\"data\":\"warn\"}\n\nevent: stdout\ndata: {\"data\":\"lo\"}\n\nevent: error\ndata: {\"code\":\"timeout\",\"message\":\"too slow\"}\n\nevent: exit\ndata: {\"process_id\":\"p9\",\"exit_code\":3,\"timed_out\":false,\"signal\":\"TERM\",\"duration_ms\":42}\n\n")
		case r.Method == "POST" && strings.HasSuffix(call.Path, "/exec"):
			writeJSON(w, 200, map[string]any{"process_id": "p1", "exit_code": 0, "stdout": "ok", "stderr": "", "truncated": true, "timed_out": false, "signal": "KILL", "duration_ms": 77})
		case strings.HasSuffix(call.Path, "/stdin"):
			w.WriteHeader(204)
		case call.Path == "/computers/c1/history":
			writeJSON(w, 200, map[string]any{"items": []json.RawMessage{json.RawMessage(commandRecordJSON)}, "next_cursor": "c2"})
		case strings.HasPrefix(call.Path, "/computers/c1/history/"):
			var record map[string]any
			_ = json.Unmarshal([]byte(commandRecordJSON), &record)
			record["id"] = strings.TrimPrefix(call.Path, "/computers/c1/history/")
			writeJSON(w, 200, record)
		default:
			writeFailure(w, 404, "invalid_request", "no route")
		}
	})
	ctx := context.Background()

	result, err := client.Computers.Exec(ctx, "c1", ComputerExecInput{Command: []string{"cat"}, Stdin: "hello", StdinOpen: true})
	if err != nil || !result.Truncated || result.Signal != "KILL" || result.DurationMS != 77 {
		t.Fatalf("exec %+v %v", result, err)
	}
	if body := fake.last().body(t); body["stdin"] != "hello" || body["stdin_open"] != true {
		t.Fatalf("exec body %v", body)
	}

	if err := client.Computers.SendStdin(ctx, "c1", "m1", []byte{0, 1, 'x'}, false); err != nil {
		t.Fatal(err)
	}
	call := fake.last()
	if call.Method != "POST" || call.Path != "/computers/c1/commands/m1/stdin" || string(call.Raw) != "\x00\x01x" || call.Headers.Get("Content-Type") != "application/octet-stream" {
		t.Fatalf("stdin call %+v", call)
	}
	if _, present := call.Query["close"]; present {
		t.Fatalf("close sent without being asked %v", call.Query)
	}
	if err := client.Computers.SendStdin(ctx, "c1", "m1", nil, true); err != nil || fake.last().Query["close"] != "true" || len(fake.last().Raw) != 0 {
		t.Fatalf("stdin close %+v %v", fake.last(), err)
	}

	var events []ComputerExecEvent
	err = client.Computers.ExecStream(ctx, "c1", ComputerExecInput{Command: []string{"make"}}, func(event ComputerExecEvent) bool {
		events = append(events, event)
		return true
	})
	if err != nil || fake.last().Headers.Get("Accept") != "text/event-stream" || len(events) != 6 {
		t.Fatalf("stream %+v %v", events, err)
	}
	var out string
	for _, event := range events {
		if event.Event == "stdout" {
			out += event.Data
		}
	}
	exit := events[5].Exit
	if events[0].Event != "start" || events[0].ProcessID != "p9" || out != "hello" || events[2].Data != "warn" || events[4].Error == nil || events[4].Error.Code != "timeout" || exit == nil || *exit.ExitCode != 3 || exit.Signal != "TERM" || exit.DurationMS != 42 {
		t.Fatalf("stream events %+v", events)
	}
	seen := 0
	if err := client.Computers.ExecStream(ctx, "c1", ComputerExecInput{Command: []string{"make"}}, func(ComputerExecEvent) bool { seen++; return false }); err != nil || seen != 1 {
		t.Fatalf("early stop %d %v", seen, err)
	}
	var apiErr *APIError
	if err := client.Computers.ExecStream(ctx, "c1", ComputerExecInput{Shell: "x"}, func(ComputerExecEvent) bool { return true }); !errors.As(err, &apiErr) || apiErr.StatusCode != 422 || apiErr.Code != "invalid_request" {
		t.Fatalf("stream failure %v", err)
	}

	page, err := client.Computers.History(ctx, "c1", "c1", 5)
	if err != nil || fake.last().Query["cursor"] != "c1" || fake.last().Query["limit"] != "5" || page.NextCursor != "c2" || len(page.Items) != 1 || page.Items[0].State != "exited" || *page.Items[0].ExitCode != 0 || page.Items[0].DurationMS != 5000 || page.Items[0].EndedAt == nil {
		t.Fatalf("history %+v %v", page, err)
	}
	if _, err := client.Computers.History(ctx, "c1", "", 0); err != nil || len(fake.last().Query) != 0 {
		t.Fatalf("plain history %v", fake.last().Query)
	}
	last, err := client.Computers.Command(ctx, "c1", "")
	if err != nil || fake.last().Path != "/computers/c1/history/last" || last.ID != "last" || last.Stdout != "hi\n" {
		t.Fatalf("last command %+v %v", last, err)
	}
	if named, err := client.Computers.Command(ctx, "c1", "m7"); err != nil || named.ID != "m7" {
		t.Fatalf("named command %+v %v", named, err)
	}
}

func TestComputerJobsLogsFollowWaitKill(t *testing.T) {
	client, fake := newControlServer(t, func(w http.ResponseWriter, r *http.Request, call controlCall) {
		switch {
		case call.Path == "/computers/c1/jobs":
			writeJSON(w, 200, map[string]any{"items": []json.RawMessage{json.RawMessage(commandRecordJSON)}})
		case call.Path == "/computers/c1/jobs/j1/logs" && call.Query["follow"] == "1":
			sse(w, "id: 4\nevent: stdout\ndata: {\"data\":\"one\"}\n\n: ping\n\nid: 8\nevent: stderr\ndata: {\"data\":\"two\"}\n\nid: 9\nevent: exit\ndata: "+commandRecordJSON+"\n\n")
		case call.Path == "/computers/c1/jobs/j1/logs":
			writeJSON(w, 200, map[string]any{"chunks": []map[string]any{{"stream": "stdout", "data": "one"}, {"stream": "stderr", "data": "two"}}, "next_offset": 9, "done": true, "job": json.RawMessage(commandRecordJSON)})
		case call.Path == "/computers/c1/jobs/j1/wait":
			if call.Query["timeout_seconds"] == "999" {
				writeFailure(w, 409, "computer_invalid_state", "computer is not running")
				return
			}
			writeJSON(w, 200, json.RawMessage(commandRecordJSON))
		case call.Path == "/computers/c1/jobs/j1/kill":
			writeJSON(w, 200, json.RawMessage(commandRecordJSON))
		default:
			writeFailure(w, 404, "invalid_request", "no route")
		}
	})
	ctx := context.Background()

	jobs, err := client.Computers.Jobs(ctx, "c1")
	if err != nil || len(jobs.Items) != 1 || !jobs.Items[0].Background {
		t.Fatalf("jobs %+v %v", jobs, err)
	}
	logs, err := client.Computers.JobLogs(ctx, "c1", "j1", 4)
	if err != nil || fake.last().Query["offset"] != "4" || len(logs.Chunks) != 2 || logs.Chunks[1].Stream != "stderr" || logs.NextOffset != 9 || !logs.Done || logs.Job.ID != "m1" {
		t.Fatalf("job logs %+v %v", logs, err)
	}
	if _, err := client.Computers.JobLogs(ctx, "c1", "j1", 0); err != nil || len(fake.last().Query) != 0 {
		t.Fatalf("job logs without offset %v", fake.last().Query)
	}

	var followed []ComputerJobLogEvent
	err = client.Computers.FollowJobLogs(ctx, "c1", "j1", 3, func(event ComputerJobLogEvent) bool {
		followed = append(followed, event)
		return true
	})
	call := fake.last()
	if err != nil || call.Query["follow"] != "1" || call.Query["offset"] != "3" || call.Headers.Get("Last-Event-ID") != "3" || call.Headers.Get("Accept") != "text/event-stream" {
		t.Fatalf("follow request %+v %v", call, err)
	}
	if len(followed) != 3 || followed[0].Event != "stdout" || followed[0].Data != "one" || followed[0].Offset != 4 || followed[1].Event != "stderr" || followed[1].Offset != 8 || followed[2].Event != "exit" || followed[2].Exit == nil || followed[2].Exit.State != "exited" || followed[2].Offset != 9 {
		t.Fatalf("followed %+v", followed)
	}
	if err := client.Computers.FollowJobLogs(ctx, "c1", "j1", 0, func(ComputerJobLogEvent) bool { return true }); err != nil || fake.last().Headers.Get("Last-Event-ID") != "" {
		t.Fatalf("follow from start %v %v", err, fake.last().Headers)
	}

	if record, err := client.Computers.WaitJob(ctx, "c1", "j1", 30); err != nil || fake.last().Query["timeout_seconds"] != "30" || fake.last().Method != "POST" || record.State != "exited" {
		t.Fatalf("wait %+v %v", record, err)
	}
	if _, err := client.Computers.WaitJob(ctx, "c1", "j1", 0); err != nil || len(fake.last().Query) != 0 {
		t.Fatalf("wait without timeout %v", fake.last().Query)
	}
	var apiErr *APIError
	if _, err := client.Computers.WaitJob(ctx, "c1", "j1", 999); !errors.As(err, &apiErr) || apiErr.Code != "computer_invalid_state" {
		t.Fatalf("wait failure %v", err)
	}
	if _, err := client.Computers.KillJob(ctx, "c1", "j1", ""); err != nil || string(fake.last().Raw) != "{}" {
		t.Fatalf("kill default %q %v", fake.last().Raw, err)
	}
	if _, err := client.Computers.KillJob(ctx, "c1", "j1", "KILL"); err != nil || fake.last().body(t)["signal"] != "KILL" {
		t.Fatalf("kill signal %q %v", fake.last().Raw, err)
	}
}

func TestComputerLeasesLimitsMetricsSSH(t *testing.T) {
	client, fake := newControlServer(t, func(w http.ResponseWriter, r *http.Request, call controlCall) {
		now := "2026-11-10T00:00:00Z"
		switch {
		case call.Path == "/computers/c1/activity-leases" && r.Method == "GET":
			writeJSON(w, 200, map[string]any{"items": []map[string]any{{"computer_id": "c1", "task": "build", "expires_at": now, "created_at": now, "updated_at": now}}})
		case strings.HasPrefix(call.Path, "/computers/c1/activity-leases/") && r.Method == "PUT":
			writeJSON(w, 200, map[string]any{"computer_id": "c1", "task": strings.TrimPrefix(call.Path, "/computers/c1/activity-leases/"), "expires_at": now, "created_at": now, "updated_at": now, "woke": true})
		case strings.HasPrefix(call.Path, "/computers/c1/activity-leases/") && r.Method == "DELETE":
			w.WriteHeader(204)
		case call.Path == "/computers/c1/limits":
			writeJSON(w, 200, map[string]any{"computer_id": "c1", "max_run_seconds": 3600, "archive_at": now, "running_since": now})
		case call.Path == "/computers/c1/metrics":
			writeJSON(w, 200, map[string]any{"computer_id": "c1", "interval_seconds": 15, "oom_kills": 1, "vmm_oom_kills": 0, "samples": []map[string]any{{"time": now, "cpu_percent": 12.5, "memory_used_bytes": 1, "memory_total_bytes": 2, "disk_used_bytes": 3, "disk_total_bytes": 4, "net_rx_bytes": 5, "net_tx_bytes": 6, "oom_kills": 1, "vmm_oom_kills": 0}}})
		case call.Path == "/computers/c1/ssh-tokens" && r.Method == "POST":
			writeJSON(w, 201, map[string]any{"id": "t1", "token": "sst_x", "ssh_command": "ssh sst_x@ssh.stackshift.cloud -p 2222", "expires_at": now})
		case call.Path == "/computers/c1/ssh-tokens" && r.Method == "DELETE":
			w.WriteHeader(204)
		case call.Path == "/computers/ssh/certificates":
			writeJSON(w, 200, map[string]any{"certificate": "ssh-ed25519-cert-v01@openssh.com AAAA", "login": "c1", "host": "ssh.stackshift.cloud", "ssh_command": "ssh c1@ssh.stackshift.cloud -p 2222", "expires_at": now})
		default:
			writeFailure(w, 404, "invalid_request", "no route")
		}
	})
	ctx := context.Background()

	leases, err := client.Computers.ActivityLeases(ctx, "c1")
	if err != nil || len(leases.Items) != 1 || leases.Items[0].Task != "build" {
		t.Fatalf("leases %+v %v", leases, err)
	}
	lease, err := client.Computers.PutActivityLease(ctx, "c1", "ci build", 120)
	if err != nil || fake.last().Path != "/computers/c1/activity-leases/ci%20build" && fake.last().Path != "/computers/c1/activity-leases/ci build" || fake.last().body(t)["ttl_seconds"] != float64(120) || !lease.Woke || lease.Task != "ci build" {
		t.Fatalf("put lease %+v %v %q", lease, err, fake.last().Path)
	}
	if err := client.Computers.ReleaseActivityLease(ctx, "c1", "build"); err != nil || fake.last().Method != "DELETE" || fake.last().Path != "/computers/c1/activity-leases/build" {
		t.Fatalf("release %v", err)
	}
	limits, err := client.Computers.Limits(ctx, "c1")
	if err != nil || limits.MaxRunSeconds != 3600 || limits.ArchiveAt == nil || limits.DeleteAt != nil || limits.RunningSince == nil {
		t.Fatalf("limits %+v %v", limits, err)
	}
	metrics, err := client.Computers.Metrics(ctx, "c1", "2026-11-09T00:00:00Z")
	if err != nil || fake.last().Query["since"] != "2026-11-09T00:00:00Z" || metrics.IntervalSeconds != 15 || metrics.OOMKills != 1 || len(metrics.Samples) != 1 || metrics.Samples[0].CPUPercent != 12.5 || metrics.Samples[0].NetTxBytes != 6 {
		t.Fatalf("metrics %+v %v", metrics, err)
	}
	if _, err := client.Computers.Metrics(ctx, "c1", ""); err != nil || len(fake.last().Query) != 0 {
		t.Fatalf("metrics without since %v", fake.last().Query)
	}
	if token, err := client.Computers.CreateSSHToken(ctx, "c1", 0); err != nil || string(fake.last().Raw) != "{}" || token.Token != "sst_x" || !strings.Contains(token.SSHCommand, "2222") {
		t.Fatalf("token default %q %+v %v", fake.last().Raw, token, err)
	}
	if _, err := client.Computers.CreateSSHToken(ctx, "c1", 600); err != nil || fake.last().body(t)["ttl_seconds"] != float64(600) {
		t.Fatalf("token ttl %q %v", fake.last().Raw, err)
	}
	if err := client.Computers.RevokeSSHTokens(ctx, "c1"); err != nil || fake.last().Method != "DELETE" {
		t.Fatalf("revoke tokens %v", err)
	}
	certificate, err := client.Computers.SSHCertificate(ctx, "c1", "ssh-ed25519 AAAA")
	if err != nil || fake.last().Method != "POST" || fake.last().body(t)["computer_id"] != "c1" || fake.last().body(t)["public_key"] != "ssh-ed25519 AAAA" || certificate.Login != "c1" || certificate.Host != "ssh.stackshift.cloud" {
		t.Fatalf("certificate %+v %v", certificate, err)
	}
}

func TestComputerAgentPromptEventsFollowSteerInterrupt(t *testing.T) {
	conversation := func(status string) map[string]any {
		return map[string]any{"id": "v1", "provider": "claude", "cwd": "/workspace", "status": status, "turns": 1, "created_at": "2026-11-10T00:00:00Z", "updated_at": "2026-11-10T00:00:00Z"}
	}
	client, fake := newControlServer(t, func(w http.ResponseWriter, r *http.Request, call controlCall) {
		switch {
		case call.Path == "/computers/c1/agent/prompt":
			writeJSON(w, 202, conversation("running"))
		case call.Path == "/computers/c1/agent/conversations" && r.Method == "GET":
			writeJSON(w, 200, map[string]any{"conversations": []map[string]any{conversation("idle")}})
		case call.Path == "/computers/c1/agent/conversations/v1/events" && call.Query["follow"] == "1":
			sse(w, "id: 3\ndata: {\"seq\":3,\"turn\":1,\"time\":\"2026-11-10T00:00:00Z\",\"kind\":\"turn.started\",\"prompt\":\"hi\"}\n\nid: 4\ndata: {\"seq\":4,\"turn\":1,\"time\":\"2026-11-10T00:00:01Z\",\"kind\":\"output\",\"text\":\"hello\"}\n\nevent: end\ndata: {}\n\nid: 5\ndata: {\"seq\":5,\"turn\":1,\"time\":\"2026-11-10T00:00:02Z\",\"kind\":\"output\",\"text\":\"never delivered\"}\n\n")
		case call.Path == "/computers/c1/agent/conversations/v1/events":
			writeJSON(w, 200, map[string]any{"conversation_id": "v1", "events": []map[string]any{{"seq": 3, "turn": 1, "time": "2026-11-10T00:00:00Z", "kind": "turn.ended", "status": "idle", "exit_code": 0}}, "next_seq": 3, "status": "idle", "exit_code": 0})
		case call.Path == "/computers/c1/agent/conversations/v1/steer", call.Path == "/computers/c1/agent/conversations/v1/interrupt":
			writeJSON(w, 200, conversation("interrupted"))
		default:
			writeFailure(w, 404, "invalid_request", "no route")
		}
	})
	ctx := context.Background()

	started, err := client.Computers.PromptAgent(ctx, "c1", ComputerAgentPrompt{Prompt: "fix the build", Provider: "claude", CWD: "/workspace/app", Env: map[string]string{"A": "1"}, SecretBindingIDs: []string{"s1"}})
	body := fake.last().body(t)
	if err != nil || started.ID != "v1" || started.Status != "running" || body["prompt"] != "fix the build" || body["cwd"] != "/workspace/app" || body["provider"] != "claude" || body["secret_binding_ids"] == nil {
		t.Fatalf("prompt %+v %v %v", started, err, body)
	}
	if _, present := body["conversation_id"]; present {
		t.Fatalf("empty conversation id sent %v", body)
	}
	conversations, err := client.Computers.AgentConversations(ctx, "c1")
	if err != nil || len(conversations) != 1 || conversations[0].Status != "idle" || conversations[0].Turns != 1 {
		t.Fatalf("conversations %+v %v", conversations, err)
	}
	page, err := client.Computers.AgentEvents(ctx, "c1", "v1", 2, 10)
	if err != nil || fake.last().Query["after_seq"] != "2" || fake.last().Query["limit"] != "10" || page.NextSeq != 3 || page.Status != "idle" || len(page.Events) != 1 || page.Events[0].Kind != "turn.ended" || *page.Events[0].ExitCode != 0 {
		t.Fatalf("events %+v %v", page, err)
	}
	var followed []ComputerAgentEvent
	err = client.Computers.FollowAgentEvents(ctx, "c1", "v1", 2, 50, func(event ComputerAgentEvent) bool {
		followed = append(followed, event)
		return true
	})
	call := fake.last()
	if err != nil || call.Query["follow"] != "1" || call.Query["after_seq"] != "2" || call.Query["limit"] != "50" || call.Headers.Get("Last-Event-ID") != "2" {
		t.Fatalf("follow request %+v %v", call, err)
	}
	if len(followed) != 2 || followed[0].Kind != "turn.started" || followed[0].Prompt != "hi" || followed[1].Text != "hello" || followed[1].Seq != 4 {
		t.Fatalf("followed %+v (events after `end` must not be delivered)", followed)
	}
	if steered, err := client.Computers.SteerAgent(ctx, "c1", "v1", "try the other fix"); err != nil || fake.last().body(t)["text"] != "try the other fix" || steered.ID != "v1" {
		t.Fatalf("steer %+v %v", steered, err)
	}
	if stopped, err := client.Computers.InterruptAgent(ctx, "c1", "v1"); err != nil || fake.last().Method != "POST" || fake.last().Path != "/computers/c1/agent/conversations/v1/interrupt" || stopped.Status != "interrupted" {
		t.Fatalf("interrupt %+v %v", stopped, err)
	}
}

func TestComputerFileFoldersTransfersClaimsAndKeyScopes(t *testing.T) {
	transfer := map[string]any{"id": "t1", "direction": "upload", "path": "a.bin", "state": "uploading", "size_bytes": 10}
	plan := map[string]any{"transfer": transfer, "upload_id": "u1", "part_size_bytes": 67108864, "expires_at": "2026-11-10T00:00:00Z", "parts": []map[string]any{{"number": 1, "offset_bytes": 0, "size_bytes": 10, "done": false, "url": "https://objects.example/up/1", "method": "PUT", "headers": map[string][]string{"Content-MD5": {"abc"}}}}}
	claim := map[string]any{"user_code": "ABCD-1234", "name": "agent", "status": "pending", "requested_scopes": []string{"computers:use"}, "approved_scopes": []string{}, "requester_ip": "1.2.3.4", "requester_user_agent": "curl", "created_at": "2026-11-10T00:00:00Z", "expires_at": "2026-11-10T00:10:00Z"}
	client, fake := newControlServer(t, func(w http.ResponseWriter, r *http.Request, call controlCall) {
		switch {
		case r.Method == "DELETE" && call.Path == "/computers/c1/files":
			w.WriteHeader(204)
		case call.Path == "/computers/c1/files/mkdir":
			w.WriteHeader(204)
		case call.Path == "/computers/c1/files/uploads":
			writeJSON(w, 201, plan)
		case call.Path == "/computers/c1/files/uploads/t1/resume":
			writeJSON(w, 200, plan)
		case call.Path == "/computers/c1/files/uploads/t1/complete":
			writeJSON(w, 202, map[string]any{"transfer": map[string]any{"id": "t1", "direction": "upload", "path": "a.bin", "state": "queued"}})
		case call.Path == "/computers/c1/files/downloads":
			writeJSON(w, 202, map[string]any{"transfer": map[string]any{"id": "t2", "direction": "download", "path": "a.bin", "state": "queued"}})
		case call.Path == "/computers/c1/files/transfers/t2":
			writeJSON(w, 200, map[string]any{"transfer": map[string]any{"id": "t2", "direction": "download", "path": "a.bin", "state": "ready", "chunk_bytes": 1024, "chunk_sha256": []string{"aa"}, "checksum_sha256": "bb"}, "download": map[string]any{"url": "https://objects.example/down/t2", "method": "GET", "headers": map[string][]string{}, "expires_at": "2026-11-10T00:05:00Z"}})
		case call.Path == "/auth/agent/claims/ABCD-1234" && r.Method == "GET":
			writeJSON(w, 200, claim)
		case call.Path == "/auth/agent/claims/ABCD-1234/decision":
			approved := map[string]any{}
			for k, v := range claim {
				approved[k] = v
			}
			approved["status"] = "approved"
			approved["approved_scopes"] = []string{"computers:read"}
			approved["project_id"] = call.body(t)["project_id"]
			writeJSON(w, 200, approved)
		case call.Path == "/computers/keys" && r.Method == "POST":
			writeJSON(w, 201, map[string]any{"id": "k1", "project_id": "p1", "name": "agent", "prefix": "csk_", "scopes": call.body(t)["scopes"], "created_at": "2026-11-10T00:00:00Z", "secret": "csk_secret"})
		default:
			writeFailure(w, 404, "invalid_request", "no route")
		}
	})
	ctx := context.Background()

	if err := client.Computers.Remove(ctx, "c1", "build", true); err != nil || fake.last().Query["path"] != "build" || fake.last().Query["recursive"] != "true" {
		t.Fatalf("remove %+v %v", fake.last(), err)
	}
	if err := client.Computers.Remove(ctx, "c1", "a.txt", false); err != nil || fake.last().Query["recursive"] != "" {
		t.Fatalf("remove file %+v %v", fake.last(), err)
	}
	if err := client.Computers.Mkdir(ctx, "c1", "a/b", true); err != nil || fake.last().body(t)["path"] != "a/b" || fake.last().body(t)["parents"] != true {
		t.Fatalf("mkdir %+v %v", fake.last(), err)
	}
	if err := client.Computers.Mkdir(ctx, "c1", "a", false); err != nil {
		t.Fatal(err)
	} else if _, present := fake.last().body(t)["parents"]; present {
		t.Fatalf("parents sent as false %v", fake.last().Raw)
	}

	mode := 0o644
	started, err := client.Computers.StartUpload(ctx, "c1", ComputerUploadInput{Path: "a.bin", SizeBytes: 10, SHA256: "ff", Mode: &mode, PartMD5: []string{"abc"}}, "up-1")
	body := fake.last().body(t)
	if err != nil || fake.last().Headers.Get("Idempotency-Key") != "up-1" || body["size_bytes"] != float64(10) || body["mode"] != float64(420) || started.UploadID != "u1" || len(started.Parts) != 1 || started.Parts[0].Headers["Content-MD5"][0] != "abc" || started.Parts[0].Method != "PUT" {
		t.Fatalf("start upload %+v %v %v", started, err, body)
	}
	resumed, err := client.Computers.ResumeUpload(ctx, "c1", "t1", ComputerResumeUploadInput{UploadID: "u1", PartMD5: []string{"abc"}})
	if err != nil || fake.last().body(t)["upload_id"] != "u1" || resumed.Transfer.ID != "t1" {
		t.Fatalf("resume upload %+v %v", resumed, err)
	}
	completed, err := client.Computers.CompleteUpload(ctx, "c1", "t1", ComputerCompleteUploadInput{UploadID: "u1", Parts: []ComputerCompletedPart{{Number: 1, ETag: "\"abc\""}}})
	parts, _ := fake.last().body(t)["parts"].([]any)
	if err != nil || completed.Transfer.State != "queued" || len(parts) != 1 || parts[0].(map[string]any)["etag"] != "\"abc\"" {
		t.Fatalf("complete upload %+v %v", completed, err)
	}
	download, err := client.Computers.StartDownload(ctx, "c1", "a.bin", "")
	if err != nil || fake.last().Headers.Get("Idempotency-Key") == "" || fake.last().body(t)["path"] != "a.bin" || download.Transfer.ID != "t2" {
		t.Fatalf("start download %+v %v", download, err)
	}
	ready, err := client.Computers.Transfer(ctx, "c1", "t2")
	if err != nil || ready.Transfer.State != "ready" || ready.Transfer.ChunkBytes != 1024 || ready.Transfer.ChunkSHA256[0] != "aa" || ready.Download == nil || ready.Download.URL != "https://objects.example/down/t2" || ready.Download.Method != "GET" {
		t.Fatalf("transfer %+v %v", ready, err)
	}

	pending, err := client.Computers.AgentClaim(ctx, "ABCD-1234")
	if err != nil || pending.Status != "pending" || pending.RequestedScopes[0] != ComputerScopeUse || pending.RequesterIP != "1.2.3.4" {
		t.Fatalf("claim %+v %v", pending, err)
	}
	decided, err := client.Computers.DecideAgentClaim(ctx, "ABCD-1234", ComputerAgentClaimDecision{Approve: true, Scopes: []ComputerKeyScope{ComputerScopeRead}, ProjectID: "p1"})
	body = fake.last().body(t)
	if err != nil || body["approve"] != true || body["project_id"] != "p1" || decided.Status != "approved" || decided.ApprovedScopes[0] != ComputerScopeRead || decided.ProjectID != "p1" {
		t.Fatalf("decide %+v %v %v", decided, err, body)
	}

	key, err := client.Computers.Keys.Create(ctx, CreateComputerKeyInput{ProjectID: "p1", Name: "agent", Scopes: []ComputerKeyScope{ComputerScopeRead, ComputerScopeUse}})
	scopes, _ := fake.last().body(t)["scopes"].([]any)
	if err != nil || len(scopes) != 2 || scopes[0] != "computers:read" || scopes[1] != "computers:use" || len(key.Scopes) != 2 || key.Scopes[1] != ComputerScopeUse || key.Secret != "csk_secret" {
		t.Fatalf("key scopes %+v %v", key, err)
	}
}

func TestComputerAgentRegistrationAndKeyPolling(t *testing.T) {
	var answers []string
	var mu sync.Mutex
	client, fake := newControlServer(t, func(w http.ResponseWriter, r *http.Request, call controlCall) {
		switch call.Path {
		case "/auth/agent":
			writeJSON(w, 200, map[string]any{"claim_token": "claim-1", "user_code": "ABCD-1234", "verification_uri": "https://app.stackshift.cloud/agent", "verification_uri_complete": "https://app.stackshift.cloud/agent?code=ABCD-1234", "expires_in": 600, "interval": 2, "scopes": []string{"computers:use"}})
		case "/auth/agent/token":
			mu.Lock()
			answer := "authorization_pending"
			if len(answers) > 0 {
				answer, answers = answers[0], answers[1:]
			}
			mu.Unlock()
			switch answer {
			case "ok":
				writeJSON(w, 200, map[string]any{"key": "csk_live", "key_id": "k1", "prefix": "csk_", "name": "agent", "project_id": "p1", "scopes": []string{"computers:use"}})
			case "slow_down":
				writeFailure(w, 400, "slow_down", "poll less often")
			case "access_denied":
				writeFailure(w, 400, "access_denied", "the owner said no")
			default:
				writeFailure(w, 400, "authorization_pending", "waiting for the owner")
			}
		default:
			writeFailure(w, 404, "invalid_request", "no route")
		}
	})
	base := client.baseURL
	ctx := context.Background()

	registration, err := RegisterComputerAgent(ctx, base, nil, ComputerAgentRegistrationInput{Name: "agent", Scopes: []ComputerKeyScope{ComputerScopeUse}})
	call := fake.last()
	if err != nil || call.Headers.Get("Authorization") != "" || call.body(t)["name"] != "agent" || registration.ClaimToken != "claim-1" || registration.UserCode != "ABCD-1234" || registration.Interval != 2 || registration.ExpiresIn != 600 || registration.VerificationURIComplete == "" {
		t.Fatalf("register %+v %v", registration, err)
	}
	mu.Lock()
	answers = []string{"ok"}
	mu.Unlock()
	issued, err := ExchangeComputerAgentClaim(ctx, base, nil, "claim-1")
	if err != nil || fake.last().Headers.Get("Authorization") != "" || fake.last().body(t)["claim_token"] != "claim-1" || issued.Key != "csk_live" || issued.Scopes[0] != ComputerScopeUse {
		t.Fatalf("exchange %+v %v", issued, err)
	}
	var apiErr *APIError
	mu.Lock()
	answers = []string{"access_denied"}
	mu.Unlock()
	if _, err := ExchangeComputerAgentClaim(ctx, base, nil, "claim-1"); !errors.As(err, &apiErr) || apiErr.Code != "access_denied" {
		t.Fatalf("exchange failure %v", err)
	}

	var waits []time.Duration
	original := computerAgentWait
	computerAgentWait = func(ctx context.Context, d time.Duration) error {
		waits = append(waits, d)
		return ctx.Err()
	}
	t.Cleanup(func() { computerAgentWait = original })

	mu.Lock()
	answers = []string{"authorization_pending", "slow_down", "authorization_pending", "ok"}
	mu.Unlock()
	before := fake.count()
	key, err := WaitForComputerAgentKey(ctx, base, nil, registration)
	if err != nil || key.Key != "csk_live" {
		t.Fatalf("wait %+v %v", key, err)
	}
	if fake.count()-before != 4 {
		t.Fatalf("polled %d times", fake.count()-before)
	}
	if want := []time.Duration{2 * time.Second, 7 * time.Second, 7 * time.Second}; fmt.Sprint(waits) != fmt.Sprint(want) {
		t.Fatalf("waits %v, want %v (slow_down adds 5s and keeps it)", waits, want)
	}
	for _, call := range fake.calls {
		if call.Headers.Get("Authorization") != "" {
			t.Fatalf("authorization sent on %s", call.Path)
		}
	}

	waits = nil
	mu.Lock()
	answers = []string{"authorization_pending", "access_denied"}
	mu.Unlock()
	if _, err := WaitForComputerAgentKey(ctx, base, nil, registration); !errors.As(err, &apiErr) || apiErr.Code != "access_denied" || len(waits) != 1 {
		t.Fatalf("denied %v waits=%v", err, waits)
	}

	waits = nil
	mu.Lock()
	answers = nil
	mu.Unlock()
	short := &ComputerAgentRegistration{ClaimToken: "claim-1", Interval: 5, ExpiresIn: 3}
	var waitErr *ComputerWaitError
	if _, err := WaitForComputerAgentKey(ctx, base, nil, short); !errors.As(err, &waitErr) || len(waits) != 0 || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expiry %v waits=%v", err, waits)
	}

	waits = nil
	zero := &ComputerAgentRegistration{ClaimToken: "claim-1", Interval: 0, ExpiresIn: 600}
	mu.Lock()
	answers = []string{"authorization_pending", "ok"}
	mu.Unlock()
	if _, err := WaitForComputerAgentKey(ctx, base, nil, zero); err != nil || len(waits) != 1 || waits[0] != time.Second {
		t.Fatalf("minimum interval %v waits=%v", err, waits)
	}

	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	computerAgentWait = original
	mu.Lock()
	answers = nil
	mu.Unlock()
	if _, err := WaitForComputerAgentKey(cancelled, base, nil, registration); err == nil {
		t.Fatal("cancelled context must stop the wait")
	}
}

func TestComputerEnsureProject(t *testing.T) {
	client, fake := newControlServer(t, func(w http.ResponseWriter, r *http.Request, call controlCall) {
		writeJSON(w, 201, map[string]any{"id": "p1", "name": "Computers", "created": true})
	})
	project, err := client.Computers.EnsureProject(context.Background())
	if err != nil || project.ID != "p1" || !project.Created || fake.last().Method != "POST" || fake.last().Path != "/computers/project" {
		t.Fatalf("ensure project %+v %v %+v", project, err, fake.last())
	}
}

func TestComputerAgentProvidersResultFanoutsAndAttach(t *testing.T) {
	client, fake := newControlServer(t, func(w http.ResponseWriter, r *http.Request, call controlCall) {
		switch {
		case call.Path == "/computers/c1/agent/providers":
			writeJSON(w, 200, map[string]any{"providers": []map[string]any{{"provider": "gemini", "name": "Gemini CLI", "installed": true, "signed_in": false, "login_command": []string{"gemini"}, "model_keys": []string{"gemini"}}}})
		case call.Path == "/computers/c1/agent/models":
			writeJSON(w, 200, map[string]any{"models": []map[string]any{{"id": "gpt-5", "name": "gpt-5"}}, "source": "provider"})
		case call.Path == "/computers/c1/agent/conversations/v1/result":
			writeJSON(w, 200, map[string]any{"conversation_id": "v1", "provider": "claude", "turn": 2, "status": "idle", "result": map[string]any{"duration_ms": 1200, "exit_code": 0, "last_message": "done", "changed_files": map[string]any{"files": 1, "stat": " a.go | 4 +++-", "insertions": 3, "deletions": 1}}})
		case call.Path == "/computers/c1/agent/prompt":
			writeJSON(w, 202, map[string]any{"id": "f1", "source_computer_id": "c1", "provider": "codex", "created_at": "2026-11-12T00:00:00Z", "forks": []map[string]any{{"computer_id": "c2", "name": "fork-1", "state": "running", "prompt": "a", "status": "starting"}}})
		case call.Path == "/computers/c1/agent/turns":
			writeJSON(w, 200, map[string]any{"turns": []map[string]any{{"id": "t1", "conversation_id": "", "turn": 1, "provider": "codex", "model": "", "prompt": "a", "status": "starting", "error": "", "result": nil, "fanout_id": "f1", "created_at": "2026-11-12T00:00:00Z", "ended_at": nil}}})
		case call.Path == "/computers/c1/agent/fanouts":
			writeJSON(w, 200, map[string]any{"fanouts": []map[string]any{{"id": "f1", "forks": []map[string]any{}}}})
		case call.Path == "/computers/c1/agent/fanouts/f1" || call.Path == "/computers/c1/agent/fanouts/f1/keep":
			writeJSON(w, 200, map[string]any{"id": "f1", "forks": []map[string]any{}})
		case call.Path == "/computers/c1/files":
			w.WriteHeader(204)
		default:
			writeFailure(w, 404, "invalid_request", "no route")
		}
	})
	ctx := context.Background()

	providers, err := client.Computers.AgentProviders(ctx, "c1", true)
	if err != nil || len(providers) != 1 || providers[0].Provider != ComputerAgentGemini || fake.last().Method != "GET" || fake.last().Query["refresh"] != "1" {
		t.Fatalf("providers %+v %v %+v", providers, err, fake.last())
	}
	list, err := client.Computers.AgentModels(ctx, "c1", ComputerAgentCodex, "b1")
	if err != nil || list.Source != "provider" || list.Models[0].ID != "gpt-5" || fake.last().Method != "GET" || fake.last().Query["provider"] != "codex" || fake.last().Query["secret_binding_id"] != "b1" {
		t.Fatalf("models %+v %v %+v", list, err, fake.last())
	}
	result, err := client.Computers.AgentResult(ctx, "c1", "v1", true)
	if err != nil || result.Turn != 2 || result.Result.ChangedFiles.Insertions != 3 || fake.last().Query["wait"] != "1" {
		t.Fatalf("result %+v %v %+v", result, err, fake.last())
	}
	fanout, err := client.Computers.PromptAgentFanout(ctx, "c1", ComputerAgentFanoutPrompt{Prompts: []string{"a", "b"}, Provider: ComputerAgentCodex})
	body := fake.last().body(t)
	if err != nil || fanout.ID != "f1" || fanout.Forks[0].ComputerID != "c2" || body["prompts"] == nil || body["forks"] != nil || body["prompt"] != nil {
		t.Fatalf("fanout %+v %v %v", fanout, err, body)
	}
	if _, err := client.Computers.PromptAgent(ctx, "c1", ComputerAgentPrompt{Prompt: "x", Attachments: []string{"/workspace/.stackshift/attachments/a/b.png"}}); err != nil || fake.last().body(t)["attachments"] == nil {
		t.Fatalf("attachments not sent %v %s", err, fake.last().Raw)
	}
	if turns, err := client.Computers.AgentTurns(ctx, "c1", 20); err != nil || len(turns) != 1 || turns[0].Status != ComputerAgentStarting || turns[0].EndedAt != nil || *turns[0].FanoutID != "f1" || fake.last().Query["limit"] != "20" {
		t.Fatalf("turns %+v %v %+v", turns, err, fake.last())
	}
	if list, err := client.Computers.AgentFanouts(ctx, "c1"); err != nil || len(list) != 1 || fake.last().Path != "/computers/c1/agent/fanouts" {
		t.Fatalf("fanouts %+v %v", list, err)
	}
	if got, err := client.Computers.AgentFanout(ctx, "c1", "f1"); err != nil || got.ID != "f1" || fake.last().Path != "/computers/c1/agent/fanouts/f1" {
		t.Fatalf("fanout %+v %v", got, err)
	}
	if kept, err := client.Computers.KeepAgentFork(ctx, "c1", "f1", "c2"); err != nil || kept.ID != "f1" || fake.last().Method != "POST" || fake.last().Path != "/computers/c1/agent/fanouts/f1/keep" || fake.last().body(t)["computer_id"] != "c2" {
		t.Fatalf("keep %+v %v", kept, err)
	}

	file := filepath.Join(t.TempDir(), "spec.md")
	if err := os.WriteFile(file, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	remote, err := client.Computers.AttachAgentFile(ctx, "c1", file)
	if err != nil || !regexp.MustCompile(`^/workspace/\.stackshift/attachments/[0-9a-f]+/spec\.md$`).MatchString(remote) || fake.last().Method != "PUT" || fake.last().Query["path"] != remote || string(fake.last().Raw) != "hello" {
		t.Fatalf("attach %q %v %+v", remote, err, fake.last())
	}
}
