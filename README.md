# StackShift Go SDK

Official Go SDK for StackShift.

## Install

```bash
go get github.com/stackshiftCloud/assets-go
```

## AI credits

AI credits are account-scoped. Use the shared client to inspect usage, start a top-up, or select managed billing versus your own provider key:

```go
client, err := stackshift.New(stackshift.Options{})
ctx := context.Background()
balance, err := client.AICredits.Balance(ctx)
usage, err := client.AICredits.Usage(ctx, url.Values{"product": {"assets"}, "limit": {"25"}})
quote, err := client.AICredits.Quote(ctx, 1_000, "USD")
checkout, err := client.AICredits.Checkout(ctx, quote.ID, uuid.NewString(), "paystack")
funding, err := client.AICredits.SetFunding(ctx, "assets", "managed")
```

Keep account API keys and checkout creation on your server.

## Partner platforms

Staff-enabled partner accounts buy domains for merchants from their StackShift wallet. Use an API token with the `partner:manage` scope (and `mail:manage` to run campaigns) on your server:

```go
domain, err := client.Partner.PurchaseDomain(ctx, stackshift.PartnerDomainPurchaseInput{
	Domain: "merchantshop.com", Years: 1, MerchantRef: "m_123",
	Registrant: stackshift.PartnerRegistrant{FirstName: "Ada", LastName: "Obi", Email: "ada@merchantshop.com", Phone: "+2348012345678", Address: "1 Marina", City: "Lagos", State: "Lagos", PostalCode: "101001", Country: "NG"},
	Brand:      stackshift.PartnerBrandInput{Name: "Merchant Shop", PostalAddress: "1 Marina, Lagos, Nigeria"},
	Replies:    stackshift.PartnerRepliesInput{Mode: "forward", ForwardTo: "ada@gmail.com"},
}, "order-m_123-1")
status, err := client.Partner.GetDomain(ctx, domain.ID)
```

`Partner` also exposes `Overview`, `UpdateSettings`, `ListDomains` (with `Query`, `Attention` and `Website` to find the merchants that need you), `ConnectDomain` and `TransferInDomain` / `TransferInConnectedDomain` / `SwitchTransferNameservers` for domains the merchant already owns, `SetWebsite`, `RemoveWebsite`, `CheckWebsite` and `ListWebsiteProjects` for the storefront, `ListMailboxes`, `CreateMailbox`, `UpdateMailbox`, `CreateMailboxPasswordLink`, `PauseMailbox`, `ResumeMailbox`, `RemoveMailbox`, `CreateAlias`, `UpdateAlias` and `RemoveAlias` for merchant mailboxes, `StartMailMigration`, `GetMailMigration`, `SwitchMailMigration`, `CancelMailMigration`, `RetryMailMigration`, `RetryMailMigrationAccount` and `SkipMailMigrationAccount` for moving old email in, `RetryDomain`, `GetDiagnosis`, `RestoreDNS`, `TransferOut`, `ResumeBrand`, `GetComplaints`, `ListComplaintSources`, `SuppressNeverOpened`, `UndoSuppression`, `Usage`, `ListRequests` and `CreateRequest`. `TransferOut` needs the `partner:transfer` scope (or `*`); `partner:manage` alone is refused. Purchases are capped at 30 per minute and, per partner, at 25 domains in a rolling 24 hours (429 `daily_purchase_limit`; staff can raise it). Validation errors are 400, state conflicts are 409.

## Connected agents

An agent that holds a connection credential (issued by the owner, not an account API key) can use the same API the MCP tools and the CLI use:

```go
client, _ := stackshift.New(stackshift.Options{APIKey: os.Getenv("STACKSHIFT_CONNECTION_CREDENTIAL")})
operation, err := client.Agents.Connections.Operations.Propose(ctx, stackshift.AgentConnectionOperationInput{
	Action: "deployment.restart", ResourceID: projectID, Reason: "Restart after the config fix", IdempotencyKey: "restart-shop-1",
})
wait, err := client.Agents.Connections.Operations.Wait(ctx, operation.ID, 0)
for err == nil && !wait.Settled {
	wait, err = client.Agents.Connections.Operations.Wait(ctx, operation.ID, 0)
}
```

