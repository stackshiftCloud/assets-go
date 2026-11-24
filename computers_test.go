package stackshift

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type computersCall struct {
	Method  string
	Path    string
	Query   map[string]string
	Headers http.Header
	Body    map[string]any
	Raw     []byte
}

type computersMock struct {
	mu         sync.Mutex
	polls      int
	calls      []computersCall
	computers  map[string]map[string]any
	reads      map[string]int
	snapshots  map[string]map[string]any
	approvals  map[string]map[string]any
	files      map[string][]byte
	idempotent map[string]map[string]any
	holder     string
	nextID     int
	settings   map[string]any
	setups     map[string]map[string]any
}

func newComputersMock(polls int) *computersMock {
	return &computersMock{polls: polls, computers: map[string]map[string]any{}, reads: map[string]int{}, snapshots: map[string]map[string]any{}, approvals: map[string]map[string]any{}, files: map[string][]byte{}, idempotent: map[string]map[string]any{}, holder: "agent", setups: map[string]map[string]any{}, settings: map[string]any{
		"plan": "pro", "can_edit": true, "running_now": 1, "spend_cap": nil, "spent_this_month": map[string]any{"amount_minor": 1250, "currency": "NGN"},
		"default_idle_policy": nil, "default_recording": true, "default_approvals": false,
		"notifications": map[string]any{"email": true, "push": false, "slack": false, "email_address": "owner@example.com", "slack_configured": false, "push_available": true, "push_devices": 1},
	}}
}

func (m *computersMock) id() string {
	m.nextID++
	return fmt.Sprintf("00000000-0000-4000-8000-%012d", m.nextID)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeFailure(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{"code": code, "message": message, "retryable": false}})
}

func (m *computersMock) view(record map[string]any) map[string]any {
	id := record["id"].(string)
	m.reads[id]++
	if record["state"] == "creating" && m.reads[id] > m.polls {
		record["state"] = "running"
	}
	copied := map[string]any{}
	for key, value := range record {
		copied[key] = value
	}
	return copied
}

func (m *computersMock) make(body map[string]any, extra map[string]any) map[string]any {
	record := map[string]any{
		"id": m.id(), "project_id": body["project_id"], "name": body["name"], "region": "eu", "template": body["template"], "template_version": "1",
		"size": body["size"], "state": "creating", "internet": map[string]any{"mode": "open"}, "idle_policy": map[string]any{"pause_after_seconds": 0, "mode": "memory"},
		"access": map[string]any{"keys": true}, "needs_you": false, "agent_paused": false, "input_holder": "agent", "ssh_command": "ssh x@y", "created_at": time.Now().UTC().Format(time.RFC3339), "updated_at": time.Now().UTC().Format(time.RFC3339),
	}
	if access, ok := body["access"].(map[string]any); ok {
		record["access"] = access
	}
	if internet, ok := body["internet"].(map[string]any); ok {
		record["internet"] = internet
	}
	for key, value := range extra {
		record[key] = value
	}
	return record
}

func (m *computersMock) mutate(w http.ResponseWriter, header http.Header, build func() map[string]any) {
	key := header.Get("Idempotency-Key")
	if key == "" {
		writeFailure(w, 400, "invalid_request", "Idempotency-Key is required")
		return
	}
	if seen, ok := m.idempotent[key]; ok {
		replay := map[string]any{}
		for k, v := range seen {
			replay[k] = v
		}
		replay["replayed"] = true
		writeJSON(w, 202, replay)
		return
	}
	answer := build()
	m.idempotent[key] = answer
	out := map[string]any{"replayed": false}
	for k, v := range answer {
		out[k] = v
	}
	writeJSON(w, 202, out)
}

var computersRoute = regexp.MustCompile(`^/api/v1/computers/([^/]+)(/.*)?$`)

