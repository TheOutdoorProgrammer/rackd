package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/TheOutdoorProgrammer/boating-accident/internal/vault"
)

// ErrNotFound is returned when a row does not exist.
var ErrNotFound = errors.New("db: not found")

// Store is the encrypted data-access layer. Every read decrypts a row's data
// blob into a typed model; every write encrypts it. The vault must be unlocked.
type Store struct {
	db *sql.DB
	v  *vault.Vault
}

func NewStore(db *sql.DB, v *vault.Vault) *Store { return &Store{db: db, v: v} }

func (s *Store) encrypt(val any) ([]byte, error) {
	j, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	return s.v.Encrypt(j)
}

func (s *Store) decrypt(blob []byte, val any) error {
	j, err := s.v.Decrypt(blob)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, val)
}

func nowStamp() string { return time.Now().UTC().Format(time.RFC3339) }

func checkAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
