package supportactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

type vector struct {
	Name     string            `json:"name"`
	Raw      string            `json:"raw_body"`
	Headers  map[string]string `json:"headers"`
	Customer string            `json:"expected_customer_external_id"`
	Error    *string           `json:"error"`
}
type fixtures struct {
	KeyID        string   `json:"key_id"`
	Secret       string   `json:"secret"`
	Now          int64    `json:"now"`
	Installation string   `json:"installation_id"`
	Actions      []Action `json:"actions"`
	Vectors      []vector `json:"vectors"`
	Receipt      vector   `json:"receipt"`
}

func fixture(t *testing.T) fixtures {
	t.Helper()
	raw, err := os.ReadFile("../support-actions-conformance/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var f fixtures
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	return f
}
func (f fixtures) options() Verification {
	return Verification{KeyID: f.KeyID, Secret: f.Secret, InstallationID: f.Installation, ExpectedCustomerExternalID: "customer-fixture", Now: time.Unix(f.Now, 0), Actions: f.Actions}
}
func id(n int) string { return fmt.Sprintf("00000000-0000-4000-8000-%012d", n) }
func TestConformance(t *testing.T) {
	f := fixture(t)
	for _, v := range f.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			options := f.options()
			options.ExpectedCustomerExternalID = v.Customer
			_, err := Verify([]byte(v.Raw), v.Headers, options)
			if v.Error != nil {
				if err == nil || err.Error() != *v.Error {
					t.Fatalf("error=%v want %s", err, *v.Error)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			var stamp int64
			_, _ = fmt.Sscan(v.Headers["x-stackshift-timestamp"], &stamp)
			headers, err := Sign([]byte(v.Raw), RequestContext, f.KeyID, f.Secret, stamp, v.Headers["x-stackshift-delivery-id"])
			if err != nil || !reflect.DeepEqual(headers, v.Headers) {
				t.Fatalf("signature mismatch: %v", err)
			}
		})
	}
	headers, err := Sign([]byte(f.Receipt.Raw), ReceiptContext, f.KeyID, f.Secret, f.Now, id(10))
	if err != nil || !reflect.DeepEqual(headers, f.Receipt.Headers) {
		t.Fatal("receipt signature mismatch", err)
	}
	if _, err := VerifyReceipt([]byte(f.Receipt.Raw), headers, f.options(), id(1), id(10)); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyReceipt([]byte(f.Receipt.Raw), headers, f.options(), id(1), id(11)); err == nil {
		t.Fatal("delivery mismatch accepted")
	}
}

// This fake is only for protocol orchestration; disk durability is tested by the SQLite example.
type fakeStore struct {
	mu      sync.Mutex
	claims  map[string]Claim
	digests map[string]string
}

