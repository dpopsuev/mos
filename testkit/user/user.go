package user

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
)

// User is a simulated governance participant with cryptographic identity.
// Each user has an ephemeral ED25519 keypair, a local working directory,
// and the ability to sign artifacts.
type User struct {
	Name        string
	PrivateKey  ed25519.PrivateKey
	PublicKey   ed25519.PublicKey
	fingerprint string
	WorkDir     string
}

// NewUser creates a test user with an ephemeral ED25519 keypair.
func NewUser(t testing.TB, name string) *User {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key for %s: %v", name, err)
	}

	h := sha256.Sum256(pub)
	fp := "ssh-ed25519:" + base64.RawStdEncoding.EncodeToString(h[:])

	workDir := t.TempDir()

	return &User{
		Name:        name,
		PrivateKey:  priv,
		PublicKey:   pub,
		fingerprint: fp,
		WorkDir:     workDir,
	}
}

// Fingerprint returns the user's public key fingerprint.
func (u *User) Fingerprint() string {
	return u.fingerprint
}

// Sign produces an ed25519 signature of the data, encoded as base64.
func (u *User) Sign(data []byte) (string, error) {
	sig := ed25519.Sign(u.PrivateKey, data)
	return base64.StdEncoding.EncodeToString(sig), nil
}

// Verify checks that a base64-encoded signature is valid for the given data.
func (u *User) Verify(data []byte, sig64 string) error {
	sig, err := base64.StdEncoding.DecodeString(sig64)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	if !ed25519.Verify(u.PublicKey, data, sig) {
		return fmt.Errorf("signature verification failed for user %s", u.Name)
	}
	return nil
}

// AllowedSignersEntry returns a line for the allowed_signers file.
func (u *User) AllowedSignersEntry() string {
	pubB64 := base64.StdEncoding.EncodeToString(u.PublicKey)
	return fmt.Sprintf("%s ssh-ed25519 %s", u.Name, pubB64)
}

// AllowedSigners generates an allowed_signers file content from a set of users.
func AllowedSigners(users ...*User) string {
	var result string
	for _, u := range users {
		result += u.AllowedSignersEntry() + "\n"
	}
	return result
}
