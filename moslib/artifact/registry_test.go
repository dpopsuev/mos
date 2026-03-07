package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefineCustomArtifactTypeInConfig(t *testing.T) {
	root := setupScaffold(t)
	configContent := `config {
  mos { version = 1 }
  backend { type = "git" }
  governance { model = "bdfl" scope = "cabinet" }

  artifact_type "initiative" {
    directory = "initiatives"

    fields {
      title    { required = true }
      status   { required = true }
      priority { enum = ["critical", "high", "medium", "low"] }
    }

    lifecycle {
      active_states  = ["proposed", "approved", "in-progress"]
      archive_states = ["delivered", "retired"]
    }
  }
}
`
	os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(configContent), 0644)

	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	td, ok := reg.Types["initiative"]
	if !ok {
		t.Fatal("initiative type not found in registry")
	}
	if td.Directory != "initiatives" {
		t.Errorf("Directory = %q, want initiatives", td.Directory)
	}
	if len(td.Fields) != 3 {
		t.Errorf("Fields count = %d, want 3", len(td.Fields))
	}
	if len(td.Lifecycle.ActiveStates) != 3 {
		t.Errorf("ActiveStates = %v, want 3 states", td.Lifecycle.ActiveStates)
	}
	if len(td.Lifecycle.ArchiveStates) != 2 {
		t.Errorf("ArchiveStates = %v, want 2 states", td.Lifecycle.ArchiveStates)
	}
}

func TestCADTypesFromConfigMos(t *testing.T) {
	root := setupScaffold(t)
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	contract, ok := reg.Types["contract"]
	if !ok {
		t.Fatal("contract not in registry from config.mos")
	}
	if contract.Directory != "contracts" {
		t.Errorf("contract.Directory = %q, want contracts", contract.Directory)
	}

	hasTitle := false
	hasStatus := false
	for _, f := range contract.Fields {
		if f.Name == "title" && f.Required {
			hasTitle = true
		}
		if f.Name == "status" && f.Required {
			hasStatus = true
		}
	}
	if !hasTitle || !hasStatus {
		t.Error("contract missing required title/status fields")
	}

	rule, ok := reg.Types["rule"]
	if !ok {
		t.Fatal("rule not in registry from config.mos")
	}
	if rule.Directory != "rules" {
		t.Errorf("rule.Directory = %q, want rules", rule.Directory)
	}

	spec, ok := reg.Types["specification"]
	if !ok {
		t.Fatal("specification not in registry from config.mos")
	}
	hasNonGoals := false
	for _, f := range spec.Fields {
		if f.Name == "non_goals" {
			hasNonGoals = true
		}
	}
	if !hasNonGoals {
		t.Error("specification should have non_goals field")
	}

	if _, ok := reg.Types["binder"]; !ok {
		t.Fatal("binder not in registry from config.mos")
	}
}

func TestCADCustomFieldsInConfigMos(t *testing.T) {
	root := setupScaffold(t)
	configContent := `config {
  mos { version = 1 }
  backend { type = "git" }
  governance { model = "bdfl" scope = "cabinet" }

  artifact_type "contract" {
    directory = "contracts"
    fields {
      title { required = true }
      status { required = true }
      priority { enum = ["p0", "p1", "p2"] }
    }
  }
}
`
	os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(configContent), 0644)

	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	contract := reg.Types["contract"]

	hasTitle := false
	hasStatus := false
	hasPriority := false
	for _, f := range contract.Fields {
		if f.Name == "title" && f.Required {
			hasTitle = true
		}
		if f.Name == "status" && f.Required {
			hasStatus = true
		}
		if f.Name == "priority" {
			hasPriority = true
		}
	}
	if !hasTitle || !hasStatus {
		t.Error("required fields must be present")
	}
	if !hasPriority {
		t.Error("extended field 'priority' should be present")
	}
}

// Feature 2: Generic CRUD

func TestCreateInstanceOfCustomArtifactType(t *testing.T) {
	root := setupScaffold(t)
	configContent := `config {
  mos { version = 1 }
  backend { type = "git" }
  governance { model = "bdfl" scope = "cabinet" }

  artifact_type "initiative" {
    directory = "initiatives"

    fields {
      title  { required = true }
      status { required = true }
    }

    lifecycle {
      active_states  = ["proposed", "approved"]
      archive_states = ["delivered"]
    }
  }
}
`
	os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(configContent), 0644)

	reg, _ := LoadRegistry(root)
	td := reg.Types["initiative"]

	path, err := GenericCreate(root, td, "INIT-001", map[string]string{
		"title":  "Q3 Platform",
		"status": "proposed",
	})
	if err != nil {
		t.Fatalf("GenericCreate: %v", err)
	}

	expected := filepath.Join(root, ".mos", "initiatives", "active", "INIT-001", "initiative.mos")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), `initiative "INIT-001"`) {
		t.Error("file should use the initiative keyword")
	}
	if !strings.Contains(string(content), `title = "Q3 Platform"`) {
		t.Error("file should contain the title field")
	}
}

