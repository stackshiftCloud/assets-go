package supportactions

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
)

// SQLiteStore uses a caller-supplied SQLite driver and persistent database file.
// Configure WAL, synchronous=FULL and a busy timeout on every driver connection.
// The SDK deliberately does not register a database driver or expire claims.
type SQLiteStore struct{ db *sql.DB }

func NewSQLiteStore(ctx context.Context, db *sql.DB) (*SQLiteStore, error) {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS support_action_claims (
 namespace TEXT NOT NULL, claim_key TEXT NOT NULL, digest TEXT NOT NULL,
 state TEXT NOT NULL, fence TEXT NOT NULL, response BLOB,
 PRIMARY KEY(namespace,claim_key))`)
	if err != nil {
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}
func (store *SQLiteStore) Claim(ctx context.Context, namespace, key, digest string) (Claim, error) {
	fenceBytes := make([]byte, 16)
	if _, err := rand.Read(fenceBytes); err != nil {
		return Claim{}, err
	}
	fence := hex.EncodeToString(fenceBytes)
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return Claim{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO support_action_claims VALUES (?, ?, ?, ?, ?, NULL)", namespace, key, digest, "running", fence)
	if err != nil {
		return Claim{}, err
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return Claim{}, err
	}
	var claim Claim
	var storedDigest string
	if err := tx.QueryRowContext(ctx, "SELECT digest,state,fence,response FROM support_action_claims WHERE namespace=? AND claim_key=?", namespace, key).Scan(&storedDigest, &claim.State, &claim.Fence, &claim.Response); err != nil {
		return Claim{}, err
	}
	if inserted == 1 {
		claim.State = "acquired"
	} else if storedDigest != digest {
		claim.State = "conflict"
	}
	if err := tx.Commit(); err != nil {
		return Claim{}, err
	}
	return claim, nil
}
func (store *SQLiteStore) Complete(ctx context.Context, namespace, key, fence string, response []byte, unknown bool) error {
	state := "completed"
	if unknown {
		state = "unknown"
	}
	result, err := store.db.ExecContext(ctx, "UPDATE support_action_claims SET state=?,response=? WHERE namespace=? AND claim_key=? AND fence=? AND state='running'", state, response, namespace, key, fence)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrFenceLost
	}
	return nil
}
