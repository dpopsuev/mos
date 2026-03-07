package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/dsl"
)

func setupReclassifyTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mosDir := filepath.Join(dir, MosDir)

	configContent := `config "mos" {
  artifact_type "contract" {
    directory = "contracts"
    prefix = "CON"
    lifecycle {
      active_states = ["draft", "active"]
      archive_states = ["complete", "abandoned"]
    }
  }

  artifact_type "specification" {
    directory = "specifications"
    prefix = "SPEC"
    lifecycle {
      active_states = ["draft", "active"]
      archive_states = ["retired"]
    }
  }
}
`
	if err := os.MkdirAll(mosDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mosDir, "config.mos"), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, d := range []string{
		"contracts/active", "contracts/archive",
		"specifications/active", "specifications/archive",
	} {
		if err := os.MkdirAll(filepath.Join(mosDir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func createTestContract(t *testing.T, root, id string) {
	t.Helper()
	mosDir := filepath.Join(root, MosDir)
	contractDir := filepath.Join(mosDir, "contracts", "active", id)
	if err := os.MkdirAll(contractDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := `contract "` + id + `" {
  title = "Test Contract"
  status = "draft"
  goal = "A goal"

  feature "Feature One" {
  }

  section "Notes" {
    text = "Important notes"
  }
}
`
	if err := os.WriteFile(filepath.Join(contractDir, "contract.mos"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReclassify_ContractToSpec(t *testing.T) {
	root := setupReclassifyTestDir(t)
	createTestContract(t, root, "CON-2026-999")

	result, err := Reclassify(root, "CON-2026-999", "specification")
	if err != nil {
		t.Fatalf("Reclassify: %v", err)
	}

	if result.OldID != "CON-2026-999" {
		t.Errorf("OldID = %q, want CON-2026-999", result.OldID)
	}
	if result.OldKind != "contract" {
		t.Errorf("OldKind = %q, want contract", result.OldKind)
	}
	if result.NewKind != "specification" {
		t.Errorf("NewKind = %q, want specification", result.NewKind)
	}
	if !strings.HasPrefix(result.NewID, "SPEC-") {
		t.Errorf("NewID = %q, expected SPEC- prefix", result.NewID)
	}

	// New artifact exists and has preserved content
	newAB, err := dsl.ReadArtifact(result.NewPath)
	if err != nil {
		t.Fatalf("reading new artifact: %v", err)
	}
	if newAB.Kind != "specification" {
		t.Errorf("new artifact kind = %q, want specification", newAB.Kind)
	}
	if newAB.Name != result.NewID {
		t.Errorf("new artifact name = %q, want %q", newAB.Name, result.NewID)
	}

	title, _ := dsl.FieldString(newAB.Items, "title")
	if title != "Test Contract" {
		t.Errorf("title not preserved: got %q", title)
	}
	goal, _ := dsl.FieldString(newAB.Items, "goal")
	if goal != "A goal" {
		t.Errorf("goal not preserved: got %q", goal)
	}

	var hasFeature, hasSection bool
	for _, item := range newAB.Items {
		if fb, ok := item.(*dsl.FeatureBlock); ok && fb.Name == "Feature One" {
			hasFeature = true
		}
		if b, ok := item.(*dsl.Block); ok && b.Name == "section" && b.Title == "Notes" {
			hasSection = true
		}
	}
	if !hasFeature {
		t.Error("feature block not preserved")
	}
	if !hasSection {
		t.Error("section block not preserved")
	}
}

func TestReclassify_TombstoneCreated(t *testing.T) {
	root := setupReclassifyTestDir(t)
	createTestContract(t, root, "CON-2026-998")

	result, err := Reclassify(root, "CON-2026-998", "specification")
	if err != nil {
		t.Fatalf("Reclassify: %v", err)
	}

	tombPath := filepath.Join(root, MosDir, "contracts", "active", "CON-2026-998", "contract.mos")
	tombAB, err := dsl.ReadArtifact(tombPath)
	if err != nil {
		t.Fatalf("reading tombstone: %v", err)
	}

	status, _ := dsl.FieldString(tombAB.Items, "status")
	if status != "reclassified" {
		t.Errorf("tombstone status = %q, want reclassified", status)
	}

	reclTo, _ := dsl.FieldString(tombAB.Items, "reclassified_to")
	if reclTo != result.NewID {
		t.Errorf("tombstone reclassified_to = %q, want %q", reclTo, result.NewID)
	}

	reclKind, _ := dsl.FieldString(tombAB.Items, "reclassified_kind")
	if reclKind != "specification" {
		t.Errorf("tombstone reclassified_kind = %q, want specification", reclKind)
	}
}

func TestReclassify_SameKind(t *testing.T) {
	root := setupReclassifyTestDir(t)
	createTestContract(t, root, "CON-2026-997")

	_, err := Reclassify(root, "CON-2026-997", "contract")
	if err == nil {
		t.Fatal("expected error when reclassifying to same kind")
	}
	if !strings.Contains(err.Error(), "already a contract") {
		t.Errorf("error = %v, expected 'already a contract'", err)
	}
}

func TestReclassify_UnknownTarget(t *testing.T) {
	root := setupReclassifyTestDir(t)
	createTestContract(t, root, "CON-2026-996")

	_, err := Reclassify(root, "CON-2026-996", "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown target kind")
	}
}

func TestReclassify_NotFound(t *testing.T) {
	root := setupReclassifyTestDir(t)

	_, err := Reclassify(root, "CON-2026-000", "specification")
	if err == nil {
		t.Fatal("expected error for missing artifact")
	}
}
