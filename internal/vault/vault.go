package vault

import (
	"database/sql"
	"errors"
	"sync"
)

var (
	// ErrNotInitialized means no vault has been created yet.
	ErrNotInitialized = errors.New("vault: not initialized")
	// ErrAlreadySetup means a vault already exists and cannot be re-created.
	ErrAlreadySetup = errors.New("vault: already initialized")
	// ErrLocked means the DEK is not currently in memory.
	ErrLocked = errors.New("vault: locked")
	// ErrBadPIN means the supplied PIN did not unwrap the DEK.
	ErrBadPIN = errors.New("vault: incorrect pin")
)

// Vault holds the data-encryption key in memory while unlocked. Single-user:
// there is exactly one vault per process. All methods are safe for concurrent
// use.
type Vault struct {
	db     *sql.DB
	params Argon2Params // used only when creating a new vault

	mu  sync.RWMutex
	dek []byte // nil when locked
}

// New constructs a Vault backed by db. params are applied only on Setup.
func New(db *sql.DB, params Argon2Params) *Vault {
	return &Vault{db: db, params: params}
}

// Initialized reports whether a vault has been set up.
func (v *Vault) Initialized() (bool, error) {
	var n int
	if err := v.db.QueryRow(`SELECT COUNT(*) FROM vault_meta`).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// Unlocked reports whether the DEK is currently held in memory.
func (v *Vault) Unlocked() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.dek != nil
}

// Setup creates a brand-new vault: generate a DEK, wrap it under a key derived
// from the PIN, persist the metadata, and leave the vault unlocked.
func (v *Vault) Setup(pin string) error {
	initialized, err := v.Initialized()
	if err != nil {
		return err
	}
	if initialized {
		return ErrAlreadySetup
	}

	salt, err := randomBytes(saltSize)
	if err != nil {
		return err
	}
	dek, err := randomBytes(dekSize)
	if err != nil {
		return err
	}
	kek := deriveKEK(pin, salt, v.params)
	wrapped, err := seal(kek, dek)
	if err != nil {
		return err
	}

	if _, err := v.db.Exec(
		`INSERT INTO vault_meta (id, salt, argon_memory, argon_time, argon_threads, wrapped_dek)
		 VALUES (1, ?, ?, ?, ?, ?)`,
		salt, v.params.MemoryKiB, v.params.Time, v.params.Threads, wrapped,
	); err != nil {
		return err
	}

	v.mu.Lock()
	v.dek = dek
	v.mu.Unlock()
	return nil
}

// Unlock derives the KEK from the PIN, unwraps the DEK, and holds it in memory.
// Returns ErrBadPIN when the PIN is wrong (the AEAD tag fails to authenticate).
func (v *Vault) Unlock(pin string) error {
	var (
		salt    []byte
		mem     uint32
		tm      uint32
		threads uint8
		wrapped []byte
	)
	err := v.db.QueryRow(
		`SELECT salt, argon_memory, argon_time, argon_threads, wrapped_dek FROM vault_meta WHERE id = 1`,
	).Scan(&salt, &mem, &tm, &threads, &wrapped)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotInitialized
	}
	if err != nil {
		return err
	}

	kek := deriveKEK(pin, salt, Argon2Params{MemoryKiB: mem, Time: tm, Threads: threads})
	dek, err := open(kek, wrapped)
	if err != nil {
		return ErrBadPIN
	}

	v.mu.Lock()
	v.dek = dek
	v.mu.Unlock()
	return nil
}

// Lock zeroes and drops the DEK. The vault must be unlocked again with the PIN.
func (v *Vault) Lock() {
	v.mu.Lock()
	for i := range v.dek {
		v.dek[i] = 0
	}
	v.dek = nil
	v.mu.Unlock()
}

// Encrypt seals application data with the DEK.
func (v *Vault) Encrypt(plaintext []byte) ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.dek == nil {
		return nil, ErrLocked
	}
	return seal(v.dek, plaintext)
}

// Decrypt opens application data with the DEK.
func (v *Vault) Decrypt(blob []byte) ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.dek == nil {
		return nil, ErrLocked
	}
	return open(v.dek, blob)
}
