package stackshift

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"
)

const eventBody = `{"id":"evt_1","type":"deploy.failed","version":1,"time":"2026-11-26T10:00:00Z","product":"hosting","subject":{"kind":"deployment","id":"dep_1"},"actor":{"type":"system"},"data":{"branch":"main"}}`

func signEvent(body string, at time.Time, secret string) string {
	timestamp := strconv.FormatInt(at.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "." + body))
	return "t=" + timestamp + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyEvent(t *testing.T) {
	now := time.Now()
	event, err := VerifyEvent(VerifyEventInput{RawBody: []byte(eventBody), SignatureHeader: signEvent(eventBody, now, "whsec"), Secret: "whsec"})
	if err != nil || event.Type != "deploy.failed" || event.Subject.ID != "dep_1" {
		t.Fatalf("event=%+v err=%v", event, err)
	}
	cases := []struct {
		name   string
		header string
		secret string
		code   string
	}{
		{"bad signature", signEvent(eventBody, now, "other"), "whsec", "invalid_signature"},
		{"expired", signEvent(eventBody, now.Add(-time.Hour), "whsec"), "whsec", "expired"},
		{"missing", "", "whsec", "missing_signature"},
	}
	for _, tc := range cases {
		_, err := VerifyEvent(VerifyEventInput{RawBody: []byte(eventBody), SignatureHeader: tc.header, Secret: tc.secret})
		var signatureErr *EventSignatureError
		if !errors.As(err, &signatureErr) || signatureErr.Code != tc.code {
			t.Errorf("%s: err=%v", tc.name, err)
		}
	}
}

func TestOnEvent(t *testing.T) {
	var seen []string
	for _, pattern := range []string{"deploy.failed", "deploy.*", "*.failed", "mail.*"} {
		handle := OnEvent(pattern, func(ctx context.Context, event Event) error {
			seen = append(seen, pattern)
			return nil
		})
		if err := handle(context.Background(), EventInvocation{TriggerType: "queue", QueueName: "_events", Payload: json.RawMessage(eventBody)}); err != nil {
			t.Fatal(err)
		}
	}
	if len(seen) != 3 {
		t.Fatalf("seen=%v", seen)
	}
	handle := OnEvent("*", func(ctx context.Context, event Event) error { t.Fatal("ran"); return nil })
	_ = handle(context.Background(), EventInvocation{QueueName: "emails", Payload: json.RawMessage(eventBody)})
}
