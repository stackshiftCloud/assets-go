package stackshift

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestMailReputationClients(t *testing.T) {
	t.Setenv("STACKSHIFT_API_KEY", "env-key")
	client, err := New(Options{HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("Authorization") != "Bearer env-key" {
			t.Fatalf("missing api key: %s", req.Header.Get("Authorization"))
		}
		switch {
		case strings.HasSuffix(req.URL.Path, "/mail/limits"):
			return jsonResponse(http.StatusOK, `{"tier":"verified","status":"healthy","dailyLimit":2000,"dailyUsed":12,"hourlyLimit":400,"hourlyUsed":3,"perMinuteLimit":30,"currentWarmupStage":3,"manualReviewStatus":"not_required"}`), nil
		case strings.HasSuffix(req.URL.Path, "/mail/reputation"):
			return jsonResponse(http.StatusOK, `{"workspace":{"status":"healthy","score":92,"tier":"verified","bounceRate":0.01,"deferralRate":0.02,"suppressionRate":0,"complaintRate":null},"domains":[]}`), nil
		case strings.HasSuffix(req.URL.Path, "/mail/reputation/events"):
			return jsonResponse(http.StatusOK, `{"data":[{"id":"rep_evt_123","type":"mail.reputation.workspace.watch","severity":"warning","message":"watch","metadata":{"bounceRate":0.06},"createdAt":"2026-05-09T00:00:00Z"}],"nextCursor":null}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		return jsonResponse(http.StatusNotFound, `{}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	limits, err := client.Mail.Limits.Get(context.Background())
	if err != nil || limits.Tier != "verified" {
		t.Fatalf("limits=%+v err=%v", limits, err)
	}
	rep, err := client.Mail.Reputation.Get(context.Background())
	if err != nil || rep.Workspace.Score != 92 {
		t.Fatalf("reputation=%+v err=%v", rep, err)
	}
	events, err := client.Mail.Reputation.Events.List(context.Background(), MailReputationEventListOptions{Severity: "warning"})
	if err != nil || len(events.Data) != 1 {
		t.Fatalf("events=%+v err=%v", events, err)
	}
}
