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
