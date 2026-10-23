package stackshift

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestMailSubscriberImportContracts(t *testing.T) {
	type call struct {
		path, method string
		body         map[string]any
	}
	calls := []call{}
	client, err := New(Options{APIKey: "sk_test", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		calls = append(calls, call{path: req.URL.EscapedPath(), method: req.Method, body: body})
		return jsonResponse(200, `{"imported":2,"skipped":2,"preview":true,"summary":{"imported":1,"updated":1,"invalid":1,"skipped":1,"pending":1},"rows":[{"row":3,"outcome":"invalid","code":"invalid_email","reason":"Invalid email"}]}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	input := MailSubscriberImportRequest{Mode: MailImportConfirmation, Members: []MailAudienceMemberImport{{Email: "ada@example.com", ConsentSource: "signup", ConsentAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)}}}
	result, err := client.Mail.Audiences.PreviewImport(context.Background(), "audience/id", input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary.Imported != 1 || result.Imported != 2 || result.Skipped != 2 || result.Rows[0].Row != 3 {
		t.Fatalf("unexpected report: %+v", result)
	}
	if _, err = client.Mail.Audiences.ImportDetailed(context.Background(), "audience/id", input); err != nil {
		t.Fatal(err)
	}
	if err = client.Mail.Audiences.SetEditableFields(context.Background(), "audience/id", nil); err != nil {
		t.Fatal(err)
	}
	if calls[0].path != "/api/v1/mail/audiences/audience%2Fid/imports/preview" || calls[0].method != "POST" || calls[0].body["mode"] != "confirmation" {
		t.Fatalf("unexpected preview request: %+v", calls[0])
	}
	if calls[1].path != "/api/v1/mail/audiences/audience%2Fid/imports" || calls[1].body["mode"] != "confirmation" {
		t.Fatalf("unexpected commit request: %+v", calls[1])
	}
	if calls[2].method != "PUT" || !reflect.DeepEqual(calls[2].body["editableFields"], []any{}) {
		t.Fatalf("field allowlist must encode as []: %+v", calls[2])
	}
}
