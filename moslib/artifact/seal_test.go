package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSealContract(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "SEAL-A", ContractOpts{Title: "A", Status: "draft"})
	t.Setenv("MOS_GPG_KEY", "TEST-KEY-123")

	err := SealContract(root, "SEAL-A", "lock", "Taking ownership")
	if err != nil {
		t.Fatalf("SealContract: %v", err)
	}

	seal, err := CheckSeal(root, "SEAL-A")
	if err != nil {
		t.Fatalf("CheckSeal: %v", err)
	}
	if seal == nil {
		t.Fatal("expected seal to be present")
	}
	if seal.Operator != "TEST-KEY-123" {
		t.Errorf("operator = %q, want TEST-KEY-123", seal.Operator)
	}
	if seal.Intent != "lock" {
		t.Errorf("intent = %q, want lock", seal.Intent)
	}
	if seal.Timestamp == "" {
		t.Error("expected timestamp to be set")
	}
}

func TestSealContractAlreadyLocked(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "SEAL-B", ContractOpts{Title: "B", Status: "draft"})
	t.Setenv("MOS_GPG_KEY", "OPERATOR-1")

	SealContract(root, "SEAL-B", "lock", "first lock")
	err := SealContract(root, "SEAL-B", "lock", "second lock")
	if err == nil {
		t.Fatal("expected error sealing already-locked contract")
	}
	if !strings.Contains(err.Error(), "already sealed") {
		t.Errorf("expected 'already sealed' in error, got: %v", err)
	}
}

func TestUnsealContract(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "UNSEAL-A", ContractOpts{Title: "A", Status: "draft"})
	t.Setenv("MOS_GPG_KEY", "MY-KEY")

	SealContract(root, "UNSEAL-A", "lock", "test")
	err := UnsealContract(root, "UNSEAL-A", false)
	if err != nil {
		t.Fatalf("UnsealContract: %v", err)
	}
	seal, _ := CheckSeal(root, "UNSEAL-A")
	if seal != nil {
		t.Error("expected seal to be removed")
	}
}

func TestUnsealContractWrongOperator(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "UNSEAL-B", ContractOpts{Title: "B", Status: "draft"})
	t.Setenv("MOS_GPG_KEY", "OPERATOR-A")
	SealContract(root, "UNSEAL-B", "lock", "test")

	t.Setenv("MOS_GPG_KEY", "OPERATOR-B")
	err := UnsealContract(root, "UNSEAL-B", false)
	if err == nil {
		t.Fatal("expected error unsealing another operator's lock")
	}
	if !strings.Contains(err.Error(), "sealed by") {
		t.Errorf("expected 'sealed by' in error, got: %v", err)
	}
}

func TestUnsealContractHigherAuthority(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "UNSEAL-C", ContractOpts{Title: "C", Status: "draft"})

	configContent := `config {
  mos {
    version = 1
  }
  backend {
    type = "git"
  }
  governance {
    model = "bdfl"
    scope = "cabinet"
  }
  authority {
    role "maintainer" { level = 100 }
    role "contributor" { level = 50 }
    operator "LOW-KEY" { role = "contributor" }
    operator "HIGH-KEY" { role = "maintainer" }
  }
}
`
	configPath := filepath.Join(root, ".mos", "config.mos")
	os.WriteFile(configPath, []byte(configContent), 0644)

	t.Setenv("MOS_GPG_KEY", "LOW-KEY")
	SealContract(root, "UNSEAL-C", "lock", "contributor lock")

	t.Setenv("MOS_GPG_KEY", "HIGH-KEY")
	err := UnsealContract(root, "UNSEAL-C", false)
	if err != nil {
		t.Fatalf("higher authority should be able to unseal: %v", err)
	}
}

func TestUnsealContractForce(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "UNSEAL-D", ContractOpts{Title: "D", Status: "draft"})
	t.Setenv("MOS_GPG_KEY", "OPERATOR-X")
	SealContract(root, "UNSEAL-D", "lock", "test")

	t.Setenv("MOS_GPG_KEY", "OPERATOR-Y")
	err := UnsealContract(root, "UNSEAL-D", true)
	if err != nil {
		t.Fatalf("force unseal should succeed: %v", err)
	}
}

func TestMutateLockedContract(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "LOCK-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "LOCK-B", ContractOpts{Title: "B", Status: "draft"})
	t.Setenv("MOS_GPG_KEY", "LOCKER")
	SealContract(root, "LOCK-A", "lock", "test")

	t.Setenv("MOS_GPG_KEY", "OTHER")

	title := "New Title"
	err := UpdateContract(root, "LOCK-A", ContractUpdateOpts{Title: &title})
	if err == nil {
		t.Error("expected error updating locked contract by non-owner")
	}

	err = LinkContract(root, "LOCK-A", "LOCK-B")
	if err == nil {
		t.Error("expected error linking locked contract by non-owner")
	}

	err = DeleteContract(root, "LOCK-A", false)
	if err == nil {
		t.Error("expected error deleting locked contract by non-owner")
	}
}

func TestMutateLockedContractByOwner(t *testing.T) {
	root := setupScaffold(t)
	CreateContract(root, "LOCKOWN-A", ContractOpts{Title: "A", Status: "draft"})
	CreateContract(root, "LOCKOWN-B", ContractOpts{Title: "B", Status: "draft"})
	t.Setenv("MOS_GPG_KEY", "OWNER-KEY")
	SealContract(root, "LOCKOWN-A", "lock", "test")

	title := "Updated by Owner"
	err := UpdateContract(root, "LOCKOWN-A", ContractUpdateOpts{Title: &title})
	if err != nil {
		t.Errorf("lock owner should be able to update: %v", err)
	}
}

func TestResolveOperatorFromEnv(t *testing.T) {
	t.Setenv("MOS_GPG_KEY", "ENV-KEY-999")
	op, err := ResolveOperator()
	if err != nil {
		t.Fatalf("ResolveOperator: %v", err)
	}
	if op != "ENV-KEY-999" {
		t.Errorf("operator = %q, want ENV-KEY-999", op)
	}
}

func TestResolveOperatorFromGitConfig(t *testing.T) {
	t.Setenv("MOS_GPG_KEY", "")
	op, err := ResolveOperator()
	if err != nil {
		t.Skipf("no git config available: %v", err)
	}
	if op == "" {
		t.Error("expected non-empty operator from git config")
	}
}

// --- primitive 11: timestamps ---
