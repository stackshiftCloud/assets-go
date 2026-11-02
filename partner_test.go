package stackshift

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
)

type partnerCall struct {
	method string
	uri    string
	key    string
	body   map[string]any
}

func TestPartnerRequests(t *testing.T) {
	t.Parallel()

	var calls []partnerCall
	status := http.StatusOK
	payload := `{"success":true,"data":{"id":"dom_1","brandId":"brand_1","domain":"merchantshop.com","state":"registering","brandStatus":{"id":"brand_1","merchantRef":"m_123","sendingPausedAt":"2026-10-23T10:00:00Z","sendingPauseReason":"complaint rate","suspendedAt":"2026-10-23T11:00:00Z"},"senders":{"marketing":"hello@news.merchantshop.com"},"warmup":{"marketingUnlocked":true,"requiredCleanDays":7},"brandPauseReason":"","expiryDate":"2027-10-23T00:00:00Z","account":{"dailyPurchaseLimit":25},"transferredOutAt":"2026-10-24T09:00:00Z","limits":{"shortWindowHours":24,"dailyPurchaseLimit":25,"warningFraction":0.5,"merchantSharePercent":50,"merchantCount":2,"dailySendingLimit":10000,"hourlySendingLimit":1000},"health":{"attempted":400,"accepted":390,"last24h":{"attempted":100,"accepted":98,"bounced":3,"bounceRate":0.03}},"transactionalHealth":{"attempted":500,"accepted":495,"last24h":{"attempted":60}},"domains":[{"id":"dom_1","health":{"attempted":10},"transactionalHealth":{"attempted":20}}],"total":1,"authCode":"AUTH"}}`
	client, err := New(Options{APIKey: "sspat_fixture", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		call := partnerCall{method: req.Method, uri: req.URL.RequestURI(), key: req.Header.Get("Idempotency-Key")}
		if req.Body != nil {
			raw, _ := io.ReadAll(req.Body)
			if len(raw) > 0 {
				_ = json.Unmarshal(raw, &call.body)
			}
		}
		calls = append(calls, call)
		return jsonResponse(status, payload), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	last := func() partnerCall { return calls[len(calls)-1] }

	overview, err := client.Partner.Overview(ctx)
	if err != nil || last().method != http.MethodGet || last().uri != "/api/v1/partner" {
		t.Fatalf("overview: %v %#v", err, last())
	}
	if overview.Account.DailyPurchaseLimit != 25 || overview.Limits.ShortWindowHours != 24 || overview.Limits.DailyPurchaseLimit != 25 || overview.Limits.WarningFraction != 0.5 || overview.Limits.MerchantSharePercent*overview.Limits.DailySendingLimit/100 != 5000 {
		t.Fatalf("overview limits: %#v", overview)
	}
	if _, err := client.Partner.UpdateSettings(ctx, PartnerSettingsInput{ComplaintPauseRate: 0.002, BouncePauseRate: 0.04}); err != nil ||
		last().method != http.MethodPut || last().uri != "/api/v1/partner/settings" || last().body["complaintPauseRate"] != 0.002 {
		t.Fatalf("settings: %v %#v", err, last())
	}

	purchase := PartnerDomainPurchaseInput{
		Domain: "merchantshop.com", Years: 1, Currency: "USD", MerchantRef: "m_123",
		Registrant: PartnerRegistrant{FirstName: "Ada", LastName: "Obi", Email: "ada@merchantshop.com", Phone: "+2348012345678", Address: "1 Marina", City: "Lagos", State: "Lagos", PostalCode: "101001", Country: "NG"},
		Brand:      PartnerBrandInput{Name: "Merchant Shop", PostalAddress: "1 Marina, Lagos"},
		Replies:    PartnerRepliesInput{Mode: "forward", ForwardTo: "ada@gmail.com"},
	}
	domain, err := client.Partner.PurchaseDomain(ctx, purchase, "order-1")
	if err != nil || domain.ID != "dom_1" || last().key != "order-1" || last().uri != "/api/v1/partner/domains" || last().method != http.MethodPost {
		t.Fatalf("purchase: %v %#v", err, last())
	}
	if last().body["merchantRef"] != "m_123" || last().body["currency"] != "USD" || last().body["registrant"].(map[string]any)["first_name"] != "Ada" || last().body["brand"].(map[string]any)["name"] != "Merchant Shop" {
		t.Fatalf("purchase body: %#v", last().body)
	}

	list, err := client.Partner.ListDomains(ctx, PartnerDomainListOptions{MerchantRef: "m 123", Limit: 20, Offset: 40})
	if err != nil || list.Total != 1 || list.Domains[0].Health == nil || list.Domains[0].TransactionalHealth == nil || list.Domains[0].TransactionalHealth.Attempted != 20 || last().uri != "/api/v1/partner/domains?limit=20&merchantRef=m+123&offset=40" {
		t.Fatalf("list: %v %#v", err, last())
	}
	if _, err := client.Partner.ListDomains(ctx, PartnerDomainListOptions{}); err != nil || last().uri != "/api/v1/partner/domains" {
		t.Fatalf("empty list: %v %#v", err, last())
	}
	if _, err := client.Partner.ListDomains(ctx, PartnerDomainListOptions{Query: "shop", Attention: true}); err != nil || last().uri != "/api/v1/partner/domains?attention=true&q=shop" {
		t.Fatalf("attention list: %v %#v", err, last())
	}

	view, err := client.Partner.GetDomain(ctx, "dom/1")
	if err != nil || last().uri != "/api/v1/partner/domains/dom%2F1" || view.BrandStatus == nil || view.BrandStatus.SendingPausedAt == nil || view.BrandStatus.MerchantRef != "m_123" || view.Senders.Marketing != "hello@news.merchantshop.com" || view.Warmup == nil || !view.Warmup.MarketingUnlocked || view.ExpiryDate == nil || view.TransferredOutAt == nil || view.Health == nil || view.Health.Accepted != 390 || view.Health.Last24h.Attempted != 100 || view.Health.Last24h.BounceRate != 0.03 || view.BrandStatus.SuspendedAt == nil || view.TransactionalHealth == nil || view.TransactionalHealth.Attempted != 500 || view.TransactionalHealth.Last24h.Attempted != 60 {
		t.Fatalf("get: %v %#v %#v", err, last(), view)
	}
	if _, err := client.Partner.RetryDomain(ctx, "dom_1"); err != nil || last().method != http.MethodPost || last().uri != "/api/v1/partner/domains/dom_1/retry" {
		t.Fatalf("retry: %v %#v", err, last())
	}
	transfer, err := client.Partner.TransferOut(ctx, "dom_1")
	if err != nil || transfer.AuthCode != "AUTH" || last().uri != "/api/v1/partner/domains/dom_1/transfer-out" {
		t.Fatalf("transfer: %v %#v", err, last())
	}
	if _, err := client.Partner.ResumeBrand(ctx, "brand_1"); err != nil || last().method != http.MethodPost || last().uri != "/api/v1/partner/brands/brand_1/resume" {
		t.Fatalf("resume: %v %#v", err, last())
	}
	if _, err := client.Partner.Usage(ctx, PartnerUsageOptions{From: "2026-10-01T00:00:00Z", To: "2026-11-01T00:00:00Z"}); err != nil ||
		last().uri != "/api/v1/partner/usage?from=2026-10-01T00%3A00%3A00Z&to=2026-11-01T00%3A00%3A00Z" {
		t.Fatalf("usage: %v %#v", err, last())
	}

	status = http.StatusPaymentRequired
	payload = `{"success":false,"message":"Your StackShift wallet does not cover this purchase.","error":{"code":"insufficient_wallet_balance","details":{"requiredMinor":1500000,"availableMinor":200000,"currency":"NGN"}}}`
	_, err = client.Partner.PurchaseDomain(ctx, purchase, "order-2")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusPaymentRequired || apiErr.Code != "insufficient_wallet_balance" || apiErr.Details["currency"] != "NGN" {
		t.Fatalf("wallet error: %#v", err)
	}

	status = http.StatusTooManyRequests
	payload = `{"success":false,"message":"You have reached today's domain purchase limit.","error":{"code":"daily_purchase_limit"}}`
	_, err = client.Partner.PurchaseDomain(ctx, purchase, "order-3")
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusTooManyRequests || apiErr.Code != "daily_purchase_limit" {
		t.Fatalf("daily limit error: %#v", err)
	}
}

func TestPartnerMerchantInsights(t *testing.T) {
	t.Parallel()

	var calls []partnerCall
	status := http.StatusOK
	payload := ""
	client, err := New(Options{APIKey: "sspat_fixture", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		call := partnerCall{method: req.Method, uri: req.URL.RequestURI(), key: req.Header.Get("Idempotency-Key")}
		if req.Body != nil {
			raw, _ := io.ReadAll(req.Body)
			if len(raw) > 0 {
				_ = json.Unmarshal(raw, &call.body)
			}
		}
		calls = append(calls, call)
		return jsonResponse(status, payload), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	last := func() partnerCall { return calls[len(calls)-1] }

	payload = `{"success":true,"data":{"domains":[{"id":"dom_1","lastAttemptAt":"2026-10-24T08:00:00Z","waitingFor":{"kind":"wallet","currency":"NGN","requiredMinor":130000,"availableMinor":20000},"refunded":true,"sentByDay":[0,0,3,4,0,1,2],"warmup":{"status":"warming","day":3,"totalDays":30,"marketingUnlocked":false,"marketingOpensInDays":4},"pause":{"reason":"complaint_rate","rate":0.004,"limit":0.003,"window":"last_7_days","campaignName":"Sale","listShare":0.8,"recoversAt":null},"failure":{"code":"wallet_short","title":"Waiting for your wallet","headline":"Waiting","explanation":"Top up.","since":"2026-10-23T11:00:00Z","facts":[{"label":"Needed","value":"1,300","tone":"bad","at":null}],"data":{},"fix":{"label":"Top up","href":"https://app.stackshift.cloud/billing/wallet"},"canRetry":true},"diagnosis":{"checkedAt":"2026-10-24T09:00:00Z","summary":"The wallet is short."}}],"total":1,"attentionTotal":1}}`
	list, err := client.Partner.ListDomains(ctx, PartnerDomainListOptions{})
	item := list.Domains[0]
	if err != nil || list.AttentionTotal != 1 || item.WaitingFor == nil || item.WaitingFor.RequiredMinor != 130000 || !item.Refunded || len(item.SentByDay) != 7 || item.Warmup == nil || item.Warmup.MarketingOpensInDays != 4 ||
		item.Pause == nil || item.Pause.Reason != "complaint_rate" || item.Pause.RecoversAt != nil || item.Failure == nil || item.Failure.Fix == nil || item.Failure.Fix.Href == "" || item.Failure.Facts[0].Tone != "bad" || item.Diagnosis == nil || item.LastAttemptAt == nil {
		t.Fatalf("list insights: %v %#v", err, list)
	}

	payload = `{"success":true,"data":{"id":"dom_1","replies":{"mode":"mailbox","mailboxRecoveryEmail":"ada@gmail.com","mailboxOwnerName":"Ada Obi"},"records":[{"domain":"news.merchantshop.com","type":"MX","name":"merchantshop.com","value":"mx.stackshift.cloud","priority":10,"verified":true,"purpose":"reply_mx"}],"refunded":true,"refund":{"amountMinor":130000,"currency":"NGN","at":"2026-10-23T12:00:00Z"},"webhookDeliveredAt":"2026-10-23T11:05:00Z","warmup":{"marketingSentToday":0,"marketingOpensInDays":4}}}`
	view, err := client.Partner.GetDomain(ctx, "dom_1")
	if err != nil || view.Replies.MailboxOwnerName == nil || *view.Replies.MailboxOwnerName != "Ada Obi" || view.Replies.MailboxRecoveryEmail != "ada@gmail.com" || len(view.Records) != 1 || view.Records[0].Purpose != "reply_mx" || *view.Records[0].Priority != 10 ||
		view.Refund == nil || view.Refund.AmountMinor != 130000 || view.WebhookDeliveredAt == nil || view.Warmup.MarketingOpensInDays != 4 {
		t.Fatalf("detail insights: %v %#v", err, view)
	}

	payload = `{"success":true,"data":{"checkedAt":"2026-10-24T09:00:00Z","summary":"The nameservers were changed."}}`
	diagnosis, err := client.Partner.GetDiagnosis(ctx, "dom/1")
	if err != nil || diagnosis.Summary != "The nameservers were changed." || last().method != http.MethodGet || last().uri != "/api/v1/partner/domains/dom%2F1/diagnosis" {
		t.Fatalf("diagnosis: %v %#v", err, last())
	}

	payload = `{"success":true,"data":{"id":"dom_1"}}`
	if _, err := client.Partner.RestoreDNS(ctx, "dom_1"); err != nil || last().method != http.MethodPost || last().uri != "/api/v1/partner/domains/dom_1/restore-dns" {
		t.Fatalf("restore dns: %v %#v", err, last())
	}

	payload = `{"success":true,"data":{"id":"dom_1","source":"connected","connectMethod":"cloudflare","mail":{"migration":"suggested"},"website":{"projectId":"prj_1","projectName":"Storefront","state":"waiting_dns","hosts":[{"hostname":"kicksbyfemi.ng","dns":"waiting","found":["23.227.38.65"],"https":"none","httpsRenewsAt":null}],"records":[{"type":"A","name":"@","value":"203.0.113.7","state":"wrong"}],"queue":{"ahead":12,"estimatedLiveAt":"2026-11-02T09:00:00Z"},"liveAt":null,"failure":null},"transfer":{"state":"moving_dns","startedAt":"2026-11-01T09:00:00Z","completedAt":null,"amountMinor":1500000,"currency":"NGN","refunded":false,"dnsCopied":[{"type":"MX","name":"@","value":"mx.example.com"}],"failure":null}}}`
	connected, err := client.Partner.ConnectDomain(ctx, PartnerDomainConnectInput{
		Domain: "kicksbyfemi.ng", MerchantRef: "m_10377",
		Brand:   PartnerBrandInput{Name: "Kicks by Femi", PostalAddress: "1 Marina, Lagos"},
		Replies: PartnerRepliesInput{Mode: "mailbox", MailboxRecoveryEmail: "femi@gmail.com"},
		Website: &PartnerWebsiteInput{ProjectID: "prj_1"},
		DNS:     PartnerDNSInput{Method: "cloudflare", CloudflareToken: "cf_token"},
	}, "connect-1")
	if err != nil || connected.Source != "connected" || connected.Website == nil || connected.Website.Queue == nil || connected.Website.Queue.Ahead != 12 || connected.Website.Hosts[0].Found[0] != "23.227.38.65" ||
		connected.Transfer == nil || connected.Transfer.State != "moving_dns" || connected.Transfer.DNSCopied[0].Type != "MX" || connected.Mail.Migration == nil || *connected.Mail.Migration != "suggested" ||
		last().method != http.MethodPost || last().uri != "/api/v1/partner/domains/connect" || last().key != "connect-1" ||
		last().body["website"].(map[string]any)["projectId"] != "prj_1" || last().body["dns"].(map[string]any)["method"] != "cloudflare" ||
		last().body["replies"].(map[string]any)["mailboxRecoveryEmail"] != "femi@gmail.com" {
		t.Fatalf("connect: %v %#v", err, last())
	}
	if _, err := client.Partner.TransferInDomain(ctx, PartnerDomainTransferInInput{
		PartnerDomainConnectInput: PartnerDomainConnectInput{Domain: "kicksbyfemi.ng", MerchantRef: "m_10377", DNS: PartnerDNSInput{Method: "manual"}},
		AuthCode:                  "Zx9-auth", ExtraHosts: []string{"shop"},
	}, "transfer-1"); err != nil || last().uri != "/api/v1/partner/domains/transfer-in" || last().key != "transfer-1" || last().body["authCode"] != "Zx9-auth" || last().body["domain"] != "kicksbyfemi.ng" || last().body["extraHosts"].([]any)[0] != "shop" {
		t.Fatalf("transfer in: %v %#v", err, last())
	}
	if _, err := client.Partner.TransferInConnectedDomain(ctx, "dom/1", PartnerConnectedTransferInInput{AuthCode: "Zx9-auth", Currency: "USD"}, "transfer-2"); err != nil ||
		last().method != http.MethodPost || last().uri != "/api/v1/partner/domains/dom%2F1/transfer-in" || last().key != "transfer-2" || last().body["currency"] != "USD" {
		t.Fatalf("transfer in connected: %v %#v", err, last())
	}
	if _, err := client.Partner.SetWebsite(ctx, "dom_1", PartnerWebsiteInput{ProjectID: "prj_2"}); err != nil || last().method != http.MethodPut || last().uri != "/api/v1/partner/domains/dom_1/website" || last().body["projectId"] != "prj_2" {
		t.Fatalf("set website: %v %#v", err, last())
	}
	if _, err := client.Partner.RemoveWebsite(ctx, "dom_1"); err != nil || last().method != http.MethodDelete || last().uri != "/api/v1/partner/domains/dom_1/website" {
		t.Fatalf("remove website: %v %#v", err, last())
	}
	if _, err := client.Partner.CheckWebsite(ctx, "dom_1"); err != nil || last().method != http.MethodPost || last().uri != "/api/v1/partner/domains/dom_1/website/check" {
		t.Fatalf("check website: %v %#v", err, last())
	}
	if _, err := client.Partner.ListDomains(ctx, PartnerDomainListOptions{Website: "pending"}); err != nil || last().uri != "/api/v1/partner/domains?website=pending" {
		t.Fatalf("website filter: %v %#v", err, last())
	}

	payload = `{"success":true,"data":[{"id":"prj_1","name":"Storefront","slug":"storefront","hostname":"storefront.stackshift.cloud","merchantWebsites":1240}]}`
	projects, err := client.Partner.ListWebsiteProjects(ctx)
	if err != nil || len(projects) != 1 || projects[0].MerchantWebsites != 1240 || last().uri != "/api/v1/partner/website-projects" {
		t.Fatalf("website projects: %v %#v", err, last())
	}

	payload = `{"success":true,"data":[{"id":"mbx_1","address":"hello@kicksbyfemi.ng","kind":"mailbox","displayName":"Femi","deliversTo":null,"status":"awaiting_password","recoveryEmail":"femi@gmail.com","passwordSetAt":null,"passwordLinkSentAt":"2026-11-01T09:00:00Z","lastSignInAt":null,"lastSignInApp":null,"storageUsedBytes":0,"quotaBytes":10737418240,"plan":"basic","price":{"amountMinor":250000,"currency":"NGN"},"renewsAt":"2026-12-01T00:00:00Z"}]}`
	mailboxes, err := client.Partner.ListMailboxes(ctx, "dom_1")
	if err != nil || len(mailboxes) != 1 || mailboxes[0].Status != "awaiting_password" || mailboxes[0].PasswordLinkSentAt == nil || mailboxes[0].Price == nil || mailboxes[0].Price.AmountMinor != 250000 || last().uri != "/api/v1/partner/domains/dom_1/mailboxes" {
		t.Fatalf("list mailboxes: %v %#v", err, last())
	}
	payload = `{"success":true,"data":{"mailbox":{"id":"mbx_2","address":"orders@kicksbyfemi.ng","kind":"mailbox","status":"awaiting_password"},"passwordLink":{"url":"https://mail.kicksbyfemi.ng/webmail/set-password#tok","expiresAt":"2026-11-08T00:00:00Z"}}}`
	sendLink := true
	created, err := client.Partner.CreateMailbox(ctx, "dom_1", PartnerMailboxInput{LocalPart: "orders", Plan: "pro", SendPasswordLink: &sendLink}, "mailbox-1")
	if err != nil || created.Mailbox.ID != "mbx_2" || created.PasswordLink.URL == "" || last().method != http.MethodPost || last().uri != "/api/v1/partner/domains/dom_1/mailboxes" || last().key != "mailbox-1" || last().body["localPart"] != "orders" || last().body["sendPasswordLink"] != true {
		t.Fatalf("create mailbox: %v %#v", err, last())
	}
	payload = `{"success":true,"data":{"id":"mbx_2"}}`
	name := "Femi A."
	if _, err := client.Partner.UpdateMailbox(ctx, "dom_1", "mbx/2", PartnerMailboxUpdate{DisplayName: &name}); err != nil || last().method != http.MethodPatch || last().uri != "/api/v1/partner/domains/dom_1/mailboxes/mbx%2F2" || last().body["displayName"] != "Femi A." {
		t.Fatalf("update mailbox: %v %#v", err, last())
	}
	if _, err := client.Partner.PauseMailbox(ctx, "dom_1", "mbx_2"); err != nil || last().method != http.MethodPost || last().uri != "/api/v1/partner/domains/dom_1/mailboxes/mbx_2/pause" {
		t.Fatalf("pause mailbox: %v %#v", err, last())
	}
	if _, err := client.Partner.ResumeMailbox(ctx, "dom_1", "mbx_2"); err != nil || last().uri != "/api/v1/partner/domains/dom_1/mailboxes/mbx_2/resume" {
		t.Fatalf("resume mailbox: %v %#v", err, last())
	}
	payload = `{"success":true,"data":{"id":"mbx_2","address":"orders@kicksbyfemi.ng","status":"removed"}}`
	if removed, err := client.Partner.RemoveMailbox(ctx, "dom_1", "mbx_2"); err != nil || removed.Status != "removed" || last().method != http.MethodDelete || last().uri != "/api/v1/partner/domains/dom_1/mailboxes/mbx_2" {
		t.Fatalf("remove mailbox: %v %#v", err, last())
	}
	payload = `{"success":true,"data":{"url":"https://mail.kicksbyfemi.ng/webmail/set-password#tok","expiresAt":"2026-11-08T00:00:00Z"}}`
	link, err := client.Partner.CreateMailboxPasswordLink(ctx, "dom_1", "mbx_2", PartnerPasswordLinkInput{Reset: true})
	if err != nil || link.URL == "" || last().uri != "/api/v1/partner/domains/dom_1/mailboxes/mbx_2/password-link" || last().body["reset"] != true || last().body["send"] != false {
		t.Fatalf("password link: %v %#v", err, last())
	}
	payload = `{"success":true,"data":{"id":"als_1","kind":"alias","deliversTo":"hello@kicksbyfemi.ng"}}`
	alias, err := client.Partner.CreateAlias(ctx, "dom_1", PartnerAliasInput{LocalPart: "orders", DeliverTo: "hello@kicksbyfemi.ng"})
	if err != nil || alias.Kind != "alias" || last().uri != "/api/v1/partner/domains/dom_1/aliases" || last().key != "" || last().body["deliverTo"] != "hello@kicksbyfemi.ng" {
		t.Fatalf("create alias: %v %#v", err, last())
	}
	if _, err := client.Partner.UpdateAlias(ctx, "dom_1", "als_1", PartnerAliasUpdate{DeliverTo: "sales@kicksbyfemi.ng"}); err != nil || last().method != http.MethodPatch || last().uri != "/api/v1/partner/domains/dom_1/aliases/als_1" {
		t.Fatalf("update alias: %v %#v", err, last())
	}
	payload = ``
	status = http.StatusNoContent
	if err := client.Partner.RemoveAlias(ctx, "dom_1", "als_1"); err != nil || last().method != http.MethodDelete || last().uri != "/api/v1/partner/domains/dom_1/aliases/als_1" {
		t.Fatalf("remove alias: %v %#v", err, last())
	}
	status = http.StatusOK

	payload = `{"success":true,"data":{"id":"mig_1","state":"copying","startedBy":"partner","source":"google","host":"","switch":{"requestedFor":"tonight","scheduledAt":null,"switchedAt":null,"answeredAt":null},"catchUpAt":null,"completedAt":null,"createdAt":"2026-11-01T09:00:00Z","updatedAt":"2026-11-01T09:05:00Z","emails":{"copied":120,"total":4000,"caughtUp":0},"mailboxes":[{"mailboxId":"mbx_1","address":"hello@kicksbyfemi.ng","username":"femi@kicksbyfemi.ng","state":"copying","copiedEmails":120,"totalEmails":4000,"copiedBytes":1000,"totalBytes":50000,"caughtUpEmails":null,"minutesLeft":12,"message":null}],"records":[{"type":"MX","name":"@","value":"mx.stackshift.cloud","priority":10,"found":false}],"failure":null}}`
	migration, err := client.Partner.StartMailMigration(ctx, "dom_1", PartnerMailMigrationInput{
		Source: "google", Switch: "tonight",
		Accounts: []PartnerMailMigrationAccountInput{{MailboxID: "mbx_1", Username: "femi@kicksbyfemi.ng", Password: "app-password"}},
	}, "migration-1")
	if err != nil || migration.State != "copying" || migration.Mailboxes[0].CopiedEmails != 120 || migration.Emails.Total != 4000 || *migration.Switch.RequestedFor != "tonight" || migration.Records[0].Found || last().method != http.MethodPost || last().uri != "/api/v1/partner/domains/dom_1/mail-migration" || last().key != "migration-1" ||
		last().body["accounts"].([]any)[0].(map[string]any)["mailboxId"] != "mbx_1" || last().body["switch"] != "tonight" {
		t.Fatalf("start migration: %v %#v", err, last())
	}
	if _, err := client.Partner.GetMailMigration(ctx, "dom_1"); err != nil || last().method != http.MethodGet || last().uri != "/api/v1/partner/domains/dom_1/mail-migration" {
		t.Fatalf("get migration: %v %#v", err, last())
	}
	payload = `{"success":true,"data":{"switch":{"requestedFor":"now","scheduledAt":"2026-11-01T09:10:00Z","switchedAt":"2026-11-01T09:10:00Z","answeredAt":null},"migration":null}}`
	if switched, err := client.Partner.SwitchMailMigration(ctx, "dom_1", PartnerMailMigrationSwitchInput{When: "now"}); err != nil || switched.Migration != nil || switched.Switch.SwitchedAt == nil || last().uri != "/api/v1/partner/domains/dom_1/mail-migration/switch" || last().body["when"] != "now" {
		t.Fatalf("switch migration: %v %#v", err, last())
	}
	payload = `{"success":true,"data":{"id":"mig_1","state":"copying"}}`
	if _, err := client.Partner.RetryMailMigration(ctx, "dom_1"); err != nil || last().method != http.MethodPost || last().uri != "/api/v1/partner/domains/dom_1/mail-migration/retry" {
		t.Fatalf("retry migration: %v %#v", err, last())
	}
	if _, err := client.Partner.RetryMailMigrationAccount(ctx, "dom_1", "mbx/1", &PartnerMailMigrationCredentials{Username: "femi@kicksbyfemi.ng", Password: "new-app-password"}); err != nil ||
		last().uri != "/api/v1/partner/domains/dom_1/mail-migration/accounts/mbx%2F1/retry" || last().body["password"] != "new-app-password" {
		t.Fatalf("retry migration account: %v %#v", err, last())
	}
	if _, err := client.Partner.RetryMailMigrationAccount(ctx, "dom_1", "mbx_1", nil); err != nil || last().body != nil {
		t.Fatalf("retry migration account without credentials: %v %#v", err, last())
	}
	if _, err := client.Partner.SwitchTransferNameservers(ctx, "dom/1"); err != nil || last().method != http.MethodPost || last().uri != "/api/v1/partner/domains/dom%2F1/transfer-in/switch-nameservers" {
		t.Fatalf("switch nameservers: %v %#v", err, last())
	}
	if _, err := client.Partner.CancelMailMigration(ctx, "dom_1"); err != nil || last().uri != "/api/v1/partner/domains/dom_1/mail-migration/cancel" {
		t.Fatalf("cancel migration: %v %#v", err, last())
	}
	if _, err := client.Partner.SkipMailMigrationAccount(ctx, "dom_1", "mbx/1"); err != nil || last().method != http.MethodPost || last().uri != "/api/v1/partner/domains/dom_1/mail-migration/accounts/mbx%2F1/skip" {
		t.Fatalf("skip migration account: %v %#v", err, last())
	}

	payload = `{"success":true,"data":{"complained":4,"suppressed":12}}`
	summary, err := client.Partner.GetComplaints(ctx, "brand_1")
	if err != nil || summary.Complained != 4 || summary.Suppressed != 12 || last().uri != "/api/v1/partner/brands/brand_1/complaints" {
		t.Fatalf("complaints: %v %#v", err, last())
	}

	payload = `{"success":true,"data":[{"audienceId":"aud_1","audienceName":"Old list","importedAt":"2026-10-01T00:00:00Z","contacts":500,"neverOpened":320,"complained":4}]}`
	sources, err := client.Partner.ListComplaintSources(ctx, "brand_1")
	if err != nil || len(sources) != 1 || sources[0].NeverOpened != 320 || last().uri != "/api/v1/partner/brands/brand_1/complaints/sources" {
		t.Fatalf("complaint sources: %v %#v", err, last())
	}

	payload = `{"success":true,"data":{"id":"bulk_1","audienceId":"aud_1","reason":"never_opened","suppressed":320}}`
	bulk, err := client.Partner.SuppressNeverOpened(ctx, "brand_1", PartnerBulkSuppressionInput{AudienceID: "aud_1", Reason: "never_opened"})
	if err != nil || bulk.Suppressed != 320 || last().method != http.MethodPost || last().uri != "/api/v1/partner/brands/brand_1/suppressions/bulk" || last().body["audienceId"] != "aud_1" || last().body["reason"] != "never_opened" {
		t.Fatalf("bulk suppress: %v %#v", err, last())
	}

	payload = `{"success":true,"data":{"restored":320}}`
	restored, err := client.Partner.UndoSuppression(ctx, "brand_1", "bulk/1")
	if err != nil || restored != 320 || last().method != http.MethodPost || last().uri != "/api/v1/partner/brands/brand_1/suppressions/bulk/bulk%2F1/undo" {
		t.Fatalf("undo suppression: %v %#v", err, last())
	}

	payload = `{"success":true,"data":{"id":"req_1","kind":"purchase_limit","message":"We onboard 100 merchants a day.","status":"open","resolvedAt":null,"createdAt":"2026-10-24T09:00:00Z","updatedAt":"2026-10-24T09:00:00Z"}}`
	request, err := client.Partner.CreateRequest(ctx, PartnerRequestInput{Kind: "purchase_limit", Message: "We onboard 100 merchants a day."})
	if err != nil || request.Status != "open" || last().method != http.MethodPost || last().uri != "/api/v1/partner/requests" || last().body["kind"] != "purchase_limit" {
		t.Fatalf("create request: %v %#v", err, last())
	}

	payload = `{"success":true,"data":{"requests":[{"id":"req_1","kind":"purchase_limit","message":"","status":"done","resolvedAt":"2026-10-25T09:00:00Z","createdAt":"2026-10-24T09:00:00Z","updatedAt":"2026-10-25T09:00:00Z"}]}}`
	requests, err := client.Partner.ListRequests(ctx)
	if err != nil || len(requests) != 1 || requests[0].Status != "done" || last().method != http.MethodGet || last().uri != "/api/v1/partner/requests" {
		t.Fatalf("list requests: %v %#v", err, last())
	}

	status = http.StatusConflict
	payload = `{"success":false,"message":"You already have an open request of this kind.","error":{"code":"request_already_open"}}`
	_, err = client.Partner.CreateRequest(ctx, PartnerRequestInput{Kind: "dedicated_addresses"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusConflict || apiErr.Code != "request_already_open" {
		t.Fatalf("open request error: %#v", err)
	}
}