func (m *computersMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	if r.Method == "GET" && call.Path == "/computers/public-pricing" {
		price := func(currency string, rate float64) map[string]any {
			return map[string]any{"currency": currency, "vcpu_hour": 0.035 * rate, "memory_gib_hour": 0.012 * rate, "disk_gb_month": 0.13 * rate, "snapshot_gb_month": 0.1 * rate, "egress_gb": 0.02 * rate, "effective_at": time.Now().UTC().Format(time.RFC3339)}
		}
		writeJSON(w, 200, map[string]any{"ngn": price("NGN", 1500), "usd": price("USD", 1), "free_plan": map[string]any{"computers": 1, "vcpus": 4, "memory_bytes": 8589934592, "running_hours_per_month": 4}, "running_limits": []map[string]any{{"plan": "free", "computers": 1}, {"plan": "pro", "computers": 25}}})
		return
	}
	if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		writeFailure(w, 401, "unauthorized", "missing bearer token")
		return
	}
	body, path, query := call.Body, call.Path, call.Query
	switch {
	case r.Method == "GET" && path == "/computers/templates":
		writeJSON(w, 200, []map[string]any{{"slug": "agent-desktop", "name": "Agent desktop", "description": "d", "version": "1", "desktop": true, "default_disk_bytes": 21474836480, "min_disk_bytes": 21474836480, "max_disk_bytes": 549755813888, "sizes": []any{}}})
		return
	case path == "/computers" && r.Method == "GET":
		var items []map[string]any
		for _, c := range m.sorted() {
			if c["project_id"] == query["project_id"] {
				items = append(items, c)
			}
		}
		limit, _ := strconv.Atoi(query["limit"])
		if limit == 0 {
			limit = 2
		}
		start, _ := strconv.Atoi(query["cursor"])
		end := start + limit
		if end > len(items) {
			end = len(items)
		}
		page := []map[string]any{}
		for _, c := range items[start:end] {
			page = append(page, m.view(c))
		}
		out := map[string]any{"items": page, "has_more": end < len(items)}
		if end < len(items) {
			out["next_cursor"] = strconv.Itoa(end)
		}
		writeJSON(w, 200, out)
		return
	case path == "/computers" && r.Method == "POST":
		if body["project_id"] == nil || body["template"] == nil || body["size"] == nil {
			writeFailure(w, 400, "invalid_request", "project_id, name, template and size are required")
			return
		}
		m.mutate(w, r.Header, func() map[string]any {
			created := m.make(body, nil)
			m.computers[created["id"].(string)] = created
			return map[string]any{"computer": created, "operation_id": m.id()}
		})
		return
	case r.Method == "GET" && path == "/computers/snapshots":
		items := []map[string]any{}
		for _, s := range m.snapshots {
			items = append(items, s)
		}
		writeJSON(w, 200, map[string]any{"items": items, "has_more": false})
		return
	case r.Method == "DELETE" && strings.HasPrefix(path, "/computers/snapshots/"):
		snapshotID := strings.TrimPrefix(path, "/computers/snapshots/")
		confirm := r.Header.Get("X-StackShift-Confirm-Delete")
		if confirm == "" || (confirm != snapshotID && (m.snapshots[snapshotID] == nil || confirm != m.snapshots[snapshotID]["name"])) {
			writeFailure(w, 428, "confirmation_required", "send the snapshot's id or name in X-StackShift-Confirm-Delete")
			return
		}
		delete(m.snapshots, snapshotID)
		w.WriteHeader(204)
		return
	case r.Method == "GET" && path == "/computers/ci":
		writeJSON(w, 200, map[string]any{
			"connections":   []map[string]any{{"id": "c1", "project_id": query["project_id"], "installation_id": 42, "account_login": "acme", "account_type": "Organization", "created_at": time.Now().UTC().Format(time.RFC3339), "repositories": []map[string]any{{"full_name": "acme/app", "private": true, "html_url": "https://github.com/acme/app"}}}},
			"installations": []map[string]any{{"installation_id": 43, "account_login": "me", "account_type": "User"}},
			"jobs":          []map[string]any{m.ciJob(query["project_id"])},
			"labels":        []string{"stackshift-small", "stackshift-small-mem", "stackshift-medium", "stackshift-medium-mem", "stackshift-large", "stackshift-xlarge"},
		})
		return
	case r.Method == "GET" && path == "/computers/ci/jobs":
		writeJSON(w, 200, map[string]any{"items": []map[string]any{m.ciJob(query["project_id"])}})
		return
	case r.Method == "POST" && path == "/computers/ci/connections":
		if body["installation_id"] == nil {
			writeFailure(w, 400, "invalid_request", "installation_id is required")
			return
		}
		writeJSON(w, 201, map[string]any{"id": "c2", "project_id": body["project_id"], "installation_id": body["installation_id"], "account_login": "me", "account_type": "User", "created_at": time.Now().UTC().Format(time.RFC3339)})
		return
	case r.Method == "POST" && path == "/computers/ci/circleci/connections":
		if body["resource_class"] == nil {
			writeFailure(w, 400, "invalid_request", "resource_class is required")
			return
		}
		writeJSON(w, 201, map[string]any{"id": "cc1", "project_id": body["project_id"], "resource_class": body["resource_class"], "size": body["size"], "max_concurrent_jobs": 2, "created_at": time.Now().UTC().Format(time.RFC3339)})
		return
	case r.Method == "DELETE" && strings.HasPrefix(path, "/computers/ci/circleci/connections/"):
		w.WriteHeader(204)
		return
	case r.Method == "DELETE" && strings.HasPrefix(path, "/computers/ci/connections/"):
		w.WriteHeader(204)
		return
	case r.Method == "GET" && strings.HasPrefix(path, "/computers/approvals/"):
		approval := m.approvals[strings.TrimPrefix(path, "/computers/approvals/")]
		polls, _ := approval["polls"].(int)
		approval["polls"] = polls + 1
		if polls+1 >= 2 {
			approval["status"] = "approved"
		}
		writeJSON(w, 200, approval)
		return
	case r.Method == "GET" && path == "/computers/pricing":
		currency := query["currency"]
		if currency == "" {
			currency = "USD"
		}
		writeJSON(w, 200, map[string]any{"currency": currency, "vcpu_hour": 0.01, "memory_gib_hour": 0.002, "disk_gb_month": 0.1, "snapshot_gb_month": 0.05, "egress_gb": 0.02, "effective_at": time.Now().UTC().Format(time.RFC3339)})
		return
	case r.Method == "GET" && path == "/computers/regions":
		writeJSON(w, 200, map[string]any{"items": []map[string]any{{"slug": "eu", "name": "Helsinki, Finland", "default": true}, {"slug": "us", "name": "Ashburn, USA", "default": false}}})
		return
	case path == "/computers/settings":
		if r.Method == "PUT" {
			if cap, ok := body["spend_cap"].(map[string]any); ok && cap["currency"] != "NGN" && cap["currency"] != "USD" {
				writeFailure(w, 400, "invalid_request", "currency must be NGN or USD")
				return
			}
			for _, field := range []string{"spend_cap", "default_idle_policy", "default_recording", "default_approvals"} {
				if value, present := body[field]; present {
					m.settings[field] = value
				}
			}
			if notifications, ok := body["notifications"].(map[string]any); ok {
				current := m.settings["notifications"].(map[string]any)
				for key, value := range notifications {
					current[key] = value
				}
			}
		}
		writeJSON(w, 200, m.settings)
		return
	case r.Method == "POST" && path == "/computers/images/import":
		if body["image"] == nil || body["name"] == nil || body["name"] == "" {
			writeFailure(w, 400, "invalid_request", "image and name are required")
			return
		}
		m.mutate(w, r.Header, func() map[string]any {
			name, _ := body["computer_name"].(string)
			if name == "" {
				name = fmt.Sprint(body["name"]) + "-builder"
			}
			builder := m.make(map[string]any{"project_id": body["project_id"], "name": name, "template": "coding", "size": body["size"], "access": body["access"]}, nil)
			if internet, ok := body["internet"].(map[string]any); ok {
				builder["internet"] = internet
			}
			m.computers[builder["id"].(string)] = builder
			keep, _ := body["keep_builder"].(bool)
			setup := map[string]any{"id": m.id(), "computer_id": builder["id"], "project_id": body["project_id"], "kind": "image", "spec": map[string]any{"image": body["image"]}, "state": "running", "steps": []map[string]any{{"name": "pull", "state": "running"}}, "image": map[string]any{"reference": body["image"], "service": true}, "import": map[string]any{"name": body["name"], "state": "waiting", "keep_builder": keep}, "log": "", "log_bytes": 0, "created_at": time.Now().UTC().Format(time.RFC3339), "updated_at": time.Now().UTC().Format(time.RFC3339)}
			m.setups[builder["id"].(string)] = setup
			return map[string]any{"computer": builder, "setup": setup}
		})
		return
	case r.Method == "GET" && path == "/computers/usage":
		if month := query["month"]; month != "" && !regexp.MustCompile(`^\d{4}-\d{2}$`).MatchString(month) {
			writeFailure(w, 400, "invalid_request", "month must be YYYY-MM")
			return
		}
		writeJSON(w, 200, map[string]any{"month": query["month"], "currency": query["currency"], "total": 1.5, "items": []map[string]any{{"computer_id": query["computer_id"], "running_seconds": 3600, "total": 1.5, "lines": []map[string]any{{"kind": "running", "detail": "4 vCPU", "amount": 1.5}}}}})
		return
	case r.Method == "POST" && strings.HasSuffix(path, "/decision") && strings.HasPrefix(path, "/computers/approvals/"):
		approval := m.approvals[strings.TrimSuffix(strings.TrimPrefix(path, "/computers/approvals/"), "/decision")]
		if approval["status"] != "pending" {
			writeFailure(w, 409, "computer_invalid_state", "already answered")
			return
		}
		if body["approved"] == true {
			approval["status"] = "approved"
		} else {
			approval["status"] = "declined"
		}
		writeJSON(w, 200, approval)
		return
	case path == "/computers/keys" && r.Method == "GET":
		writeJSON(w, 200, []map[string]any{{"id": m.id(), "project_id": query["project_id"], "name": "k", "prefix": "csk_", "created_at": time.Now().UTC().Format(time.RFC3339)}})
		return
	case path == "/computers/keys" && r.Method == "POST":
		writeJSON(w, 201, map[string]any{"id": m.id(), "project_id": body["project_id"], "name": body["name"], "prefix": "csk_", "created_at": time.Now().UTC().Format(time.RFC3339), "secret": "csk_secret"})
		return
	case r.Method == "DELETE" && strings.HasPrefix(path, "/computers/keys/"):
		w.WriteHeader(204)
		return
	}
	match := computersRoute.FindStringSubmatch(r.URL.Path)
	if match == nil || m.computers[match[1]] == nil {
		writeFailure(w, 404, "computer_not_found", "computer not found")
		return
	}
	target := m.computers[match[1]]
	sub := match[2]
	result := func() map[string]any { return map[string]any{"computer": m.view(target), "operation_id": m.id()} }
	switch {
	case sub == "" && r.Method == "GET":
		writeJSON(w, 200, m.view(target))
	case sub == "" && r.Method == "PATCH":
		for key, value := range body {
			target[key] = value
		}
		out := result()
		out["replayed"] = false
		writeJSON(w, 200, out)
	case sub == "" && r.Method == "DELETE":
		if confirm := r.Header.Get("X-StackShift-Confirm-Delete"); confirm != target["id"] && confirm != target["name"] || confirm == "" {
			writeFailure(w, 428, "confirmation_required", "send the computer's id or name in X-StackShift-Confirm-Delete")
			return
		}
		m.mutate(w, r.Header, func() map[string]any { target["state"] = "deleting"; return result() })
	case sub == "/fork":
		m.mutate(w, r.Header, func() map[string]any {
			forks := []map[string]any{}
			names, _ := body["names"].([]any)
			for i := 0; i < int(body["count"].(float64)); i++ {
				name := fmt.Sprintf("%v-%d", target["name"], i+1)
				if i < len(names) {
					name = names[i].(string)
				}
				fork := m.make(map[string]any{"project_id": target["project_id"], "name": name, "template": target["template"], "size": target["size"]}, map[string]any{"parent_id": target["id"], "fork_mode": "live"})
				m.computers[fork["id"].(string)] = fork
				forks = append(forks, fork)
			}
			return map[string]any{"source": m.view(target), "forks": forks, "operation_id": m.id()}
		})
	case sub == "/pause":
		m.mutate(w, r.Header, func() map[string]any { target["state"] = "paused"; return result() })
	case sub == "/resume":
		m.mutate(w, r.Header, func() map[string]any { target["state"] = "running"; return result() })
	case sub == "/restore":
		m.mutate(w, r.Header, func() map[string]any {
			target["state"] = "running"
			target["from_snapshot_id"] = body["snapshot_id"]
			return result()
		})
	case sub == "/snapshots":
		if r.Header.Get("Idempotency-Key") == "" {
			writeFailure(w, 400, "invalid_request", "Idempotency-Key is required")
			return
		}
		snapshot := map[string]any{"id": m.id(), "project_id": target["project_id"], "computer_id": target["id"], "name": body["name"], "mode": body["mode"], "status": "pending", "template": target["template"], "template_version": "1", "size": target["size"], "stored_bytes": 0, "computers_started": 0, "created_at": time.Now().UTC().Format(time.RFC3339)}
		m.snapshots[snapshot["id"].(string)] = snapshot
		writeJSON(w, 202, snapshot)
	case sub == "/exec":
		text, _ := body["shell"].(string)
		writeJSON(w, 200, map[string]any{"process_id": "p0", "exit_code": 0, "stdout": "ran " + text, "stderr": "", "timed_out": false})
	case sub == "/files" && r.Method == "GET":
		content, ok := m.files[query["path"]]
		if !ok {
			writeFailure(w, 404, "invalid_request", "no such file")
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(content)
	case sub == "/files" && r.Method == "PUT":
		m.files[query["path"]] = raw
		w.WriteHeader(204)
	case sub == "/files/list":
		entries := []map[string]any{}
		for path, content := range m.files {
			entries = append(entries, map[string]any{"name": path, "path": path, "type": "file", "size_bytes": len(content)})
		}
		writeJSON(w, 200, entries)
	case sub == "/ports" && r.Method == "POST":
		writeJSON(w, 201, map[string]any{"port": body["port"], "visibility": body["visibility"], "url": fmt.Sprintf("https://%v-x.sandbox.stackshift.cloud", body["port"])})
	case strings.HasPrefix(sub, "/ports/") && r.Method == "DELETE":
		w.WriteHeader(204)
	case sub == "/actions":
		if (m.holder == "human" || target["agent_paused"] == true) && body["actor"] == "agent" {
			writeFailure(w, 423, "input_locked", "a person holds the controls")
			return
		}
		results := []map[string]any{}
		for _, item := range body["actions"].([]any) {
			action := item.(map[string]any)
			entry := map[string]any{"type": action["type"], "ok": action["type"] != "focus_window"}
			if action["type"] == "cursor_position" {
				entry["cursor"] = map[string]any{"x": 10, "y": 20}
			}
			if action["type"] == "focus_window" {
				entry["error"] = map[string]any{"code": "window_not_found", "message": "no window"}
			}
			if action["screenshot"] == true {
				entry["screenshot"] = map[string]any{"media_type": "image/png", "width": 1024, "height": 768, "scale": 1, "data": "bW9jaw==", "taken_at": time.Now().UTC().Format(time.RFC3339)}
			}
			results = append(results, entry)
		}
		writeJSON(w, 200, map[string]any{"step_id": body["step_id"], "results": results})
	case sub == "/screenshot":
		writeJSON(w, 200, map[string]any{"media_type": "image/png", "width": 1280, "height": 800, "scale": 1, "data": "bW9jaw==", "taken_at": time.Now().UTC().Format(time.RFC3339)})
	case sub == "/take-over":
		m.holder = "human"
		writeJSON(w, 200, map[string]any{"holder": "human"})
	case sub == "/hand-back":
		m.holder = "agent"
		writeJSON(w, 200, map[string]any{"holder": "agent"})
	case sub == "/steps":
		next, _ := strconv.Atoi(firstNonEmpty(query["before_seq"], query["after_seq"], "0"))
		writeJSON(w, 200, map[string]any{"steps": []map[string]any{{"seq": 1, "step_id": "s", "actor": "agent", "action": map[string]any{"type": "click"}, "ok": true, "started_at": time.Now().UTC().Format(time.RFC3339), "duration_ms": 3}}, "next_seq": next + 1})
	case sub == "/recordings":
		writeJSON(w, 200, []map[string]any{{"recording_id": "r1", "path": "/r1.webm", "started_at": time.Now().UTC().Format(time.RFC3339), "size_bytes": 3}})
	case strings.HasPrefix(sub, "/recordings/") && r.Method == "GET":
		_, _ = w.Write([]byte{7, 8, 9})
	case sub == "/system" && r.Method == "GET":
		live := target["state"] == "running"
		out := map[string]any{"computer_id": target["id"], "live": live, "region": "eu", "region_name": "Helsinki, Finland", "os_name": "Debian GNU/Linux", "os_version": "13", "kernel": "6.12.0", "services": []any{}, "listening": []any{}, "idle": false}
		if live {
			out["services"] = []map[string]any{{"unit": "ssh.service", "description": "OpenBSD Secure Shell server", "active": "active", "sub": "running", "enabled": true}}
			out["listening"] = []map[string]any{{"port": 3000, "address": "0.0.0.0", "protocol": "tcp", "process": "node", "published": true, "url": "https://3000-x.sandbox.stackshift.cloud"}}
			out["read_at"] = time.Now().UTC().Format(time.RFC3339)
		}
		writeJSON(w, 200, out)
	case sub == "/windows" && r.Method == "GET":
		if target["state"] != "running" {
			writeFailure(w, 409, "computer_invalid_state", "computer is not running")
			return
		}
		writeJSON(w, 200, map[string]any{"windows": []map[string]any{{"id": "w1", "title": "t", "app": "a", "x": 0, "y": 0, "width": 1, "height": 1, "focused": true}}})
	case sub == "/task" && r.Method == "PUT":
		for _, field := range []string{"task", "next_step"} {
			if text, present := body[field].(string); present {
				if len(text) > 2000 {
					writeFailure(w, 400, "invalid_request", "text is too long")
					return
				}
				target[field] = text
			}
		}
		target["task_updated_at"] = time.Now().UTC().Format(time.RFC3339)
		task, _ := target["task"].(string)
		next, _ := target["next_step"].(string)
		writeJSON(w, 200, map[string]any{"task": task, "next_step": next, "updated_at": target["task_updated_at"]})
	case sub == "/task" && r.Method == "DELETE":
		delete(target, "task")
		delete(target, "next_step")
		delete(target, "task_updated_at")
		w.WriteHeader(204)
	case sub == "/setup" && r.Method == "POST":
		chosen := 0
		for _, key := range []string{"script", "devcontainer", "source"} {
			if body[key] != nil {
				chosen++
			}
		}
		if chosen != 1 {
			writeFailure(w, 400, "invalid_request", "exactly one of script, devcontainer or source")
			return
		}
		kind := "script"
		for _, key := range []string{"devcontainer", "source"} {
			if body[key] != nil {
				kind = key
			}
		}
		if r.Header.Get("Idempotency-Key") == "" {
			writeFailure(w, 400, "invalid_request", "Idempotency-Key is required")
			return
		}
		setup := map[string]any{"id": m.id(), "computer_id": target["id"], "project_id": target["project_id"], "kind": kind, "spec": body, "state": "pending", "steps": []any{}, "log": "", "log_bytes": 0, "created_at": time.Now().UTC().Format(time.RFC3339), "updated_at": time.Now().UTC().Format(time.RFC3339)}
		m.setups[target["id"].(string)] = setup
		writeJSON(w, 202, setup)
	case sub == "/setup" && r.Method == "GET":
		setup := m.setups[target["id"].(string)]
		if setup == nil {
			writeFailure(w, 404, "invalid_request", "no setup")
			return
		}
		writeJSON(w, 200, setup)
	case sub == "/pause-agent":
		target["agent_paused"] = true
		writeJSON(w, 200, m.view(target))
	case sub == "/resume-agent":
		target["agent_paused"] = false
		writeJSON(w, 200, m.view(target))
	case sub == "/processes":
		writeJSON(w, 200, map[string]any{"processes": []map[string]any{{"pid": 1, "name": "chrome", "cpu_percent": 12.5, "memory_bytes": 1024}}, "memory_used_bytes": 1, "memory_total_bytes": 2, "disk_used_bytes": 3, "disk_total_bytes": 4})
	case sub == "/logs":
		entries := []map[string]any{{"time": "2026-11-09T00:00:00Z", "source": "console", "line": "boot"}, {"time": "2026-11-09T00:00:00Z", "source": "agent", "line": "hello"}}
		kept := []map[string]any{}
		for _, entry := range entries {
			if query["source"] == "" || entry["source"] == query["source"] {
				kept = append(kept, entry)
			}
		}
		writeJSON(w, 200, map[string]any{"entries": kept, "next_cursor": "c2"})
	case sub == "/approvals" && r.Method == "GET":
		items := []map[string]any{}
		for _, a := range m.approvals {
			if a["computer_id"] == target["id"] && (query["status"] == "" || a["status"] == query["status"]) {
				items = append(items, a)
			}
		}
		writeJSON(w, 200, items)
	case sub == "/approvals":
		kind, _ := body["kind"].(string)
		if kind == "" {
			kind = "approve"
		}
		access, _ := target["access"].(map[string]any)
		if kind == "approve" && access["approvals"] != true {
			writeFailure(w, 409, "approvals_disabled", "approvals are off")
			return
		}
		approval := map[string]any{"id": m.id(), "computer_id": target["id"], "kind": kind, "question": body["question"], "page_url": body["page_url"], "next_step": body["next_step"], "status": "pending", "created_at": time.Now().UTC().Format(time.RFC3339), "notified_at": time.Now().UTC().Format(time.RFC3339), "notified_channels": []string{"in_app", "email"}}
		m.approvals[approval["id"].(string)] = approval
		target["needs_you"] = true
		writeJSON(w, 201, approval)
	case sub == "/sessions":
		writeJSON(w, 201, map[string]any{"id": m.id(), "computer_id": target["id"], "kind": body["kind"], "url": fmt.Sprintf("/api/v1/computers/sessions/%v/tok", body["kind"]), "token": "tok", "expires_at": time.Now().UTC().Format(time.RFC3339)})
	case sub == "/events":
		after, _ := strconv.Atoi(query["after_seq"])
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, ": keepalive\n\n")
		for seq := 1; seq <= 3; seq++ {
			if seq > after {
				fmt.Fprintf(w, "id: %d\nevent: state_changed\ndata: {\"seq\":%d,\"type\":\"state_changed\",\"computer_id\":%q,\"state\":\"running\",\"time\":\"2026-11-09T00:00:00Z\"}\n\n", seq, seq, target["id"])
			}
		}
	default:
		writeFailure(w, 404, "invalid_request", "no route")
	}
}