func TestApplyWorksForCustomTypes(t *testing.T) {
	root := setupScaffold(t)
	configContent := `config {
  mos { version = 1 }
  backend { type = "git" }
  governance { model = "bdfl" scope = "cabinet" }

  artifact_type "initiative" {
    directory = "initiatives"

    fields {
      title  { required = true }
      status { required = true }
    }

    lifecycle {
      active_states  = ["proposed", "approved"]
      archive_states = ["delivered"]
    }
  }
}
`
	os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(configContent), 0644)

	content := []byte(`initiative "INIT-002" {
  title = "Q4 Migration"
  status = "proposed"
}
`)
	resultPath, err := ApplyArtifact(root, content)
	if err != nil {
		t.Fatalf("ApplyArtifact: %v", err)
	}

	expected := filepath.Join(root, ".mos", "initiatives", "active", "INIT-002", "initiative.mos")
	if resultPath != expected {
		t.Errorf("path = %q, want %q", resultPath, expected)
	}
}

func TestDeleteCustomArtifactInstance(t *testing.T) {
	root := setupScaffold(t)
	configContent := `config {
  mos { version = 1 }
  backend { type = "git" }
  governance { model = "bdfl" scope = "cabinet" }

  artifact_type "initiative" {
    directory = "initiatives"

    fields {
      title  { required = true }
      status { required = true }
    }

    lifecycle {
      active_states  = ["proposed"]
      archive_states = ["delivered"]
    }
  }
}
`
	os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(configContent), 0644)

	reg, _ := LoadRegistry(root)
	td := reg.Types["initiative"]
	GenericCreate(root, td, "DEL-001", map[string]string{
		"title": "To Delete", "status": "proposed",
	})

	if err := GenericDelete(root, td, "DEL-001"); err != nil {
		t.Fatalf("GenericDelete: %v", err)
	}

	dir := filepath.Join(root, ".mos", "initiatives", "active", "DEL-001")
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("directory should be removed after delete")
	}
}

// Feature 3: Dynamic Linter Validation

func TestLinterValidatesAgainstSchemaDefinedFields(t *testing.T) {
	root := setupScaffold(t)
	configContent := `config {
  mos { version = 1 }
  backend { type = "git" }
  governance { model = "bdfl" scope = "cabinet" }

  artifact_type "initiative" {
    directory = "initiatives"

    fields {
      title    { required = true }
      status   { required = true }
      priority { enum = ["critical", "high", "medium", "low"] }
    }

    lifecycle {
      active_states  = ["proposed"]
      archive_states = ["delivered"]
    }
  }
}
`
	os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(configContent), 0644)

	// Create an initiative missing required title and with bad priority
	dir := filepath.Join(root, ".mos", "initiatives", "active", "BAD-001")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "initiative.mos"), []byte(`initiative "BAD-001" {
  status = "proposed"
  priority = "urgent"
}
`), 0644)

	diags, err := LintAll(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}

	hasRequired := false
	hasEnum := false
	for _, d := range diags {
		if d.Rule == "schema/required-field" && strings.Contains(d.Message, "title") {
			hasRequired = true
		}
		if d.Rule == "schema/enum" && strings.Contains(d.Message, "priority") {
			hasEnum = true
		}
	}
	if !hasRequired {
		t.Error("expected required-field error for missing title")
	}
	if !hasEnum {
		t.Error("expected enum error for invalid priority")
	}
}

func TestUnknownFieldsProduceWarnings(t *testing.T) {
	root := setupScaffold(t)
	configContent := `config {
  mos { version = 1 }
  backend { type = "git" }
  governance { model = "bdfl" scope = "cabinet" }

  artifact_type "initiative" {
    directory = "initiatives"

    fields {
      title  { required = true }
      status { required = true }
    }

    lifecycle {
      active_states  = ["proposed"]
      archive_states = ["delivered"]
    }
  }
}
`
	os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(configContent), 0644)

	dir := filepath.Join(root, ".mos", "initiatives", "active", "UNK-001")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "initiative.mos"), []byte(`initiative "UNK-001" {
  title = "Test"
  status = "proposed"
  extra_field = "surprise"
}
`), 0644)

	diags, err := LintAll(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}

	hasWarning := false
	hasError := false
	for _, d := range diags {
		if d.Rule == "schema/unknown-field" {
			if d.Severity == "warning" {
				hasWarning = true
			}
			if d.Severity == "error" {
				hasError = true
			}
		}
	}
	if !hasWarning {
		t.Error("expected warning for unknown field")
	}
	if hasError {
		t.Error("unknown field should be warning, not error")
	}
}

