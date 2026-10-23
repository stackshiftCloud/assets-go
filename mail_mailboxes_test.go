package stackshift

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type recordedMailboxCall struct {
	method string
	path   string
	query  string
	body   string
}

func newMailboxTestClient(t *testing.T, calls *[]recordedMailboxCall, response string) *Client {
	t.Helper()
	client, err := New(Options{APIKey: "sk_test", HTTPClient: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var body string
		if request.Body != nil {
			raw, _ := io.ReadAll(request.Body)
			body = string(raw)
		}
		*calls = append(*calls, recordedMailboxCall{method: request.Method, path: request.URL.Path, query: request.URL.RawQuery, body: body})
		if request.Method == http.MethodDelete {
			return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}}, nil
		}
		return jsonResponse(http.StatusOK, response), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestMailboxDomainsAndMailboxes(t *testing.T) {
	var calls []recordedMailboxCall
	client := newMailboxTestClient(t, &calls, `{"data":[{"id":"mbx_1","address":"ada@example.com","status":"active"}],"id":"mbx_1","address":"ada@example.com","password":"pw","appliedNow":true,"secret":"s3cret","retried":2}`)
	ctx := context.Background()

	days := 30
	if _, err := client.Mail.MailboxDomains.Update(ctx, "dom_1", MailboxDomainUpdateRequest{RetentionDays: &days, ClearCatchAll: true}); err != nil {
		t.Fatal(err)
	}
	last := calls[len(calls)-1]
	if last.method != http.MethodPatch || last.path != "/api/v1/mail/mailbox-domains/dom_1" || !strings.Contains(last.body, `"catchAll":null`) || !strings.Contains(last.body, `"retentionDays":30`) {
		t.Fatalf("call=%+v", last)
	}
	if err := client.Mail.MailboxDomains.Delete(ctx, "dom_1"); err != nil {
		t.Fatal(err)
	}

	created, err := client.Mail.Mailboxes.Create(ctx, MailboxCreateRequest{DomainID: "dom_1", LocalPart: "ada", OwnerEmail: "ada@personal.test"})
	if err != nil || created.Address != "ada@example.com" {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	if last = calls[len(calls)-1]; last.body != `{"domainId":"dom_1","localPart":"ada","ownerEmail":"ada@personal.test"}` {
		t.Fatalf("body=%s", last.body)
	}
	items, err := client.Mail.Mailboxes.List(ctx, "dom_1")
	if err != nil || len(items) != 1 || calls[len(calls)-1].query != "domainId=dom_1" {
		t.Fatalf("items=%+v calls=%+v err=%v", items, calls, err)
	}
	reset, err := client.Mail.Mailboxes.ResetPassword(ctx, "mbx_1")
	if err != nil || !reset.AppliedNow || calls[len(calls)-1].path != "/api/v1/mail/mailboxes/mbx_1/password-reset" {
		t.Fatalf("reset=%+v err=%v", reset, err)
	}
	if _, err := client.Mail.Mailboxes.Suspend(ctx, "mbx_1"); err != nil || calls[len(calls)-1].path != "/api/v1/mail/mailboxes/mbx_1/suspend" {
		t.Fatalf("suspend err=%v", err)
	}
	if err := client.Mail.Mailboxes.RevokeDevice(ctx, "mbx_1", "app_password", "dev_1"); err != nil || calls[len(calls)-1].path != "/api/v1/mail/mailboxes/mbx_1/devices/app_password/dev_1" {
		t.Fatalf("revoke err=%v", err)
	}
	password, err := client.Mail.Mailboxes.CreateAppPassword(ctx, "mbx_1", "phone")
	if err != nil || password.Secret != "s3cret" {
		t.Fatalf("password=%+v err=%v", password, err)
	}
}

func TestMailboxAliasesGroupsRoutesAndImports(t *testing.T) {
	var calls []recordedMailboxCall
	client := newMailboxTestClient(t, &calls, `{"id":"res_1","retried":2,"issue":null,"data":[]}`)
	ctx := context.Background()

	if _, err := client.Mail.Aliases.Update(ctx, "als_1", false); err != nil || calls[len(calls)-1].body != `{"sendAs":false}` {
		t.Fatalf("alias call=%+v err=%v", calls[len(calls)-1], err)
	}
	members := []string{}
	if _, err := client.Mail.Groups.Update(ctx, "grp_1", MailboxGroupUpdateRequest{MemberIDs: &members, ClearModerator: true}); err != nil {
		t.Fatal(err)
	}
	if body := calls[len(calls)-1].body; !strings.Contains(body, `"memberIds":[]`) || !strings.Contains(body, `"moderatorId":null`) {
		t.Fatalf("group body=%s", body)
	}
	if _, err := client.Mail.Routes.Create(ctx, MailboxRouteCreateRequest{DomainID: "dom_1", Pattern: "orders-*", Targets: []MailboxRouteTarget{{Kind: "app", WebhookURL: "https://app.test/inbound"}}}); err != nil {
		t.Fatal(err)
	}
	if body := calls[len(calls)-1].body; !strings.Contains(body, `"kind":"app"`) {
		t.Fatalf("route body=%s", body)
	}
	retried, err := client.Mail.Routes.RetryDeliveries(ctx, "rte_1")
	if err != nil || retried != 2 {
		t.Fatalf("retried=%d err=%v", retried, err)
	}
	issue, err := client.Mail.Routes.AppHealth(ctx, "dom_1")
	if err != nil || issue != nil || calls[len(calls)-1].query != "domainId=dom_1" {
		t.Fatalf("issue=%+v err=%v", issue, err)
	}
	if _, err := client.Mail.MailboxImports.Start(ctx, MailboxImportStartRequest{DomainID: "dom_1", Source: "imap", Host: "imap.old.test", Accounts: []MailboxImportAccount{{MailboxID: "mbx_1", Username: "ada", Password: "x"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Mail.MailboxImports.Switch(ctx, "imp_1", "tonight"); err != nil || calls[len(calls)-1].body != `{"when":"tonight"}` {
		t.Fatalf("switch call=%+v err=%v", calls[len(calls)-1], err)
	}
}
