package stackshift

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestMailSendUsesPublicMailBaseURL(t *testing.T) {
	var seen *http.Request
	client, err := New(Options{
		APIKey: "sk_test",
		HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			seen = req
			return jsonResponse(http.StatusAccepted, `{"id":"msg_123","status":"queued","idempotencyStatus":"created"}`), nil
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Mail.Send(context.Background(), MailSendRequest{
		From: "StackShift <noreply@mail.stackshift.cloud>",
		To:   "ada@example.com", Subject: "Welcome", Text: "Welcome",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != "msg_123" || resp.IdempotencyStatus != "created" {
		t.Fatalf("unexpected send response: %+v", resp)
	}
	if seen.URL.String() != "https://api.stackshift.cloud/v1/mail/send" {
		t.Fatalf("unexpected URL: %s", seen.URL.String())
	}
	if seen.Header.Get("Authorization") != "Bearer sk_test" {
		t.Fatalf("missing auth header")
	}
}

func TestMailMessagesAndPublicErrors(t *testing.T) {
	client, err := New(Options{
		APIKey: "sk_test",
		HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch {
			case strings.Contains(req.URL.Path, "/bounces"):
				return jsonResponse(http.StatusOK, `{"data":[{"id":"bounce_123","messageId":"msg_123","recipientEmail":"bad@example.com","bounceType":"hard","reason":"mailbox_not_found","smtpStatusCode":"550","diagnosticCode":"smtp; 550 user unknown","receivedAt":"2026-05-08T00:00:00Z"}]}`), nil
			case strings.Contains(req.URL.Path, "/attempts"):
				return jsonResponse(http.StatusOK, `{"data":[{"id":"attempt_123","attemptNumber":1,"status":"accepted_by_mta","startedAt":"2026-05-08T00:00:00Z"}]}`), nil
			case strings.Contains(req.URL.Path, "/logs"):
				return jsonResponse(http.StatusOK, `{"data":[{"level":"info","event":"mail.message.queued","message":"queued","createdAt":"2026-05-08T00:00:00Z"}]}`), nil
			case strings.HasSuffix(req.URL.Path, "/msg_123"):
				return jsonResponse(http.StatusOK, `{"id":"msg_123","from":{"email":"noreply@mail.stackshift.cloud"},"to":["ada@example.com"],"cc":[],"bcc":[],"subject":"Welcome","status":"mta_accepted","createdAt":"2026-05-08T00:00:00Z"}`), nil
			case strings.HasSuffix(req.URL.Path, "/mail/messages"):
				return jsonResponse(http.StatusOK, `{"data":[{"id":"msg_123","from":"StackShift <noreply@mail.stackshift.cloud>","to":["ada@example.com"],"subject":"Welcome","status":"mta_accepted","createdAt":"2026-05-08T00:00:00Z"}]}`), nil
			default:
				return jsonResponse(http.StatusBadRequest, `{"error":{"code":"invalid_recipient","message":"Invalid recipient."}}`), nil
			}
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	list, err := client.Mail.Messages.List(context.Background(), MailMessageListOptions{Status: "mta_accepted", Limit: 20})
	if err != nil || len(list.Data) != 1 {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	message, err := client.Mail.Messages.Get(context.Background(), "msg_123")
	if err != nil || message.Status != "mta_accepted" {
		t.Fatalf("message=%+v err=%v", message, err)
	}
	attempts, err := client.Mail.Messages.Attempts(context.Background(), "msg_123")
	if err != nil || attempts[0].Status != "accepted_by_mta" {
		t.Fatalf("attempts=%+v err=%v", attempts, err)
	}
	logs, err := client.Mail.Messages.Logs(context.Background(), "msg_123")
	if err != nil || logs[0].Event != "mail.message.queued" {
		t.Fatalf("logs=%+v err=%v", logs, err)
	}
	bounces, err := client.Mail.Messages.Bounces(context.Background(), "msg_123")
	if err != nil || bounces[0].BounceType != "hard" {
		t.Fatalf("bounces=%+v err=%v", bounces, err)
	}
	_, err = client.Mail.Messages.Get(context.Background(), "bad")
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != "invalid_recipient" || apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected public API error, got %#v", err)
	}
}

func TestMailSuppressionsAndBounces(t *testing.T) {
	var calls []string
	client, err := New(Options{
		APIKey: "sk_test",
		HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls = append(calls, req.Method+" "+req.URL.String())
			switch {
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/suppressions"):
				return jsonResponse(http.StatusCreated, suppressionJSON("manual", "user")), nil
			case req.Method == http.MethodDelete && strings.HasSuffix(req.URL.Path, "/mail/suppressions/sup_123"):
				return jsonResponse(http.StatusNoContent, ``), nil
			case strings.HasSuffix(req.URL.Path, "/mail/suppressions"):
				return jsonResponse(http.StatusOK, `{"data":[`+suppressionJSON("hard_bounce", "bounce_processor")+`],"nextCursor":null}`), nil
			case strings.HasSuffix(req.URL.Path, "/mail/bounces"):
				return jsonResponse(http.StatusOK, `{"data":[`+bounceJSON()+`],"nextCursor":null}`), nil
			default:
				return jsonResponse(http.StatusNotFound, `{"error":{"code":"not_found","message":"Not found."}}`), nil
			}
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	suppressions, err := client.Mail.Suppressions.List(context.Background(), MailSuppressionListOptions{Email: "bad@example.com"})
	if err != nil || suppressions.Data[0].Reason != "hard_bounce" {
		t.Fatalf("suppressions=%+v err=%v", suppressions, err)
	}
	created, err := client.Mail.Suppressions.Create(context.Background(), MailSuppressionCreateRequest{Email: "bad@example.com", Reason: "manual"})
	if err != nil || created.Source != "user" {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	if err := client.Mail.Suppressions.Delete(context.Background(), "sup_123"); err != nil {
		t.Fatal(err)
	}
	bounces, err := client.Mail.Bounces.List(context.Background(), MailBounceListOptions{Type: "hard", MessageID: "msg_123"})
	if err != nil || bounces.Data[0].Reason != "mailbox_not_found" {
		t.Fatalf("bounces=%+v err=%v", bounces, err)
	}
	if !strings.Contains(calls[0], "email=bad%40example.com") || !strings.Contains(calls[3], "messageId=msg_123") {
		t.Fatalf("unexpected calls: %#v", calls)
	}
}

func TestMailDomains(t *testing.T) {
	var paths []string
	client, err := New(Options{
		APIKey: "sk_test",
		HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			paths = append(paths, req.Method+" "+req.URL.Path)
			switch {
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/domains"):
				return jsonResponse(http.StatusCreated, domainJSON("pending")), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/domains"):
				return jsonResponse(http.StatusOK, `{"data":[`+domainJSON("verified")+`],"nextCursor":null}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/domains/domain_123"):
				return jsonResponse(http.StatusOK, domainJSON("pending")), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/domains/domain_123/verify"):
				return jsonResponse(http.StatusOK, domainJSON("verified")), nil
			case req.Method == http.MethodDelete && strings.HasSuffix(req.URL.Path, "/mail/domains/domain_123"):
				return jsonResponse(http.StatusNoContent, ``), nil
			default:
				return jsonResponse(http.StatusNotFound, `{"error":{"code":"domain_not_found","message":"Mail domain not found."}}`), nil
			}
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	created, err := client.Mail.Domains.Create(context.Background(), "acme.com")
	if err != nil || created.Domain != "acme.com" || created.Records[0].Type != "TXT" {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	list, err := client.Mail.Domains.List(context.Background())
	if err != nil || list.Data[0].Status != "verified" {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	got, err := client.Mail.Domains.Get(context.Background(), "domain_123")
	if err != nil || got.ID != "domain_123" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	verified, err := client.Mail.Domains.Verify(context.Background(), "domain_123")
	if err != nil || verified.Status != "verified" {
		t.Fatalf("verified=%+v err=%v", verified, err)
	}
	if err := client.Mail.Domains.Delete(context.Background(), "domain_123"); err != nil {
		t.Fatal(err)
	}
	if len(paths) != 5 {
		t.Fatalf("unexpected calls: %#v", paths)
	}
}

func TestMailOTP(t *testing.T) {
	var calls []string
	client, err := New(Options{
		APIKey: "sk_test",
		HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls = append(calls, req.Method+" "+req.URL.String())
			switch {
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/otp/send"):
				return jsonResponse(http.StatusAccepted, otpSendJSON("pending")), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/otp/verify"):
				return jsonResponse(http.StatusOK, `{"verified":true,"challengeId":"otp_123","status":"verified"}`), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/otp/challenges/otp_123/cancel"):
				return jsonResponse(http.StatusOK, otpChallengeJSON("canceled")), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/otp/challenges/otp_123"):
				return jsonResponse(http.StatusOK, otpChallengeJSON("pending")), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/otp/challenges"):
				return jsonResponse(http.StatusOK, `{"data":[`+otpChallengeJSON("verified")+`],"nextCursor":null}`), nil
			default:
				return jsonResponse(http.StatusNotFound, `{"error":{"code":"otp_not_found","message":"OTP challenge not found."}}`), nil
			}
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	sent, err := client.Mail.OTP.Send(context.Background(), MailOTPSendRequest{
		To: "ada@example.com", Purpose: "login", IdempotencyKey: "otp:login:ada@example.com",
	})
	if err != nil || sent.ID != "otp_123" {
		t.Fatalf("sent=%+v err=%v", sent, err)
	}
	verified, err := client.Mail.OTP.Verify(context.Background(), MailOTPVerifyRequest{
		To: "ada@example.com", Purpose: "login", Code: "123456",
	})
	if err != nil || !verified.Verified {
		t.Fatalf("verified=%+v err=%v", verified, err)
	}
	list, err := client.Mail.OTP.Challenges.List(context.Background(), MailOTPChallengeListOptions{Purpose: "login", Status: "verified"})
	if err != nil || list.Data[0].Status != "verified" {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	got, err := client.Mail.OTP.Challenges.Get(context.Background(), "otp_123")
	if err != nil || got.To != "ada@example.com" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	canceled, err := client.Mail.OTP.Challenges.Cancel(context.Background(), "otp_123")
	if err != nil || canceled.Status != "canceled" {
		t.Fatalf("canceled=%+v err=%v", canceled, err)
	}
	if !strings.Contains(calls[2], "purpose=login") || !strings.Contains(calls[2], "status=verified") {
		t.Fatalf("unexpected OTP calls: %#v", calls)
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func domainJSON(status string) string {
	return `{"id":"domain_123","domain":"acme.com","status":"` + status + `","spfStatus":"verified","dkimStatus":"verified","dmarcStatus":"missing","returnPathStatus":"verified","createdAt":"2026-05-08T00:00:00Z","verifiedAt":null,"records":[{"type":"TXT","name":"acme.com","value":"v=spf1 include:_spf.stackshift.cloud ~all","status":"verified","required":true}]}`
}

func suppressionJSON(reason, source string) string {
	return `{"id":"sup_123","email":"bad@example.com","reason":"` + reason + `","source":"` + source + `","messageId":"msg_123","createdAt":"2026-05-08T00:00:00Z","expiresAt":null}`
}

func bounceJSON() string {
	return `{"id":"bounce_123","messageId":"msg_123","recipientEmail":"bad@example.com","bounceType":"hard","reason":"mailbox_not_found","smtpStatusCode":"550","diagnosticCode":"smtp; 550 user unknown","receivedAt":"2026-05-08T00:00:00Z"}`
}

func otpSendJSON(status string) string {
	return `{"id":"otp_123","to":"ada@example.com","purpose":"login","status":"` + status + `","expiresAt":"2026-05-08T00:10:00Z","attemptsRemaining":5,"resendAvailableAt":"2026-05-08T00:01:00Z","messageId":"msg_123","idempotencyStatus":"created"}`
}

func otpChallengeJSON(status string) string {
	return `{"id":"otp_123","to":"ada@example.com","purpose":"login","status":"` + status + `","attempts":1,"maxAttempts":5,"attemptsRemaining":4,"expiresAt":"2026-05-08T00:10:00Z","resendAvailableAt":"2026-05-08T00:01:00Z","messageId":"msg_123","metadata":{"userId":"user_123"},"createdAt":"2026-05-08T00:00:00Z","verifiedAt":null,"failedAt":null}`
}
