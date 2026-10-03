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
