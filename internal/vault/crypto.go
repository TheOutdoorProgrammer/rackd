// Package vault is the in-memory guardian of the data-encryption key (DEK).
//
// The vault is unlocked with a PIN, which is stretched via Argon2id into a
// key-encryption key (KEK). The KEK unwraps a random 256-bit DEK that encrypts
// all application data. The DEK exists only in memory while unlocked, so a
// process restart always re-locks the vault.
package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	dekSize  = 32 // AES-256 data-encryption key
	saltSize = 16 // Argon2id salt
	keySize  = 32 // derived key-encryption key (AES-256)
)

// ErrDecrypt is returned whenever AEAD decryption fails. During unlock this means
// "wrong PIN"; elsewhere it means corrupted or tampered ciphertext.
var ErrDecrypt = errors.New("vault: decryption failed")

// Argon2Params captures the Argon2id cost parameters. They are persisted with the
// vault so a future change to defaults never invalidates an existing vault.
type Argon2Params struct {
	MemoryKiB uint32
	Time      uint32
	Threads   uint8
}

// deriveKEK stretches a low-entropy PIN into a 32-byte key-encryption key. It is
// deliberately slow and memory-hard to blunt offline brute-force of the small
// PIN space.
func deriveKEK(pin string, salt []byte, p Argon2Params) []byte {
	return argon2.IDKey([]byte(pin), salt, p.Time, p.MemoryKiB, p.Threads, keySize)
}

// seal encrypts plaintext with key using AES-256-GCM, returning nonce||ciphertext.
func seal(key, plaintext []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("vault: generate nonce: %w", err)
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// open reverses seal. A failure (wrong key or tampering) returns ErrDecrypt.
func open(key, blob []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	ns := gcm.NonceSize()
	if len(blob) < ns {
		return nil, ErrDecrypt
	}
	nonce, ciphertext := blob[:ns], blob[ns:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecrypt
	}
	return plaintext, nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("vault: new cipher: %w", err)
	}
	return cipher.NewGCM(block)
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, fmt.Errorf("vault: random bytes: %w", err)
	}
	return b, nil
}
