package stackshift

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

type databaseCall struct {
	method string
	uri    string
	key    string
	body   map[string]any
}

const (
	databaseParent = "db_main"
	databaseBranch = "db_branch"
)

func databaseEnvelope(data string) string {
	return `{"success":true,"status":"OK","message":"ok","data":` + data + `}`
}

func databaseOperation(id, kind, status string) string {
	return fmt.Sprintf(`{"id":%q,"branch_database_id":%q,"parent_database_id":%q,"kind":%q,"status":%q,"phase":"provisioning"}`, id, databaseBranch, databaseParent, kind, status)
}

const databaseCredentialsJSON = `{"connection_string":"postgresql://legacy","endpoints":[
{"network":"private","role":"direct","connection_string":"postgresql://private-direct"},
{"network":"private","role":"pooled","connection_string":"postgresql://private-pooled"},
{"network":"public","role":"pooled","connection_string":"postgresql://public-pooled"},
{"network":"public","role":"direct","connection_string":"postgresql://public-direct"}]}`

type databaseReply struct {
	status int
	body   string
}

func databaseClient(t *testing.T, router func(call databaseCall) databaseReply) (*Client, *[]databaseCall) {
	t.Helper()
	var calls []databaseCall
	client, err := New(Options{APIKey: "sspat_fixture", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		call := databaseCall{method: req.Method, uri: strings.TrimPrefix(req.URL.RequestURI(), "/api/v1"), key: req.Header.Get("Idempotency-Key")}
		if req.Body != nil {
			if raw, _ := io.ReadAll(req.Body); len(raw) > 0 {
				_ = json.Unmarshal(raw, &call.body)
			}
		}
		calls = append(calls, call)
		reply := router(call)
		return jsonResponse(reply.status, reply.body), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	return client, &calls
}

func countDatabaseCalls(calls []databaseCall, method string) int {
	count := 0
	for _, call := range calls {
		if call.method == method {
			count++
		}
	}
	return count
}

func withBranchRouter(createStatus string) func(databaseCall) databaseReply {
	return withBranchRouterOptions(createStatus, false, 0)
}

func withBranchRouterOptions(createStatus string, deleteFails bool, creatingPolls int) func(databaseCall) databaseReply {
	createPolls := 0
	return func(call databaseCall) databaseReply {
		if call.method+" "+call.uri == "GET /databases/db_main/branches/operations/op_create" && createStatus == "succeeded" {
			createPolls++
			if createPolls <= creatingPolls {
				return databaseReply{200, databaseEnvelope(databaseOperation("op_create", "create", "running"))}
			}
		}
		if deleteFails && call.method+" "+call.uri == "DELETE /databases/db_main/branches/db_branch" {
			return databaseReply{409, `{"success":false,"message":"The branch is busy.","error":{"code":"busy"}}`}
		}
		switch call.method + " " + call.uri {
		case "POST /databases/db_main/branches":
			return databaseReply{202, databaseEnvelope(databaseOperation("op_create", "create", "pending"))}
		case "GET /databases/db_main/branches/operations/op_create":
			operation := databaseOperation("op_create", "create", createStatus)
			if createStatus == "failed" {
				operation = strings.TrimSuffix(operation, "}") + `,"error_code":"no_capacity","error_message":"No room."}`
			}
			return databaseReply{200, databaseEnvelope(operation)}
		case "GET /databases/db_main/branches/db_branch":
			return databaseReply{200, databaseEnvelope(`{"id":"db_branch","name":"fix-checkout","state":"ready"}`)}
		case "GET /databases/db_branch/credentials":
			return databaseReply{200, databaseEnvelope(databaseCredentialsJSON)}
		case "DELETE /databases/db_main/branches/db_branch":
			return databaseReply{202, databaseEnvelope(databaseOperation("op_delete", "delete", "pending"))}
		case "GET /databases/db_main/branches/operations/op_delete":
			return databaseReply{200, databaseEnvelope(databaseOperation("op_delete", "delete", "succeeded"))}
		}
		return databaseReply{500, `{"message":"unexpected ` + call.method + " " + call.uri + `"}`}
	}
}

func TestDatabaseBranchCreateWaitAndIdempotency(t *testing.T) {
	t.Parallel()

	polls := 0
	client, calls := databaseClient(t, func(call databaseCall) databaseReply {
		if call.method == http.MethodGet {
			polls++
			status := "running"
			if polls >= 3 {
				status = "succeeded"
			}
			return databaseReply{200, databaseEnvelope(databaseOperation("op_1", "create", status))}
		}
		return databaseReply{202, databaseEnvelope(databaseOperation("op_1", "create", "pending"))}
	})
	ctx := context.Background()
	last := func() databaseCall { return (*calls)[len(*calls)-1] }

	created, err := client.Databases.Branches.Create(ctx, databaseParent, DatabaseBranchCreateInput{
		Name: "fix-checkout", From: map[string]string{"time": "2026-10-29T10:00:00Z"}, TTL: "7d", SleepAfterMinutes: 15, CreatedBy: "pr", SourceRef: "pr:142",
	})
	if err != nil || created.Status != "pending" || last().method != http.MethodPost || last().uri != "/databases/db_main/branches" || last().key != "" {
		t.Fatalf("create: %v %#v", err, last())
	}
	if last().body["name"] != "fix-checkout" || last().body["ttl"] != "7d" || last().body["sleep_after_minutes"] != float64(15) || last().body["from"].(map[string]any)["time"] != "2026-10-29T10:00:00Z" {
		t.Fatalf("create body: %#v", last().body)
	}
	done, err := client.Databases.Branches.WaitForOperation(ctx, databaseParent, created.ID, DatabaseBranchWaitOptions{Interval: time.Millisecond})
	if err != nil || done.Status != "succeeded" || polls != 3 || last().uri != "/databases/db_main/branches/operations/op_1" {
		t.Fatalf("wait: %v %#v polls=%d", err, done, polls)
	}

	if _, err := client.Databases.Branches.Create(ctx, databaseParent, DatabaseBranchCreateInput{Name: "keyed", IdempotencyKey: "create-key"}); err != nil || last().key != "create-key" {
		t.Fatalf("create key: %v %#v", err, last())
	}
	if _, err := client.Databases.Branches.Reset(ctx, databaseParent, databaseBranch, "reset-key"); err != nil || last().key != "reset-key" || last().uri != "/databases/db_main/branches/db_branch/reset" {
		t.Fatalf("reset: %v %#v", err, last())
	}
	generated := regexp.MustCompile(`^[0-9a-f]{32}$`)
	if _, err := client.Databases.Branches.Reset(ctx, databaseParent, databaseBranch, ""); err != nil || !generated.MatchString(last().key) {
		t.Fatalf("reset generated key: %v %#v", err, last())
	}
	if _, err := client.Databases.Branches.Retry(ctx, databaseParent, databaseBranch, DatabaseBranchRetryInput{ConfirmChargeMinor: 2192, IdempotencyKey: "retry-key"}); err != nil || last().key != "retry-key" || last().body["confirm_charge_minor"] != float64(2192) || last().method != http.MethodPost || last().uri != "/databases/db_main/branches/db_branch/retry" {
		t.Fatalf("retry: %v %#v", err, last())
	}
	if _, err := client.Databases.Branches.Retry(ctx, databaseParent, databaseBranch, DatabaseBranchRetryInput{}); err != nil || !generated.MatchString(last().key) {
		t.Fatalf("retry generated key: %v %#v", err, last())
	}
	if _, err := client.Databases.Branches.Promote(ctx, databaseParent, databaseBranch, DatabaseBranchPromoteInput{ConfirmName: "orders", IdempotencyKey: "promote-key"}); err != nil ||
		last().key != "promote-key" || last().uri != "/databases/db_main/branches/db_branch/promote" || last().body["confirm_name"] != "orders" {
		t.Fatalf("promote: %v %#v", err, last())
	}
	if _, err := client.Databases.Branches.Promote(ctx, databaseParent, databaseBranch, DatabaseBranchPromoteInput{ConfirmName: "orders"}); err != nil || !generated.MatchString(last().key) {
		t.Fatalf("promote generated key: %v %#v", err, last())
	}
	if _, err := client.Databases.Branches.Delete(ctx, databaseParent, databaseBranch, DatabaseBranchDeleteOptions{Cascade: true, ConfirmName: "fix checkout", IdempotencyKey: "delete-key"}); err != nil ||
		last().method != http.MethodDelete || last().uri != "/databases/db_main/branches/db_branch?cascade=true&confirm=fix+checkout" || last().key != "delete-key" {
		t.Fatalf("delete: %v %#v", err, last())
	}
}

func TestDatabaseBranchOptionsChangesAndScheduledJobs(t *testing.T) {
	t.Parallel()

	payload := ``
	status := 200
	client, calls := databaseClient(t, func(databaseCall) databaseReply { return databaseReply{status, payload} })
	ctx := context.Background()
	last := func() databaseCall { return (*calls)[len(*calls)-1] }

	allowance := `{"used":3,"limit":10,"account_used":5,"account_limit":25,"awake_used":1,"awake_limit":5,"max_ttl_hours":720}`
	payload = databaseEnvelope(`{"branches":[{"id":"db_branch","name":"fix-checkout","parent_name":"shop-db","created_by_user":{"id":"u_1","name":"Ada Okafor"}}],"allowance":` + allowance + `}`)
	listed, err := client.Databases.Branches.List(ctx, databaseParent)
	if err != nil || listed.Allowance.Limit != 10 || listed.Allowance.MaxTTLHours != 720 || listed.Branches[0].ParentName != "shop-db" || listed.Branches[0].CreatedByUser.Name != "Ada Okafor" {
		t.Fatalf("list: %v %#v", err, listed)
	}

	payload = databaseEnvelope(`{"id":"db_branch","name":"fix-checkout","changed_bytes":1048576,"changed_bytes_at":"2026-10-29T12:00:00Z","database":{"id":"db_branch","branch":{"mechanism":"instant"},"sleep":{"enabled":true,"after_minutes":5,"state":"awake","allow_scheduled_jobs":false,"last_wake_failure":{"path":"pooled","outcome":"failed","reason":"storage_full","started_at":"x","held_count":1}}}}`)
	detail, err := client.Databases.Branches.Get(ctx, databaseParent, databaseBranch)
	if err != nil || detail.Database.Branch.Mechanism != "instant" || *detail.ChangedBytes != 1048576 || detail.ChangedBytesAt == "" || detail.Database.Sleep.LastWakeFailure.Reason != "storage_full" {
		t.Fatalf("detail: %v %#v", err, detail)
	}

	payload = databaseEnvelope(`{"starting_points":{"now":{"available":true,"mechanism":"instant"},"time":{"available":false,"unavailable_code":"outside_window"},"restore_point":{"available":true,"mechanism":"instant","choices":[{"id":"rp_1","at":"2026-10-29T09:00:00Z","name":"before-migration","kind":"manual"}]},"backup":{"available":true,"mechanism":"backup","choices":[{"id":"bk_1","at":"2026-10-29T08:00:00Z","size_bytes":79691776}]}},"allowance":` + allowance + `,"default_ttl_hours":168,"default_sleep_after_minutes":5,"copy_limit_bytes":10737418240}`)
	options, err := client.Databases.Branches.Options(ctx, databaseParent)
	if err != nil || last().uri != "/databases/db_main/branches/options" || options.StartingPoints.Time.UnavailableCode != "outside_window" || options.DefaultTTLHours != 168 || options.CopyLimitBytes != 10737418240 ||
		options.StartingPoints.RestorePoint.Choices[0].Name != "before-migration" || options.StartingPoints.Backup.Choices[0].SizeBytes != 79691776 {
		t.Fatalf("options: %v %#v", err, options)
	}

	payload = databaseEnvelope(`{"data_bytes":79691776,"changed_bytes":null,"rows_written":1204,"rows_written_by_table":[{"table":"public.orders","rows":1100}],"table_changes":[{"object":"public.orders.coupon_code","kind":"column","change":"added"}],"table_changes_complete":true,"parent_rows_written_since":null,"paused":{"scheduled_jobs":[{"id":3,"name":"nightly","schedule":"0 1 * * *","active":false}],"subscriptions":["billing_feed"],"outside_servers":[]}}`)
	changes, err := client.Databases.Branches.Changes(ctx, databaseParent, databaseBranch)
	if err != nil || last().uri != "/databases/db_main/branches/db_branch/changes" || changes.RowsWritten != 1204 || changes.ParentRowsWrittenSince != nil || changes.ChangedBytes != nil || changes.Paused.ScheduledJobs[0].ID != 3 {
		t.Fatalf("changes: %v %#v", err, changes)
	}

	payload = databaseEnvelope(`{"scheduled_jobs":[{"id":3,"name":"nightly","schedule":"0 1 * * *","active":true}]}`)
	resumed, err := client.Databases.Branches.ResumeScheduledJobs(ctx, databaseParent, databaseBranch, DatabaseBranchResumeScheduledJobsInput{IDs: []int{3}})
	if err != nil || last().method != http.MethodPost || last().uri != "/databases/db_main/branches/db_branch/scheduled-jobs/resume" || !resumed.ScheduledJobs[0].Active {
		t.Fatalf("resume: %v %#v", err, last())
	}
	if ids, _ := last().body["ids"].([]any); len(ids) != 1 || ids[0] != float64(3) {
		t.Fatalf("resume body: %#v", last().body)
	}

	payload = databaseEnvelope(`{"scheduled_jobs":[],"count":0,"active_count":0}`)
	if _, err := client.Databases.ScheduledJobs(ctx, databaseParent); err != nil || last().uri != "/databases/db_main/scheduled-jobs" {
		t.Fatalf("scheduled jobs: %v %#v", err, last())
	}

	payload = databaseEnvelope(`{"wake_events":[{"path":"pooled","outcome":"failed","reason":"plan_limit","started_at":"x","held_count":4}]}`)
	events, err := client.Databases.WakeEvents(ctx, databaseParent, DatabaseWakeEventOptions{})
	if err != nil || last().uri != "/databases/db_main/wake-events" || events.WakeEvents[0].Reason != "plan_limit" || events.WakeEvents[0].HeldCount != 4 {
		t.Fatalf("wake events: %v %#v", err, events)
	}
	if _, err := client.Databases.WakeEvents(ctx, databaseParent, DatabaseWakeEventOptions{Failed: true, Limit: 1}); err != nil || last().uri != "/databases/db_main/wake-events?failed=true&limit=1" {
		t.Fatalf("wake events filtered: %v %#v", err, last())
	}

	cases := []struct {
		status  int
		code    string
		details string
		call    func() error
	}{
		{409, "asleep", ``, func() error {
			_, err := client.Databases.Branches.Changes(ctx, databaseParent, databaseBranch)
			return err
		}},
		{502, "read_failed", ``, func() error { _, err := client.Databases.ScheduledJobs(ctx, databaseParent); return err }},
		{502, "write_failed", ``, func() error {
			_, err := client.Databases.Branches.ResumeScheduledJobs(ctx, databaseParent, databaseBranch, DatabaseBranchResumeScheduledJobsInput{IDs: []int{3}})
			return err
		}},
		{409, "not_failed", ``, func() error {
			_, err := client.Databases.Branches.Retry(ctx, databaseParent, databaseBranch, DatabaseBranchRetryInput{})
			return err
		}},
		{422, "too_large", `{"size_bytes":20000000000,"limit_bytes":10737418240}`, func() error {
			_, err := client.Databases.Branches.Create(ctx, databaseParent, DatabaseBranchCreateInput{Name: "big"})
			return err
		}},
	}
	for _, item := range cases {
		status = item.status
		payload = `{"success":false,"message":"Refused.","error":{"code":"` + item.code + `"`
		if item.details != "" {
			payload += `,"details":` + item.details
		}
		payload += `}}`
		var apiErr *APIError
		if err := item.call(); !errors.As(err, &apiErr) || apiErr.StatusCode != item.status || apiErr.Code != item.code {
			t.Fatalf("%s: %#v", item.code, err)
		}
		if item.code == "too_large" && (apiErr.Details["size_bytes"] != float64(20000000000) || apiErr.Details["limit_bytes"] != float64(10737418240)) {
			t.Fatalf("too_large details: %#v", apiErr.Details)
		}
	}
}

func TestDatabaseBranchPriceQuoteRoundTrip(t *testing.T) {
	t.Parallel()

	const quote = `{"current_monthly_minor":0,"new_monthly_minor":2192,"charge_minor":2192,"credit_minor":0,"currency":"NGN","billing":"hourly","branch":{"hourly_minor":2192,"storage_gb_month_minor":15000,"currency":"NGN","usd":{"monthly_minor":1231,"hourly":"0.0169","storage_gb_month_minor":12}}}`
	client, calls := databaseClient(t, func(call databaseCall) databaseReply {
		if confirmed, _ := call.body["confirm_charge_minor"].(float64); confirmed >= 2192 {
			return databaseReply{202, databaseEnvelope(databaseOperation("op_dedicated", "create", "pending"))}
		}
		return databaseReply{409, `{"success":false,"error":"price_confirmation_required","message":"Confirm the price first.","quote":` + quote + `}`}
	})
	ctx := context.Background()

	_, err := client.Databases.Branches.Create(ctx, databaseParent, DatabaseBranchCreateInput{Name: "dedicated-copy", InstanceClass: "s-2vcpu-4gb"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 409 || apiErr.Code != "price_confirmation_required" || apiErr.Message != "Confirm the price first." {
		t.Fatalf("quote error: %#v", err)
	}
	if (*calls)[0].body["instance_class"] != "s-2vcpu-4gb" {
		t.Fatalf("instance class: %#v", (*calls)[0].body)
	}
	var parsed DatabaseBranchQuote
	raw, _ := json.Marshal(apiErr.Details["quote"])
	if err := json.Unmarshal(raw, &parsed); err != nil || parsed.ChargeMinor != 2192 || parsed.Currency != "NGN" || parsed.Branch.USD.Hourly != "0.0169" {
		t.Fatalf("quote: %v %#v", err, apiErr.Details)
	}
	created, err := client.Databases.Branches.Create(ctx, databaseParent, DatabaseBranchCreateInput{Name: "dedicated-copy", ConfirmChargeMinor: int(parsed.ChargeMinor)})
	if err != nil || created.ID != "op_dedicated" || (*calls)[len(*calls)-1].body["confirm_charge_minor"] != float64(2192) {
		t.Fatalf("confirmed: %v %#v", err, created)
	}
	if _, err := client.Databases.Branches.Retry(ctx, databaseParent, databaseBranch, DatabaseBranchRetryInput{}); !errors.As(err, &apiErr) || apiErr.Code != "price_confirmation_required" {
		t.Fatalf("retry quote: %#v", err)
	}
	if retried, err := client.Databases.Branches.Retry(ctx, databaseParent, databaseBranch, DatabaseBranchRetryInput{ConfirmChargeMinor: 2192}); err != nil || retried.ID != "op_dedicated" {
		t.Fatalf("retry confirmed: %v %#v", err, retried)
	}
}

func TestDatabaseBranchPlanLimitBodyWithoutTheStandardEnvelope(t *testing.T) {
	t.Parallel()

	body := `{"error":"branch limit reached","code":"plan_limit","message":"You are at 3 Branches.","limit":3,"current":3,"upgrade_url":"/dashboard/billing"}`
	status := 402
	client, _ := databaseClient(t, func(databaseCall) databaseReply { return databaseReply{status, body} })
	ctx := context.Background()

	_, err := client.Databases.Branches.Create(ctx, databaseParent, DatabaseBranchCreateInput{Name: "over"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 402 || apiErr.Code != "plan_limit" || apiErr.Message != "You are at 3 Branches." ||
		apiErr.Details["limit"] != float64(3) || apiErr.Details["current"] != float64(3) || apiErr.Details["upgrade_url"] != "/dashboard/billing" {
		t.Fatalf("plan limit: %#v", err)
	}
	if _, present := apiErr.Details["error"]; present {
		t.Fatalf("details repeat the error field: %#v", apiErr.Details)
	}

	body = `{"success":false,"message":"The plan is unpaid.","error":{"code":"payment_required"}}`
	_, err = client.Databases.Branches.Create(ctx, databaseParent, DatabaseBranchCreateInput{Name: "unpaid"})
	if !errors.As(err, &apiErr) || apiErr.Code != "payment_required" || apiErr.StatusCode != 402 {
		t.Fatalf("payment required: %#v", err)
	}

	status = 500
	body = `{"error":"Something broke"}`
	_, err = client.Databases.Branches.Create(ctx, databaseParent, DatabaseBranchCreateInput{Name: "broken"})
	if !errors.As(err, &apiErr) || apiErr.Code != "" || apiErr.Details != nil {
		t.Fatalf("plain error: %#v", err)
	}
}

func TestDatabaseBillingAndBranchCost(t *testing.T) {
	t.Parallel()

	payload := databaseEnvelope(`{"plan":"dedicated","branch_usage":{"currency":"NGN","owed_minor":0,"stopped":false},"sleep_cost":{"billing":"hourly","currency":"NGN","monthly_minor":160000,"asleep_monthly_minor":110000,"sleep_saves":true,"asleep_hours_per_day":16,"estimate_basis":"assumed","estimate_monthly_minor":120000}}`)
	client, calls := databaseClient(t, func(databaseCall) databaseReply { return databaseReply{200, payload} })
	ctx := context.Background()
	billing, err := client.Databases.Billing(ctx, databaseParent)
	if err != nil || (*calls)[0].uri != "/databases/db_main/billing" || billing.SleepCost == nil || !billing.SleepCost.SleepSaves || billing.SleepCost.EstimateBasis != "assumed" || billing.SleepCost.AsleepHoursPerDay != 16 || billing.Raw["plan"] != "dedicated" || billing.BranchUsage == nil || billing.BranchUsage.Stopped {
		t.Fatalf("billing: %v %#v", err, billing)
	}
	payload = databaseEnvelope(`{"currency":"NGN","owed_minor":5400,"stopped":true,"stopped_at":"2026-10-29T10:00:00Z","retire_at":"2026-11-05T10:00:00Z"}`)
	usage, err := client.Databases.BranchUsage(ctx, databaseParent)
	if err != nil || (*calls)[len(*calls)-1].uri != "/databases/db_main/branch-usage" || usage.OwedMinor != 5400 || !usage.Stopped || usage.RetireAt == "" {
		t.Fatalf("branch usage: %v %#v", err, usage)
	}
	payload = databaseEnvelope(`{"status":"payment_required","message":"Pay to continue.","payment_url":"https://pay.example/abc","payment_reference":"ref_1","amount_minor":5400,"wallet_minor":1000,"currency":"NGN"}`)
	checkout, err := client.Databases.PayBranchUsage(ctx, databaseParent)
	if last := (*calls)[len(*calls)-1]; err != nil || last.method != http.MethodPost || last.uri != "/databases/db_main/branch-usage/pay" || checkout.Status != "payment_required" || checkout.PaymentReference != "ref_1" || checkout.AmountMinor != 5400 {
		t.Fatalf("pay: %v %#v", err, checkout)
	}
	payload = databaseEnvelope(`{"plan":"dedicated"}`)
	if billing, err := client.Databases.Billing(ctx, databaseParent); err != nil || billing.SleepCost != nil {
		t.Fatalf("billing without a plan: %v %#v", err, billing)
	}

	payload = databaseEnvelope(`{"id":"db_branch","name":"fix-checkout","changed_bytes":3565158,"cost":{"currency":"NGN","so_far_minor":2411,"hourly_minor":2192,"awake_seconds":3960,"compute_minor":2411,"changed_bytes":3565158,"storage_monthly_minor":15000,"storage_charged_minor":0,"charged_minor":2411,"usd":{"monthly_minor":1231,"hourly":"0.0169","storage_gb_month_minor":12}}}`)
	branch, err := client.Databases.Branches.Get(ctx, databaseParent, databaseBranch)
	if err != nil || branch.Cost == nil || branch.Cost.ChargedMinor != 2411 || branch.Cost.SoFarMinor != 2411 || branch.Cost.AwakeSeconds != 3960 || *branch.Cost.ChangedBytes != 3565158 || branch.Cost.USD.Hourly != "0.0169" {
		t.Fatalf("branch cost: %v %#v", err, branch)
	}
}

func TestDatabaseBranchDeleteNeedsConfirmationAndAlwaysSendsAKey(t *testing.T) {
	t.Parallel()

	client, calls := databaseClient(t, func(databaseCall) databaseReply {
		return databaseReply{202, databaseEnvelope(databaseOperation("op_delete", "delete", "pending"))}
	})
	ctx := context.Background()
	if _, err := client.Databases.Branches.Delete(ctx, databaseParent, databaseBranch, DatabaseBranchDeleteOptions{Cascade: true}); err == nil || !strings.Contains(err.Error(), "ConfirmName") || len(*calls) != 0 {
		t.Fatalf("cascade without a name: %v calls=%d", err, len(*calls))
	}
	keys := map[string]bool{}
	generated := regexp.MustCompile(`^[0-9a-f]{32}$`)
	for index := 0; index < 3; index++ {
		for _, send := range []func() error{
			func() error {
				_, err := client.Databases.Branches.Reset(ctx, databaseParent, databaseBranch, "")
				return err
			},
			func() error {
				_, err := client.Databases.Branches.Retry(ctx, databaseParent, databaseBranch, DatabaseBranchRetryInput{})
				return err
			},
			func() error {
				_, err := client.Databases.Branches.Promote(ctx, databaseParent, databaseBranch, DatabaseBranchPromoteInput{ConfirmName: "orders"})
				return err
			},
			func() error {
				_, err := client.Databases.Branches.Delete(ctx, databaseParent, databaseBranch, DatabaseBranchDeleteOptions{})
				return err
			},
		} {
			if err := send(); err != nil {
				t.Fatal(err)
			}
			key := (*calls)[len(*calls)-1].key
			if !generated.MatchString(key) {
				t.Fatalf("key: %q", key)
			}
			keys[key] = true
		}
	}
	if len(keys) != 12 {
		t.Fatalf("keys were reused: %d distinct of 12", len(keys))
	}
}

func TestDatabaseWithBranchCleanupKeepsTheOriginalErrorAndWaitsForCreate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	wait := DatabaseBranchWaitOptions{Interval: time.Millisecond}

	client, calls := databaseClient(t, withBranchRouterOptions("succeeded", true, 0))
	boom := errors.New("test failed")
	err := client.Databases.Branches.WithBranch(ctx, databaseParent, func(context.Context, DatabaseWithBranch) error { return boom }, DatabaseWithBranchOptions{Wait: wait})
	var apiErr *APIError
	if !errors.Is(err, boom) || !errors.As(err, &apiErr) || apiErr.Code != "busy" || countDatabaseCalls(*calls, http.MethodDelete) != 1 {
		t.Fatalf("delete failure: %v", err)
	}
	if first := strings.SplitN(err.Error(), "\n", 2)[0]; first != "test failed" {
		t.Fatalf("original error is not first: %q", err.Error())
	}

	client, _ = databaseClient(t, withBranchRouterOptions("succeeded", true, 0))
	err = client.Databases.Branches.WithBranch(ctx, databaseParent, func(context.Context, DatabaseWithBranch) error { return nil }, DatabaseWithBranchOptions{Wait: wait})
	if !errors.As(err, &apiErr) || apiErr.Code != "busy" {
		t.Fatalf("cleanup failure after success: %v", err)
	}

	client, calls = databaseClient(t, withBranchRouterOptions("succeeded", false, 4))
	err = client.Databases.Branches.WithBranch(ctx, databaseParent, func(context.Context, DatabaseWithBranch) error {
		t.Fatal("fn must not run")
		return nil
	}, DatabaseWithBranchOptions{Wait: DatabaseBranchWaitOptions{Timeout: time.Nanosecond, Interval: time.Millisecond}})
	if err == nil || !strings.Contains(err.Error(), "still running") || countDatabaseCalls(*calls, http.MethodDelete) != 1 {
		t.Fatalf("slow create: %v", err)
	}
	createPolls, deleteAt, lastCreatePoll := 0, -1, -1
	for index, call := range *calls {
		if call.uri == "/databases/db_main/branches/operations/op_create" {
			createPolls++
			lastCreatePoll = index
		}
		if call.method == http.MethodDelete {
			deleteAt = index
		}
	}
	if createPolls != 5 || deleteAt < lastCreatePoll || (*calls)[len(*calls)-1].uri != "/databases/db_main/branches/operations/op_delete" {
		t.Fatalf("order: polls=%d delete=%d last poll=%d", createPolls, deleteAt, lastCreatePoll)
	}

	client, calls = databaseClient(t, withBranchRouterOptions("succeeded", false, 1000))
	err = client.Databases.Branches.WithBranch(ctx, databaseParent, func(context.Context, DatabaseWithBranch) error {
		t.Fatal("fn must not run")
		return nil
	}, DatabaseWithBranchOptions{Wait: DatabaseBranchWaitOptions{Timeout: time.Nanosecond, Interval: time.Millisecond}, CleanupTimeout: time.Nanosecond})
	if err == nil || strings.Count(err.Error(), "still running") != 2 || countDatabaseCalls(*calls, http.MethodDelete) != 0 {
		t.Fatalf("cleanup timeout: %v", err)
	}
}

func TestDatabaseBranchPayerFieldsAndRefusalCodes(t *testing.T) {
	t.Parallel()

	payload := databaseEnvelope(`{"currency":"NGN","owed_minor":5400,"stopped":true,"payer_user_id":"u_1","payer_name":"Ada Okafor"}`)
	status := 200
	client, _ := databaseClient(t, func(databaseCall) databaseReply { return databaseReply{status, payload} })
	ctx := context.Background()
	usage, err := client.Databases.BranchUsage(ctx, databaseParent)
	if err != nil || usage.PayerUserID != "u_1" || usage.PayerName != "Ada Okafor" {
		t.Fatalf("usage: %v %#v", err, usage)
	}
	payload = databaseEnvelope(`{"status":"paid","message":"Paid.","amount_minor":5400,"wallet_minor":5400,"currency":"NGN","payer_user_id":"u_1","payer_name":"Ada Okafor"}`)
	checkout, err := client.Databases.PayBranchUsage(ctx, databaseParent)
	if err != nil || checkout.PayerName != "Ada Okafor" || checkout.PayerUserID != "u_1" {
		t.Fatalf("checkout: %v %#v", err, checkout)
	}
	for _, item := range []struct {
		status int
		code   string
	}{{403, "owner_required"}, {429, "too_soon"}, {409, "not_running"}, {409, "wake_unavailable"}, {400, "confirm_name"}} {
		status = item.status
		payload = `{"success":false,"message":"Refused.","error":{"code":"` + item.code + `"}}`
		var apiErr *APIError
		_, err := client.Databases.Branches.Promote(ctx, databaseParent, databaseBranch, DatabaseBranchPromoteInput{ConfirmName: "x"})
		if !errors.As(err, &apiErr) || apiErr.StatusCode != item.status || apiErr.Code != item.code {
			t.Fatalf("%s: %v", item.code, err)
		}
	}
}

func TestDatabaseBranchWaitTimesOut(t *testing.T) {
	t.Parallel()

	client, _ := databaseClient(t, func(databaseCall) databaseReply {
		return databaseReply{200, databaseEnvelope(databaseOperation("op_slow", "create", "running"))}
	})
	_, err := client.Databases.Branches.WaitForOperation(context.Background(), databaseParent, "op_slow", DatabaseBranchWaitOptions{Timeout: 20 * time.Millisecond, Interval: 5 * time.Millisecond})
	if err == nil || !strings.Contains(err.Error(), "still running") {
		t.Fatalf("timeout: %v", err)
	}
}

func TestDatabaseBranchRefusalsCarryTheErrorCode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		status  int
		code    string
		details string
	}{
		{402, "plan_limit", `{"limit":3}`},
		{409, "no_capacity", ``},
		{422, "outside_window", ``},
		{422, "too_large", ``},
		{409, "price_confirmation_required", `{"price":1000}`},
	}
	for _, item := range cases {
		body := `{"success":false,"message":"Refused: ` + item.code + `","error":{"code":"` + item.code + `"`
		if item.details != "" {
			body += `,"details":` + item.details
		}
		body += `}}`
		client, _ := databaseClient(t, func(databaseCall) databaseReply { return databaseReply{item.status, body} })
		_, err := client.Databases.Branches.Create(context.Background(), databaseParent, DatabaseBranchCreateInput{Name: "refused"})
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != item.status || apiErr.Code != item.code || apiErr.Message != "Refused: "+item.code {
			t.Fatalf("%s: %#v", item.code, err)
		}
		if item.code == "plan_limit" && apiErr.Details["limit"] != float64(3) {
			t.Fatalf("plan_limit details: %#v", apiErr.Details)
		}
	}
}

func TestDatabaseBranchURLPrefersThePooledConnection(t *testing.T) {
	t.Parallel()

	credentials := databaseCredentialsJSON
	client, calls := databaseClient(t, func(databaseCall) databaseReply { return databaseReply{200, databaseEnvelope(credentials)} })
	ctx := context.Background()
	if got, err := client.Databases.Branches.URL(ctx, databaseParent, databaseBranch, false); err != nil || got != "postgresql://public-pooled" || (*calls)[0].uri != "/databases/db_branch/credentials" {
		t.Fatalf("pooled: %v %q %#v", err, got, *calls)
	}
	if got, err := client.Databases.Branches.URL(ctx, databaseParent, databaseBranch, true); err != nil || got != "postgresql://public-direct" {
		t.Fatalf("direct: %v %q", err, got)
	}
	credentials = `{"connection_string":"postgresql://legacy"}`
	if got, err := client.Databases.Branches.URL(ctx, databaseParent, databaseBranch, true); err != nil || got != "postgresql://legacy" {
		t.Fatalf("legacy direct: %v %q", err, got)
	}
	if _, err := client.Databases.Branches.URL(ctx, databaseParent, databaseBranch, false); err == nil || !strings.Contains(err.Error(), "no shared connection string") {
		t.Fatalf("missing pooled: %v", err)
	}
}

func TestDatabaseSleepWakeAndMove(t *testing.T) {
	t.Parallel()

	client, calls := databaseClient(t, func(databaseCall) databaseReply {
		return databaseReply{200, databaseEnvelope(`{"enabled":true,"after_minutes":15,"state":"awake","allow_scheduled_jobs":false}`)}
	})
	ctx := context.Background()
	last := func() databaseCall { return (*calls)[len(*calls)-1] }
	sleep, err := client.Databases.Sleep(ctx, databaseParent, DatabaseSleepInput{Enabled: true, AfterMinutes: 15})
	if err != nil || sleep.AfterMinutes != 15 || last().method != http.MethodPut || last().uri != "/databases/db_main/sleep" || last().body["after_minutes"] != float64(15) {
		t.Fatalf("sleep: %v %#v", err, last())
	}
	if _, err := client.Databases.Sleep(ctx, databaseParent, DatabaseSleepInput{Enabled: false}); err != nil || last().body["enabled"] != false || len(last().body) != 1 {
		t.Fatalf("sleep off: %v %#v", err, last())
	}
	if _, err := client.Databases.Wake(ctx, databaseParent, false); err != nil || last().method != http.MethodPost || last().uri != "/databases/db_main/wake" {
		t.Fatalf("wake: %v %#v", err, last())
	}
	if _, err := client.Databases.Wake(ctx, databaseParent, true); err != nil || last().uri != "/databases/db_main/wake?wait=true" {
		t.Fatalf("wake wait: %v %#v", err, last())
	}
	if _, err := client.Databases.MoveToInstantBranches(ctx, databaseParent); err != nil || last().method != http.MethodPost || last().uri != "/databases/db_main/move-to-instant-branches" {
		t.Fatalf("move: %v %#v", err, last())
	}
	if _, err := client.Databases.InstantBranchesMove(ctx, databaseParent); err != nil || last().method != http.MethodGet || last().uri != "/databases/db_main/move-to-instant-branches" {
		t.Fatalf("move status: %v %#v", err, last())
	}
}

func TestDatabaseWithBranchDeletesAfterSuccessAndFailure(t *testing.T) {
	t.Parallel()

	wait := DatabaseBranchWaitOptions{Interval: time.Millisecond}
	ctx := context.Background()

	client, calls := databaseClient(t, withBranchRouter("succeeded"))
	err := client.Databases.Branches.WithBranch(ctx, databaseParent, func(_ context.Context, branch DatabaseWithBranch) error {
		if branch.Branch.ID != databaseBranch || branch.URL != "postgresql://public-pooled" || countDatabaseCalls(*calls, http.MethodDelete) != 0 {
			t.Fatalf("branch: %#v calls=%#v", branch, *calls)
		}
		return nil
	}, DatabaseWithBranchOptions{Wait: wait})
	if err != nil || countDatabaseCalls(*calls, http.MethodDelete) != 1 || (*calls)[len(*calls)-1].uri != "/databases/db_main/branches/operations/op_delete" {
		t.Fatalf("success: %v %#v", err, *calls)
	}
	if (*calls)[0].body["ttl"] != "24h" || !regexp.MustCompile(`^test-[0-9a-f]{8}$`).MatchString((*calls)[0].body["name"].(string)) {
		t.Fatalf("defaults: %#v", (*calls)[0].body)
	}

	client, calls = databaseClient(t, withBranchRouter("succeeded"))
	err = client.Databases.Branches.WithBranch(ctx, databaseParent, func(_ context.Context, branch DatabaseWithBranch) error {
		if branch.URL != "postgresql://public-direct" {
			t.Fatalf("direct url: %q", branch.URL)
		}
		return nil
	}, DatabaseWithBranchOptions{Create: DatabaseBranchCreateInput{Name: "ci-run", TTL: "7d"}, Direct: true, Wait: wait})
	if err != nil || (*calls)[0].body["name"] != "ci-run" || (*calls)[0].body["ttl"] != "7d" {
		t.Fatalf("options: %v %#v", err, *calls)
	}

	client, calls = databaseClient(t, withBranchRouter("succeeded"))
	boom := errors.New("test failed")
	err = client.Databases.Branches.WithBranch(ctx, databaseParent, func(context.Context, DatabaseWithBranch) error { return boom }, DatabaseWithBranchOptions{Wait: wait})
	if !errors.Is(err, boom) || countDatabaseCalls(*calls, http.MethodDelete) != 1 || (*calls)[len(*calls)-1].uri != "/databases/db_main/branches/operations/op_delete" {
		t.Fatalf("failure: %v %#v", err, *calls)
	}

	client, calls = databaseClient(t, withBranchRouter("failed"))
	err = client.Databases.Branches.WithBranch(ctx, databaseParent, func(context.Context, DatabaseWithBranch) error {
		t.Fatal("fn must not run")
		return nil
	}, DatabaseWithBranchOptions{Wait: wait})
	var failed *DatabaseOperationFailedError
	if !errors.As(err, &failed) || failed.Code() != "no_capacity" || failed.Error() != "No room." || countDatabaseCalls(*calls, http.MethodDelete) != 1 {
		t.Fatalf("create failed: %v %#v", err, *calls)
	}

	client, calls = databaseClient(t, func(databaseCall) databaseReply {
		return databaseReply{402, `{"success":false,"message":"You are at the Branch limit.","error":{"code":"plan_limit","details":{"limit":2}}}`}
	})
	err = client.Databases.Branches.WithBranch(ctx, databaseParent, func(context.Context, DatabaseWithBranch) error {
		t.Fatal("fn must not run")
		return nil
	})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "plan_limit" || len(*calls) != 1 {
		t.Fatalf("refused: %v %#v", err, *calls)
	}
}
