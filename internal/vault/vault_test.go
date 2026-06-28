package vault

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// fast Argon2 params so tests don't take seconds.
var testParams = Argon2Params{MemoryKiB: 8 * 1024, Time: 1, Threads: 1}

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE vault_meta (
		id            INTEGER PRIMARY KEY CHECK (id = 1),
		salt          BLOB    NOT NULL,
		argon_memory  INTEGER NOT NULL,
		argon_time    INTEGER NOT NULL,
		argon_threads INTEGER NOT NULL,
		wrapped_dek   BLOB    NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestSetupUnlockRoundTrip(t *testing.T) {
	v := New(testDB(t), testParams)

	if init, err := v.Initialized(); err != nil || init {
		t.Fatalf("expected uninitialized, got init=%v err=%v", init, err)
	}
	if err := v.Setup("123456"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if !v.Unlocked() {
		t.Fatal("expected unlocked after setup")
	}

	secret := []byte("Glock 19 — serial ABC123")
	ct, err := v.Encrypt(secret)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	v.Lock()
	if v.Unlocked() {
		t.Fatal("expected locked after Lock")
	}
	if _, err := v.Decrypt(ct); err != ErrLocked {
		t.Fatalf("expected ErrLocked while locked, got %v", err)
	}

	if err := v.Unlock("000000"); err != ErrBadPIN {
		t.Fatalf("expected ErrBadPIN for wrong pin, got %v", err)
	}
	if err := v.Unlock("123456"); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	pt, err := v.Decrypt(ct)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(pt) != string(secret) {
		t.Fatalf("round-trip mismatch: got %q", pt)
	}
}

func TestSetupTwiceFails(t *testing.T) {
	v := New(testDB(t), testParams)
	if err := v.Setup("123456"); err != nil {
		t.Fatalf("first setup: %v", err)
	}
	if err := v.Setup("654321"); err != ErrAlreadySetup {
		t.Fatalf("expected ErrAlreadySetup, got %v", err)
	}
}

func TestLockoutBackoff(t *testing.T) {
	l := NewLockout()
	now := time.Unix(0, 0)
	l.now = func() time.Time { return now }

	if l.Retry() != 0 {
		t.Fatal("fresh lockout should permit immediately")
	}
	for i := 0; i < l.freeAttempts; i++ {
		l.Fail()
	}
	if l.Retry() != 0 {
		t.Fatal("within free attempts there should be no delay")
	}
	l.Fail() // first failure beyond the free allowance
	if l.Retry() <= 0 {
		t.Fatal("expected a backoff delay after exhausting free attempts")
	}
	l.Reset()
	if l.Retry() != 0 {
		t.Fatal("reset should clear the lockout")
	}
}
