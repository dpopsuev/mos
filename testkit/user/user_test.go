package user_test

import (
	"testing"

	"github.com/dpopsuev/mos/testkit/user"
)

func TestNewUser(t *testing.T) {
	u := user.NewUser(t, "alice")

	if u.Name != "alice" {
		t.Errorf("name = %q, want alice", u.Name)
	}
	if u.PublicKey == nil {
		t.Fatal("public key is nil")
	}
	if u.PrivateKey == nil {
		t.Fatal("private key is nil")
	}
	if u.Fingerprint() == "" {
		t.Fatal("fingerprint is empty")
	}
	if u.WorkDir == "" {
		t.Fatal("work dir is empty")
	}
}

func TestSignAndVerify(t *testing.T) {
	u := user.NewUser(t, "bob")
	data := []byte("test payload")

	sig, err := u.Sign(data)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if sig == "" {
		t.Fatal("signature is empty")
	}

	if err := u.Verify(data, sig); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

func TestVerifyRejectsTamperedData(t *testing.T) {
	u := user.NewUser(t, "charlie")

	sig, err := u.Sign([]byte("original"))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if err := u.Verify([]byte("tampered"), sig); err == nil {
		t.Fatal("expected verification to fail for tampered data")
	}
}

func TestDifferentUsersHaveDifferentFingerprints(t *testing.T) {
	a := user.NewUser(t, "alice")
	b := user.NewUser(t, "bob")

	if a.Fingerprint() == b.Fingerprint() {
		t.Fatal("different users should have different fingerprints")
	}
}

func TestAllowedSigners(t *testing.T) {
	a := user.NewUser(t, "alice")
	b := user.NewUser(t, "bob")

	content := user.AllowedSigners(a, b)
	if content == "" {
		t.Fatal("allowed_signers content is empty")
	}
}