func newFake() *fakeStore {
	return &fakeStore{claims: map[string]Claim{}, digests: map[string]string{}}
}
func (s *fakeStore) Claim(_ context.Context, ns, key, digest string) (Claim, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key = ns + ":" + key
	if prior, ok := s.claims[key]; ok {
		if s.digests[key] != digest {
			prior.State = "conflict"
		}
		return prior, nil
	}
	claim := Claim{State: "running", Fence: key}
	s.claims[key] = claim
	s.digests[key] = digest
	claim.State = "acquired"
	return claim, nil
}
func (s *fakeStore) Complete(_ context.Context, ns, key, fence string, raw []byte, unknown bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key = ns + ":" + key
	prior := s.claims[key]
	if prior.State != "running" || prior.Fence != fence {
		return ErrFenceLost
	}
	prior.State = "completed"
	if unknown {
		prior.State = "unknown"
	}
	prior.Response = append([]byte(nil), raw...)
	s.claims[key] = prior
	return nil
}
func TestReceiverDeduplicationAndReconciliation(t *testing.T) {
	f := fixture(t)
	store := newFake()
	calls, reconciliations := 0, 0
	execute := func(context.Context, Request) (Receipt, error) {
		calls++
		return Receipt{State: "succeeded", Result: map[string]any{"updated": true}}, nil
	}
	reconcile := func(context.Context, Request) (Receipt, error) {
		reconciliations++
		return Receipt{State: "succeeded", Result: map[string]any{"updated": true}}, nil
	}
	send := func(raw string, n int) (SignedResponse, error) {
		headers, _ := Sign([]byte(raw), RequestContext, f.KeyID, f.Secret, f.Now, id(n))
		return Receive(context.Background(), []byte(raw), headers, f.options(), store, execute, reconcile)
	}
	first, err := send(f.Vectors[0].Raw, 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := send(f.Vectors[0].Raw, 10); err == nil || err.Error() != "delivery_replay" {
		t.Fatal("duplicate delivery", err)
	}
	replay, err := send(f.Vectors[0].Raw, 11)
	if err != nil || string(first.Body) != string(replay.Body) {
		t.Fatal("replay", err)
	}
	if _, err := send(f.Vectors[0].Raw+" ", 12); err == nil || err.Error() != "idempotency_conflict" {
		t.Fatal("raw digest conflict", err)
	}
	if _, err := send(f.Vectors[len(f.Vectors)-1].Raw, 13); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || reconciliations != 1 {
		t.Fatal(calls, reconciliations)
	}
}
func TestUnknownNeverRedispatched(t *testing.T) {
	f := fixture(t)
	store := newFake()
	calls := 0
	handler := func(context.Context, Request) (Receipt, error) {
		calls++
		return Receipt{}, errors.New("private provider response")
	}
	for _, n := range []int{10, 11} {
		headers, _ := Sign([]byte(f.Vectors[0].Raw), RequestContext, f.KeyID, f.Secret, f.Now, id(n))
		response, err := Receive(context.Background(), []byte(f.Vectors[0].Raw), headers, f.options(), store, handler, handler)
		if err != nil {
			t.Fatal(err)
		}
		var receipt Receipt
		if err := json.Unmarshal(response.Body, &receipt); err != nil || receipt.State != "unknown" || receipt.ErrorCode != "handler_outcome_unknown" {
			t.Fatal(receipt, err)
		}
	}
	if calls != 1 {
		t.Fatal(calls)
	}
}
func TestConcurrentRunningNeverRedispatched(t *testing.T) {
	f := fixture(t)
	store := newFake()
	started, release := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	handler := func(context.Context, Request) (Receipt, error) {
		close(started)
		<-release
		return Receipt{State: "succeeded", Result: map[string]any{"updated": true}}, nil
	}
	send := func(n int) (SignedResponse, error) {
		headers, _ := Sign([]byte(f.Vectors[0].Raw), RequestContext, f.KeyID, f.Secret, f.Now, id(n))
		return Receive(context.Background(), []byte(f.Vectors[0].Raw), headers, f.options(), store, handler, handler)
	}
	go func() { _, err := send(40); done <- err }()
	<-started
	response, err := send(41)
	if err != nil {
		t.Fatal(err)
	}
	var receipt Receipt
	_ = json.Unmarshal(response.Body, &receipt)
	if receipt.State != "accepted" {
		t.Fatal(receipt)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
func TestCatalogValidation(t *testing.T) {
	f := fixture(t)
	if err := ValidateCatalog("1.0.0", f.Actions); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func([]Action){func(a []Action) { a[0].ReconciliationEndpoint = "" }, func(a []Action) { a[0].VerificationOperation = "missing" }, func(a []Action) {
		maximum := float64(9007199254740992)
		a[0].Input["amount"] = Field{Type: "integer", Required: true, Maximum: &maximum}
	}, func(a []Action) {
		limit := 100
		a[0].Input["customer_id"] = Field{Type: "string", Required: true, MaxLength: &limit}
	}} {
		f := fixture(t)
		change(f.Actions)
		if err := ValidateCatalog("1.0.0", f.Actions); err == nil {
			t.Fatal("invalid catalog accepted")
		}
	}
}

func TestReceiptRejectsUnboundedOrNonJSONOutput(t *testing.T) {
	result := map[string]any{}
	nested := result
	for i := 0; i < 34; i++ {
		next := map[string]any{}
		nested["nested"] = next
		nested = next
	}
	if _, err := EncodeReceipt(Receipt{Schema: ReceiptContext, RequestID: id(1), State: "succeeded", Result: result}); err == nil {
		t.Fatal("deep result accepted")
	}
}
