package stackshift

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type recordedAgreementCall struct {
	method string
	path   string
	query  string
	body   string
	key    string
}

func newAgreementTestClient(t *testing.T, calls *[]recordedAgreementCall, responses ...string) *Client {
	t.Helper()
	next := 0
	client, err := New(Options{APIKey: "sk_test", HTTPClient: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var body string
		if request.Body != nil {
			raw, _ := io.ReadAll(request.Body)
			body = string(raw)
		}
		*calls = append(*calls, recordedAgreementCall{method: request.Method, path: request.URL.Path, query: request.URL.RawQuery, body: body, key: request.Header.Get("Idempotency-Key")})
		response := responses[len(responses)-1]
		if next < len(responses) {
			response = responses[next]
			next++
		}
		return jsonResponse(http.StatusOK, response), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestAgreementsListGetAndRecord(t *testing.T) {
	var calls []recordedAgreementCall
	client := newAgreementTestClient(t, &calls,
		`{"threads":[{"id":"thr_1","inboxId":"inb_1","state":"waiting_for_you","counterparty":{"address":"ada@bakes.test","check":"signed"},"what":"200 kg flour","amount":{"amount":"180000000","currency":"NGN"},"direction":"buy","openOffer":{"id":"off_1","type":"quote","expiresAt":"2026-11-08T09:00:00Z","fromUs":false},"waiting":{"kind":"approval","approvalId":"apr_1"}}],"nextCursor":"c2","counts":{"open":1,"waiting_for_you":2,"agreed":3,"ended":4}}`,
		`{"thread":{"id":"thr_1","state":"agreed","recordKind":"two_signatures"},"messages":[{"id":"msg_1","direction":"in","type":"quote","check":"signed","items":[{"description":"flour","quantity":"200","unit":"kg"}],"total":{"amount":"180000000","currency":"NGN"},"note":"hello"}],"record":{"id":"rec_1","kind":"two_signatures","hash":"abc","check":{"result":"checks_out_two","signers":[{"orgName":"Mills","kid":"k1"}]}},"replies":[{"mailMessageId":"m1","text":"yes","offerId":"off_1"}],"changeCursor":"cur_9"}`,
		`{"id":"rec_1","kind":"two_signatures","seal":"jws"}`,
	)
	ctx := context.Background()

	list, err := client.Mail.Agreements.List(ctx, ListAgreementsParams{InboxID: "inb_1", State: "open", Counterparty: "bakes.test", Cursor: "c1", Limit: 20})
	if err != nil || len(list.Threads) != 1 || list.NextCursor != "c2" || list.Counts.WaitingForYou != 2 || list.Threads[0].OpenOffer.ID != "off_1" || list.Threads[0].Waiting.ApprovalID != "apr_1" {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	last := calls[len(calls)-1]
	if last.method != http.MethodGet || last.path != "/api/v1/mail/agreements" || last.query != "counterparty=bakes.test&cursor=c1&inboxId=inb_1&limit=20&state=open" {
		t.Fatalf("call=%+v", last)
	}

	inspect, err := client.Mail.Agreements.Get(ctx, "thr_1")
	if err != nil || inspect.ChangeCursor != "cur_9" || inspect.Record.Check.Result != "checks_out_two" || inspect.Messages[0].Items[0].Unit != "kg" || inspect.Replies[0].OfferID != "off_1" {
		t.Fatalf("inspect=%+v err=%v", inspect, err)
	}
	if last = calls[len(calls)-1]; last.path != "/api/v1/mail/agreements/thr_1" {
		t.Fatalf("call=%+v", last)
	}

	record, err := client.Mail.Agreements.Record(ctx, "thr_1")
	if err != nil || last.method != http.MethodGet || calls[len(calls)-1].path != "/api/v1/mail/agreements/thr_1/record" {
		t.Fatalf("calls=%+v err=%v", calls, err)
	}
	var decoded map[string]string
	if err := json.Unmarshal(record, &decoded); err != nil || decoded["seal"] != "jws" {
		t.Fatalf("record=%s err=%v", record, err)
	}
}

func TestAgreementsWait(t *testing.T) {
	var calls []recordedAgreementCall
	client := newAgreementTestClient(t, &calls,
		`{"timedOut":true,"changeCursor":"cur_1"}`,
		`{"thread":{"id":"thr_1","state":"open"},"messages":[],"record":null,"replies":[],"changeCursor":"cur_2"}`,
	)
	ctx := context.Background()

	timedOut, err := client.Mail.Agreements.Wait(ctx, "thr_1", "cur_0", 30)
	if err != nil || !timedOut.TimedOut || timedOut.ChangeCursor != "cur_1" {
		t.Fatalf("timedOut=%+v err=%v", timedOut, err)
	}
	if last := calls[len(calls)-1]; last.path != "/api/v1/mail/agreements/thr_1/wait" || last.query != "after=cur_0&timeout=30" {
		t.Fatalf("call=%+v", last)
	}
	changed, err := client.Mail.Agreements.Wait(ctx, "thr_1", "cur_1", 0)
	if err != nil || changed.TimedOut || changed.Thread.ID != "thr_1" || changed.ChangeCursor != "cur_2" || calls[len(calls)-1].query != "after=cur_1" {
		t.Fatalf("changed=%+v calls=%+v err=%v", changed, calls, err)
	}
}

func TestAgreementsWaitForChange(t *testing.T) {
	var calls []recordedAgreementCall
	client := newAgreementTestClient(t, &calls,
		`{"timedOut":true,"changeCursor":"cur_1"}`,
		`{"timedOut":true,"changeCursor":"cur_2"}`,
		`{"thread":{"id":"thr_1","state":"agreed"},"messages":[],"record":null,"replies":[],"changeCursor":"cur_3"}`,
	)

	inspect, err := client.Mail.Agreements.WaitForChange(context.Background(), "thr_1", "cur_0")
	if err != nil || inspect.Thread.State != "agreed" || inspect.ChangeCursor != "cur_3" {
		t.Fatalf("inspect=%+v err=%v", inspect, err)
	}
	if len(calls) != 3 || calls[0].query != "after=cur_0" || calls[1].query != "after=cur_1" || calls[2].query != "after=cur_2" {
		t.Fatalf("calls=%+v", calls)
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.Mail.Agreements.WaitForChange(cancelled, "thr_1", "cur_3"); err == nil {
		t.Fatal("expected context error")
	}
}

func TestAgreementsPropose(t *testing.T) {
	var calls []recordedAgreementCall
	client := newAgreementTestClient(t, &calls, `{"threadId":"thr_1","messageId":"msg_1","state":"waiting_for_you","approvalId":"apr_1"}`)
	ctx := context.Background()

	result, err := client.Mail.Agreements.Propose(ctx, "inb_1", ProposeAgreementInput{
		Type:           "quote",
		To:             "ada@bakes.test",
		Items:          []AgreementItem{{Description: "flour", Quantity: "200", Unit: "kg", UnitPrice: &AgreementMoney{Amount: "900000", Currency: "NGN"}}},
		Total:          &AgreementMoney{Amount: "180000000", Currency: "NGN"},
		DeliveryBy:     "2026-11-10",
		OpenForSeconds: 86400,
	}, AgreementSendOptions{IdempotencyKey: "idem-1"})
	if err != nil || result.ThreadID != "thr_1" || result.State != "waiting_for_you" || result.ApprovalID != "apr_1" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	last := calls[len(calls)-1]
	want := `{"type":"quote","to":"ada@bakes.test","items":[{"description":"flour","quantity":"200","unit":"kg","unit_price":{"amount":"900000","currency":"NGN"}}],"total":{"amount":"180000000","currency":"NGN"},"deliveryBy":"2026-11-10","openForSeconds":86400}`
	if last.method != http.MethodPost || last.path != "/api/v1/mail/agent-inboxes/inb_1/agreements" || last.key != "idem-1" || last.body != want {
		t.Fatalf("call=%+v", last)
	}

	before := len(calls)
	if _, err := client.Mail.Agreements.Propose(ctx, "inb_1", ProposeAgreementInput{Type: "ask_quote"}, AgreementSendOptions{}); err == nil || len(calls) != before {
		t.Fatalf("expected missing key error without a request, err=%v", err)
	}
}

func TestAgreementsRespondCancelAndCountReply(t *testing.T) {
	var calls []recordedAgreementCall
	client := newAgreementTestClient(t, &calls, `{"threadId":"thr_1","messageId":"msg_2","state":"sent"}`)
	ctx := context.Background()

	slot := 0
	if _, err := client.Mail.Agreements.Accept(ctx, "thr_1", "off_1", &slot, AgreementSendOptions{IdempotencyKey: "k-accept"}); err != nil {
		t.Fatal(err)
	}
	last := calls[len(calls)-1]
	if last.method != http.MethodPost || last.path != "/api/v1/mail/agreements/thr_1/respond" || last.key != "k-accept" || last.body != `{"offerId":"off_1","decision":"accept","slot":0}` {
		t.Fatalf("accept=%+v", last)
	}

	if _, err := client.Mail.Agreements.Decline(ctx, "thr_1", "off_1", "price", "too high", AgreementSendOptions{IdempotencyKey: "k-decline"}); err != nil {
		t.Fatal(err)
	}
	if last = calls[len(calls)-1]; last.key != "k-decline" || last.body != `{"offerId":"off_1","decision":"decline","reason":"price","note":"too high"}` {
		t.Fatalf("decline=%+v", last)
	}

	if _, err := client.Mail.Agreements.Cancel(ctx, "thr_1", CancelAgreementInput{OfferID: "off_1", Note: "changed plans"}, AgreementSendOptions{IdempotencyKey: "k-cancel"}); err != nil {
		t.Fatal(err)
	}
	if last = calls[len(calls)-1]; last.path != "/api/v1/mail/agreements/thr_1/cancel" || last.key != "k-cancel" || last.body != `{"offerId":"off_1","note":"changed plans"}` {
		t.Fatalf("cancel=%+v", last)
	}

	result, err := client.Mail.Agreements.CountReply(ctx, "thr_1", "off_1", "m1", AgreementSendOptions{IdempotencyKey: "k-count"})
	if err != nil || result.State != "sent" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if last = calls[len(calls)-1]; last.path != "/api/v1/mail/agreements/thr_1/count-reply" || last.key != "k-count" || last.body != `{"offerId":"off_1","replyMessageId":"m1"}` {
		t.Fatalf("count=%+v", last)
	}
}

func TestAgreementsLimitsAndCheck(t *testing.T) {
	var calls []recordedAgreementCall
	client := newAgreementTestClient(t, &calls,
		`{"inboxId":"inb_1","inboxAddress":"buyer@agents.test","enabled":true,"allowedTypes":["quote","decline"],"money":{"NGN":{"buyPerDeal":"150000000","buyPerDay":"500000000","sellPerDeal":"0","sellPerDay":"0"}},"counterparties":"listed","counterpartyList":["ada@bakes.test"],"maxOpenSeconds":86400,"timeWindow":{"weekdays":[1,2,3],"from":"09:00","until":"17:00","timeZone":"Africa/Lagos","earliestDaysAhead":1,"latestDaysAhead":14,"maxSlotMinutes":60},"version":4,"usedToday":{"NGN":{"buy":"1000","sell":"0"}},"sentLastHour":1,"sentToday":2,"newAccount":true}`,
		`{"result":"checks_out_two","summary":"Checks out.","signers":[{"orgName":"Mills","party":"p1","kind":"stackshift_workspace","kid":"k1","revoked":false}],"terms":{"type":"quote"},"warnings":[]}`,
		`{"result":"doesnt_check_out","reason":"bad seal"}`,
	)
	ctx := context.Background()

	limits, err := client.Mail.Agreements.GetLimits(ctx, "inb_1")
	if err != nil || !limits.Enabled || limits.Money["NGN"].BuyPerDeal != "150000000" || limits.TimeWindow.TimeZone != "Africa/Lagos" || limits.UsedToday["NGN"].Buy != "1000" || !limits.NewAccount || limits.Version != 4 {
		t.Fatalf("limits=%+v err=%v", limits, err)
	}
	if last := calls[len(calls)-1]; last.method != http.MethodGet || last.path != "/api/v1/mail/agent-inboxes/inb_1/agreement-limits" {
		t.Fatalf("call=%+v", last)
	}

	checked, err := client.Mail.Agreements.Check(ctx, json.RawMessage(`{"id":"rec_1","seal":"jws"}`))
	if err != nil || checked.Result != "checks_out_two" || checked.Signers[0].OrgName != "Mills" || string(checked.Terms) != `{"type":"quote"}` {
		t.Fatalf("checked=%+v err=%v", checked, err)
	}
	if last := calls[len(calls)-1]; last.method != http.MethodPost || last.path != "/api/v1/public/agreements/check" || last.key != "" || !strings.Contains(last.body, `"seal":"jws"`) {
		t.Fatalf("call=%+v", last)
	}

	failed, err := client.Mail.Agreements.CheckJWS(ctx, "a.b.c")
	if err != nil || failed.Result != "doesnt_check_out" || calls[len(calls)-1].body != `{"jws":"a.b.c"}` {
		t.Fatalf("failed=%+v calls=%+v err=%v", failed, calls, err)
	}
}
