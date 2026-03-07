package linter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePolicies_MatchingContract(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")

	for _, d := range []string{
		filepath.Join(mosDir, "rules", "mechanical"),
		filepath.Join(mosDir, "contracts", "active", "CON-001"),
	} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	rule := `rule "test-policy" {
  name = "Test Policy"
  type = "mechanical"
  scope = "project"
  enforcement = "warning"

  when {
    artifact_kind = "contract"
  }

  harness {
    command = "echo ok"
    timeout = "10s"
  }
}
`
	contract := `contract "CON-001" {
  title = "Test Contract"
  status = "active"
  kind = "feature"
}
`
	if err := os.WriteFile(filepath.Join(mosDir, "rules", "mechanical", "test-policy.mos"), []byte(rule), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mosDir, "contracts", "active", "CON-001", "contract.mos"), []byte(contract), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := &ProjectContext{
		Root:        root,
		RuleIDs:     map[string]string{"test-policy": filepath.Join(mosDir, "rules", "mechanical", "test-policy.mos")},
		ContractIDs: map[string]string{"CON-001": filepath.Join(mosDir, "contracts", "active", "CON-001", "contract.mos")},
		ArtifactIDs: map[string]map[string]string{},
	}

	diags := validatePolicies(ctx)

	found := false
	for _, d := range diags {
		if d.Rule == "policy-match" && d.Severity == SeverityWarning {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected policy-match warning diagnostic for matching contract")
	}
}

func TestValidatePolicies_NonMatchingArtifact(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")

	for _, d := range []string{
		filepath.Join(mosDir, "rules", "mechanical"),
		filepath.Join(mosDir, "specifications", "active", "SPEC-001"),
	} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	rule := `rule "contract-only" {
  name = "Contract Only"
  type = "mechanical"
  scope = "project"
  enforcement = "error"

  when {
    artifact_kind = "contract"
  }
}
`
	spec := `specification "SPEC-001" {
  title = "Test Spec"
  status = "active"
}
`
	if err := os.WriteFile(filepath.Join(mosDir, "rules", "mechanical", "contract-only.mos"), []byte(rule), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mosDir, "specifications", "active", "SPEC-001", "specification.mos"), []byte(spec), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := &ProjectContext{
		Root:        root,
		RuleIDs:     map[string]string{"contract-only": filepath.Join(mosDir, "rules", "mechanical", "contract-only.mos")},
		ContractIDs: map[string]string{},
		ArtifactIDs: map[string]map[string]string{
			"specification": {"SPEC-001": filepath.Join(mosDir, "specifications", "active", "SPEC-001", "specification.mos")},
		},
	}

	diags := validatePolicies(ctx)

	for _, d := range diags {
		if d.Rule == "policy-match" {
			t.Errorf("expected no policy-match diagnostic for specification, got: %s", d.Message)
		}
	}
}

func TestValidatePolicies_EnforcementSeverity(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")

	for _, d := range []string{
		filepath.Join(mosDir, "rules", "mechanical"),
		filepath.Join(mosDir, "contracts", "active", "CON-001"),
	} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	for _, tc := range []struct {
		enforcement string
		expected    Severity
	}{
		{"error", SeverityError},
		{"warning", SeverityWarning},
		{"info", SeverityInfo},
	} {
		t.Run(tc.enforcement, func(t *testing.T) {
			rule := `rule "sev-test" {
  name = "Severity Test"
  type = "mechanical"
  scope = "project"
  enforcement = "` + tc.enforcement + `"

  when {
    artifact_kind = "contract"
  }
}
`
			contract := `contract "CON-001" {
  title = "Test"
  status = "active"
}
`
			if err := os.WriteFile(filepath.Join(mosDir, "rules", "mechanical", "sev-test.mos"), []byte(rule), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(mosDir, "contracts", "active", "CON-001", "contract.mos"), []byte(contract), 0644); err != nil {
				t.Fatal(err)
			}

			ctx := &ProjectContext{
				Root:        root,
				RuleIDs:     map[string]string{"sev-test": filepath.Join(mosDir, "rules", "mechanical", "sev-test.mos")},
				ContractIDs: map[string]string{"CON-001": filepath.Join(mosDir, "contracts", "active", "CON-001", "contract.mos")},
				ArtifactIDs: map[string]map[string]string{},
			}

			diags := validatePolicies(ctx)

			found := false
			for _, d := range diags {
				if d.Rule == "policy-match" {
					if d.Severity != tc.expected {
						t.Errorf("expected severity %v, got %v", tc.expected, d.Severity)
					}
					found = true
				}
			}
			if !found {
				t.Error("expected policy-match diagnostic")
			}
		})
	}
}

func TestValidatePolicies_RuleWithoutWhen(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")

	for _, d := range []string{
		filepath.Join(mosDir, "rules", "mechanical"),
		filepath.Join(mosDir, "contracts", "active", "CON-001"),
	} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	rule := `rule "global-rule" {
  name = "Global"
  type = "mechanical"
  scope = "project"
  enforcement = "error"

  harness {
    command = "echo ok"
  }
}
`
	contract := `contract "CON-001" {
  title = "Test"
  status = "active"
}
`
	if err := os.WriteFile(filepath.Join(mosDir, "rules", "mechanical", "global-rule.mos"), []byte(rule), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mosDir, "contracts", "active", "CON-001", "contract.mos"), []byte(contract), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := &ProjectContext{
		Root:        root,
		RuleIDs:     map[string]string{"global-rule": filepath.Join(mosDir, "rules", "mechanical", "global-rule.mos")},
		ContractIDs: map[string]string{"CON-001": filepath.Join(mosDir, "contracts", "active", "CON-001", "contract.mos")},
		ArtifactIDs: map[string]map[string]string{},
	}

	diags := validatePolicies(ctx)

	for _, d := range diags {
		if d.Rule == "policy-match" {
			t.Errorf("expected no policy-match for rule without when block, got: %s", d.Message)
		}
	}
}