`Agents.Connections` has `Identity`, `Resources`, `Capabilities`, `Plan`, `TaskFacts`, `History` and `LowerLevel`; `Agents.Connections.Operations` has `List`, `Propose`, `Get`, `Wait` and `Approval`. `Propose` needs an `IdempotencyKey`: keep it and send the same one to retry. `Wait` returns `Settled: false` when the owner has not decided yet; call it again instead of proposing again. `Approval` says whether the owner has decided, with `ReviewURL` for the owner. `LowerLevel` only makes the connection stricter; raising a level returns `*APIError` with `Code` `session_required`.

## Databases: Branches and Sleep when idle

Branches give you a copy of a database for a test run, a preview or a pull request. `WithBranch` makes one, waits until it is ready, runs your function and always deletes it afterwards:

```go
err := client.Databases.Branches.WithBranch(ctx, databaseID, func(ctx context.Context, branch stackshift.DatabaseWithBranch) error {
	return runMigrationsAndTests(ctx, branch.URL)
})
```

If your function returns an error, that error comes first in the result (a failed cleanup is joined to it); a slow create is waited out before the branch is deleted (`CleanupTimeout`, default 10 minutes). `URL` is the shared (pooled) connection string; set `Direct` in `DatabaseWithBranchOptions` for the direct one. `Databases.Branches` also has `List`, `Create`, `Get`, `Update`, `Delete`, `Reset`, `Retry` (a failed Branch), `Promote`, `Operation`, `WaitForOperation`, `URL`, `Options` (what a new Branch can start from), `Changes` and `ResumeScheduledJobs`. `List` also returns your plan `Allowance`. `Create`, `Reset`, `Retry`, `Promote` and `Delete` return an operation; `WaitForOperation` polls it until it succeeds or fails. `Reset`, `Retry`, `Promote` and `Delete` send an `Idempotency-Key` (a new one for each call if you leave it empty). Deleting with `Cascade` also needs `ConfirmName` (the branch's name), or an error is returned before the call is sent. `Promote` accepts the parent's name or, when the parent is a branch, its short name. `Databases.Sleep`, `Databases.Wake`, `Databases.Billing`, `Databases.BranchUsage`, `Databases.PayBranchUsage`, `Databases.ScheduledJobs`, `Databases.WakeEvents`, `Databases.MoveToInstantBranches` and `Databases.InstantBranchesMove` cover Sleep when idle and the move to instant Branches. Refusals return `*APIError` with `Code` set to values such as `plan_limit`, `no_capacity`, `outside_window`, `too_large` and `payment_required`. On a Dedicated database, `Create` and `Retry` first return `price_confirmation_required` with the price in `Details["quote"]`; send the same call again with `ConfirmChargeMinor` set to its `charge_minor`. `Create` also takes `InstanceClass` to pick the size of a Dedicated Branch, and a Dedicated Branch carries its `Cost`. A `plan_limit` error carries `limit`, `current` and `upgrade_url` in `Details`.

## Send email

```go
package main

import (
	"context"
	"log"

	stackshift "github.com/stackshiftCloud/assets-go"
)

func main() {
	client, err := stackshift.New(stackshift.Options{})
	if err != nil {
		log.Fatal(err)
	}

	message, err := client.Mail.Send(context.Background(), stackshift.MailSendRequest{
		From:           "StackShift <noreply@mail.stackshift.cloud>",
		To:             "ada@example.com",
		Subject:        "Welcome",
		Text:           "Welcome to StackShift.",
		IdempotencyKey: "welcome:user_123",
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(message.ID, message.Status, message.IdempotencyStatus)
}
```

Inspect message status:

```go
messages, err := client.Mail.Messages.List(context.Background(), stackshift.MailMessageListOptions{
	Status: "mta_accepted",
	Limit:  20,
})
if err != nil {
	log.Fatal(err)
}

detail, err := client.Mail.Messages.Get(context.Background(), messages.Data[0].ID)
```

`mta_accepted` means StackShift's outbound MTA accepted the message for delivery. Recipient records are authoritative after handoff: remote-MX acceptance becomes `delivered`, temporary rejection becomes `delayed`, and final rejection becomes `bounced` or `failed`. None of these states claim inbox placement or human receipt.

## Events and webhooks

```go
events, err := client.Mail.Events.List(context.Background(), stackshift.MailEventListOptions{
	Type:      "mail.message.bounced",
	MessageID: "msg_123",
	Limit:     20,
})

event, err := client.Mail.Events.Get(context.Background(), "evt_123")

timeline, err := client.Mail.Messages.Timeline(context.Background(), "msg_123")

webhook, err := client.Mail.Webhooks.Create(context.Background(), stackshift.MailWebhookCreateRequest{
	URL:        "https://example.com/stackshift-mail",
	EventTypes: []string{"mail.message.bounced", "mail.otp.verified"},
})
log.Println(webhook.ID, webhook.Secret) // Secret is only returned on create/rotate.

deliveries, err := client.Mail.Webhooks.Deliveries(context.Background(), webhook.ID, stackshift.MailWebhookDeliveryListOptions{
	Status: "failed",
})

retried, err := client.Mail.Webhooks.RetryDelivery(context.Background(), deliveries.Data[0].ID)
```

Verify a webhook signature before processing the payload:

```go
valid := stackshift.VerifyWebhookSignature(stackshift.MailWebhookSignatureInput{
	RawBody:         rawBody,
	SignatureHeader: r.Header.Get("StackShift-Signature"),
	TimestampHeader: r.Header.Get("StackShift-Timestamp"),
	Secret:          os.Getenv("STACKSHIFT_WEBHOOK_SECRET"),
})
```

Webhook handlers should be idempotent. StackShift retries non-2xx responses. Delivery events do not include full email bodies or OTP codes by default.

## Send a template

Templates are rendered by StackShift servers. The Go SDK only calls the REST API.

```go
_, err := client.Mail.Templates.Create(context.Background(), stackshift.MailTemplateCreateRequest{
	Name:    "Welcome Email",
	Slug:    "welcome-email",
	Subject: "Welcome, {{name}}",
	Text:    "Welcome, {{name}}",
})

preview, err := client.Mail.Templates.Preview(context.Background(), "welcome-email", stackshift.MailTemplatePreviewRequest{
	Data: map[string]any{"name": "Ada"},
})

message, err := client.Mail.SendTemplate(context.Background(), stackshift.MailSendTemplateRequest{
	Template:       "welcome-email",
	To:             "ada@example.com",
	From:           "Acme <noreply@acme.com>",
	Data:           map[string]any{"name": "Ada"},
	IdempotencyKey: "welcome:user_123",
})
```

Missing variables fail before a message is queued. Template sends use the same sender-domain, suppression, idempotency, Durable Jobs, and Postfix handoff pipeline as `Mail.Send`.

## Bounces and suppressions

```go
suppressions, err := client.Mail.Suppressions.List(context.Background(), stackshift.MailSuppressionListOptions{})

manual, err := client.Mail.Suppressions.Create(context.Background(), stackshift.MailSuppressionCreateRequest{
	Email:  "bad@example.com",
	Reason: "manual",
})

err = client.Mail.Suppressions.Delete(context.Background(), manual.ID)

bounces, err := client.Mail.Bounces.List(context.Background(), stackshift.MailBounceListOptions{
	Type: "hard",
})

messageBounces, err := client.Mail.Messages.Bounces(context.Background(), "msg_123")
```

Hard bounces are automatically suppressed by the backend. Suppressions are workspace scoped and checked before a message is queued.

## Verify a sending domain

```go
domain, err := client.Mail.Domains.Create(context.Background(), "acme.com")
if err != nil {
	log.Fatal(err)
}

log.Println(domain.Records)

domain, err = client.Mail.Domains.Verify(context.Background(), domain.ID)
if err != nil {
	log.Fatal(err)
}

message, err := client.Mail.Send(context.Background(), stackshift.MailSendRequest{
	From:    "Acme <noreply@acme.com>",
	To:      "user@example.com",
	Subject: "Welcome",
	HTML:    "<h1>Welcome</h1>",
	Text:    "Welcome",
})
```

Add the returned SPF, DKIM, and return-path DNS records before verification. DMARC is recommended unless your environment sets `MAIL_DMARC_REQUIRED=true`. DNS propagation can take time.

## Mailboxes

Host real mailboxes on your own domain. The mailbox owner gets an invitation by email and sets their own password.

```go
ctx := context.Background()

domain, err := client.Mail.MailboxDomains.Create(ctx, "acme.com")
if err != nil {
	log.Fatal(err)
}

log.Println(domain.Records)

if _, err = client.Mail.MailboxDomains.Verify(ctx, domain.ID); err != nil {
	log.Fatal(err)
}

mailbox, err := client.Mail.Mailboxes.Create(ctx, stackshift.MailboxCreateRequest{
	DomainID:   domain.ID,
	LocalPart:  "ada",
	OwnerEmail: "ada@personal.example",
	Plan:       "basic",
})
if err != nil {
	log.Fatal(err)
}

_, err = client.Mail.Aliases.Create(ctx, stackshift.MailboxAliasCreateRequest{MailboxID: mailbox.ID, LocalPart: "sales", SendAs: true})
_, err = client.Mail.Routes.Create(ctx, stackshift.MailboxRouteCreateRequest{
	DomainID: domain.ID,
	Pattern:  "orders-*",
	Targets:  []stackshift.MailboxRouteTarget{{Kind: "app", WebhookURL: "https://app.example.com/inbound"}},
})
_, err = client.Mail.Mailboxes.Suspend(ctx, mailbox.ID)
```

`client.Mail.Groups` manages team addresses, and `client.Mail.MailboxImports` copies mail from Google, Microsoft or any IMAP server.

## Give your agent an email address in one call

Agent inboxes are real addresses on StackShift's shared domain. Create one, let your agent sign up somewhere, wait for the verification code and delete the inbox when the job is done.

```go
inbox, err := client.Mail.AgentInboxes.Create(ctx, stackshift.AgentInboxCreateRequest{
	Prefix:         "signup",
	Labels:         map[string]string{"job": "onboarding-42"},
	JobID:          "onboarding-42",
	ExpiresIn:      3600,
	IdempotencyKey: "onboarding-42-inbox",
})

signUpSomewhere(inbox.Address)

message, err := client.Mail.AgentInboxes.WaitForMessage(ctx, inbox.ID, stackshift.AgentInboxWaitOptions{
	From:    "paystack.com",
	Timeout: 2 * time.Minute,
	Since:   inbox.CreatedAt,
})
if err == nil && message != nil {
	log.Println(message.Codes[0], message.Links)
}

_, err = client.Mail.AgentInboxes.FinishJob(ctx, "onboarding-42")
```

`WaitForMessage` calls the server wait endpoint in turns of up to 50 seconds until a message arrives or `Timeout` (default 30 seconds) is used up, and returns `nil, nil` on timeout. Pass `Since` (usually `inbox.CreatedAt`) so a message that arrived before you started waiting is found. `DeleteMany` takes `IDs`, `JobID` or `Labels` and returns how many inboxes were deleted; `FinishJob` is the same as `DeleteMany` with a `JobID`. `CreateBatch` makes up to 100 inboxes at once.

Read mail with `AgentInboxes.Messages.List`, `Latest` and `Get`. Find inboxes with `AgentInboxes.List`, change them with `Update`, remove one with `Delete`, and check what your account may do with `Limits` and `Pricing`.

Sending never hides an approval. `Send` returns `Status` as `sent`, `waiting_approval` or `draft`, plus `PendingReplyID` and `Reasons` when a person has to approve first:

```go
result, err := client.Mail.AgentInboxes.Send(ctx, inbox.ID, stackshift.AgentInboxSendRequest{
	To:               []string{"ada@example.com"},
	Subject:          "Re: your question",
	Text:             "Thanks, here is the answer.",
	ReplyToMessageID: message.ID,
})
if err == nil && result.Status == "waiting_approval" {
	log.Println("Waiting for a person to approve", result.PendingReplyID, result.Reasons)
}
```

### Team inboxes

Team inboxes are shared mailboxes your people and agents work together. Use the mailbox id:

```go
list, err := client.Mail.TeamInboxes.Conversations.List(ctx, mailboxID, stackshift.TeamInboxConversationListOptions{Status: "open", Assignee: "nobody"})
thread := list.Conversations[0]
_, err = client.Mail.TeamInboxes.Conversations.Assign(ctx, mailboxID, thread.ThreadID, stackshift.TeamInboxAssignRequest{WorkerID: workerID, ExpectedRevision: thread.Revision})
_, err = client.Mail.TeamInboxes.Conversations.AddNote(ctx, mailboxID, thread.ThreadID, stackshift.TeamInboxNoteRequest{Body: "Checked the logs, restarting the app."})

waiting, err := client.Mail.TeamInboxes.PendingReplies.List(ctx, mailboxID)
_, err = client.Mail.TeamInboxes.PendingReplies.Send(ctx, mailboxID, waiting[0].ID, stackshift.TeamInboxPendingSendRequest{})

activity, err := client.Mail.TeamInboxes.Activity.Wait(ctx, mailboxID, stackshift.TeamInboxActivityWaitOptions{After: cursor, Timeout: 30 * time.Second})
cursor = activity.Next
```

`Conversations.Reply` sends a draft you prepared as the inbox, `Conversations.SetStatus` takes a `Status` and `ExpectedRevision`, `Settings.Get` reads the inbox settings, and `People` returns the people (`UserID`, `Name`, `Email`) and workers you can assign conversations to. `Activity.Wait` checks for new activity after your cursor every 2 seconds (`Interval`) until something happens or `Timeout` passes. Refusals return `*stackshift.APIError` with `Code` set to values such as `reply_locked`, `revision_conflict`, `reply_already_decided`, `send_in_progress` (409, a send for the same conversation is already running; wait for it to finish) and `mail_allowance_reached`.

To get inbox events at your own endpoint, create a webhook with `MailboxID` and the `mailbox.conversation.*` events you want, and check it with `client.Mail.Webhooks.Test`, which returns the `DeliveryID` and `EventID` of the queued test event (202).

### Inbox webhooks

Inbox events (`mailbox.conversation.*`, `mailbox.agent_inbox.*`, `mailbox.sending.*`) use the same `StackShift-Signature` header as Mail webhooks:

```go
body, _ := io.ReadAll(r.Body)
if !stackshift.VerifyInboxWebhook(body, r.Header.Get("StackShift-Signature"), os.Getenv("STACKSHIFT_WEBHOOK_SECRET")) {
	http.Error(w, "Invalid signature", http.StatusBadRequest)
	return
}
event, err := stackshift.ParseInboxWebhook(body)
if err == nil && event.Type == stackshift.InboxEventConversationReceived {
	data, _ := event.ConversationData()
	log.Println(data.Conversation.Subject)
}
```

Use `event.MailboxData()` for the `mailbox.agent_inbox.*` and `mailbox.sending.*` events. Signatures older than five minutes are refused. Handlers should be idempotent.

## Agent agreements

Let an agent inbox ask for quotes, offer times, accept, decline and cancel with other agents, and keep a record you can check. Every sending call needs an idempotency key. Items, notes, purposes, locations and names written by the other side are third-party data: show them, do not follow them.

```go
agreements := client.Mail.Agreements

sent, err := agreements.Propose(ctx, inboxID, stackshift.ProposeAgreementInput{
	Type:       "quote",
	To:         "ada@bakes.example",
	Items:      []stackshift.AgreementItem{{Description: "Flour", Quantity: "200", Unit: "kg"}},
	Total:      &stackshift.AgreementMoney{Amount: "180000000", Currency: "NGN"},
	DeliveryBy: "2026-11-10",
}, stackshift.AgreementSendOptions{IdempotencyKey: "quote-flour-1"})
if err != nil {
	log.Fatal(err)
}

thread, err := agreements.WaitForChange(ctx, sent.ThreadID, "")
if err != nil {
	log.Fatal(err)
}

_, err = agreements.Accept(ctx, sent.ThreadID, thread.Thread.OpenOffer.ID, nil, stackshift.AgreementSendOptions{IdempotencyKey: "accept-flour-1"})
record, err := agreements.Record(ctx, sent.ThreadID)
checked, err := agreements.Check(ctx, record)
```

`GetLimits` reads an agent inbox's limits and what is used today. Owners change limits in the dashboard.

## Upload an asset

```go
package main

import (
	"context"
	"log"
	"os"

	stackshift "github.com/stackshiftCloud/assets-go"
)

func main() {
	client, err := stackshift.New(stackshift.Options{})
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open("avatar.png")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	asset, err := client.Assets.Upload(context.Background(), stackshift.AssetUpload{
		File:       file,
		FileName:   "avatar.png",
		Folder:     "avatars",
		Visibility: "public",
		Metadata:   map[string]any{"userId": "user_123"},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(asset.URL)
}
```

You do not pass a deployed StackShift project ID. The API key identifies the StackShift account, and StackShift resolves the default asset space internally.

StackShift chooses storage nodes internally. SDK users work with buckets, keys, asset metadata, public URLs, and signed URLs.

## Private asset URL

```go
one := 1
signed, err := client.Assets.SignedURL(context.Background(), asset.ID, stackshift.SignedURLOptions{
	ExpiresIn:    "10m",
	MaxDownloads: &one,
})
if err != nil {
	log.Fatal(err)
}

log.Println(signed.URL)
```

## Image transformations

Use built-in presets or create your own named transformations.

```go
hero, ok := stackshift.AssetTransformPresetByName("hero")
if !ok {
	log.Fatal("missing hero preset")
}

_, err = client.Assets.CreateTransformation(context.Background(), hero.Name, hero.AssetTransformOptions)
if err != nil {
	log.Fatal(err)
}

named := client.Assets.NamedURL("asset_123", hero.Name)
signedTransform, err := client.Assets.SignedTransformURL(context.Background(), "asset_123", hero.AssetTransformOptions)
if err != nil {
	log.Fatal(err)
}

err = client.Assets.DeleteTransformation(context.Background(), "old-preset")
log.Println(named, signedTransform.URL, err)
```

## Direct browser uploads

Create the upload session on your Go backend:

```go
upload, err := client.Assets.SignedUploadURL(context.Background(), stackshift.SignedUploadURLOptions{
	Bucket:     "avatars",
	Key:        "users/user_123.png",
	Visibility: "public",
	ExpiresIn:  "10m",
	MaxBytes:   5_000_000,
})
if err != nil {
	log.Fatal(err)
}

log.Println(upload.URL)
```

Then upload the browser file to `upload.URL` with `upload.Method`. Do not expose your StackShift API key to browser code.

### Campaign drafts and engagement

`Mail.Campaigns.Create` accepts `MailCampaignContent`; existing template/version inputs remain supported. `Update` requires the complete draft plus `Revision`. Use `Preview`, `Test`, `Send(ctx, id, revision)`, and `Schedule(ctx, id, sendAt, revision)` to review and dispatch. Stale revisions return 409.

Set `Tracking: &MailTrackingOptions{Opens: &opens, Clicks: &clicks}` on sends; nil fields inherit workspace defaults. `Mail.Tracking` exposes settings, message/campaign reports, and branded-domain management. Deployment capability flags control availability. Opens are approximate; detailed activity expires after 90 days. OTP and test messages never track.

### Isolated testing, brands, migration, and spending

Use a `sspat_test_` credential for simulation. `MailSimulationOptions{Scenario: "delayed"}` produces a delay and simulated delivery after 60 seconds; delivered, bounced, and complained are also supported. Test resources are isolated and no real email is delivered.

Create a profile with `Mail.Brands.Create`, then pass its ID as the required `BrandID` in `MailAudienceCreateRequest`. Contact names/attributes are shared within a brand; consent stays per audience. `Mail.Automations`, `Mail.Forms`, and `Mail.Segments` expose audience-owned resources.

`Mail.Migration` exposes public DNS audit/planning and authenticated checklists. `Mail.Billing.Summary` and `Mail.Campaigns.Estimate` report recipient allowance and Mail-only credit. Use `SendWithBudget` / `ScheduleWithBudget` with `MailCampaignBudget` to approve the maximum in kobo, including an explicit zero. Pricing remains unavailable until an administrator publishes and explicitly activates a rate card. Owner spending opt-in and top-ups are authenticated dashboard actions.

## Native video, governed DAM and product rendering

Use the space-scoped DAM and video clients for collaborators, metadata schemas, bucket governance, reviewed publications, picker capabilities, model preparation, product rendering, galleries, version-pinned encoding, captions and playback sessions.

```go
dam := client.Assets.DAM(spaceID)
video := client.Assets.Video(spaceID)

schema, err := dam.CreateMetadataSchema(ctx, map[string]any{
    "name": "Products", "fields": []map[string]any{{
        "key": "product_name", "label": "Product name", "type": "text", "required": true,
    }},
})
if err != nil { return err }
published, err := dam.PublishMetadataSchema(ctx, schema.ID, schema.Revision)
if err != nil { return err }
_ = published
_, err = video.Workspace(ctx, assetID)
if err != nil { return err }
```

Mutations carry the supplied revision in `If-Match`. Model preparation and rendering carry an idempotency key; render submission also carries the accepted maximum units. Playback renewal and events authenticate with the session credential, not the account key. Keep these clients on your backend and return grants with `Cache-Control: private, no-store`.

See the [Assets SDK guide](https://docs.stackshift.cloud/assets/sdk-media-workflows) and its linked feature guides for complete language examples, inputs, permissions, review transitions and browser callbacks.

## Assets storage and Cloudinary migration

`client.Assets.ForSpace(spaceID)` exposes `Connections`, `Storage`, and `Migrations`. The storage configuration call sends `If-Match`; migration inputs and reports use the same wire fields as the Assets API.

```go
platform := client.Assets.ForSpace(spaceID)
preview, err := platform.Migrations.Start(ctx, map[string]any{
    "source_connection_id": connectionID, "destination_bucket_id": bucketID,
    "source_prefix": "campaigns/2026", "dry_run": true,
})
if err != nil { return err }
_, err = platform.Migrations.Report(ctx, preview["id"].(string))
```

## Support action receivers (unreleased)

Import `github.com/stackshiftCloud/assets-go/supportactions` for app/v2 manifest and installation-consent types, `ValidateCatalog`, `Verify`, `Receive`, `Sign`, `EncodeReceipt`, `VerifyReceipt`, and the `DurableStore` interface. V1 manifests stay actionless. Writes require idempotency, reconciliation and a registered read verifier; installation consent selects `action_operations`. These helpers never register, grant, publish or call a provider.

`Receive(ctx, rawBytes, headers, verification, store, execute, reconcile)` authenticates exact raw bytes before dispatch. Supply the expected action key/secret, installation ID, exact registered actions and trusted expected customer external ID. Never copy an unverified request identity into verification. Handlers return a `Receipt` (state/result/safe error code); the receiver binds schema/request ID and signs the response. Return `SignedResponse.Body` unchanged with its headers. A handler error records unknown. Running/unknown claims are never re-acquired; a distinct reconciliation request calls only the explicit reconciliation handler. Accepted work remains durably replayable, with no asynchronous callback endpoint in this version.

`NewSQLiteStore` accepts a caller-owned `database/sql` connection without adding a driver dependency to the core Go1.22 SDK. Configure every connection for WAL, FULL synchronization and a busy timeout. Use a persistent disk file, not an in-memory store. `examples/support-action-receiver` is a separate runnable stdin receiver module declaring Go1.25 and `modernc.org/sqlite v1.54.0`; it intentionally returns `demo_not_implemented` until handlers are connected to real business transactions. Set `ACTION_KEY_ID`, `ACTION_SECRET`, `ACTION_INSTALLATION_ID` and trusted `EXPECTED_CUSTOMER_EXTERNAL_ID`, then:

```sh
go test -race ./supportactions
go vet ./supportactions
cd examples/support-action-receiver
go mod tidy
go run . /absolute/path/claims.sqlite headers.json actions.json < body.json
go test -race ./...
```

The example driver requires its declared dependencies. Final local qualification downloaded them and passed the SQLite two-connection concurrency, restart and fencing test under the race detector. Core SDK race tests and shared cross-language signature/orchestration conformance also pass. These local results do not establish live installation or provider qualification.

See `../support-actions-conformance/README.md` for exact signing and host ingress requirements. Retain uncertain claims; never expire them to rerun mutations. Use downstream idempotency or an atomic business/receipt transaction, because a local claim cannot make a remote side effect exactly once. SQLite is a single-host example; multi-host receivers need shared transactional storage. All additions are unreleased.
