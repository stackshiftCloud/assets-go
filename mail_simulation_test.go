package stackshift

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestMailSimulationContract(t *testing.T) {
	client, err := New(Options{APIKey: "sspat_test_example", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var payload struct {
			Simulation MailSimulationOptions `json:"simulation"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Simulation.Scenario != "delayed" {
			t.Fatalf("simulation option lost: %+v", payload)
		}
		return jsonResponse(200, `{"id":"message","mode":"test","status":"queued"}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Mail.Send(context.Background(), MailSendRequest{From: "hello@example.com", To: "test@example.com", Subject: "Simulation", Text: "Test", Simulation: &MailSimulationOptions{Scenario: "delayed"}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Mode != MailTest {
		t.Fatalf("mode lost: %+v", result)
	}
}
