package stackshift

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMailAdvancedClients(t *testing.T) {
	var calls []string
	client, err := New(Options{
		APIKey: "sk_test",
		HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls = append(calls, req.Method+" "+req.URL.String())
			switch {
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/scheduled"):
				return jsonResponse(http.StatusCreated, `{"id":"sch_123","status":"scheduled","sendAt":"2026-05-11T10:00:00Z"}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/scheduled"):
				return jsonResponse(http.StatusOK, `{"data":[{"id":"sch_123","status":"scheduled","sendAt":"2026-05-11T10:00:00Z"}]}`), nil
			case strings.HasSuffix(req.URL.Path, "/mail/scheduled/sch_123/cancel"):
				return jsonResponse(http.StatusOK, `{"id":"sch_123","status":"canceled","sendAt":"2026-05-11T10:00:00Z"}`), nil
			case strings.HasSuffix(req.URL.Path, "/mail/scheduled/sch_123"):
				return jsonResponse(http.StatusOK, `{"id":"sch_123","status":"scheduled","sendAt":"2026-05-11T10:00:00Z"}`), nil
			case strings.HasSuffix(req.URL.Path, "/mail/batch-template"):
				return jsonResponse(http.StatusAccepted, `{"id":"batch_tpl_123","status":"queued","messageCount":1}`), nil
			case strings.HasSuffix(req.URL.Path, "/mail/batch"):
				return jsonResponse(http.StatusAccepted, `{"id":"batch_123","status":"queued","messageCount":1}`), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/inbound/domains"):
				return jsonResponse(http.StatusCreated, `{"id":"ind_123","domain":"inbound.example.com","status":"pending"}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/inbound/domains"):
				return jsonResponse(http.StatusOK, `{"data":[{"id":"ind_123","domain":"inbound.example.com","status":"verified"}]}`), nil
			case strings.HasSuffix(req.URL.Path, "/mail/inbound/domains/ind_123/verify"):
				return jsonResponse(http.StatusOK, `{"id":"ind_123","domain":"inbound.example.com","status":"verified"}`), nil
			case strings.HasSuffix(req.URL.Path, "/mail/inbound/messages/inm_123"):
				return jsonResponse(http.StatusOK, `{"id":"inm_123","fromEmail":"ada@example.com","toEmails":["support@example.com"]}`), nil
			case strings.HasSuffix(req.URL.Path, "/mail/inbound/messages"):
				return jsonResponse(http.StatusOK, `{"data":[{"id":"inm_123","fromEmail":"ada@example.com","toEmails":["support@example.com"]}]}`), nil
			case strings.HasSuffix(req.URL.Path, "/mail/analytics"):
				return jsonResponse(http.StatusOK, `{"summary":{"sent":1},"series":[{"bucket":"2026-05-11","sent":1}]}`), nil
			default:
				return jsonResponse(http.StatusNotFound, `{"error":{"code":"not_found","message":"Not found."}}`), nil
			}
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	sendAt := time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC)
	scheduled, err := client.Mail.Scheduled.Create(context.Background(), MailScheduledCreateRequest{
		SendAt:  sendAt,
		Message: &MailSendRequest{From: "StackShift <noreply@mail.stackshift.cloud>", To: "ada@example.com", Subject: "Hi", Text: "Hi"},
	})
	if err != nil || scheduled.ID != "sch_123" {
		t.Fatalf("scheduled=%+v err=%v", scheduled, err)
	}
	if list, err := client.Mail.Scheduled.List(context.Background()); err != nil || len(list.Data) != 1 {
		t.Fatalf("scheduled list=%+v err=%v", list, err)
	}
	if got, err := client.Mail.Scheduled.Get(context.Background(), "sch_123"); err != nil || got.Status != "scheduled" {
		t.Fatalf("scheduled get=%+v err=%v", got, err)
	}
	if got, err := client.Mail.Scheduled.Cancel(context.Background(), "sch_123"); err != nil || got.Status != "canceled" {
		t.Fatalf("scheduled cancel=%+v err=%v", got, err)
	}
	if batch, err := client.Mail.Batch.Send(context.Background(), MailBatchSendRequest{From: "StackShift <noreply@mail.stackshift.cloud>", Messages: []MailSendRequest{{To: "ada@example.com", Subject: "Hi", Text: "Hi"}}}); err != nil || batch.MessageCount != 1 {
		t.Fatalf("batch=%+v err=%v", batch, err)
	}
	if batch, err := client.Mail.Batch.SendTemplate(context.Background(), MailBatchTemplateRequest{Template: "welcome", From: "StackShift <noreply@mail.stackshift.cloud>", Recipients: []MailBatchTemplateRecipient{{To: "ada@example.com"}}}); err != nil || batch.ID != "batch_tpl_123" {
		t.Fatalf("template batch=%+v err=%v", batch, err)
	}
	if domain, err := client.Mail.Inbound.Domains.Create(context.Background(), "inbound.example.com"); err != nil || domain.ID != "ind_123" {
		t.Fatalf("inbound domain=%+v err=%v", domain, err)
	}
	if domains, err := client.Mail.Inbound.Domains.List(context.Background()); err != nil || domains.Data[0].Status != "verified" {
		t.Fatalf("inbound domains=%+v err=%v", domains, err)
	}
	if domain, err := client.Mail.Inbound.Domains.Verify(context.Background(), "ind_123"); err != nil || domain.Status != "verified" {
		t.Fatalf("inbound verify=%+v err=%v", domain, err)
	}
	if messages, err := client.Mail.Inbound.Messages.List(context.Background()); err != nil || messages.Data[0].ID != "inm_123" {
		t.Fatalf("inbound messages=%+v err=%v", messages, err)
	}
	if message, err := client.Mail.Inbound.Messages.Get(context.Background(), "inm_123"); err != nil || message.FromEmail != "ada@example.com" {
		t.Fatalf("inbound message=%+v err=%v", message, err)
	}
	analytics, err := client.Mail.Analytics.Get(context.Background(), MailAnalyticsOptions{Range: "7d", Interval: "day", Domain: "example.com"})
	if err != nil || analytics.Summary["sent"] != 1 {
		t.Fatalf("analytics=%+v err=%v", analytics, err)
	}
	if !strings.Contains(calls[len(calls)-1], "domain=example.com") {
		t.Fatalf("analytics query not sent: %s", calls[len(calls)-1])
	}
}
