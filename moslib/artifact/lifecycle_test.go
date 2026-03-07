package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCON030_FieldDefOrderedFlag(t *testing.T) {
	root := setupScaffold(t)
	configPath := filepath.Join(root, ".mos", "config.mos")
	data, _ := os.ReadFile(configPath)
	extra := `
  artifact_type "task" {
    directory = "tasks"
    fields {
      priority {
        enum = ["low", "medium", "high"]
        ordered = true
      }
    }
    lifecycle {
      active_states = ["open"]
      archive_states = ["closed"]
    }
  }
`
	patched := strings.Replace(string(data), "\n}\n", extra+"\n}\n", 1)
	os.WriteFile(configPath, []byte(patched), 0644)

	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	td := reg.Types["task"]
	if len(td.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(td.Fields))
	}
	if !td.Fields[0].Ordered {
		t.Error("expected priority field to be ordered")
	}
	if len(td.Fields[0].Enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(td.Fields[0].Enum))
	}
}

func TestCON030_ForwardOnlyEnforcement(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	err := ValidateScenarioTransition(reg, "implemented", "pending")
	if err == nil {
		t.Error("expected error for backward transition from implemented to pending")
	}

	err = ValidateScenarioTransition(reg, "pending", "implemented")
	if err != nil {
		t.Errorf("expected no error for forward transition, got: %v", err)
	}
}

func TestCON030_TransitionGateParsing(t *testing.T) {
	root := setupScaffold(t)
	configPath := filepath.Join(root, ".mos", "config.mos")
	data, _ := os.ReadFile(configPath)
	extra := `
  artifact_type "workflow" {
    directory = "workflows"
    fields {
      phase {
        enum = ["draft", "review", "approved"]
        ordered = true
        transitions {
          review {
            to = "approved"
            verified_by = "harness"
          }
        }
      }
    }
    lifecycle {
      active_states = ["open"]
      archive_states = ["done"]
    }
  }
`
	patched := strings.Replace(string(data), "\n}\n", extra+"\n}\n", 1)
	os.WriteFile(configPath, []byte(patched), 0644)

	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	td := reg.Types["workflow"]
	fd := td.Fields[0]
	if len(fd.Transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(fd.Transitions))
	}
	tr := fd.Transitions[0]
	if tr.From != "review" || tr.To != "approved" || tr.VerifiedBy != "harness" {
		t.Errorf("unexpected transition: %+v", tr)
	}
}

func TestCON030_VerifyRunsHarnessAndTransitions(t *testing.T) {
	root := setupScaffold(t)
	_, err := CreateRule(root, "RUL-VERIFY-001", RuleOpts{
		Name:        "Verify Test Rule",
		Type:        "mechanical",
		Scope:       "unit",
		Enforcement: "error",
		HarnessCmd:  "echo ok",
	})
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	createHookTestContract(t, root, "CON-VER-001", "active", map[string]string{
		"Alpha": "implemented",
	})

	results, err := VerifyContract(root, "CON-VER-001")
	if err != nil {
		t.Fatalf("VerifyContract: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Pass {
		t.Errorf("expected scenario to pass verification")
	}

	scenarios, _ := ListScenarios(root, "CON-VER-001")
	for _, s := range scenarios {
		if s.Name == "Alpha" && s.Status != "verified" {
			t.Errorf("expected Alpha status 'verified', got %q", s.Status)
		}
	}
}

func TestCON030_FailedHarnessKeepsImplemented(t *testing.T) {
	root := setupScaffold(t)
	_, err := CreateRule(root, "RUL-VERIFY-002", RuleOpts{
		Name:        "Failing Harness Rule",
		Type:        "mechanical",
		Scope:       "unit",
		Enforcement: "error",
		HarnessCmd:  "exit 1",
	})
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	createHookTestContract(t, root, "CON-VER-002", "active", map[string]string{
		"Beta": "implemented",
	})

	results, err := VerifyContract(root, "CON-VER-002")
	if err != nil {
		t.Fatalf("VerifyContract: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Pass {
		t.Error("expected scenario to fail verification")
	}

	scenarios, _ := ListScenarios(root, "CON-VER-002")
	for _, s := range scenarios {
		if s.Name == "Beta" && s.Status != "implemented" {
			t.Errorf("expected Beta status 'implemented' after failed verify, got %q", s.Status)
		}
	}
}

func TestCON030_DefaultScenarioSchemaHasOrderedEnum(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["contract"]
	if len(td.ScenarioFields) == 0 {
		t.Fatal("expected contract type to have scenario fields defined")
	}
	sf := td.ScenarioFields[0]
	if sf.Name != "status" {
		t.Fatalf("expected first scenario field to be 'status', got %q", sf.Name)
	}
	if !sf.Ordered {
		t.Error("expected scenario status to be ordered")
	}
	if len(sf.Enum) != 3 || sf.Enum[0] != "pending" || sf.Enum[1] != "implemented" || sf.Enum[2] != "verified" {
		t.Errorf("unexpected scenario enum: %v", sf.Enum)
	}
	if len(sf.Transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(sf.Transitions))
	}
	tr := sf.Transitions[0]
	if tr.From != "implemented" || tr.To != "verified" || tr.VerifiedBy != "harness" {
		t.Errorf("unexpected transition: %+v", tr)
	}
}

// --- CON-2026-031: Ledger Primitive ---
