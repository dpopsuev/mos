package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupMigrationScaffold(t *testing.T) string {
	t.Helper()
	root := setupScaffold(t)
	configPath := filepath.Join(root, ".mos", "config.mos")
	data, _ := os.ReadFile(configPath)
	extra := `
  artifact_type "ticket" {
    directory = "tickets"
    version = "2"
    fields {
      title {
        required = true
      }
      priority {
        default = "medium"
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

	os.MkdirAll(filepath.Join(root, ".mos", "tickets", "active"), 0755)
	os.MkdirAll(filepath.Join(root, ".mos", "tickets", "archive"), 0755)
	return root
}

func TestCON028_FieldDefDefault(t *testing.T) {
	root := setupMigrationScaffold(t)
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	td := reg.Types["ticket"]
	for _, f := range td.Fields {
		if f.Name == "priority" && f.Default != "medium" {
			t.Errorf("expected priority default 'medium', got %q", f.Default)
		}
	}
}

func TestCON028_ArtifactTypeDefVersion(t *testing.T) {
	root := setupMigrationScaffold(t)
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	td := reg.Types["ticket"]
	if td.Version != "2" {
		t.Errorf("expected version '2', got %q", td.Version)
	}
}

func TestCON028_DryRunListsMissingFields(t *testing.T) {
	root := setupMigrationScaffold(t)

	instanceDir := filepath.Join(root, ".mos", "tickets", "active", "TIK-001")
	os.MkdirAll(instanceDir, 0755)
	os.WriteFile(filepath.Join(instanceDir, "ticket.mos"), []byte(`ticket "TIK-001" {
  title = "Missing priority"
  status = "open"
}
`), 0644)

	reg, _ := LoadRegistry(root)
	diffs, err := ComputeMigration(root, reg)
	if err != nil {
		t.Fatalf("ComputeMigration: %v", err)
	}

	found := false
	for _, d := range diffs {
		if d.ID == "TIK-001" && !d.UpToDate {
			found = true
			if len(d.Missing) != 1 || d.Missing[0].Name != "priority" {
				t.Errorf("expected missing priority field, got %v", d.Missing)
			}
		}
	}
	if !found {
		t.Error("expected TIK-001 to need migration")
	}
}

func TestCON028_ApplyInsertsMissingFields(t *testing.T) {
	root := setupMigrationScaffold(t)

	instanceDir := filepath.Join(root, ".mos", "tickets", "active", "TIK-002")
	os.MkdirAll(instanceDir, 0755)
	instancePath := filepath.Join(instanceDir, "ticket.mos")
	os.WriteFile(instancePath, []byte(`ticket "TIK-002" {
  title = "Needs migration"
  status = "open"
}
`), 0644)

	reg, _ := LoadRegistry(root)
	diffs, _ := ComputeMigration(root, reg)
	count, err := ApplyMigration(diffs)
	if err != nil {
		t.Fatalf("ApplyMigration: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 migration, got %d", count)
	}

	data, _ := os.ReadFile(instancePath)
	if !strings.Contains(string(data), `priority = "medium"`) {
		t.Error("expected priority field to be backfilled")
	}
}

func TestCON028_UpToDateInstancesSkipped(t *testing.T) {
	root := setupMigrationScaffold(t)

	instanceDir := filepath.Join(root, ".mos", "tickets", "active", "TIK-003")
	os.MkdirAll(instanceDir, 0755)
	instancePath := filepath.Join(instanceDir, "ticket.mos")
	os.WriteFile(instancePath, []byte(`ticket "TIK-003" {
  title = "Already has priority"
  status = "open"
  priority = "high"
}
`), 0644)

	reg, _ := LoadRegistry(root)
	diffs, _ := ComputeMigration(root, reg)

	for _, d := range diffs {
		if d.ID == "TIK-003" && !d.UpToDate {
			t.Error("expected TIK-003 to be up to date")
		}
	}

	before, _ := os.ReadFile(instancePath)
	ApplyMigration(diffs)
	after, _ := os.ReadFile(instancePath)
	if string(before) != string(after) {
		t.Error("expected up-to-date instance to not be modified")
	}
}

func TestCON028_BuiltInTypesCanDefineDefaults(t *testing.T) {
	root := setupMigrationScaffold(t)

	configPath := filepath.Join(root, ".mos", "config.mos")
	data, _ := os.ReadFile(configPath)
	extra := `
  artifact_type "contract" {
    fields {
      kind {
        default = "feature"
      }
    }
  }
`
	patched := strings.Replace(string(data), "\n}\n", extra+"\n}\n", 1)
	os.WriteFile(configPath, []byte(patched), 0644)

	CreateContract(root, "CON-NKIND", ContractOpts{Title: "No kind", Status: "draft"})

	reg, _ := LoadRegistry(root)
	diffs, _ := ComputeMigration(root, reg)

	var found bool
	for _, d := range diffs {
		if d.ID == "CON-NKIND" && !d.UpToDate {
			found = true
		}
	}
	if !found {
		t.Error("expected CON-NKIND to need migration for kind field")
	}

	ApplyMigration(diffs)

	contractPath := filepath.Join(root, ".mos", "contracts", "active", "CON-NKIND", "contract.mos")
	content, _ := os.ReadFile(contractPath)
	if !strings.Contains(string(content), `kind = "feature"`) {
		t.Error("expected kind field to be backfilled with default 'feature'")
	}
}

// --- CON-2026-029: Lifecycle Hooks ---