func (m *computersMock) ciJob(projectID string) map[string]any {
	return map[string]any{"provider": "github", "id": "j1", "project_id": projectID, "github_job_id": 11, "github_run_id": 22, "repository": "acme/app", "repository_private": true, "workflow_name": "ci", "job_name": "test", "head_branch": "main", "head_sha": "abc", "labels": []string{"stackshift-small-mem"}, "size": "small-mem", "state": "running", "runner_name": "r1", "queued_at": time.Now().UTC().Format(time.RFC3339)}
}

func (m *computersMock) sorted() []map[string]any {
	var out []map[string]any
	for i := 1; i <= m.nextID; i++ {
		id := fmt.Sprintf("00000000-0000-4000-8000-%012d", i)
		if c, ok := m.computers[id]; ok {
			out = append(out, c)
		}
	}
	return out
}

func (m *computersMock) last() computersCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls[len(m.calls)-1]
}

const computersProject = "3f0c6a52-7d1e-4f53-9a3c-0d2b7f8e1a44"

func computersFixture(t *testing.T, polls int) (*Client, *computersMock) {
	t.Helper()
	mock := newComputersMock(polls)
	server := httptest.NewServer(mock)
	t.Cleanup(server.Close)
	client, err := New(Options{APIKey: "csk_fixture", BaseURL: server.URL + "/api/v1"})
	if err != nil {
		t.Fatal(err)
	}
	return client, mock
}

