package stackshift

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestMailTemplates(t *testing.T) {
	var calls []string
	client, err := New(Options{
		APIKey: "sk_test",
		HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls = append(calls, req.Method+" "+req.URL.String())
			switch {
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/send-template"):
				return jsonResponse(http.StatusAccepted, `{"id":"msg_tpl_123","status":"queued","templateId":"tpl_123","templateVersionId":"tplv_123","idempotencyStatus":"created"}`), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/templates"):
				return jsonResponse(http.StatusCreated, templateJSON()), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/templates"):
				return jsonResponse(http.StatusOK, `{"data":[`+templateJSON()+`],"nextCursor":null}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/templates/welcome-email"):
				return jsonResponse(http.StatusOK, templateDetailJSON()), nil
			case req.Method == http.MethodPatch && strings.HasSuffix(req.URL.Path, "/mail/templates/tpl_123"):
				return jsonResponse(http.StatusOK, templateJSON()), nil
			case req.Method == http.MethodDelete && strings.HasSuffix(req.URL.Path, "/mail/templates/tpl_123"):
				return jsonResponse(http.StatusNoContent, ``), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/templates/tpl_123/versions"):
				return jsonResponse(http.StatusOK, `{"data":[`+templateVersionJSON()+`]}`), nil
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/mail/templates/tpl_123/versions/tplv_123"):
				return jsonResponse(http.StatusOK, templateVersionJSON()), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/templates/tpl_123/versions/tplv_123/activate"):
				return jsonResponse(http.StatusOK, templateJSON()), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/templates/tpl_123/preview"):
				return jsonResponse(http.StatusOK, `{"subject":"Welcome, Ada","html":null,"text":"Welcome, Ada","missingVariables":[]}`), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/mail/templates/tpl_123/test-send"):
				return jsonResponse(http.StatusAccepted, `{"messageId":"msg_123","status":"queued"}`), nil
			default:
				return jsonResponse(http.StatusNotFound, `{"error":{"code":"template_not_found","message":"Template not found."}}`), nil
			}
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	created, err := client.Mail.Templates.Create(context.Background(), MailTemplateCreateRequest{Name: "Welcome", Slug: "welcome-email", Subject: "Welcome, {{name}}", Text: "Welcome, {{name}}"})
	if err != nil || created.Slug != "welcome-email" {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	list, err := client.Mail.Templates.List(context.Background(), MailTemplateListOptions{Search: "welcome", Limit: 10})
	if err != nil || list.Data[0].Variables[0] != "name" {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	detail, err := client.Mail.Templates.Get(context.Background(), "welcome-email")
	if err != nil || detail.ActiveVersion.ID != "tplv_123" {
		t.Fatalf("detail=%+v err=%v", detail, err)
	}
	if _, err = client.Mail.Templates.Update(context.Background(), "tpl_123", MailTemplateUpdateRequest{Subject: "Welcome to Acme, {{name}}"}); err != nil {
		t.Fatal(err)
	}
	versions, err := client.Mail.Templates.Versions(context.Background(), "tpl_123")
	if err != nil || versions[0].VersionNumber != 1 {
		t.Fatalf("versions=%+v err=%v", versions, err)
	}
	if _, err = client.Mail.Templates.GetVersion(context.Background(), "tpl_123", "tplv_123"); err != nil {
		t.Fatal(err)
	}
	if _, err = client.Mail.Templates.ActivateVersion(context.Background(), "tpl_123", "tplv_123"); err != nil {
		t.Fatal(err)
	}
	preview, err := client.Mail.Templates.Preview(context.Background(), "tpl_123", MailTemplatePreviewRequest{Data: map[string]any{"name": "Ada"}})
	if err != nil || preview.Subject != "Welcome, Ada" {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
	testSend, err := client.Mail.Templates.TestSend(context.Background(), "tpl_123", MailTemplateTestSendRequest{To: "ada@example.com", Data: map[string]any{"name": "Ada"}})
	if err != nil || testSend.MessageID != "msg_123" {
		t.Fatalf("testSend=%+v err=%v", testSend, err)
	}
	if err := client.Mail.Templates.Delete(context.Background(), "tpl_123"); err != nil {
		t.Fatal(err)
	}
	sent, err := client.Mail.SendTemplate(context.Background(), MailSendTemplateRequest{Template: "welcome-email", To: "ada@example.com", Data: map[string]any{"name": "Ada"}})
	if err != nil || sent.TemplateVersionID != "tplv_123" {
		t.Fatalf("sent=%+v err=%v", sent, err)
	}
	if !strings.Contains(calls[1], "search=welcome") {
		t.Fatalf("unexpected calls: %#v", calls)
	}
}

func templateJSON() string {
	return `{"id":"tpl_123","name":"Welcome","slug":"welcome-email","description":null,"status":"active","activeVersionId":"tplv_123","variables":["name"],"createdAt":"2026-05-08T00:00:00Z","updatedAt":"2026-05-08T00:00:00Z"}`
}

func templateVersionJSON() string {
	return `{"id":"tplv_123","templateId":"tpl_123","versionNumber":1,"subject":"Welcome, {{name}}","html":null,"text":"Welcome, {{name}}","variables":["name"],"createdAt":"2026-05-08T00:00:00Z"}`
}

func templateDetailJSON() string {
	return `{"id":"tpl_123","name":"Welcome","slug":"welcome-email","description":null,"status":"active","activeVersionId":"tplv_123","variables":["name"],"activeVersion":` + templateVersionJSON() + `,"versions":[` + templateVersionJSON() + `],"createdAt":"2026-05-08T00:00:00Z","updatedAt":"2026-05-08T00:00:00Z"}`
}
