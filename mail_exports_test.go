package stackshift

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestMailExportsClient(t *testing.T) {
	client, err := New(Options{APIKey: "sk_test", HTTPClient: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch {
		case request.Method == http.MethodPost && strings.HasSuffix(request.URL.Path, "/mail/exports"):
			return jsonResponse(http.StatusAccepted, exportJSON("queued")), nil
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/mail/exports"):
			return jsonResponse(http.StatusOK, `{"data":[`+exportJSON("ready")+`]}`), nil
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/mail/exports/exp_123"):
			return jsonResponse(http.StatusOK, exportJSON("ready")), nil
		default:
			return jsonResponse(http.StatusNotFound, `{}`), nil
		}
	})})
	if err != nil {
		t.Fatal(err)
	}
	created, err := client.Mail.Exports.Create(context.Background(), MailExportCreateRequest{})
	if err != nil || created.Status != "queued" {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	items, err := client.Mail.Exports.List(context.Background())
	if err != nil || len(items) != 1 || items[0].RecordCount != 12 {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	item, err := client.Mail.Exports.Get(context.Background(), "exp_123")
	if err != nil || item.Checksum != "sha256" {
		t.Fatalf("item=%+v err=%v", item, err)
	}
	if client.Mail.Exports.DownloadPath("exp_123") != "/mail/exports/exp_123/download" {
		t.Fatal("unexpected download path")
	}
}

func exportJSON(status string) string {
	return `{"id":"exp_123","kind":"recipient_delivery","format":"jsonl","status":"` + status + `","from":"2026-08-01T00:00:00Z","to":"2026-08-11T00:00:00Z","recordCount":12,"sizeBytes":1024,"checksum":"sha256","createdAt":"2026-08-11T00:00:00Z","expiresAt":"2026-08-18T00:00:00Z"}`
}