// Feature 4: Lifecycle State Machine

func TestCustomLifecycleStatesPerArtifactType(t *testing.T) {
	root := setupScaffold(t)
	configContent := `config {
  mos { version = 1 }
  backend { type = "git" }
  governance { model = "bdfl" scope = "cabinet" }

  artifact_type "initiative" {
    directory = "initiatives"

    fields {
      title  { required = true }
      status { required = true }
    }

    lifecycle {
      active_states  = ["proposed", "approved", "in-progress"]
      archive_states = ["delivered", "retired"]
    }
  }
}
`
	os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(configContent), 0644)

	reg, _ := LoadRegistry(root)
	td := reg.Types["initiative"]

	GenericCreate(root, td, "LIFE-001", map[string]string{
		"title": "Lifecycle Test", "status": "proposed",
	})

	// Transition to approved (still active)
	if err := GenericUpdateStatus(root, td, "LIFE-001", "approved"); err != nil {
		t.Fatalf("status -> approved: %v", err)
	}
	activePath := filepath.Join(root, ".mos", "initiatives", "active", "LIFE-001", "initiative.mos")
	if _, err := os.Stat(activePath); err != nil {
		t.Error("approved should be in active directory")
	}

	// Transition to delivered (archive)
	if err := GenericUpdateStatus(root, td, "LIFE-001", "delivered"); err != nil {
		t.Fatalf("status -> delivered: %v", err)
	}
	archivePath := filepath.Join(root, ".mos", "initiatives", "archive", "LIFE-001", "initiative.mos")
	if _, err := os.Stat(archivePath); err != nil {
		t.Error("delivered should be in archive directory")
	}
	if _, err := os.Stat(activePath); !os.IsNotExist(err) {
		t.Error("old active directory should be removed")
	}
}

func TestRejectInvalidStatusTransition(t *testing.T) {
	root := setupScaffold(t)
	configContent := `config {
  mos { version = 1 }
  backend { type = "git" }
  governance { model = "bdfl" scope = "cabinet" }

  artifact_type "initiative" {
    directory = "initiatives"

    fields {
      title  { required = true }
      status { required = true }
    }

    lifecycle {
      active_states  = ["proposed", "approved"]
      archive_states = ["delivered"]
    }
  }
}
`
	os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(configContent), 0644)

	reg, _ := LoadRegistry(root)
	td := reg.Types["initiative"]

	GenericCreate(root, td, "INV-001", map[string]string{
		"title": "Invalid Test", "status": "proposed",
	})

	err := GenericUpdateStatus(root, td, "INV-001", "cancelled")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("error should mention invalid status, got: %v", err)
	}
	if !strings.Contains(err.Error(), "proposed") {
		t.Errorf("error should list valid states, got: %v", err)
	}
}

// Feature 5: Backward Compatibility

func TestExistingContractsAndRulesWorkUnchanged(t *testing.T) {
	root := setupScaffold(t)

	// No artifact_type blocks -- vanilla config
	contractPath, err := CreateContract(root, "BC-001", ContractOpts{Title: "Backward Compat"})
	if err != nil {
		t.Fatalf("CreateContract: %v", err)
	}
	assertParses(t, contractPath)

	contracts, err := ListContracts(root, ListOpts{})
	if err != nil {
		t.Fatalf("ListContracts: %v", err)
	}
	if len(contracts) != 1 || contracts[0].ID != "BC-001" {
		t.Errorf("expected [BC-001], got %v", contracts)
	}

	rulePath, err := CreateRule(root, "bc-rule", RuleOpts{
		Name: "backward-compat", Type: "interpretive", Enforcement: "error", Scope: "project",
	})
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}
	assertParses(t, rulePath)

	_, err = LintAll(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
}