func TestComputersLifecycleExecFilesPortsAndActions(t *testing.T) {
	client, mock := computersFixture(t, 2)
	ctx := context.Background()

	templates, err := client.Computers.Templates(ctx, computersProject)
	if err != nil || templates[0].Slug != "agent-desktop" || templates[0].MinDiskBytes != 20<<30 || templates[0].MaxDiskBytes != 512<<30 {
		t.Fatalf("templates %v %v", templates, err)
	}
	input := CreateComputerInput{ProjectID: computersProject, Name: "inbox-helper", Template: "agent-desktop", Size: NewComputerSize(4, 8, 20)}
	created, err := client.Computers.Create(ctx, input, "create-1")
	if err != nil || created.Computer.State != "creating" || created.Replayed {
		t.Fatalf("create %+v %v", created, err)
	}
	if got := mock.calls[len(mock.calls)-1].Headers.Get("Idempotency-Key"); got != "create-1" {
		t.Fatalf("idempotency key %q", got)
	}
	if size := mock.calls[len(mock.calls)-1].Body["size"].(map[string]any); size["memory_bytes"].(float64) != 8*(1<<30) {
		t.Fatalf("size %v", size)
	}
	replay, err := client.Computers.Create(ctx, input, "create-1")
	if err != nil || !replay.Replayed {
		t.Fatalf("replay %+v %v", replay, err)
	}
	id := created.Computer.ID
	computer, err := client.Computers.Wait(ctx, id, "running", time.Millisecond)
	if err != nil || computer.State != "running" {
		t.Fatalf("wait %+v %v", computer, err)
	}

	execResult, err := client.Computers.Exec(ctx, id, ComputerExecInput{Shell: "npm test", CWD: "/workspace/app"})
	if err != nil || execResult.Stdout != "ran npm test" || *execResult.ExitCode != 0 {
		t.Fatalf("exec %+v %v", execResult, err)
	}
	if err := client.Computers.WriteFile(ctx, id, "a.txt", []byte("héllo"), 493); err != nil {
		t.Fatal(err)
	}
	if call := mock.last(); call.Method != "PUT" || call.Query["mode"] != "493" || call.Headers.Get("Content-Type") != "application/octet-stream" || string(call.Raw) != "héllo" {
		t.Fatalf("write %+v", call)
	}
	content, err := client.Computers.ReadFile(ctx, id, "a.txt")
	if err != nil || string(content) != "héllo" {
		t.Fatalf("read %q %v", content, err)
	}
	if _, err := client.Computers.ReadFile(ctx, id, "missing"); err == nil {
		t.Fatal("missing file must fail")
	}
	files, err := client.Computers.ListFiles(ctx, id, ".")
	if err != nil || len(files) != 1 {
		t.Fatalf("list %v %v", files, err)
	}

	port, err := client.Computers.PublishPort(ctx, id, 3000, "public")
	if err != nil || port.Port != 3000 || !strings.Contains(port.URL, "3000") {
		t.Fatalf("port %+v %v", port, err)
	}
	if err := client.Computers.UnpublishPort(ctx, id, 3000); err != nil || !strings.HasSuffix(mock.last().Path, "/ports/3000") {
		t.Fatalf("unpublish %v", err)
	}

	response, err := client.Computers.Act(ctx, id, []ComputerAction{ComputerClick(0, 0).WithScreenshot(), ComputerType("hello"), ComputerKeys("ctrl+c"), ComputerScroll(5, 5, 0, 3)}, "", "step-1")
	if err != nil || len(response.Results) != 4 || response.Results[0].Screenshot == nil || response.StepID != "step-1" {
		t.Fatalf("act %+v %v", response, err)
	}
	actions := mock.last().Body["actions"].([]any)
	if first := actions[0].(map[string]any); first["x"].(float64) != 0 || first["y"].(float64) != 0 || first["screenshot"] != true {
		t.Fatalf("zero coordinates must be sent: %v", first)
	}
	if mock.last().Body["actor"] != "agent" {
		t.Fatalf("actor %v", mock.last().Body["actor"])
	}
	failed, err := client.Computers.Act(ctx, id, []ComputerAction{{Type: "focus_window", WindowID: "x"}}, "", "")
	if err != nil || failed.Results[0].Err() == nil {
		t.Fatalf("focus %+v %v", failed, err)
	}
	var actionErr *APIError
	if !errors.As(failed.Results[0].Err(), &actionErr) || actionErr.Code != "window_not_found" {
		t.Fatalf("action error %v", failed.Results[0].Err())
	}

	if _, err := client.Computers.TakeOver(ctx, id); err != nil {
		t.Fatal(err)
	}
	_, err = client.Computers.Act(ctx, id, []ComputerAction{ComputerClick(1, 1)}, "", "")
	var locked *APIError
	if !errors.As(err, &locked) || locked.StatusCode != 423 || locked.Code != "input_locked" {
		t.Fatalf("locked %v", err)
	}
	if _, err := client.Computers.HandBack(ctx, id); err != nil {
		t.Fatal(err)
	}
	if shot, err := client.Computers.Screenshot(ctx, id, ComputerScreenshotOptions{MaxWidth: 640}); err != nil || shot.Width != 1280 || mock.last().Query["max_width"] != "640" {
		t.Fatalf("screenshot %+v %v", shot, err)
	}
}

