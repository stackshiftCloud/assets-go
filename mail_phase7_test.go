package stackshift

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMailEventsTimelineAndWebhooks(t *testing.T) {
	var calls []string
	client, err := New(Options{
		APIKey: "sk_test",
		HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls = append(calls, req.Method+" "+req.URL.String())
			switch {
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/events"):
				if req.URL.Query().Get("type") != "mail.message.sent" || req.URL.Query().Get("messageId") != "msg_123" {
					t.Fatalf("unexpected event query: %s", req.URL.RawQuery)
				}
				return jsonResponse(http.StatusOK, `{"data":[`+phase7EventJSON()+`],"nextCursor":null}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/events/evt_123"):
				return jsonResponse(http.StatusOK, phase7EventJSON()), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/messages/msg_123/timeline"):
				return jsonResponse(http.StatusOK, `{"data":[{"type":"mail.message.queued","label":"Message queued","createdAt":"2026-05-09T12:00:00Z"},{"type":"mail.message.sent","label":"Accepted by StackShift outbound MTA","createdAt":"2026-05-09T12:01:00Z"}]}`), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/webhooks"):
				return jsonResponse(http.StatusCreated, phase7WebhookJSON(`,"secret":"whsec_test"`)), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/webhooks"):
				return jsonResponse(http.StatusOK, `{"data":[`+phase7WebhookJSON("")+`],"nextCursor":null}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123"):
				return jsonResponse(http.StatusOK, phase7WebhookJSON("")), nil
			case req.Method == http.MethodPatch && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123"):
				return jsonResponse(http.StatusOK, phase7WebhookJSON("")), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123/rotate-secret"):
				return jsonResponse(http.StatusOK, `{"secret":"whsec_new"}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123/deliveries"):
				if req.URL.Query().Get("status") != "failed" {
					t.Fatalf("unexpected delivery query: %s", req.URL.RawQuery)
				}
				return jsonResponse(http.StatusOK, `{"data":[`+phase7DeliveryJSON("failed")+`],"nextCursor":null}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/webhook-deliveries/whd_123"):
				return jsonResponse(http.StatusOK, phase7DeliveryDetailJSON("failed")), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/webhook-deliveries/whd_123/retry"):
				return jsonResponse(http.StatusAccepted, phase7DeliveryJSON("pending")), nil
			case req.Method == http.MethodDelete && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123"):
				return jsonResponse(http.StatusNoContent, ``), nil
			default:
				return jsonResponse(http.StatusNotFound, `{"error":{"code":"not_found","message":"Not found."}}`), nil
			}
		}),
	})
	if err != nil {
		t.Fatal(err)
	}

	events, err := client.Mail.Events.List(context.Background(), MailEventListOptions{Type: "mail.message.sent", MessageID: "msg_123", Limit: 20})
	if err != nil || events.Data[0].ID != "evt_123" {
		t.Fatalf("events=%+v err=%v", events, err)
	}
	event, err := client.Mail.Events.Get(context.Background(), "evt_123")
	if err != nil || event.Type != "mail.message.sent" {
		t.Fatalf("event=%+v err=%v", event, err)
	}
	timeline, err := client.Mail.Messages.Timeline(context.Background(), "msg_123")
	if err != nil || timeline[1].Label != "Accepted by StackShift outbound MTA" {
		t.Fatalf("timeline=%+v err=%v", timeline, err)
	}

	created, err := client.Mail.Webhooks.Create(context.Background(), MailWebhookCreateRequest{
		URL: "https://example.com/stackshift", EventTypes: []string{"mail.message.sent"},
	})
	if err != nil || created.Secret != "whsec_test" {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	hooks, err := client.Mail.Webhooks.List(context.Background(), MailWebhookListOptions{Limit: 10})
	if err != nil || hooks.Data[0].ID != "wh_123" {
		t.Fatalf("hooks=%+v err=%v", hooks, err)
	}
	hook, err := client.Mail.Webhooks.Get(context.Background(), "wh_123")
	if err != nil || hook.Status != "active" {
		t.Fatalf("hook=%+v err=%v", hook, err)
	}
	status := "disabled"
	updated, err := client.Mail.Webhooks.Update(context.Background(), "wh_123", MailWebhookUpdateRequest{Status: &status})
	if err != nil || updated.ID != "wh_123" {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	secret, err := client.Mail.Webhooks.RotateSecret(context.Background(), "wh_123")
	if err != nil || secret != "whsec_new" {
		t.Fatalf("secret=%q err=%v", secret, err)
	}
	deliveries, err := client.Mail.Webhooks.Deliveries(context.Background(), "wh_123", MailWebhookDeliveryListOptions{Status: "failed"})
	if err != nil || deliveries.Data[0].Status != "failed" {
		t.Fatalf("deliveries=%+v err=%v", deliveries, err)
	}
	delivery, err := client.Mail.Webhooks.GetDelivery(context.Background(), "whd_123")
	if err != nil || len(delivery.Attempts) != 1 {
		t.Fatalf("delivery=%+v err=%v", delivery, err)
	}
	retried, err := client.Mail.Webhooks.RetryDelivery(context.Background(), "whd_123")
	if err != nil || retried.Status != "pending" {
		t.Fatalf("retried=%+v err=%v", retried, err)
	}
	if err := client.Mail.Webhooks.Delete(context.Background(), "wh_123"); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 12 {
		t.Fatalf("unexpected calls: %#v", calls)
	}
}

func TestMailWebhookSignatureVerification(t *testing.T) {
	body := []byte(`{"id":"evt_123","type":"mail.otp.verified"}`)
	timestamp := "1778328000"
	mac := hmac.New(sha256.New, []byte("whsec_test"))
	mac.Write([]byte(timestamp + "."))
	mac.Write(body)
	header := "t=" + timestamp + ",v1=" + hex.EncodeToString(mac.Sum(nil))
	input := MailWebhookSignatureInput{
		RawBody: body, SignatureHeader: header, TimestampHeader: timestamp,
		Secret: "whsec_test", Now: time.Unix(1778328000, 0),
	}
	if !VerifyWebhookSignature(input) {
		t.Fatalf("expected valid signature")
	}
	input.SignatureHeader = "t=" + timestamp + ",v1=bad"
	if VerifyWebhookSignature(input) {
		t.Fatalf("expected invalid signature to fail")
	}
	input.SignatureHeader = header
	input.Now = time.Unix(1778328601, 0)
	if clientOK := (&MailWebhooksClient{}).VerifySignature(input); clientOK {
		t.Fatalf("expected old timestamp to fail")
	}
}

func phase7EventJSON() string {
	return `{"id":"evt_123","type":"mail.message.sent","resourceType":"message","resourceId":"msg_123","messageId":"msg_123","payload":{"messageId":"msg_123","status":"sent"},"occurredAt":"2026-05-09T12:01:00Z","createdAt":"2026-05-09T12:01:00Z"}`
}

func phase7WebhookJSON(extra string) string {
	return `{"id":"wh_123","url":"https://example.com/stackshift","description":"Production mail events","status":"active","eventTypes":["mail.message.sent"],"failureCount":0,"lastSuccessAt":null,"lastFailureAt":null,"createdAt":"2026-05-09T12:00:00Z"` + extra + `}`
}

func phase7DeliveryJSON(status string) string {
	return `{"id":"whd_123","endpointId":"wh_123","eventId":"evt_123","status":"` + status + `","attemptCount":1,"nextAttemptAt":null,"lastAttemptAt":"2026-05-09T12:01:00Z","responseStatus":500,"responseBodySnippet":"failed","errorMessage":"non-2xx response","createdAt":"2026-05-09T12:00:00Z","completedAt":null}`
}

func phase7DeliveryDetailJSON(status string) string {
	return phase7DeliveryJSON(status)[:len(phase7DeliveryJSON(status))-1] + `,"attempts":[{"id":"whda_123","attemptNumber":1,"responseStatus":500,"responseBodySnippet":"failed","errorMessage":"non-2xx response","durationMs":12,"createdAt":"2026-05-09T12:01:00Z"}]}`
}
