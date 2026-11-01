package stackshift

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type agentInboxCall struct {
	method  string
	path    string
	query   url.Values
	headers http.Header
	body    string
}

type agentInboxFake struct {
	calls     []agentInboxCall
	responses map[string]string
	waits     []string
	clock     time.Time
}

func newAgentInboxClient(t *testing.T, fake *agentInboxFake) *Client {
	t.Helper()
	fake.clock = time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	client, err := New(Options{APIKey: "sk_test", HTTPClient: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var body string
		if request.Body != nil {
			raw, _ := io.ReadAll(request.Body)
			body = string(raw)
		}
		fake.calls = append(fake.calls, agentInboxCall{method: request.Method, path: request.URL.Path, query: request.URL.Query(), headers: request.Header, body: body})
		if request.Method == http.MethodDelete {
			return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}}, nil
		}
		if strings.HasSuffix(request.URL.Path, "/messages/wait") {
			seconds, _ := time.ParseDuration(request.URL.Query().Get("timeout") + "s")
			fake.clock = fake.clock.Add(seconds)
			next := `{"message":null,"timedOut":true}`
			if len(fake.waits) > 0 {
				next, fake.waits = fake.waits[0], fake.waits[1:]
			}
			return jsonResponse(http.StatusOK, next), nil
		}
		if strings.Contains(request.URL.Path, "/messages/latest") && request.URL.Query().Get("from") == "nobody" {
			return jsonResponse(http.StatusNotFound, `{"error":{"code":"no_messages","message":"No messages yet."}}`), nil
		}
		for suffix, response := range fake.responses {
			if strings.HasSuffix(request.URL.Path, suffix) {
				return jsonResponse(http.StatusOK, response), nil
			}
		}
		return jsonResponse(http.StatusOK, `{"id":"agi_1"}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	client.Mail.AgentInboxes.now = func() time.Time { return fake.clock }
	return client
}

const agentInboxMessageJSON = `{"id":"msg_1","threadId":"thr_1","from":{"email":"no-reply@paystack.com"},"subject":"Your code","codes":["482913"],"links":[]}`

func TestAgentInboxCreateAndBatch(t *testing.T) {
	fake := &agentInboxFake{}
	client := newAgentInboxClient(t, fake)
	ctx := context.Background()

	created, err := client.Mail.AgentInboxes.Create(ctx, AgentInboxCreateRequest{
		Prefix: "task", Domain: "shared", Labels: map[string]string{"job": "j1"}, JobID: "j1", ClientRef: "run-1",
		ExpiresIn: 3600, SendMode: "replies", IdempotencyKey: "run-1-inbox",
	})
	if err != nil || created.ID != "agi_1" {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	call := fake.calls[0]
	if call.method != http.MethodPost || call.path != "/api/v1/mail/agent-inboxes" || call.headers.Get("Idempotency-Key") != "run-1-inbox" || call.headers.Get("Authorization") != "Bearer sk_test" {
		t.Fatalf("call=%+v", call)
	}
	for _, want := range []string{`"prefix":"task"`, `"domain":"shared"`, `"labels":{"job":"j1"}`, `"jobId":"j1"`, `"clientRef":"run-1"`, `"expiresIn":3600`, `"sendMode":"replies"`} {
		if !strings.Contains(call.body, want) {
			t.Fatalf("body %s missing %s", call.body, want)
		}
	}
	if strings.Contains(call.body, "IdempotencyKey") || strings.Contains(call.body, "run-1-inbox") {
		t.Fatalf("idempotency key leaked into body: %s", call.body)
	}

	if _, err := client.Mail.AgentInboxes.Create(ctx, AgentInboxCreateRequest{}); err != nil {
		t.Fatal(err)
	}
	if fake.calls[1].headers.Get("Idempotency-Key") != "" || fake.calls[1].body != "{}" {
		t.Fatalf("call=%+v", fake.calls[1])
	}

	fake.responses = map[string]string{"/batch": `{"data":[{"id":"agi_1"},{"id":"agi_2"}]}`}
	batch, err := client.Mail.AgentInboxes.CreateBatch(ctx, []AgentInboxCreateRequest{{Prefix: "a", IdempotencyKey: "ignored"}, {Prefix: "b"}}, "batch-1")
	if err != nil || len(batch) != 2 {
		t.Fatalf("batch=%+v err=%v", batch, err)
	}
	call = fake.calls[2]
	if call.path != "/api/v1/mail/agent-inboxes/batch" || call.headers.Get("Idempotency-Key") != "batch-1" || call.body != `{"inboxes":[{"prefix":"a"},{"prefix":"b"}]}` {
		t.Fatalf("call=%+v", call)
	}
}

func TestAgentInboxListGetUpdateDelete(t *testing.T) {
	fake := &agentInboxFake{responses: map[string]string{"/agent-inboxes": `{"data":[{"id":"agi_1"}],"nextCursor":"cur_2"}`}}
	client := newAgentInboxClient(t, fake)
	ctx := context.Background()

	page, err := client.Mail.AgentInboxes.List(ctx, AgentInboxListOptions{
		Labels: map[string]string{"team": "qa", "job": "j1"}, JobID: "j1", Query: "task", Domain: "shared", Limit: 25, Cursor: "cur_1",
	})
	if err != nil || page.NextCursor != "cur_2" || len(page.Data) != 1 {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	query := fake.calls[0].query
	if got := strings.Join(query["label"], ","); got != "job:j1,team:qa" || query.Get("jobId") != "j1" || query.Get("q") != "task" || query.Get("domain") != "shared" || query.Get("limit") != "25" || query.Get("cursor") != "cur_1" {
		t.Fatalf("query=%v", query)
	}
	if _, err := client.Mail.AgentInboxes.List(ctx, AgentInboxListOptions{}); err != nil || len(fake.calls[1].query) != 0 {
		t.Fatalf("query=%v err=%v", fake.calls[1].query, err)
	}

	if _, err := client.Mail.AgentInboxes.Get(ctx, "agi_1"); err != nil || fake.calls[2].path != "/api/v1/mail/agent-inboxes/agi_1" {
		t.Fatalf("call=%+v err=%v", fake.calls[2], err)
	}

	mode := "any"
	labels := map[string]string{"job": "j2"}
	if _, err := client.Mail.AgentInboxes.Update(ctx, "agi_1", AgentInboxUpdateRequest{Labels: &labels, SendMode: &mode, ClearExpiry: true}); err != nil {
		t.Fatal(err)
	}
	call := fake.calls[3]
	if call.method != http.MethodPatch || call.body != `{"labels":{"job":"j2"},"clearExpiry":true,"sendMode":"any"}` {
		t.Fatalf("call=%+v", call)
	}

	if err := client.Mail.AgentInboxes.Delete(ctx, "agi_1"); err != nil {
		t.Fatal(err)
	}
	if fake.calls[4].method != http.MethodDelete || fake.calls[4].path != "/api/v1/mail/agent-inboxes/agi_1" {
		t.Fatalf("call=%+v", fake.calls[4])
	}
}

func TestAgentInboxDeleteManyAndFinishJob(t *testing.T) {
	fake := &agentInboxFake{responses: map[string]string{"/delete": `{"deleted":3}`}}
	client := newAgentInboxClient(t, fake)
	ctx := context.Background()

	deleted, err := client.Mail.AgentInboxes.DeleteMany(ctx, AgentInboxDeleteRequest{IDs: []string{"agi_1", "agi_2"}})
	if err != nil || deleted != 3 {
		t.Fatalf("deleted=%d err=%v", deleted, err)
	}
	if fake.calls[0].method != http.MethodPost || fake.calls[0].path != "/api/v1/mail/agent-inboxes/delete" || fake.calls[0].body != `{"ids":["agi_1","agi_2"]}` {
		t.Fatalf("call=%+v", fake.calls[0])
	}
	if _, err := client.Mail.AgentInboxes.DeleteMany(ctx, AgentInboxDeleteRequest{Labels: map[string]string{"team": "qa"}}); err != nil || fake.calls[1].body != `{"labels":{"team":"qa"}}` {
		t.Fatalf("call=%+v err=%v", fake.calls[1], err)
	}
	if _, err := client.Mail.AgentInboxes.FinishJob(ctx, "j1"); err != nil || fake.calls[2].body != `{"jobId":"j1"}` {
		t.Fatalf("call=%+v err=%v", fake.calls[2], err)
	}
	if _, err := client.Mail.AgentInboxes.DeleteMany(ctx, AgentInboxDeleteRequest{}); err == nil {
		t.Fatal("empty request should be refused")
	}
	if _, err := client.Mail.AgentInboxes.DeleteMany(ctx, AgentInboxDeleteRequest{IDs: []string{"a"}, JobID: "j"}); err == nil {
		t.Fatal("two selectors should be refused")
	}
	if len(fake.calls) != 3 {
		t.Fatalf("refused requests were sent: %d calls", len(fake.calls))
	}
}

func TestAgentInboxMessages(t *testing.T) {
	fake := &agentInboxFake{responses: map[string]string{
		"/messages":        `{"data":[` + agentInboxMessageJSON + `],"nextCursor":"m_2"}`,
		"/messages/latest": agentInboxMessageJSON,
		"/messages/msg_1":  agentInboxMessageJSON,
	}}
	client := newAgentInboxClient(t, fake)
	ctx := context.Background()

	page, err := client.Mail.AgentInboxes.Messages.List(ctx, "agi_1", AgentInboxMessageListOptions{After: "m_1", Limit: 10})
	if err != nil || page.Data[0].Codes[0] != "482913" || page.NextCursor != "m_2" {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	if fake.calls[0].path != "/api/v1/mail/agent-inboxes/agi_1/messages" || fake.calls[0].query.Get("after") != "m_1" || fake.calls[0].query.Get("limit") != "10" {
		t.Fatalf("call=%+v", fake.calls[0])
	}
	latest, err := client.Mail.AgentInboxes.Messages.Latest(ctx, "agi_1", AgentInboxMessageFilter{From: "paystack.com", Subject: "code"})
	if err != nil || latest.ID != "msg_1" || fake.calls[1].query.Get("from") != "paystack.com" || fake.calls[1].query.Get("subject") != "code" {
		t.Fatalf("latest=%+v err=%v call=%+v", latest, err, fake.calls[1])
	}
	_, err = client.Mail.AgentInboxes.Messages.Latest(ctx, "agi_1", AgentInboxMessageFilter{From: "nobody"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound || apiErr.Code != "no_messages" {
		t.Fatalf("err=%v", err)
	}
	message, err := client.Mail.AgentInboxes.Messages.Get(ctx, "agi_1", "msg_1")
	if err != nil || message.ID != "msg_1" || fake.calls[3].path != "/api/v1/mail/agent-inboxes/agi_1/messages/msg_1" {
		t.Fatalf("message=%+v err=%v", message, err)
	}
}

func TestAgentInboxWaitForMessage(t *testing.T) {
	fake := &agentInboxFake{waits: []string{`{"message":` + agentInboxMessageJSON + `,"timedOut":false}`}}
	client := newAgentInboxClient(t, fake)
	ctx := context.Background()
	since := time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)

	found, err := client.Mail.AgentInboxes.WaitForMessage(ctx, "agi_1", AgentInboxWaitOptions{From: "paystack.com", Subject: "code", Timeout: 120 * time.Second, Since: since})
	if err != nil || found == nil || found.Codes[0] != "482913" {
		t.Fatalf("found=%+v err=%v", found, err)
	}
	query := fake.calls[0].query
	if len(fake.calls) != 1 || fake.calls[0].method != http.MethodGet || fake.calls[0].path != "/api/v1/mail/agent-inboxes/agi_1/messages/wait" ||
		query.Get("since") != "2026-10-30T09:00:00Z" || query.Get("from") != "paystack.com" || query.Get("subject") != "code" || query.Get("timeout") != "50" {
		t.Fatalf("calls=%+v", fake.calls)
	}

	fake.calls = nil
	fake.waits = []string{`{"message":null,"timedOut":true}`, `{"message":null,"timedOut":true}`, `{"message":` + agentInboxMessageJSON + `,"timedOut":false}`}
	found, err = client.Mail.AgentInboxes.WaitForMessage(ctx, "agi_1", AgentInboxWaitOptions{Timeout: 200 * time.Second})
	if err != nil || found == nil || found.ID != "msg_1" || len(fake.calls) != 3 {
		t.Fatalf("found=%+v err=%v calls=%d", found, err, len(fake.calls))
	}
	for _, call := range fake.calls {
		if call.query.Get("timeout") != "50" || call.query.Get("since") != fake.calls[0].query.Get("since") {
			t.Fatalf("call=%+v", call)
		}
	}
}

func TestAgentInboxWaitForMessageTimesOut(t *testing.T) {
	fake := &agentInboxFake{}
	client := newAgentInboxClient(t, fake)

	found, err := client.Mail.AgentInboxes.WaitForMessage(context.Background(), "agi_1", AgentInboxWaitOptions{Timeout: 120 * time.Second})
	if err != nil || found != nil {
		t.Fatalf("found=%+v err=%v", found, err)
	}
	var timeouts []string
	for _, call := range fake.calls {
		timeouts = append(timeouts, call.query.Get("timeout"))
	}
	if strings.Join(timeouts, ",") != "50,50,20" {
		t.Fatalf("timeouts=%v", timeouts)
	}
}

func TestAgentInboxSendReturnsEveryStatus(t *testing.T) {
	fake := &agentInboxFake{responses: map[string]string{"/send": `{"status":"waiting_approval","pendingReplyId":"pr_1","reasons":["New contact"]}`}}
	client := newAgentInboxClient(t, fake)
	ctx := context.Background()

	waiting, err := client.Mail.AgentInboxes.Send(ctx, "agi_1", AgentInboxSendRequest{To: []string{"ada@example.com"}, Subject: "Hi", Text: "Hello", ReplyToMessageID: "msg_1"})
	if err != nil || waiting.Status != "waiting_approval" || waiting.PendingReplyID != "pr_1" || len(waiting.Reasons) != 1 {
		t.Fatalf("waiting=%+v err=%v", waiting, err)
	}
	call := fake.calls[0]
	if call.method != http.MethodPost || call.path != "/api/v1/mail/agent-inboxes/agi_1/send" || call.body != `{"to":["ada@example.com"],"subject":"Hi","text":"Hello","replyToMessageId":"msg_1"}` {
		t.Fatalf("call=%+v", call)
	}

	fake.responses["/send"] = `{"status":"draft","reasons":["draft only"]}`
	draft, err := client.Mail.AgentInboxes.Send(ctx, "agi_1", AgentInboxSendRequest{To: []string{"ada@example.com"}, Subject: "Hi", HTML: "<p>Hi</p>"})
	if err != nil || draft.Status != "draft" {
		t.Fatalf("draft=%+v err=%v", draft, err)
	}
	fake.responses["/send"] = `{"status":"sent","messageId":"out_1"}`
	sent, err := client.Mail.AgentInboxes.Send(ctx, "agi_1", AgentInboxSendRequest{To: []string{"ada@example.com"}, Subject: "Hi", Text: "Hello"})
	if err != nil || sent.Status != "sent" || sent.MessageID != "out_1" {
		t.Fatalf("sent=%+v err=%v", sent, err)
	}
}

func TestAgentInboxLimitsAndPricing(t *testing.T) {
	fake := &agentInboxFake{responses: map[string]string{
		"/limits":  `{"trustTier":"new","included":3,"live":1,"sendModes":["off","replies"]}`,
		"/pricing": `{"includedByPlan":{"free":1},"bands":[{"from":4,"to":10,"monthlyKobo":100,"dailyKobo":4}],"quotaBytes":1,"sharedDomain":"agents.stackshift.cloud"}`,
	}}
	client := newAgentInboxClient(t, fake)
	ctx := context.Background()

	limits, err := client.Mail.AgentInboxes.Limits(ctx)
	if err != nil || limits.TrustTier != "new" || limits.Included != 3 || fake.calls[0].path != "/api/v1/mail/agent-inboxes/limits" {
		t.Fatalf("limits=%+v err=%v", limits, err)
	}
	pricing, err := client.Mail.AgentInboxes.Pricing(ctx)
	if err != nil || pricing.SharedDomain != "agents.stackshift.cloud" || pricing.Bands[0].MonthlyKobo != 100 || fake.calls[1].path != "/api/v1/mail/agent-inboxes/pricing" {
		t.Fatalf("pricing=%+v err=%v", pricing, err)
	}
}