func TestComputersForkPauseSnapshotApprovalSessionsEventsKeys(t *testing.T) {
	client, mock := computersFixture(t, 0)
	ctx := context.Background()
	computer, err := client.Computers.CreateAndWait(ctx, CreateComputerInput{ProjectID: computersProject, Name: "c", Template: "coding", Size: NewComputerSize(2, 4, 20)}, "")
	if err != nil {
		t.Fatal(err)
	}
	forks, err := client.Computers.Fork(ctx, computer.ID, ComputerForkInput{Count: 2, Names: []string{"a", "b"}}, "")
	if err != nil || len(forks.Forks) != 2 || forks.Forks[0].Name != "a" || forks.Forks[1].ParentID != computer.ID {
		t.Fatalf("fork %+v %v", forks, err)
	}
	if mock.last().Headers.Get("Idempotency-Key") == "" {
		t.Fatal("generated idempotency key expected")
	}

	paused, err := client.Computers.Pause(ctx, computer.ID, "memory", "")
	if err != nil || paused.Computer.State != "paused" {
		t.Fatalf("pause %+v %v", paused, err)
	}
	if resumed, err := client.Computers.Resume(ctx, computer.ID, ""); err != nil || resumed.Computer.State != "running" {
		t.Fatalf("resume %+v %v", resumed, err)
	}
	snapshot, err := client.Computers.Snapshot(ctx, computer.ID, "before", "machine", "")
	if err != nil || snapshot.Mode != "machine" {
		t.Fatalf("snapshot %+v %v", snapshot, err)
	}
	page, err := client.Computers.Snapshots.List(ctx, computersProject, computer.ID, "", 0)
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("snapshots %+v %v", page, err)
	}
	if _, err := client.Computers.Restore(ctx, computer.ID, snapshot.ID, ""); err != nil {
		t.Fatal(err)
	}
	if err := client.Computers.Snapshots.Delete(ctx, snapshot.ID); err != nil {
		t.Fatal(err)
	}

	_, err = client.Computers.AskOwner(ctx, computer.ID, ComputerAskInput{Question: "Send it?"})
	var refused *APIError
	if !errors.As(err, &refused) || refused.Code != "approvals_disabled" || refused.StatusCode != 409 {
		t.Fatalf("approvals %v", err)
	}
	on := true
	if _, err := client.Computers.Update(ctx, computer.ID, UpdateComputerInput{Access: &ComputerAccess{Approvals: &on}}); err != nil {
		t.Fatal(err)
	}
	approval, err := client.Computers.AskOwner(ctx, computer.ID, ComputerAskInput{Question: "Send it?", Details: "to bob", PageURL: "https://example.com/pay"})
	if err != nil || approval.Status != "pending" || approval.Kind != "approve" || mock.last().Body["page_url"] != "https://example.com/pay" {
		t.Fatalf("ask %+v %v", approval, err)
	}
	decided, err := client.Computers.WaitForApproval(ctx, approval.ID, time.Millisecond)
	if err != nil || decided.Status != "approved" {
		t.Fatalf("decided %+v %v", decided, err)
	}

	help, err := client.Computers.AskOwner(ctx, computer.ID, ComputerAskInput{Kind: "help", Question: "Solve this captcha"})
	if err != nil || help.Kind != "help" || mock.last().Body["kind"] != "help" {
		t.Fatalf("help %+v %v", help, err)
	}
	if refreshed, err := client.Computers.Get(ctx, computer.ID); err != nil || !refreshed.NeedsYou {
		t.Fatalf("needs_you %+v %v", refreshed, err)
	}
	if pending, err := client.Computers.Approvals(ctx, computer.ID, "pending"); err != nil || len(pending) != 1 || mock.last().Query["status"] != "pending" {
		t.Fatalf("approvals %v %v", pending, err)
	}
	if decided, err := client.Computers.DecideApproval(ctx, help.ID, true); err != nil || decided.Status != "approved" || mock.last().Body["approved"] != true {
		t.Fatalf("decide %+v %v", decided, err)
	}
	_, err = client.Computers.DecideApproval(ctx, help.ID, false)
	var answered *APIError
	if !errors.As(err, &answered) || answered.StatusCode != 409 || answered.Code != "computer_invalid_state" {
		t.Fatalf("answered %v", err)
	}
	if paused, err := client.Computers.PauseAgent(ctx, computer.ID); err != nil || !paused.AgentPaused {
		t.Fatalf("pause agent %+v %v", paused, err)
	}
	_, err = client.Computers.Act(ctx, computer.ID, []ComputerAction{ComputerClick(1, 1)}, "", "")
	var agentLocked *APIError
	if !errors.As(err, &agentLocked) || agentLocked.StatusCode != 423 {
		t.Fatalf("paused agent act %v", err)
	}
	if resumed, err := client.Computers.ResumeAgent(ctx, computer.ID); err != nil || resumed.AgentPaused {
		t.Fatalf("resume agent %+v %v", resumed, err)
	}
	steps, err := client.Computers.Steps(ctx, computer.ID, ComputerStepOptions{Order: "latest", BeforeSeq: 10, Limit: 5})
	if q := mock.last().Query; err != nil || steps.NextSeq != 11 || q["order"] != "latest" || q["before_seq"] != "10" || q["limit"] != "5" {
		t.Fatalf("steps %+v %v", steps, err)
	}
	recordings, err := client.Computers.Recordings(ctx, computer.ID)
	if err != nil || len(recordings) != 1 {
		t.Fatalf("recordings %v %v", recordings, err)
	}
	if video, err := client.Computers.DownloadRecording(ctx, computer.ID, recordings[0].RecordingID); err != nil || fmt.Sprint(video) != "[7 8 9]" {
		t.Fatalf("video %v %v", video, err)
	}
	if processes, err := client.Computers.Processes(ctx, computer.ID); err != nil || processes.Processes[0].Name != "chrome" || processes.DiskTotalBytes != 4 {
		t.Fatalf("processes %+v %v", processes, err)
	}
	logs, err := client.Computers.Logs(ctx, computer.ID, ComputerLogOptions{Source: "agent", Limit: 10})
	if err != nil || len(logs.Entries) != 1 || logs.Entries[0].Source != "agent" || logs.NextCursor != "c2" {
		t.Fatalf("logs %+v %v", logs, err)
	}
	usage, err := client.Computers.Usage(ctx, computersProject, ComputerUsageOptions{Month: "2026-11", Currency: "NGN"})
	if err != nil || usage.Month != "2026-11" || usage.Items[0].Lines[0].Kind != "running" {
		t.Fatalf("usage %+v %v", usage, err)
	}
	_, err = client.Computers.Usage(ctx, computersProject, ComputerUsageOptions{Month: "nov"})
	var badMonth *APIError
	if !errors.As(err, &badMonth) || badMonth.Code != "invalid_request" {
		t.Fatalf("bad month %v", err)
	}
	if pricing, err := client.Computers.Pricing(ctx, computersProject, "NGN"); err != nil || pricing.Currency != "NGN" || pricing.VCPUHour != 0.01 {
		t.Fatalf("pricing %+v %v", pricing, err)
	}

	cdp, err := client.Computers.CDPURL(ctx, computer.ID)
	if err != nil || !strings.HasPrefix(cdp, "ws://127.0.0.1:") || !strings.HasSuffix(cdp, "/api/v1/computers/sessions/cdp/tok") {
		t.Fatalf("cdp %q %v", cdp, err)
	}
	if mock.last().Body["kind"] != "cdp" {
		t.Fatalf("kind %v", mock.last().Body["kind"])
	}
	terminal, err := client.Computers.TerminalURL(ctx, computer.ID, 30, 100)
	if err != nil || !strings.Contains(terminal, "rows=30") || !strings.Contains(terminal, "columns=100") {
		t.Fatalf("terminal %q %v", terminal, err)
	}
	if secure, _ := (&ComputersClient{client: &Client{baseURL: "https://api.stackshift.cloud/api/v1"}}).SocketURL("/api/v1/computers/sessions/live/x", 0, 0); secure != "wss://api.stackshift.cloud/api/v1/computers/sessions/live/x" {
		t.Fatalf("socket %q", secure)
	}

	var seen []int64
	if err := client.Computers.Events(ctx, computer.ID, 0, false, func(event ComputerEvent) bool { seen = append(seen, event.Seq); return true }); err != nil || fmt.Sprint(seen) != "[1 2 3]" {
		t.Fatalf("events %v %v", seen, err)
	}
	seen = nil
	if err := client.Computers.Events(ctx, computer.ID, 2, false, func(event ComputerEvent) bool { seen = append(seen, event.Seq); return true }); err != nil || fmt.Sprint(seen) != "[3]" {
		t.Fatalf("resumed events %v %v", seen, err)
	}
	seen = nil
	if err := client.Computers.Events(ctx, computer.ID, 0, true, func(event ComputerEvent) bool { seen = append(seen, event.Seq); return false }); err != nil || len(seen) != 1 {
		t.Fatalf("stopped events %v %v", seen, err)
	}

	key, err := client.Computers.Keys.Create(ctx, CreateComputerKeyInput{ProjectID: computersProject, Name: "agent", Limits: &ComputerKeyLimits{MaxRunning: 2}})
	if err != nil || key.Secret != "csk_secret" {
		t.Fatalf("key %+v %v", key, err)
	}
	if keys, err := client.Computers.Keys.List(ctx, computersProject); err != nil || len(keys) != 1 {
		t.Fatalf("keys %v %v", keys, err)
	}
	if err := client.Computers.Keys.Revoke(ctx, key.ID); err != nil {
		t.Fatal(err)
	}

	all, err := client.Computers.ListAll(ctx, computersProject, "")
	if err != nil || len(all) != 3 {
		t.Fatalf("list all %d %v", len(all), err)
	}
	deleted, err := client.Computers.Delete(ctx, computer.ID, "")
	if err != nil || deleted.Computer.State != "deleting" {
		t.Fatalf("delete %+v %v", deleted, err)
	}
	_, err = client.Computers.Get(ctx, "00000000-0000-4000-8000-999999999999")
	var missing *APIError
	if !errors.As(err, &missing) || missing.StatusCode != 404 || missing.Code != "computer_not_found" {
		t.Fatalf("missing %v", err)
	}
}

