package main

import (
	"context"
	"errors"
	"github.com/stackshiftCloud/assets-go/supportactions"
	"path/filepath"
	"sync"
	"testing"
)

func TestDiskClaimsConcurrencyRestartAndFencing(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "claims.sqlite")
	first, store, err := openStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	second, other, err := openStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	var wg sync.WaitGroup
	claims := make(chan supportactions.Claim, 2)
	errs := make(chan error, 2)
	for _, s := range []*supportactions.SQLiteStore{store, other} {
		wg.Add(1)
		go func(s *supportactions.SQLiteStore) {
			defer wg.Done()
			claim, err := s.Claim(ctx, "execution", "same", "digest")
			claims <- claim
			errs <- err
		}(s)
	}
	wg.Wait()
	close(claims)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	acquired, running := 0, 0
	var fence string
	for claim := range claims {
		switch claim.State {
		case "acquired":
			acquired++
			fence = claim.Fence
		case "running":
			running++
		}
	}
	if acquired != 1 || running != 1 {
		t.Fatalf("acquired=%d running=%d", acquired, running)
	}
	if err := store.Complete(ctx, "execution", "same", "wrong", []byte("receipt"), false); !errors.Is(err, supportactions.ErrFenceLost) {
		t.Fatal(err)
	}
	if err := store.Complete(ctx, "execution", "same", fence, []byte("receipt"), true); err != nil {
		t.Fatal(err)
	}
	first.Close()
	reopened, again, err := openStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	claim, err := again.Claim(ctx, "execution", "same", "digest")
	if err != nil || claim.State != "unknown" || string(claim.Response) != "receipt" {
		t.Fatal(claim, err)
	}
	conflict, err := again.Claim(ctx, "execution", "same", "changed")
	if err != nil || conflict.State != "conflict" {
		t.Fatal(conflict, err)
	}
	if err := again.Complete(ctx, "execution", "same", fence, []byte("changed"), false); !errors.Is(err, supportactions.ErrFenceLost) {
		t.Fatal(err)
	}
}
