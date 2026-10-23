// A stdin/file receiver example, intentionally without a live business mutation.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/stackshiftCloud/assets-go/supportactions"
	"io"
	_ "modernc.org/sqlite"
	"net/url"
	"os"
)

func openStore(ctx context.Context, path string) (*sql.DB, *supportactions.SQLiteStore, error) {
	if path == "" || path == ":memory:" {
		return nil, nil, fmt.Errorf("persistent database path required")
	}
	u := url.URL{Scheme: "file", Path: path}
	q := u.Query()
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "synchronous(FULL)")
	q.Add("_pragma", "busy_timeout(10000)")
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, nil, err
	}
	store, err := supportactions.NewSQLiteStore(ctx, db)
	if err != nil {
		db.Close()
		return nil, nil, err
	}
	return db, store, nil
}
func run() error {
	if len(os.Args) != 4 {
		return fmt.Errorf("usage: receiver persistent.sqlite headers.json actions.json < body.json")
	}
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, supportactions.RequestLimit+1))
	if err != nil {
		return err
	}
	headersRaw, err := os.ReadFile(os.Args[2])
	if err != nil {
		return err
	}
	var headers map[string]string
	if err := json.Unmarshal(headersRaw, &headers); err != nil {
		return err
	}
	catalogRaw, err := os.ReadFile(os.Args[3])
	if err != nil {
		return err
	}
	var actions []supportactions.Action
	if err := json.Unmarshal(catalogRaw, &actions); err != nil {
		return err
	}
	if len(actions) == 0 {
		return fmt.Errorf("empty catalog")
	}
	if err := supportactions.ValidateCatalog(actions[0].Version, actions); err != nil {
		return err
	}
	ctx := context.Background()
	db, store, err := openStore(ctx, os.Args[1])
	if err != nil {
		return err
	}
	defer db.Close()
	options := supportactions.Verification{KeyID: os.Getenv("ACTION_KEY_ID"), Secret: os.Getenv("ACTION_SECRET"), InstallationID: os.Getenv("ACTION_INSTALLATION_ID"), ExpectedCustomerExternalID: os.Getenv("EXPECTED_CUSTOMER_EXTERNAL_ID"), Actions: actions}
	execute := func(context.Context, supportactions.Request) (supportactions.Receipt, error) {
		return supportactions.Receipt{State: "failed", Result: map[string]any{}, ErrorCode: "demo_not_implemented"}, nil
	}
	reconcile := func(context.Context, supportactions.Request) (supportactions.Receipt, error) {
		return supportactions.Receipt{State: "unknown", Result: map[string]any{}, ErrorCode: "demo_not_implemented"}, nil
	}
	response, err := supportactions.Receive(ctx, raw, headers, options, store, execute, reconcile)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(response)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "receiver failed:", err)
		os.Exit(1)
	}
}
