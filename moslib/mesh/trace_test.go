package mesh

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestContractIDExtraction(t *testing.T) {
	tests := []struct {
		msg  string
		want []string
	}{
		{"Fix CON-2026-100 bug", []string{"CON-2026-100"}},
		{"BUG-2026-001 and CON-2026-200", []string{"BUG-2026-001", "CON-2026-200"}},
		{"no contract here", nil},
		{"SPEC-2026-033 addresses C5", []string{"SPEC-2026-033"}},
	}
	for _, tt := range tests {
		got := contractIDRe.FindAllString(tt.msg, -1)
		if len(got) != len(tt.want) {
			t.Errorf("msg=%q: got %v, want %v", tt.msg, got, tt.want)
			continue
		}
		for i, g := range got {
			if g != tt.want[i] {
				t.Errorf("msg=%q: index %d: got %q, want %q", tt.msg, i, g, tt.want[i])
			}
		}
	}
}

func TestGoPackageDir(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"cmd/mos/main.go", "cmd/mos"},
		{"README.md", ""},
		{"main.go", ""},
		{"moslib/store/store.go", "moslib/store"},
	}
	for _, tt := range tests {
		got := goPackageDir(tt.path)
		if got != tt.want {
			t.Errorf("goPackageDir(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestMapContractsToSpecs(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")

	os.MkdirAll(filepath.Join(mos, "specifications", "active", "SPEC-2026-001"), 0o755)
	os.WriteFile(filepath.Join(mos, "specifications", "active", "SPEC-2026-001", "specification.mos"), []byte(`specification "SPEC-2026-001" {
  title    = "Test spec"
  status   = "active"
  satisfies = "NEED-2026-001"
}
`), 0o644)

	os.MkdirAll(filepath.Join(mos, "contracts", "active", "CON-2026-100"), 0o755)
	os.WriteFile(filepath.Join(mos, "contracts", "active", "CON-2026-100", "contract.mos"), []byte(`contract "CON-2026-100" {
  title    = "Test contract"
  status   = "active"
  justifies = "NEED-2026-001"
}
`), 0o644)

	result, err := MapContractsToSpecs(root)
	if err != nil {
		t.Fatalf("MapContractsToSpecs: %v", err)
	}
	specs, ok := result["CON-2026-100"]
	if !ok {
		t.Fatal("expected CON-2026-100 in mapping")
	}
	if len(specs) != 1 || specs[0] != "SPEC-2026-001" {
		t.Errorf("expected [SPEC-2026-001], got %v", specs)
	}
}

func TestTraceCommitContracts(t *testing.T) {
	root := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = root
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = root
	cmd.Run()

	os.WriteFile(filepath.Join(root, "pkg", "foo.go"), []byte("package pkg\n"), 0o644)
	os.MkdirAll(filepath.Join(root, "pkg"), 0o755)
	os.WriteFile(filepath.Join(root, "pkg", "foo.go"), []byte("package pkg\n"), 0o644)

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = root
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "CON-2026-100: initial")
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	traces, err := TraceCommitContracts(root)
	if err != nil {
		t.Fatalf("TraceCommitContracts: %v", err)
	}
	if len(traces) == 0 {
		t.Fatal("expected at least 1 trace")
	}
	if len(traces[0].ContractIDs) != 1 || traces[0].ContractIDs[0] != "CON-2026-100" {
		t.Errorf("expected [CON-2026-100], got %v", traces[0].ContractIDs)
	}
}

func TestUniqueStrings(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b"}
	got := uniqueStrings(input)
	if len(got) != 3 {
		t.Errorf("expected 3 unique, got %d: %v", len(got), got)
	}
}
