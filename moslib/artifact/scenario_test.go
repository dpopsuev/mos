package artifact

import (
	"testing"

	"github.com/dpopsuev/mos/moslib/dsl"
)

func createContractWithScenarios(t *testing.T, root string) {
	t.Helper()
	content := []byte(`contract "CON-SCEN-001" {
  title = "Test Scenarios"
  status = "active"
  kind = "feature"

  feature "Widget Management" {

    scenario "Create widget" {
      status = "done"
      given {
        a valid workspace
      }
      when {
        widget create is run
      }
      then {
        widget is created
      }
    }

    scenario "Delete widget" {
      given {
        an existing widget
      }
      when {
        widget delete is run
      }
      then {
        widget is removed
      }
    }

    scenario "Update widget" {
      status = "done"
      given {
        an existing widget
      }
      when {
        widget update is run
      }
      then {
        widget is updated
      }
    }
  }
}
`)
	if _, err := ApplyArtifact(root, content); err != nil {
		t.Fatalf("creating contract with scenarios: %v", err)
	}
}

func TestCON025_ListScenariosWithStatus(t *testing.T) {
	root := setupScaffold(t)
	createContractWithScenarios(t, root)

	infos, err := ListScenarios(root, "CON-SCEN-001")
	if err != nil {
		t.Fatalf("ListScenarios: %v", err)
	}

	if len(infos) != 3 {
		t.Fatalf("expected 3 scenarios, got %d", len(infos))
	}

	expected := []ScenarioInfo{
		{Name: "Create widget", Status: "done"},
		{Name: "Delete widget", Status: "pending"},
		{Name: "Update widget", Status: "done"},
	}
	for i, want := range expected {
		if infos[i].Name != want.Name {
			t.Errorf("scenario %d: expected name %q, got %q", i, want.Name, infos[i].Name)
		}
		if infos[i].Status != want.Status {
			t.Errorf("scenario %d %q: expected status %q, got %q", i, want.Name, want.Status, infos[i].Status)
		}
	}
}

func TestCON025_MarkScenarioDone(t *testing.T) {
	root := setupScaffold(t)
	createContractWithScenarios(t, root)

	if err := SetScenarioStatus(root, "CON-SCEN-001", "Delete widget", "done"); err != nil {
		t.Fatalf("SetScenarioStatus: %v", err)
	}

	infos, err := ListScenarios(root, "CON-SCEN-001")
	if err != nil {
		t.Fatalf("ListScenarios: %v", err)
	}

	for _, s := range infos {
		if s.Name == "Delete widget" && s.Status != "done" {
			t.Errorf("expected Delete widget to be done, got %q", s.Status)
		}
	}

	// Verify other scenarios unchanged
	for _, s := range infos {
		if s.Name == "Create widget" && s.Status != "done" {
			t.Errorf("Create widget should still be done")
		}
		if s.Name == "Update widget" && s.Status != "done" {
			t.Errorf("Update widget should still be done")
		}
	}
}

func TestCON025_RevertScenarioToPending(t *testing.T) {
	root := setupScaffold(t)
	createContractWithScenarios(t, root)

	if err := SetScenarioStatus(root, "CON-SCEN-001", "Create widget", "pending"); err != nil {
		t.Fatalf("SetScenarioStatus: %v", err)
	}

	infos, err := ListScenarios(root, "CON-SCEN-001")
	if err != nil {
		t.Fatalf("ListScenarios: %v", err)
	}

	for _, s := range infos {
		if s.Name == "Create widget" && s.Status != "pending" {
			t.Errorf("expected Create widget to be pending, got %q", s.Status)
		}
	}
}

func TestCON025_MarkAllScenariosDone(t *testing.T) {
	root := setupScaffold(t)
	createContractWithScenarios(t, root)

	count, err := SetAllScenariosStatus(root, "CON-SCEN-001", "done")
	if err != nil {
		t.Fatalf("SetAllScenariosStatus: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 scenarios marked, got %d", count)
	}

	infos, err := ListScenarios(root, "CON-SCEN-001")
	if err != nil {
		t.Fatalf("ListScenarios: %v", err)
	}
	for _, s := range infos {
		if s.Status != "done" {
			t.Errorf("expected scenario %q to be done, got %q", s.Name, s.Status)
		}
	}
}

func TestCON025_FilterScenariosByStatus(t *testing.T) {
	root := setupScaffold(t)
	createContractWithScenarios(t, root)

	infos, err := ListScenarios(root, "CON-SCEN-001")
	if err != nil {
		t.Fatalf("ListScenarios: %v", err)
	}

	var pending []ScenarioInfo
	for _, s := range infos {
		if s.Status == "pending" {
			pending = append(pending, s)
		}
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending scenario, got %d", len(pending))
	}
	if pending[0].Name != "Delete widget" {
		t.Errorf("expected pending scenario to be 'Delete widget', got %q", pending[0].Name)
	}
}

// --- CON-2026-027: Rule-to-Contract Kind Binding ---

func TestScenarioLabelsOnDSL(t *testing.T) {
	src := `contract "CON-LBL" {
  title = "Label test"
  status = "draft"

  feature "Auth" {
    scenario "login success" {
      labels = ["happy_path", "smoke"]
      given {
        user exists
      }
      when {
        they login
      }
      then {
        dashboard shown
      }
    }
    scenario "login fail" {
      given {
        user exists
      }
      when {
        wrong password
      }
      then {
        error shown
      }
    }
  }
}
`
	f, err := dsl.Parse(src, nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	ab := f.Artifact.(*dsl.ArtifactBlock)
	var infos []ScenarioInfo
	collectScenarios(ab.Items, &infos)

	if len(infos) != 2 {
		t.Fatalf("expected 2 scenarios, got %d", len(infos))
	}
	if len(infos[0].Labels) != 2 || infos[0].Labels[0] != "happy_path" || infos[0].Labels[1] != "smoke" {
		t.Errorf("expected labels [happy_path, smoke], got %v", infos[0].Labels)
	}
	if len(infos[1].Labels) != 0 {
		t.Errorf("expected no labels on second scenario, got %v", infos[1].Labels)
	}
}

func TestActorFieldOnScenario(t *testing.T) {
	src := `contract "CON-ACT" {
  title = "Actor test"
  status = "draft"

  personas {
    contributor = "regular user"
    admin = "elevated user"
  }

  feature "Auth" {
    scenario "contributor drafts" {
      actor = "contributor"
      labels = ["happy_path"]
      given {
        a contract exists
      }
      when {
        contributor creates a draft
      }
      then {
        draft is created
      }
    }
    scenario "admin completes" {
      actor = "admin"
      labels = ["happy_path"]
      given {
        a contract in active
      }
      when {
        admin transitions to complete
      }
      then {
        contract is complete
      }
    }
    scenario "no actor scenario" {
      given {
        any precondition
      }
      when {
        something happens
      }
      then {
        result
      }
    }
  }
}
`
	f, err := dsl.Parse(src, nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	ab := f.Artifact.(*dsl.ArtifactBlock)
	var infos []ScenarioInfo
	collectScenarios(ab.Items, &infos)

	if len(infos) != 3 {
		t.Fatalf("expected 3 scenarios, got %d", len(infos))
	}
	if infos[0].Actor != "contributor" {
		t.Errorf("expected actor 'contributor', got %q", infos[0].Actor)
	}
	if infos[1].Actor != "admin" {
		t.Errorf("expected actor 'admin', got %q", infos[1].Actor)
	}
	if infos[2].Actor != "" {
		t.Errorf("expected empty actor on third scenario, got %q", infos[2].Actor)
	}
}