func TestGradualAdoption(t *testing.T) {
	root := setupScaffold(t)
	configContent := `config {
  mos { version = 1 }
  backend { type = "git" }
  governance { model = "bdfl" scope = "cabinet" }

  artifact_type "initiative" {
    directory = "initiatives"

    fields {
      title  { required = true }
      status { required = true }
    }

    lifecycle {
      active_states  = ["proposed"]
      archive_states = ["delivered"]
    }
  }
}
`
	os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(configContent), 0644)

	// Create a contract (built-in type)
	CreateContract(root, "MIX-001", ContractOpts{Title: "Mixed"})

	// Create an initiative (custom type)
	reg, _ := LoadRegistry(root)
	td := reg.Types["initiative"]
	GenericCreate(root, td, "INIT-MIX", map[string]string{
		"title": "Custom Initiative", "status": "proposed",
	})

	// Both should coexist -- contract list works
	contracts, _ := ListContracts(root, ListOpts{})
	if len(contracts) != 1 {
		t.Errorf("expected 1 contract, got %d", len(contracts))
	}

	// Initiative list works
	initiatives, _ := GenericList(root, td, "")
	if len(initiatives) != 1 {
		t.Errorf("expected 1 initiative, got %d", len(initiatives))
	}

	// Lint validates both
	diags, err := LintAll(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("unexpected lint error: %s: %s", d.Rule, d.Message)
		}
	}
}

func TestAddProject(t *testing.T) {
	root := setupScaffold(t)
	seedProjectConfig(t, root)

	if err := AddProject(root, "incidents", "INC"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}

	projects, err := LoadProjects(root)
	if err != nil {
		t.Fatalf("LoadProjects: %v", err)
	}
	var found bool
	for _, p := range projects {
		if p.Name == "incidents" && p.Prefix == "INC" && p.Sequence == 0 {
			found = true
		}
	}
	if !found {
		t.Errorf("project 'incidents' not found after AddProject")
	}
}

func TestAddProjectDuplicate(t *testing.T) {
	root := setupScaffold(t)
	seedProjectConfig(t, root)

	err := AddProject(root, "contracts", "CON")
	if err == nil {
		t.Fatal("expected error adding duplicate project, got nil")
	}
}

func TestAddArtifactType(t *testing.T) {
	root := setupScaffold(t)
	seedProjectConfig(t, root)

	if err := AddArtifactType(root, "report", ""); err != nil {
		t.Fatalf("AddArtifactType: %v", err)
	}

	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	td, ok := reg.Types["report"]
	if !ok {
		t.Fatal("artifact_type 'report' not found after AddArtifactType")
	}
	if td.Directory != "reports" {
		t.Errorf("directory = %q, want 'reports'", td.Directory)
	}
	if len(td.Lifecycle.ActiveStates) != 2 {
		t.Errorf("expected 2 active states, got %d", len(td.Lifecycle.ActiveStates))
	}
	if len(td.Lifecycle.ArchiveStates) != 2 {
		t.Errorf("expected 2 archive states, got %d", len(td.Lifecycle.ArchiveStates))
	}
}

func TestAddArtifactTypeDuplicate(t *testing.T) {
	root := setupScaffold(t)
	seedProjectConfig(t, root)

	if err := AddArtifactType(root, "report", ""); err != nil {
		t.Fatalf("first AddArtifactType: %v", err)
	}
	err := AddArtifactType(root, "report", "")
	if err == nil {
		t.Fatal("expected error adding duplicate artifact_type, got nil")
	}
}

func TestAddArtifactTypeCustomDir(t *testing.T) {
	root := setupScaffold(t)
	seedProjectConfig(t, root)

	if err := AddArtifactType(root, "review", "code-reviews"); err != nil {
		t.Fatalf("AddArtifactType: %v", err)
	}

	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	td, ok := reg.Types["review"]
	if !ok {
		t.Fatal("artifact_type 'review' not found")
	}
	if td.Directory != "code-reviews" {
		t.Errorf("directory = %q, want 'code-reviews'", td.Directory)
	}
}

func TestResolveKindFromID(t *testing.T) {
	root := setupScaffold(t)
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	kind, err := reg.ResolveKindFromID("CON-2026-125")
	if err != nil {
		t.Fatalf("resolve CON: %v", err)
	}
	if kind != "contract" {
		t.Errorf("expected %q, got %q", "contract", kind)
	}

	kind, err = reg.ResolveKindFromID("SPEC-2026-001")
	if err != nil {
		t.Fatalf("resolve SPEC: %v", err)
	}
	if kind != "specification" {
		t.Errorf("expected specification, got %q", kind)
	}

	kind, err = reg.ResolveKindFromID("BUG-2026-001")
	if err != nil {
		t.Fatalf("resolve BUG: %v", err)
	}
	if kind != "contract" {
		t.Errorf("expected %q for BUG prefix (shares contract directory), got %q", "contract", kind)
	}

	_, err = reg.ResolveKindFromID("UNKNOWN-2026-001")
	if err == nil {
		t.Error("expected error for unknown prefix")
	}
}
