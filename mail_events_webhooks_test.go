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
				if req.URL.Query().Get("type") != "mail.message.mta_accepted" || req.URL.Query().Get("messageId") != "msg_123" {
					t.Fatalf("unexpected event query: %s", req.URL.RawQuery)
				}
				return jsonResponse(http.StatusOK, `{"data":[`+eventsWebhookEventJSON()+`],"nextCursor":null}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/events/evt_123"):
				return jsonResponse(http.StatusOK, eventsWebhookEventJSON()), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/messages/msg_123/timeline"):
				return jsonResponse(http.StatusOK, `{"data":[{"type":"mail.message.queued","label":"Message queued","createdAt":"2026-05-09T12:00:00Z"},{"type":"mail.message.mta_accepted","label":"Accepted by StackShift outbound MTA","createdAt":"2026-05-09T12:01:00Z"}]}`), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/webhooks"):
				return jsonResponse(http.StatusCreated, eventsWebhookWebhookJSON(`,"secret":"whsec_test"`)), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/webhooks"):
				return jsonResponse(http.StatusOK, `{"data":[`+eventsWebhookWebhookJSON("")+`],"nextCursor":null}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123"):
				return jsonResponse(http.StatusOK, eventsWebhookWebhookJSON("")), nil
			case req.Method == http.MethodPatch && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123"):
				return jsonResponse(http.StatusOK, eventsWebhookWebhookJSON("")), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123/rotate-secret"):
				return jsonResponse(http.StatusOK, `{"secret":"whsec_new"}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123/deliveries"):
				if req.URL.Query().Get("status") != "failed" {
					t.Fatalf("unexpected delivery query: %s", req.URL.RawQuery)
				}
				return jsonResponse(http.StatusOK, `{"data":[`+eventsWebhookDeliveryJSON("failed")+`],"nextCursor":null}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/webhook-deliveries/whd_123"):
				return jsonResponse(http.StatusOK, eventsWebhookDeliveryDetailJSON("failed")), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/webhook-deliveries/whd_123/retry"):
				return jsonResponse(http.StatusAccepted, eventsWebhookDeliveryJSON("pending")), nil
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

	events, err := client.Mail.Events.List(context.Background(), MailEventListOptions{Type: "mail.message.mta_accepted", MessageID: "msg_123", Limit: 20})
	if err != nil || events.Data[0].ID != "evt_123" {
		t.Fatalf("events=%+v err=%v", events, err)
	}
	event, err := client.Mail.Events.Get(context.Background(), "evt_123")
	if err != nil || event.Type != "mail.message.mta_accepted" {
		t.Fatalf("event=%+v err=%v", event, err)
	}
	timeline, err := client.Mail.Messages.Timeline(context.Background(), "msg_123")
	if err != nil || timeline[1].Label != "Accepted by StackShift outbound MTA" {
		t.Fatalf("timeline=%+v err=%v", timeline, err)
	}

	created, err := client.Mail.Webhooks.Create(context.Background(), MailWebhookCreateRequest{
		URL: "https://example.com/stackshift", EventTypes: []string{"mail.message.mta_accepted"},
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

func eventsWebhookEventJSON() string {
	return `{"id":"evt_123","type":"mail.message.mta_accepted","resourceType":"message","resourceId":"msg_123","messageId":"msg_123","payload":{"messageId":"msg_123","status":"mta_accepted"},"occurredAt":"2026-05-09T12:01:00Z","createdAt":"2026-05-09T12:01:00Z"}`
}

func eventsWebhookWebhookJSON(extra string) string {
	return `{"id":"wh_123","url":"https://example.com/stackshift","description":"Production mail events","status":"active","eventTypes":["mail.message.mta_accepted"],"failureCount":0,"lastSuccessAt":null,"lastFailureAt":null,"createdAt":"2026-05-09T12:00:00Z"` + extra + `}`
}

func eventsWebhookDeliveryJSON(status string) string {
	return `{"id":"whd_123","endpointId":"wh_123","eventId":"evt_123","status":"` + status + `","attemptCount":1,"nextAttemptAt":null,"lastAttemptAt":"2026-05-09T12:01:00Z","responseStatus":500,"responseBodySnippet":"failed","errorMessage":"non-2xx response","createdAt":"2026-05-09T12:00:00Z","completedAt":null}`
}

func eventsWebhookDeliveryDetailJSON(status string) string {
	return eventsWebhookDeliveryJSON(status)[:len(eventsWebhookDeliveryJSON(status))-1] + `,"attempts":[{"id":"whda_123","attemptNumber":1,"responseStatus":500,"responseBodySnippet":"failed","errorMessage":"non-2xx response","durationMs":12,"createdAt":"2026-05-09T12:01:00Z"}]}`
}

func TestMailWebhookInsights(t *testing.T) {
	client, err := New(Options{
		APIKey: "sk_test",
		HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch {
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123"):
				return jsonResponse(http.StatusOK, eventsWebhookWebhookJSON(`,"waitingSummary":{"count":4,"capped":false,"oldestAt":"2026-05-09T11:00:00Z","merchantUpdates":2,"highlights":[{"type":"partner.brand.paused","shop":"Merchant Shop"}]}`)), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123/deliveries"):
				return jsonResponse(http.StatusOK, `{"data":[`+eventsWebhookDeliveryJSON("failed")[:len(eventsWebhookDeliveryJSON("failed"))-1]+`,"durationMs":840,"merchant":{"partnerDomainId":"pd_1","merchantRef":"m_1","shop":"Merchant Shop"}}],"nextCursor":null}`), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123/reveal-secret"):
				return jsonResponse(http.StatusOK, `{"secret":"whsec_revealed"}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123/diagnosis"):
				return jsonResponse(http.StatusOK, `{"diagnosis":null}`), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/webhooks/wh_123/diagnosis"):
				return jsonResponse(http.StatusOK, `{"diagnosis":{"checkedAt":"2026-05-09T12:00:00Z","repliesRead":3,"summary":"The endpoint answers 500 to every delivery."}}`), nil
			default:
				return jsonResponse(http.StatusNotFound, `{"error":{"code":"not_found","message":"Not found."}}`), nil
			}
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	hook, err := client.Mail.Webhooks.Get(ctx, "wh_123")
	if err != nil || hook.WaitingSummary == nil || hook.WaitingSummary.Count != 4 || hook.WaitingSummary.MerchantUpdates != 2 || hook.WaitingSummary.Highlights[0].Type != "partner.brand.paused" {
		t.Fatalf("hook=%+v err=%v", hook, err)
	}
	deliveries, err := client.Mail.Webhooks.Deliveries(ctx, "wh_123", MailWebhookDeliveryListOptions{})
	if err != nil || deliveries.Data[0].DurationMs == nil || *deliveries.Data[0].DurationMs != 840 || deliveries.Data[0].Merchant == nil || deliveries.Data[0].Merchant.Shop != "Merchant Shop" {
		t.Fatalf("deliveries=%+v err=%v", deliveries, err)
	}
	secret, err := client.Mail.Webhooks.RevealSecret(ctx, "wh_123")
	if err != nil || secret != "whsec_revealed" {
		t.Fatalf("secret=%q err=%v", secret, err)
	}
	stored, err := client.Mail.Webhooks.Diagnosis(ctx, "wh_123")
	if err != nil || stored != nil {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
	diagnosis, err := client.Mail.Webhooks.Diagnose(ctx, "wh_123")
	if err != nil || diagnosis.RepliesRead != 3 || diagnosis.Summary != "The endpoint answers 500 to every delivery." {
		t.Fatalf("diagnosis=%+v err=%v", diagnosis, err)
	}
}
