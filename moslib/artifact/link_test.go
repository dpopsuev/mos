package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupLinkFieldScaffold(t *testing.T) string {
	t.Helper()
	root := setupScaffold(t)
	configPath := filepath.Join(root, ".mos", "config.mos")
	data, _ := os.ReadFile(configPath)
	extra := `
  artifact_type "task" {
    directory = "tasks"
    fields {
      title { required = true }
      assigned_to {
        link = true
        ref_kind = "team"
      }
      depends_on {
        link = true
        ref_kind = "task"
      }
      priority {}
      status {
        required = true
        enum = ["open", "done"]
      }
    }
    lifecycle {
      active_states = ["open"]
      archive_states = ["done"]
    }
  }

  artifact_type "team" {
    directory = "teams"
    fields {
      title { required = true }
      status {
        required = true
        enum = ["active", "retired"]
      }
    }
    lifecycle {
      active_states = ["active"]
      archive_states = ["retired"]
    }
  }
`
	patched := strings.Replace(string(data), "\n}\n", extra+"\n}\n", 1)
	os.WriteFile(configPath, []byte(patched), 0644)

	os.MkdirAll(filepath.Join(root, ".mos", "tasks", "active"), 0755)
	os.MkdirAll(filepath.Join(root, ".mos", "tasks", "archive"), 0755)
	os.MkdirAll(filepath.Join(root, ".mos", "teams", "active"), 0755)
	os.MkdirAll(filepath.Join(root, ".mos", "teams", "archive"), 0755)
	return root
}

func TestCON125_FieldDefLinkFlag(t *testing.T) {
	root := setupLinkFieldScaffold(t)
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	td := reg.Types["task"]
	var found bool
	for _, f := range td.Fields {
		if f.Name == "assigned_to" {
			found = true
			if !f.Link {
				t.Error("expected assigned_to.Link = true")
			}
			if f.RefKind != "team" {
				t.Errorf("expected assigned_to.RefKind = 'team', got %q", f.RefKind)
			}
		}
	}
	if !found {
		t.Error("field 'assigned_to' not found in task type")
	}
}

func TestCON125_FieldDefRefKind(t *testing.T) {
	root := setupLinkFieldScaffold(t)
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	td := reg.Types["task"]
	for _, f := range td.Fields {
		if f.Name == "depends_on" {
			if !f.Link {
				t.Error("expected depends_on.Link = true")
			}
			if f.RefKind != "task" {
				t.Errorf("expected depends_on.RefKind = 'task', got %q", f.RefKind)
			}
			return
		}
	}
	t.Error("field 'depends_on' not found in task type")
}

func TestCON125_LinkFieldsHelper(t *testing.T) {
	root := setupLinkFieldScaffold(t)
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	td := reg.Types["task"]
	links := td.LinkFields()
	if len(links) != 2 {
		t.Fatalf("expected 2 link fields, got %d: %v", len(links), links)
	}
	linkSet := make(map[string]bool)
	for _, l := range links {
		linkSet[l] = true
	}
	if !linkSet["assigned_to"] || !linkSet["depends_on"] {
		t.Errorf("expected assigned_to and depends_on in link fields, got %v", links)
	}
}

func TestCON125_AllLinkFieldsRegistry(t *testing.T) {
	root := setupLinkFieldScaffold(t)
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	all := reg.AllLinkFields()
	allSet := make(map[string]bool)
	for _, f := range all {
		allSet[f] = true
	}
	if !allSet["assigned_to"] || !allSet["depends_on"] {
		t.Errorf("expected assigned_to and depends_on in AllLinkFields, got %v", all)
	}
}

func TestCON125_DefaultRegistryContractLinkFields(t *testing.T) {
	reg := DefaultRegistry()
	td := reg.Types["contract"]
	links := td.LinkFields()
	linkSet := make(map[string]bool)
	for _, l := range links {
		linkSet[l] = true
	}
	for _, expected := range []string{"justifies", "implements", "documents", "sprint", "batch", "parent", "depends_on"} {
		if !linkSet[expected] {
			t.Errorf("expected %q in default contract link fields", expected)
		}
	}
}

func TestCON125_TraceFieldsExcludesSelfAndOrg(t *testing.T) {
	reg := DefaultRegistry()
	td := reg.Types["contract"]
	trace := td.TraceFields()
	traceSet := make(map[string]bool)
	for _, f := range trace {
		traceSet[f] = true
	}
	if !traceSet["justifies"] {
		t.Error("expected justifies in TraceFields")
	}
	if !traceSet["implements"] {
		t.Error("expected implements in TraceFields")
	}
	if !traceSet["documents"] {
		t.Error("expected documents in TraceFields")
	}
	if traceSet["sprint"] {
		t.Error("sprint should not be in TraceFields")
	}
	if traceSet["batch"] {
		t.Error("batch should not be in TraceFields")
	}
	if traceSet["parent"] {
		t.Error("parent should not be in TraceFields (self-referential)")
	}
	if traceSet["depends_on"] {
		t.Error("depends_on should not be in TraceFields (self-referential)")
	}
}

