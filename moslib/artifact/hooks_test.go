package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createHookTestContract(t *testing.T, root, id, status string, scenarioStatuses map[string]string) {
	t.Helper()
	var scenarioBlocks string
	for name, st := range scenarioStatuses {
		statusLine := ""
		if st != "" {
			statusLine = fmt.Sprintf("      status = %q\n", st)
		}
		scenarioBlocks += fmt.Sprintf(`    scenario %q {
%s      given {
        precondition
      }
      when {
        action
      }
      then {
        result
      }
    }
`, name, statusLine)
	}
	content := fmt.Sprintf(`contract %q {
  title = "Hook Test"
  status = %q
  kind = "feature"

  feature "Hook Feature" {
%s  }
}
`, id, status, scenarioBlocks)
	if _, err := ApplyArtifact(root, []byte(content)); err != nil {
		t.Fatalf("creating hook test contract: %v", err)
	}
}

func TestCON029_ParseHooksFromLifecycle(t *testing.T) {
	root := setupScaffold(t)
	configPath := filepath.Join(root, ".mos", "config.mos")
	data, _ := os.ReadFile(configPath)
	extra := `
  artifact_type "ticket" {
    directory = "tickets"
    fields {}
    lifecycle {
      active_states = ["open"]
      archive_states = ["closed"]
      hooks {
        on_all {
          watch_field = "status"
          threshold = "resolved"
          set_field = "status"
          set_value = "closed"
        }
      }
    }
  }
`
	patched := strings.Replace(string(data), "\n}\n", extra+"\n}\n", 1)
	os.WriteFile(configPath, []byte(patched), 0644)

	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	td := reg.Types["ticket"]
	if len(td.Lifecycle.Hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(td.Lifecycle.Hooks))
	}
	h := td.Lifecycle.Hooks[0]
	if h.Trigger != "on_all" || h.WatchField != "status" || h.Threshold != "resolved" || h.SetField != "status" || h.SetValue != "closed" {
		t.Errorf("unexpected hook: %+v", h)
	}
}

func TestCON029_OnAnyFiresOnFirstMatch(t *testing.T) {
	root := setupScaffold(t)
	createHookTestContract(t, root, "CON-HOOK-001", "draft", map[string]string{
		"Scenario A": "pending",
		"Scenario B": "pending",
		"Scenario C": "pending",
	})

	SetScenarioStatus(root, "CON-HOOK-001", "Scenario A", "implemented")

	status, err := GetContractStatus(root, "CON-HOOK-001")
	if err != nil {
		t.Fatalf("GetContractStatus: %v", err)
	}
	if status != "active" {
		t.Errorf("expected contract status 'active' after on_any hook, got %q", status)
	}
}

func TestCON029_OnAllFiresWhenAllMatch(t *testing.T) {
	root := setupScaffold(t)
	createHookTestContract(t, root, "CON-HOOK-002", "active", map[string]string{
		"Scenario A": "verified",
		"Scenario B": "verified",
	})

	SetScenarioStatus(root, "CON-HOOK-002", "Scenario B", "verified")

	status, err := GetContractStatus(root, "CON-HOOK-002")
	if err != nil {
		t.Fatalf("GetContractStatus: %v", err)
	}
	if status != "complete" {
		t.Errorf("expected contract status 'complete' after on_all hook, got %q", status)
	}
}

func TestCON029_HookDoesNotFireBelowThreshold(t *testing.T) {
	root := setupScaffold(t)
	createHookTestContract(t, root, "CON-HOOK-003", "active", map[string]string{
		"Scenario A": "verified",
		"Scenario B": "pending",
		"Scenario C": "pending",
	})

	reg, _ := LoadRegistry(root)
	EvaluateHooks(root, "CON-HOOK-003", reg)

	status, err := GetContractStatus(root, "CON-HOOK-003")
	if err != nil {
		t.Fatalf("GetContractStatus: %v", err)
	}
	if status != "active" {
		t.Errorf("expected contract status to remain 'active', got %q", status)
	}
}

func TestCON029_HooksAreIdempotent(t *testing.T) {
	root := setupScaffold(t)
	createHookTestContract(t, root, "CON-HOOK-004", "active", map[string]string{
		"Scenario A": "implemented",
		"Scenario B": "pending",
	})

	reg, _ := LoadRegistry(root)
	EvaluateHooks(root, "CON-HOOK-004", reg)
	EvaluateHooks(root, "CON-HOOK-004", reg)

	status, _ := GetContractStatus(root, "CON-HOOK-004")
	if status != "active" {
		t.Errorf("expected status to remain 'active' after idempotent hook eval, got %q", status)
	}
}

func TestCON029_DefaultContractTypeHasSaneHooks(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["contract"]
	if len(td.Lifecycle.Hooks) != 2 {
		t.Fatalf("expected 2 default hooks on contract type, got %d", len(td.Lifecycle.Hooks))
	}
	onAny := td.Lifecycle.Hooks[0]
	if onAny.Trigger != "on_any" || onAny.WatchField != "status" || onAny.Threshold != "implemented" || onAny.SetValue != "active" {
		t.Errorf("unexpected on_any hook: %+v", onAny)
	}
	onAll := td.Lifecycle.Hooks[1]
	if onAll.Trigger != "on_all" || onAll.WatchField != "status" || onAll.Threshold != "verified" || onAll.SetValue != "complete" {
		t.Errorf("unexpected on_all hook: %+v", onAll)
	}
}

// --- CON-2026-030: Ordered Enums and Harness-Gated Transitions ---