func TestComputersWaitStopsWithContext(t *testing.T) {
	client, _ := computersFixture(t, 1000)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	created, err := client.Computers.Create(context.Background(), CreateComputerInput{ProjectID: computersProject, Name: "c", Template: "coding", Size: NewComputerSize(2, 4, 20)}, "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Computers.Wait(ctx, created.Computer.ID, "running", 10*time.Millisecond)
	var waitErr *ComputerWaitError
	if !errors.As(err, &waitErr) {
		t.Fatalf("wait %v", err)
	}
}

func TestComputersGapsSystemTaskRegionsSettingsPricingImport(t *testing.T) {
	client, mock := computersFixture(t, 0)
	ctx := context.Background()
	computer, err := client.Computers.CreateAndWait(ctx, CreateComputerInput{ProjectID: computersProject, Name: "c", Template: "coding", Size: NewComputerSize(2, 8, 20), Internet: &ComputerInternet{Mode: ComputerInternetNone}}, "")
	if err != nil {
		t.Fatal(err)
	}

	system, err := client.Computers.System(ctx, computer.ID)
	if err != nil || mock.last().Path != "/computers/"+computer.ID+"/system" || !system.Live || system.OSName != "Debian GNU/Linux" || system.Services[0].Unit != "ssh.service" || !system.Listening[0].Published || system.Listening[0].Port != 3000 || system.ReadAt == nil {
		t.Fatalf("system %+v %v", system, err)
	}
	windows, err := client.Computers.ListWindows(ctx, computer.ID)
	if err != nil || len(windows.Windows) != 1 || windows.Windows[0].ID != "w1" {
		t.Fatalf("windows %+v %v", windows, err)
	}

	task, next := "Book the flight", "Open the airline site"
	saved, err := client.Computers.SetTask(ctx, computer.ID, ComputerTaskInput{Task: &task, NextStep: &next})
	if err != nil || saved.Task != task || saved.NextStep != next || mock.last().Body["task"] != task || mock.last().Body["next_step"] != next {
		t.Fatalf("set task %+v %v", saved, err)
	}
	seat := "Pick a seat"
	if _, err := client.Computers.SetTask(ctx, computer.ID, ComputerTaskInput{NextStep: &seat}); err != nil {
		t.Fatal(err)
	}
	if _, present := mock.last().Body["task"]; present || mock.last().Body["next_step"] != seat {
		t.Fatalf("partial task body %v", mock.last().Body)
	}
	tasked, err := client.Computers.Get(ctx, computer.ID)
	if err != nil || tasked.Task != task || tasked.NextStep != seat || tasked.TaskUpdatedAt == nil || tasked.Internet.Mode != ComputerInternetNone {
		t.Fatalf("tasked %+v %v", tasked, err)
	}
	long := strings.Repeat("x", 2001)
	var rejected *APIError
	if _, err := client.Computers.SetTask(ctx, computer.ID, ComputerTaskInput{Task: &long}); !errors.As(err, &rejected) || rejected.Code != ComputerErrorInvalidRequest {
		t.Fatalf("long task %v", err)
	}
	if err := client.Computers.ClearTask(ctx, computer.ID); err != nil || mock.last().Method != "DELETE" {
		t.Fatalf("clear task %v", err)
	}
	cleared, err := client.Computers.Get(ctx, computer.ID)
	if err != nil || cleared.Task != "" || cleared.TaskUpdatedAt != nil {
		t.Fatalf("cleared %+v %v", cleared, err)
	}

	if _, err := client.Computers.ActWith(ctx, computer.ID, []ComputerAction{ComputerClick(2, 2)}, ComputerActOptions{Task: "Pay", NextStep: "Confirm"}); err != nil {
		t.Fatal(err)
	}
	if mock.last().Body["task"] != "Pay" || mock.last().Body["next_step"] != "Confirm" || mock.last().Body["actor"] != "agent" {
		t.Fatalf("act with %v", mock.last().Body)
	}
	if _, err := client.Computers.Act(ctx, computer.ID, []ComputerAction{ComputerClick(3, 3)}, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, present := mock.last().Body["task"]; present {
		t.Fatalf("plain act carried a task %v", mock.last().Body)
	}

	approval, err := client.Computers.AskOwner(ctx, computer.ID, ComputerAskInput{Kind: "help", Question: "Wire the money?", NextStep: "Send the receipt"})
	if err != nil || mock.last().Body["next_step"] != "Send the receipt" || approval.NextStep != "Send the receipt" || approval.NotifiedAt == nil || len(approval.NotifiedChannels) != 2 || approval.NotifiedChannels[1] != "email" {
		t.Fatalf("ask %+v %v", approval, err)
	}

	regions, err := client.Computers.Regions(ctx, computersProject)
	if err != nil || mock.last().Query["project_id"] != computersProject || len(regions.Items) != 2 || regions.Items[0].Slug != "eu" || !regions.Items[0].Default {
		t.Fatalf("regions %+v %v", regions, err)
	}

	settings, err := client.Computers.Settings(ctx, computersProject)
	if err != nil || settings.Plan != "pro" || settings.SpendCap != nil || !settings.Notifications.PushAvailable || settings.SpentThisMonth.AmountMinor != 1250 {
		t.Fatalf("settings %+v %v", settings, err)
	}
	yes := true
	saveCap, err := client.Computers.SaveSettings(ctx, computersProject, SaveComputerSettingsInput{SpendCap: &ComputerMoney{AmountMinor: 500000, Currency: "NGN"}, DefaultApprovals: &yes, Notifications: &ComputerNotificationUpdate{Push: &yes}})
	if err != nil || mock.last().Method != "PUT" || mock.last().Query["project_id"] != computersProject {
		t.Fatalf("save %+v %v", saveCap, err)
	}
	body := mock.last().Body
	if cap, _ := body["spend_cap"].(map[string]any); cap["amount_minor"] != float64(500000) || cap["currency"] != "NGN" || body["default_approvals"] != true || len(body) != 3 {
		t.Fatalf("save body %v", body)
	}
	if saveCap.SpendCap == nil || saveCap.SpendCap.AmountMinor != 500000 || !saveCap.DefaultApprovals || !saveCap.Notifications.Push || !saveCap.Notifications.Email {
		t.Fatalf("saved settings %+v", saveCap)
	}
	removed, err := client.Computers.SaveSettings(ctx, computersProject, SaveComputerSettingsInput{ClearSpendCap: true, ClearDefaultIdlePolicy: true})
	if err != nil || removed.SpendCap != nil || removed.DefaultIdlePolicy != nil {
		t.Fatalf("remove cap %+v %v", removed, err)
	}
	if value, present := mock.last().Body["spend_cap"]; !present || value != nil {
		t.Fatalf("spend_cap should be an explicit null %v", mock.last().Body)
	}
	if _, err := client.Computers.SaveSettings(ctx, computersProject, SaveComputerSettingsInput{SpendCap: &ComputerMoney{AmountMinor: 1, Currency: "EUR"}}); !errors.As(err, &rejected) || rejected.Code != ComputerErrorInvalidRequest {
		t.Fatalf("bad currency %v", err)
	}
	if _, err := client.Computers.SaveSettings(ctx, computersProject, SaveComputerSettingsInput{}); err != nil || string(mock.last().Raw) != "{}" {
		t.Fatalf("empty save %q %v", mock.last().Raw, err)
	}

	prices, err := client.Computers.PublicPricing(ctx)
	if err != nil || mock.last().Headers.Get("Authorization") != "" || mock.last().Path != "/computers/public-pricing" || prices.NGN.Currency != "NGN" || prices.USD.EgressGB != 0.02 || prices.FreePlan.RunningHoursPerMonth != 4 || prices.FreePlan.MemoryBytes != 8589934592 || prices.RunningLimits[1].Computers != 25 {
		t.Fatalf("public pricing %+v %v", prices, err)
	}
	if priced, err := client.Computers.Pricing(ctx, computersProject, "USD"); err != nil || priced.EgressGB != 0.02 {
		t.Fatalf("pricing %+v %v", priced, err)
	}

	imported, err := client.Computers.ImportImage(ctx, ImportComputerImageInput{
		ProjectID: computersProject, Image: "node:22", Name: "node-22", KeepBuilder: true, ComputerName: "node-22-builder",
		Internet: &ComputerInternet{Mode: ComputerInternetNone}, IdlePolicy: &ComputerIdlePolicy{PauseAfterSeconds: 600, Mode: "disk"}, Access: &ComputerAccess{Approvals: &yes},
		Logins: []string{"github"}, MaxRunSeconds: 3600, ArchiveAfterSeconds: 86400, DeleteAfterSeconds: 604800,
	}, "import-1")
	if err != nil || mock.last().Headers.Get("Idempotency-Key") != "import-1" {
		t.Fatalf("import %+v %v", imported, err)
	}
	importBody := mock.last().Body
	internet, _ := importBody["internet"].(map[string]any)
	logins, _ := importBody["logins"].([]any)
	if internet["mode"] != "none" || len(logins) != 1 || logins[0] != "github" || importBody["delete_after_seconds"] != float64(604800) || importBody["archive_after_seconds"] != float64(86400) || importBody["max_run_seconds"] != float64(3600) || importBody["computer_name"] != "node-22-builder" {
		t.Fatalf("import body %v", importBody)
	}
	if imported.Computer.Name != "node-22-builder" || imported.Computer.Internet.Mode != ComputerInternetNone || imported.Setup.Kind != "image" || imported.Setup.Import == nil || !imported.Setup.Import.KeepBuilder || imported.Setup.Image.Reference != "node:22" {
		t.Fatalf("import result %+v", imported)
	}
	setup, err := client.Computers.Setup(ctx, imported.Computer.ID)
	if err != nil || setup.ID != imported.Setup.ID || setup.Steps[0].Name != "pull" {
		t.Fatalf("setup %+v %v", setup, err)
	}
	if _, err := client.Computers.ImportImage(ctx, ImportComputerImageInput{ProjectID: computersProject, Image: "node:22"}, ""); !errors.As(err, &rejected) || rejected.Code != ComputerErrorInvalidRequest {
		t.Fatalf("import without name %v", err)
	}
}

func TestComputersSetupCIDeleteConfirmAndKeylessPricing(t *testing.T) {
	client, mock := computersFixture(t, 0)
	ctx := context.Background()
	hour := ComputerSeconds(3600)
	created, err := client.Computers.Create(ctx, CreateComputerInput{
		ProjectID: computersProject, Name: "c", Template: "coding", Size: NewComputerSize(2, 4, 20),
		Source:        &ComputerSource{Repository: "https://github.com/acme/app", Token: "ghp_x"},
		Setup:         &ComputerSetupInput{Script: &ComputerSetupScript{Content: "npm ci", CWD: "/workspace/app"}},
		MaxRunSeconds: hour, ArchiveAfterSeconds: ComputerSeconds(86400), DeleteAfterSeconds: ComputerSeconds(0),
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	body := mock.last().Body
	source, _ := body["source"].(map[string]any)
	script, _ := body["setup"].(map[string]any)["script"].(map[string]any)
	if source["token"] != "ghp_x" || script["content"] != "npm ci" || script["cwd"] != "/workspace/app" || body["max_run_seconds"] != float64(3600) || body["archive_after_seconds"] != float64(86400) {
		t.Fatalf("create body %v", body)
	}
	if value, present := body["delete_after_seconds"]; !present || value != float64(0) {
		t.Fatalf("zero delete_after_seconds must be sent %v", body)
	}
	if _, err := client.Computers.Update(ctx, created.Computer.ID, UpdateComputerInput{DeleteAfterSeconds: ComputerSeconds(0), MaxRunSeconds: ComputerSeconds(7200)}); err != nil {
		t.Fatal(err)
	}
	if value, present := mock.last().Body["delete_after_seconds"]; !present || value != float64(0) || mock.last().Body["max_run_seconds"] != float64(7200) || len(mock.last().Body) != 2 {
		t.Fatalf("update body %v", mock.last().Body)
	}

	started, err := client.Computers.StartSetup(ctx, created.Computer.ID, ComputerSetupInput{Devcontainer: &ComputerSetupDevcontainer{WorkspaceFolder: "/workspace/app", ConfigPath: ".devcontainer.json"}}, "setup-1")
	if err != nil || started.Kind != "devcontainer" || started.State != "pending" || mock.last().Headers.Get("Idempotency-Key") != "setup-1" {
		t.Fatalf("devcontainer setup %+v %v", started, err)
	}
	devcontainer, _ := mock.last().Body["devcontainer"].(map[string]any)
	if devcontainer["workspace_folder"] != "/workspace/app" || devcontainer["config_path"] != ".devcontainer.json" || len(mock.last().Body) != 1 {
		t.Fatalf("devcontainer body %v", mock.last().Body)
	}
	sourced, err := client.Computers.StartSetup(ctx, created.Computer.ID, ComputerSetupInput{Source: &ComputerSetupSource{Repository: "acme/app", Revision: "main", Path: "setup.sh", Directory: "/workspace/app", Token: "ghp_x"}}, "")
	if err != nil || sourced.Kind != "source" || mock.last().Headers.Get("Idempotency-Key") == "" {
		t.Fatalf("source setup %+v %v", sourced, err)
	}
	if current, err := client.Computers.Setup(ctx, created.Computer.ID); err != nil || current.ID != sourced.ID {
		t.Fatalf("latest setup %+v %v", current, err)
	}
	var rejected *APIError
	if _, err := client.Computers.StartSetup(ctx, created.Computer.ID, ComputerSetupInput{}, ""); !errors.As(err, &rejected) || rejected.Code != ComputerErrorInvalidRequest {
		t.Fatalf("empty setup %v", err)
	}
	both := ComputerSetupInput{Script: &ComputerSetupScript{Content: "x"}, Source: &ComputerSetupSource{Repository: "r"}}
	if _, err := client.Computers.StartSetup(ctx, created.Computer.ID, both, ""); !errors.As(err, &rejected) || rejected.StatusCode != 400 {
		t.Fatalf("two setups %v", err)
	}

	overview, err := client.Computers.CI.Overview(ctx, computersProject)
	if err != nil || mock.last().Path != "/computers/ci" || mock.last().Query["project_id"] != computersProject {
		t.Fatalf("ci overview %+v %v", overview, err)
	}
	if len(overview.Connections) != 1 || overview.Connections[0].InstallationID != 42 || overview.Connections[0].Repositories[0].FullName != "acme/app" || !overview.Connections[0].Repositories[0].Private || overview.Installations[0].InstallationID != 43 || overview.Jobs[0].Size != "small-mem" || overview.Jobs[0].GitHubRunID != 22 {
		t.Fatalf("ci overview body %+v", overview)
	}
	labels, err := client.Computers.CI.Labels(ctx, computersProject)
	if err != nil || len(labels) != 6 || labels[1] != "stackshift-small-mem" || labels[3] != "stackshift-medium-mem" {
		t.Fatalf("ci labels %v %v", labels, err)
	}
	jobs, err := client.Computers.CI.Jobs(ctx, computersProject, 5)
	if err != nil || mock.last().Path != "/computers/ci/jobs" || mock.last().Query["limit"] != "5" || len(jobs) != 1 || jobs[0].State != "running" || jobs[0].Labels[0] != "stackshift-small-mem" {
		t.Fatalf("ci jobs %+v %v", jobs, err)
	}
	connection, err := client.Computers.CI.Connect(ctx, computersProject, 43)
	if err != nil || connection.InstallationID != 43 || mock.last().Body["installation_id"] != float64(43) || mock.last().Body["project_id"] != computersProject {
		t.Fatalf("ci connect %+v %v", connection, err)
	}
	if err := client.Computers.CI.Disconnect(ctx, connection.ID); err != nil || mock.last().Method != "DELETE" || mock.last().Path != "/computers/ci/connections/"+connection.ID {
		t.Fatalf("ci disconnect %v", err)
	}
	circle, err := client.Computers.CI.ConnectCircleCI(ctx, ComputerCircleCIConnectInput{ProjectID: computersProject, ResourceClass: "acme/stackshift", Size: "small", APIToken: "at"})
	if err != nil || circle.ID != "cc1" || circle.MaxConcurrentJobs != 2 || mock.last().Path != "/computers/ci/circleci/connections" || mock.last().Body["resource_class"] != "acme/stackshift" || mock.last().Body["api_token"] != "at" {
		t.Fatalf("circleci connect %+v %v", circle, err)
	}
	if err := client.Computers.CI.DisconnectCircleCI(ctx, circle.ID); err != nil || mock.last().Method != "DELETE" || mock.last().Path != "/computers/ci/circleci/connections/"+circle.ID {
		t.Fatalf("circleci disconnect %v", err)
	}

	snapshot, err := client.Computers.Snapshot(ctx, created.Computer.ID, "keep", "machine", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Computers.Snapshots.DeleteConfirmed(ctx, snapshot.ID, "other"); !errors.As(err, &rejected) || rejected.StatusCode != 428 || rejected.Code != "confirmation_required" {
		t.Fatalf("wrong snapshot confirm %v", err)
	}
	if err := client.Computers.Snapshots.Delete(ctx, snapshot.ID); err != nil || mock.last().Headers.Get("X-StackShift-Confirm-Delete") != snapshot.ID {
		t.Fatalf("snapshot delete %v", err)
	}
	named, err := client.Computers.Snapshot(ctx, created.Computer.ID, "by-name", "machine", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Computers.Snapshots.DeleteConfirmed(ctx, named.ID, "by-name"); err != nil || mock.last().Headers.Get("X-StackShift-Confirm-Delete") != "by-name" {
		t.Fatalf("snapshot delete by name %v", err)
	}
	if _, err := client.Computers.DeleteConfirmed(ctx, created.Computer.ID, "wrong", ""); !errors.As(err, &rejected) || rejected.StatusCode != 428 || rejected.Code != "confirmation_required" {
		t.Fatalf("wrong computer confirm %v", err)
	}
	if still, err := client.Computers.Get(ctx, created.Computer.ID); err != nil || still.State == "deleting" {
		t.Fatalf("refused delete changed state %+v %v", still, err)
	}
	deleted, err := client.Computers.Delete(ctx, created.Computer.ID, "")
	if err != nil || deleted.Computer.State != "deleting" || mock.last().Headers.Get("X-StackShift-Confirm-Delete") != created.Computer.ID {
		t.Fatalf("delete %+v %v", deleted, err)
	}
	other, err := client.Computers.Create(ctx, CreateComputerInput{ProjectID: computersProject, Name: "named", Template: "coding", Size: NewComputerSize(2, 4, 20)}, "")
	if err != nil {
		t.Fatal(err)
	}
	if byName, err := client.Computers.DeleteConfirmed(ctx, other.Computer.ID, "named", ""); err != nil || byName.Computer.State != "deleting" {
		t.Fatalf("delete by name %+v %v", byName, err)
	}
}

func TestComputerPublicPricingNeedsNoKeyAndFailsAsAPIError(t *testing.T) {
	var seen http.Header
	var failing bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Clone()
		if r.URL.Path != "/api/v1/computers/public-pricing" {
			writeFailure(w, 404, "invalid_request", "no route")
			return
		}
		if failing {
			writeFailure(w, 503, "node_unavailable", "prices are not published yet")
			return
		}
		writeJSON(w, 200, map[string]any{"ngn": map[string]any{"currency": "NGN", "vcpu_hour": 52, "egress_gb": 30}, "usd": map[string]any{"currency": "USD", "vcpu_hour": 0.035, "egress_gb": 0.02}, "free_plan": map[string]any{"computers": 1, "vcpus": 4, "memory_bytes": 8589934592, "running_hours_per_month": 4}, "running_limits": []map[string]any{{"plan": "enterprise", "computers": 100}}})
	}))
	t.Cleanup(server.Close)
	ctx := context.Background()
	prices, err := FetchComputerPublicPricing(ctx, server.URL+"/api/v1/", server.Client())
	if err != nil || prices.NGN.EgressGB != 30 || prices.USD.VCPUHour != 0.035 || prices.FreePlan.Computers != 1 || prices.RunningLimits[0].Computers != 100 {
		t.Fatalf("keyless pricing %+v %v", prices, err)
	}
	if seen.Get("Authorization") != "" {
		t.Fatalf("authorization sent: %q", seen.Get("Authorization"))
	}
	if prices, err := FetchComputerPublicPricing(ctx, server.URL+"/api/v1", nil); err != nil || prices.NGN.Currency != "NGN" {
		t.Fatalf("default http client %+v %v", prices, err)
	}
	failing = true
	var apiErr *APIError
	if _, err := FetchComputerPublicPricing(ctx, server.URL+"/api/v1", nil); !errors.As(err, &apiErr) || apiErr.StatusCode != 503 || apiErr.Code != ComputerErrorNodeUnavailable || apiErr.Message != "prices are not published yet" {
		t.Fatalf("keyless failure %v", err)
	}
	client, err := New(Options{APIKey: "csk_secret", BaseURL: server.URL + "/api/v1"})
	if err != nil {
		t.Fatal(err)
	}
	failing = false
	if prices, err := client.Computers.PublicPricing(ctx); err != nil || prices.USD.Currency != "USD" || seen.Get("Authorization") != "" {
		t.Fatalf("client pricing %+v %v auth=%q", prices, err, seen.Get("Authorization"))
	}
}
