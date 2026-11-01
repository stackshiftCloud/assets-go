package stackshift

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

const inboxWebhookSecret = "whsec_inbox"

const inboxWebhookPayload = `{"id":"evt_1","type":"mailbox.agent_inbox.deleted","createdAt":"2026-10-30T10:00:00Z","occurredAt":"2026-10-30T10:00:00Z","workspaceId":"ws_1","data":{"mailboxId":"mbx_1","address":"task-1@agents.stackshift.cloud","labels":{"job":"j1"},"jobId":"j1","reason":"job_finished"}}`

func signInboxWebhook(body string, at time.Time, secret string) string {
	timestamp := strconv.FormatInt(at.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "." + body))
	return "t=" + timestamp + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyInboxWebhook(t *testing.T) {
	now := time.Now()
	payload := []byte(inboxWebhookPayload)
	valid := signInboxWebhook(inboxWebhookPayload, now, inboxWebhookSecret)

	cases := []struct {
		name    string
		payload []byte
		header  string
		secret  string
		want    bool
	}{
		{"valid", payload, valid, inboxWebhookSecret, true},
		{"tampered payload", []byte(strings.Replace(inboxWebhookPayload, "j1", "j2", 1)), valid, inboxWebhookSecret, false},
		{"wrong secret", payload, valid, "whsec_other", false},
		{"signed with another secret", payload, signInboxWebhook(inboxWebhookPayload, now, "whsec_other"), inboxWebhookSecret, false},
		{"stale timestamp", payload, signInboxWebhook(inboxWebhookPayload, now.Add(-time.Hour), inboxWebhookSecret), inboxWebhookSecret, false},
		{"future timestamp", payload, signInboxWebhook(inboxWebhookPayload, now.Add(time.Hour), inboxWebhookSecret), inboxWebhookSecret, false},
		{"no timestamp", payload, "v1=" + strings.SplitN(valid, "v1=", 2)[1], inboxWebhookSecret, false},
		{"empty header", payload, "", inboxWebhookSecret, false},
	}
	for _, tc := range cases {
		if got := VerifyInboxWebhook(tc.payload, tc.header, tc.secret); got != tc.want {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

func TestParseInboxWebhook(t *testing.T) {
	event, err := ParseInboxWebhook([]byte(inboxWebhookPayload))
	if err != nil || event.Type != InboxEventAgentInboxDeleted {
		t.Fatalf("event=%+v err=%v", event, err)
	}
	data, err := event.MailboxData()
	if err != nil || data.JobID != "j1" || data.Reason != "job_finished" || data.Labels["job"] != "j1" {
		t.Fatalf("data=%+v err=%v", data, err)
	}

	conversation, err := ParseInboxWebhook([]byte(`{"id":"evt_2","type":"mailbox.conversation.received","data":{"mailboxId":"mbx_1","threadId":"thr_1","conversation":{"threadId":"thr_1","subject":"Help","status":"open"},"actor":{"kind":"system"},"activity":{"id":9,"kind":"received","data":{}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	detail, err := conversation.ConversationData()
	if err != nil || detail.Conversation.Subject != "Help" || detail.Actor.Kind != "system" || detail.Activity.ID != 9 {
		t.Fatalf("detail=%+v err=%v", detail, err)
	}
}

func TestMailboxScopedWebhookAndTestEvent(t *testing.T) {
	var bodies []string
	var paths []string
	client, err := New(Options{APIKey: "sk_test", HTTPClient: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var body string
		if request.Body != nil {
			raw, _ := io.ReadAll(request.Body)
			body = string(raw)
		}
		bodies = append(bodies, body)
		paths = append(paths, request.Method+" "+request.URL.Path)
		if strings.HasSuffix(request.URL.Path, "/test") {
			return jsonResponse(http.StatusAccepted, `{"deliveryId":"whd_1","eventId":"evt_1"}`), nil
		}
		return jsonResponse(http.StatusCreated, `{"id":"wh_1","mailboxId":"mbx_1","secret":"whsec_1"}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	hook, err := client.Mail.Webhooks.Create(ctx, MailWebhookCreateRequest{URL: "https://agent.example.com/hook", MailboxID: "mbx_1", EventTypes: []string{InboxEventConversationReceived}})
	if err != nil || hook.MailboxID == nil || *hook.MailboxID != "mbx_1" || !strings.Contains(bodies[0], `"mailboxId":"mbx_1"`) {
		t.Fatalf("hook=%+v err=%v body=%s", hook, err, bodies[0])
	}
	result, err := client.Mail.Webhooks.Test(ctx, "wh_1")
	if err != nil || result.DeliveryID != "whd_1" || result.EventID != "evt_1" || paths[1] != "POST /api/v1/mail/webhooks/wh_1/test" {
		t.Fatalf("result=%+v err=%v paths=%v", result, err, paths)
	}
}