func TestCON125_QueryReferencesBatch(t *testing.T) {
	root := setupScaffold(t)
	configPath := filepath.Join(root, ".mos", "config.mos")
	data, _ := os.ReadFile(configPath)
	extra := `
  artifact_type "batch" {
    directory = "batches"
    fields {
      title { required = true }
      goal { required = true }
      status {
        required = true
        enum = ["planned", "ready", "promoted", "cancelled"]
      }
    }
    lifecycle {
      active_states = ["planned", "ready"]
      archive_states = ["promoted", "cancelled"]
    }
  }
`
	patched := strings.Replace(string(data), "\n}\n", extra+"\n}\n", 1)
	os.WriteFile(configPath, []byte(patched), 0644)

	os.MkdirAll(filepath.Join(root, ".mos", "batches", "active"), 0755)
	os.MkdirAll(filepath.Join(root, ".mos", "batches", "active", "BAT-001"), 0755)
	os.WriteFile(filepath.Join(root, ".mos", "batches", "active", "BAT-001", "batch.mos"),
		[]byte(`batch "BAT-001" { title = "Batch One" goal = "test" status = "planned" }`), 0644)

	cdir := filepath.Join(root, ".mos", "contracts", "active", "CON-001")
	os.MkdirAll(cdir, 0755)
	os.WriteFile(filepath.Join(cdir, "contract.mos"),
		[]byte(`contract "CON-001" { title = "Batched contract" status = "draft" batch = "BAT-001" }`), 0644)

	results, err := QueryArtifacts(root, QueryOpts{References: "BAT-001"})
	if err != nil {
		t.Fatalf("QueryArtifacts: %v", err)
	}
	found := false
	for _, r := range results {
		if r.ID == "CON-001" {
			found = true
		}
	}
	if !found {
		t.Error("expected CON-001 to be found when querying references to BAT-001")
	}
}

func TestCON125_WriteTimeValidation_RejectWrongKind(t *testing.T) {
	root := setupLinkFieldScaffold(t)

	teamDir := filepath.Join(root, ".mos", "teams", "active", "TEAM-001")
	os.MkdirAll(teamDir, 0755)
	os.WriteFile(filepath.Join(teamDir, "team.mos"),
		[]byte(`team "TEAM-001" { title = "Alpha" status = "active" }`), 0644)

	taskDir := filepath.Join(root, ".mos", "tasks", "active", "TASK-001")
	os.MkdirAll(taskDir, 0755)
	os.WriteFile(filepath.Join(taskDir, "task.mos"),
		[]byte(`task "TASK-001" { title = "First" status = "open" }`), 0644)

	reg, _ := LoadRegistry(root)
	td := reg.Types["task"]

	err := GenericUpdate(root, td, "TASK-001", map[string]string{
		"assigned_to": "TASK-001",
	})
	if err == nil {
		t.Error("expected error when assigning a task ID to a team-typed link field")
	}
}

func TestCON125_WriteTimeValidation_AcceptCorrectKind(t *testing.T) {
	root := setupLinkFieldScaffold(t)

	teamDir := filepath.Join(root, ".mos", "teams", "active", "TEAM-001")
	os.MkdirAll(teamDir, 0755)
	os.WriteFile(filepath.Join(teamDir, "team.mos"),
		[]byte(`team "TEAM-001" { title = "Alpha" status = "active" }`), 0644)

	taskDir := filepath.Join(root, ".mos", "tasks", "active", "TASK-001")
	os.MkdirAll(taskDir, 0755)
	os.WriteFile(filepath.Join(taskDir, "task.mos"),
		[]byte(`task "TASK-001" { title = "First" status = "open" }`), 0644)

	reg, _ := LoadRegistry(root)
	td := reg.Types["task"]

	err := GenericUpdate(root, td, "TASK-001", map[string]string{
		"assigned_to": "TEAM-001",
	})
	if err != nil {
		t.Errorf("expected no error for valid link, got: %v", err)
	}
}

func TestCON125_WriteTimeValidation_TargetNotExist(t *testing.T) {
	root := setupLinkFieldScaffold(t)

	taskDir := filepath.Join(root, ".mos", "tasks", "active", "TASK-001")
	os.MkdirAll(taskDir, 0755)
	os.WriteFile(filepath.Join(taskDir, "task.mos"),
		[]byte(`task "TASK-001" { title = "First" status = "open" }`), 0644)

	reg, _ := LoadRegistry(root)
	td := reg.Types["task"]

	err := GenericUpdate(root, td, "TASK-001", map[string]string{
		"assigned_to": "TEAM-NONEXISTENT",
	})
	if err == nil {
		t.Error("expected error for nonexistent target artifact")
	}
}

// TestCON125_SprintTransitiveMembership moved to governance/audit/audit_test.go
// to avoid import cycles (test needs audit.RunAudit, which imports governance).

func TestCON125_MergeFieldDefsPreservesDefaults(t *testing.T) {
	root := setupScaffold(t)
	configPath := filepath.Join(root, ".mos", "config.mos")
	data, _ := os.ReadFile(configPath)
	extra := `
  artifact_type "contract" {
    directory = "contracts"
    fields {
      title { required = true }
      kind { default = "feature" }
      status { required = true }
    }
  }
`
	patched := strings.Replace(string(data), "\n}\n", extra+"\n}\n", 1)
	os.WriteFile(configPath, []byte(patched), 0644)

	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	td := reg.Types["contract"]
	links := td.LinkFields()
	linkSet := make(map[string]bool)
	for _, l := range links {
		linkSet[l] = true
	}
	if !linkSet["justifies"] {
		t.Error("expected default link field 'justifies' to be preserved after merge")
	}
	if !linkSet["sprint"] {
		t.Error("expected default link field 'sprint' to be preserved after merge")
	}
	if !linkSet["batch"] {
		t.Error("expected default link field 'batch' to be preserved after merge")
	}
}
