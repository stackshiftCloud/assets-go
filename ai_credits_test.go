package stackshift

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestAICreditRequests(t *testing.T) {
	t.Parallel()

	requests := 0
	client, err := New(Options{APIKey: "account-key", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		if req.Header.Get("Authorization") != "Bearer account-key" || req.Header.Get("X-Asset-Space-ID") != "" {
			t.Fatal("AI credit request crossed the account credential boundary")
		}

		var body []byte
		if req.Body != nil {
			body, _ = io.ReadAll(req.Body)
		}
		switch req.URL.RequestURI() {
		case "/api/v1/billing/ai-credits":
			return jsonResponse(http.StatusOK, `{"success":true,"data":{"available_microcredits":100,"monthly_microcredits":80,"purchased_microcredits":20,"reserved_microcredits":0,"shortfall_microcredits":0,"next_reset":"2026-10-01T00:00:00Z","grants":[{"id":"grant","kind":"monthly","source":"subscription:pro","policy_revision":3,"amount_microcredits":80,"remaining_microcredits":80,"available_at":"2026-09-01T00:00:00Z","expires_at":null}]}}`), nil
		case "/api/v1/billing/ai-credits/usage?limit=10&product=agents&user=user%2Fid":
			return jsonResponse(http.StatusOK, `{"success":true,"data":{"items":[{"id":"usage","operation_key":"run","user_id":"user","product":"agents","feature":"completion","funding_mode":"managed","state":"settled","created_at":"2026-09-08T00:00:00Z","reserved_microcredits":10,"used_microcredits":8,"returned_microcredits":2}],"totals_microcredits":{"agents":8}}}`), nil
		case "/api/v1/billing/ai-credits/top-ups/quote":
			assertAICreditJSONBody(t, body, map[string]any{"credits": float64(1000), "currency": "USD"})
			return jsonResponse(http.StatusOK, `{"success":true,"data":{"id":"quote","catalog_revision":3,"credit_microcredits":1000000000,"amount_minor":100,"currency":"USD","expires_at":"2026-09-08T00:15:00Z"}}`), nil
		case "/api/v1/billing/ai-credits/top-ups/checkout":
			assertAICreditJSONBody(t, body, map[string]any{"quote_id": "quote", "idempotency_key": "retry-key", "provider": "paystack"})
			return jsonResponse(http.StatusOK, `{"success":true,"data":{"reference":"ai_ref","status":"pending","authorization_url":"https://checkout.example"}}`), nil
		case "/api/v1/billing/ai-credits/funding":
			if req.Method == http.MethodPut {
				assertAICreditJSONBody(t, body, map[string]any{"product": "assets", "mode": "byok"})
				return jsonResponse(http.StatusOK, `{"success":true,"data":{"assets":"byok","agents":"managed","workflows":"managed"}}`), nil
			}
			return jsonResponse(http.StatusOK, `{"success":true,"data":{"assets":"managed","agents":"managed","workflows":"managed"}}`), nil
		default:
			t.Fatalf("unexpected AI credit request: %s %s", req.Method, req.URL.RequestURI())
			return nil, nil
		}
	})})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	balance, err := client.AICredits.Balance(ctx)
	if err != nil || balance.Grants[0].Source != "subscription:pro" || balance.Grants[0].PolicyRevision != 3 {
		t.Fatalf("balance=%+v err=%v", balance, err)
	}
	usage, err := client.AICredits.Usage(ctx, url.Values{"product": {"agents"}, "user": {"user/id"}, "limit": {"10"}})
	if err != nil || len(usage.Items) != 1 || usage.TotalsMicrocredits["agents"] != 8 {
		t.Fatalf("usage=%+v err=%v", usage, err)
	}
	quote, err := client.AICredits.Quote(ctx, 1000, "USD")
	if err != nil || quote.ID != "quote" {
		t.Fatalf("quote=%+v err=%v", quote, err)
	}
	checkout, err := client.AICredits.Checkout(ctx, "quote", "retry-key", "paystack")
	if err != nil || checkout.Reference != "ai_ref" {
		t.Fatalf("checkout=%+v err=%v", checkout, err)
	}
	funding, err := client.AICredits.Funding(ctx)
	if err != nil || funding["assets"] != "managed" {
		t.Fatalf("funding=%+v err=%v", funding, err)
	}
	funding, err = client.AICredits.SetFunding(ctx, "assets", "byok")
	if err != nil || funding["assets"] != "byok" {
		t.Fatalf("funding=%+v err=%v", funding, err)
	}
	if requests != 6 {
		t.Fatalf("requests=%d", requests)
	}
}

func assertAICreditJSONBody(t *testing.T, body []byte, want map[string]any) {
	t.Helper()
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("body=%v want=%v", got, want)
	}
	for key, value := range want {
		if got[key] != value {
			t.Fatalf("body[%q]=%v want=%v", key, got[key], value)
		}
	}
}
