# Support action protocol conformance (unreleased)

`vectors.json` is synthetic, contains no live credentials, and is shared by the TypeScript, Python, and Go tests. `raw_body` must be UTF-8 encoded exactly; do not parse and reserialize before signature verification or execution hashing.

Four headers authenticate requests and receipts:

- `x-stackshift-action-key-id`
- `x-stackshift-timestamp`: positive decimal Unix seconds
- `x-stackshift-delivery-id`: a fresh UUID for each delivery attempt
- `x-stackshift-action-signature`: `v1=` followed by lowercase hexadecimal HMAC-SHA256

The signed bytes are `context + "\n" + keyID + "\n" + timestamp + "\n" + deliveryID + "\n" + rawBody`. The key is the exact UTF-8 installation `action_secret`. Request context is `stackshift.support.action/v1`; receipt context is `stackshift.support.action-receipt/v1`. Receipts echo the request delivery ID. Webhook credentials are not action credentials.

Verify the expected installation/key, exact allowed action name/version, trusted expected customer external ID, request expiry, signature freshness (at most 300 seconds), body size, and primitive argument schema before invoking business code. Integer arguments and bounds are limited to the cross-language safe range ±9,007,199,254,740,991; revisions and timestamps use its nonnegative subset. Resolve expected customer identity from trusted host configuration or an authoritative mapping; never copy an unverified request's identity into verification options. The receiver must also enforce its own current authorization and customer ownership before accessing business records. SDK validation does not establish a new customer identity or grant.

An execution has one immutable request ID and exact body bytes, including issue/expiry times. Every retry uses a new delivery ID/timestamp but the same body. Changing whitespace alone conflicts. Execution identity is `(installation, action.name, action.version, request_id, mode)` and the payload digest is SHA-256 of raw bytes. Delivery claims and execution claims are separate durable records. Delivery replay rejects; execution replay returns the stored receipt. Running or unknown executions never automatically expire into a new mutation attempt. Terminal writes compare the acquisition fence.

A handler exception, invalid success output, or interrupted result persistence is uncertain. Reconciliation uses a new request ID, `mode: reconcile`, the original `target_request_id`, and `arguments: {}`. It invokes only the explicit reconciliation handler, which queries authoritative downstream state. It must not call the original mutation. The host must verify that the target request belongs to the same installation/action/version and original customer before looking up its outcome; a changed mapping or version requires explicit recovery, not reassociation. `accepted` can describe work already durably queued by the application; replay remains accepted until a separate reconciliation establishes the result. There is no callback/reporting endpoint in this version. Exported receipt encoding/signing helpers support the synchronous response contract.

SQLite examples use disk databases with WAL, FULL synchronization, and fenced updates. They are single-host examples, not distributed coordination. Keep the database on durable storage and retain claims for the lifetime of requests/reconciliation; do not delete uncertain claims to retry. Use a shared transactional store for multiple hosts. Deduplication does not make an external side effect and a local receipt one atomic transaction: use the downstream idempotency key or commit business changes and receipts in one transaction. A crash after a side effect can remain uncertain and requires reconciliation or human review.

Host HTTP adapters must read bounded raw bytes before JSON parsing, reject duplicate security headers, map protocol errors to safe 4xx responses, set request deadlines, propagate cancellation, and return the signed response bytes unchanged. Frameworks that flatten duplicate headers must reject them at ingress. Do not log secrets or full action/customer payloads. No endpoint, registration, publication, or provider calls are performed by these helpers.

Qualification: shared signatures and protocol orchestration pass in all languages. Disk replay/restart/concurrency/fencing pass in Python, TypeScript and the Go SQLite example; Go core race tests pass. The Go example uses its declared driver dependencies and generated go.sum. No live developer installation or downstream provider is qualified.
