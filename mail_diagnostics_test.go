package stackshift

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestMailTemplatesClient_Diagnostics(t *testing.T) {
	client, err := New(Options{APIKey: "sk_test", HTTPClient: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost || !strings.HasSuffix(request.URL.Path, "/mail/templates/tpl_123/diagnostics") {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		return jsonResponse(http.StatusOK, `{"status":"warning","stream":"broadcast","subjectLength":20,"htmlBytes":1200,"textBytes":400,"linkCount":2,"remoteImageCount":1,"missingAltCount":0,"findings":[{"code":"remote_images","severity":"warning","field":"html","message":"Remote image","recommendation":"Keep fallback content"}]}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Mail.Templates.Diagnostics(context.Background(), "tpl_123", MailTemplateDiagnosticRequest{Stream: "broadcast"})
	if err != nil || result.Status != "warning" || len(result.Findings) != 1 || result.Findings[0].Code != "remote_images" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
