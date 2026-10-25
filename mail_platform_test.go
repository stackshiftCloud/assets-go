package stackshift

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestMailPlatformContracts(t *testing.T) {
	var path string
	var body map[string]any
	client, err := New(Options{APIKey: "sspat_test_fixture", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		path = req.URL.Path
		body = map[string]any{}
		if req.Body != nil {
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
		}
		return jsonResponse(200, `{"id":"fixture","brandId":"brand","mode":"test","status":"blocked","approvedMaxKobo":0}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err = client.Mail.Segments.Create(ctx, MailSegmentInput{AudienceID: "audience", Name: "Compound", Match: "any", Conditions: []MailSegmentCondition{{Field: "plan", Operator: "equals", Value: "pro"}, {Field: "name", Operator: "exists"}}}); err != nil {
		t.Fatal(err)
	}
	if body["match"] != "any" || len(body["conditions"].([]any)) != 2 {
		t.Fatalf("compound rules lost: %#v", body)
	}
	if _, err = client.Mail.Audiences.Create(ctx, MailAudienceCreateRequest{BrandID: "brand", Name: "News"}); err != nil || body["brandId"] != "brand" {
		t.Fatalf("brand contract: %v %#v", err, body)
	}
	if _, err = client.Mail.Automations.Create(ctx, MailAutomationInput{AudienceID: "audience", Simulation: &MailSimulationOptions{Scenario: "bounced"}}); err != nil {
		t.Fatal(err)
	}
	if body["simulation"].(map[string]any)["scenario"] != "bounced" {
		t.Fatal("lost automation scenario")
	}
	campaign, err := client.Mail.Campaigns.SendWithBudget(ctx, "campaign", MailCampaignBudget{Revision: 2, ApprovedMaxKobo: 0})
	if err != nil || campaign.ApprovedMaxKobo == nil || *campaign.ApprovedMaxKobo != 0 || body["approvedMaxKobo"] != float64(0) {
		t.Fatalf("zero approval lost: %v %#v", err, body)
	}
	if _, err = client.Mail.Migration.Plan(ctx, "example.com", "resend", nil); err != nil || !strings.HasPrefix(path, "/api/v1/") {
		t.Fatalf("public route %s: %v", path, err)
	}
	if _, err = client.Mail.Migration.SaveChecklist(ctx, MailMigrationChecklist{Domain: "example.com", Provider: "resend", Completed: []string{"audit"}, Revision: 1, UpdatedAt: "response-only"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["updatedAt"]; ok {
		t.Fatal("sent response-only checklist timestamp")
	}
}

func TestMailBrandScopedReads(t *testing.T) {
	var uri string
	client, err := New(Options{APIKey: "sspat_test_fixture", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		uri = req.URL.RequestURI()
		switch req.URL.Path {
		case "/api/v1/mail/billing":
			return jsonResponse(200, `{"period":"2026-09","planSlug":"growth","currency":"NGN","periodStart":"2026-09-01T00:00:00Z","periodEnd":"2026-10-01T00:00:00Z","monthlyPriceKobo":500000,"included":50000,"used":41200,"spentKobo":0}`), nil
		case "/api/v1/mail/brands":
			return jsonResponse(200, `{"data":[{"id":"brand_1","warmup":{"day":9,"totalDays":30,"marketingUnlocked":true,"status":"warming"}},{"id":"brand_2","warmup":null}]}`), nil
		}
		return jsonResponse(200, `{"data":[]}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	usage, err := client.Mail.Billing.PeriodUsage(ctx, "2026-09")
	if err != nil || uri != "/api/v1/mail/billing?period=2026-09" || usage.Used != 41200 || usage.PlanSlug != "growth" {
		t.Fatalf("period usage: %v %s %#v", err, uri, usage)
	}
	brands, err := client.Mail.Brands.List(ctx)
	if err != nil || brands[0].Warmup == nil || brands[0].Warmup.Day != 9 || !brands[0].Warmup.MarketingUnlocked || brands[1].Warmup != nil {
		t.Fatalf("brand warmup: %v %#v", err, brands)
	}
	if _, err = client.Mail.Audiences.ListForBrand(ctx, "brand_1"); err != nil || uri != "/api/v1/mail/audiences?brandId=brand_1" {
		t.Fatalf("audiences by brand: %v %s", err, uri)
	}
	if _, err = client.Mail.Campaigns.ListForBrand(ctx, "brand_1"); err != nil || uri != "/api/v1/mail/campaigns?brandId=brand_1" {
		t.Fatalf("campaigns by brand: %v %s", err, uri)
	}
	if _, err = client.Mail.Suppressions.List(ctx, MailSuppressionListOptions{BrandID: "brand_1", Reason: "complaint"}); err != nil || uri != "/api/v1/mail/suppressions?brandId=brand_1&reason=complaint" {
		t.Fatalf("suppressions by brand: %v %s", err, uri)
	}
}
