package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/dsl"
)

func TestCON041_TransitionGateBlockedByCriteriaCoverage(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["need"]

	_, err := GenericCreate(root, td, "NEED-GATED-001", map[string]string{
		"title": "Gated need", "sensation": "Pain", "status": "validated",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	needPath, _ := FindGenericPath(root, td, "NEED-GATED-001")
	data, _ := os.ReadFile(needPath)
	f, _ := dsl.Parse(string(data), nil)
	ab := f.Artifact.(*dsl.ArtifactBlock)
	ab.Items = append(ab.Items, &dsl.Block{
		Name: "acceptance",
		Items: []dsl.Node{
			&dsl.Block{Name: "criterion", Title: "fast", Items: []dsl.Node{
				&dsl.Field{Key: "description", Value: &dsl.StringVal{Text: "Must be fast"}},
				&dsl.Field{Key: "verified_by", Value: &dsl.StringVal{Text: "harness"}},
			}},
		},
	})
	writeArtifact(needPath, f)

	err = GenericUpdateStatus(root, td, "NEED-GATED-001", "addressed")
	if err == nil {
		t.Fatal("expected transition to be blocked by criteria_coverage gate")
	}
	if !strings.Contains(err.Error(), "criteria not covered") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCON041_TransitionGateAllowedWhenCovered(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	needTD := reg.Types["need"]
	specTD := reg.Types["specification"]

	_, err := GenericCreate(root, needTD, "NEED-GATED-002", map[string]string{
		"title": "Covered need", "sensation": "Pain", "status": "validated",
	})
	if err != nil {
		t.Fatalf("create need: %v", err)
	}

	needPath, _ := FindGenericPath(root, needTD, "NEED-GATED-002")
	data, _ := os.ReadFile(needPath)
	f, _ := dsl.Parse(string(data), nil)
	ab := f.Artifact.(*dsl.ArtifactBlock)
	ab.Items = append(ab.Items, &dsl.Block{
		Name: "acceptance",
		Items: []dsl.Node{
			&dsl.Block{Name: "criterion", Title: "fast", Items: []dsl.Node{
				&dsl.Field{Key: "description", Value: &dsl.StringVal{Text: "Must be fast"}},
				&dsl.Field{Key: "verified_by", Value: &dsl.StringVal{Text: "harness"}},
			}},
		},
	})
	writeArtifact(needPath, f)

	_, err = GenericCreate(root, specTD, "SPEC-COV-001", map[string]string{
		"title": "Cover the fast criterion", "enforcement": "warn",
		"satisfies": "NEED-GATED-002", "addresses": "fast",
	})
	if err != nil {
		t.Fatalf("create spec: %v", err)
	}

	err = GenericUpdateStatus(root, needTD, "NEED-GATED-002", "addressed")
	if err != nil {
		t.Fatalf("expected transition to succeed: %v", err)
	}
}

func TestCON041_TransitionGateSkippedForNonGatedTransition(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["need"]

	_, err := GenericCreate(root, td, "NEED-NONGATED", map[string]string{
		"title": "Non-gated need", "sensation": "Pain", "status": "identified",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	err = GenericUpdateStatus(root, td, "NEED-NONGATED", "validated")
	if err != nil {
		t.Errorf("identified -> validated should not be gated: %v", err)
	}
}

func TestCON041_NeedWithoutCriteriaPassesGate(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["need"]

	_, err := GenericCreate(root, td, "NEED-NOCRIT", map[string]string{
		"title": "No criteria need", "sensation": "Pain", "status": "validated",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	err = GenericUpdateStatus(root, td, "NEED-NOCRIT", "addressed")
	if err != nil {
		t.Errorf("need without criteria should pass gate: %v", err)
	}
}

func TestCON041_TransitionGateParsedFromRegistry(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["need"]

	if len(td.Lifecycle.Gates) == 0 {
		t.Fatal("need lifecycle should have at least one gate")
	}
	gate := td.Lifecycle.Gates[0]
	if gate.From != "validated" || gate.To != "addressed" || gate.Gate != "criteria_coverage" {
		t.Errorf("unexpected gate: %+v", gate)
	}
}

// --- CON-2026-042: Urgency Propagation ---

func writeTestConfig(t *testing.T, mosDir string) {
	t.Helper()
	os.MkdirAll(mosDir, 0o755)
	os.WriteFile(filepath.Join(mosDir, "config.mos"), []byte(`config {
  mos { version = 1 }
  backend { type = "git" }
  artifact_type "specification" {
    directory = "specifications"
    prefix = "SPEC"
  }
}
`), 0o644)
}

func TestTestMatrixGateBlocks(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeTestConfig(t, mos)

	specDir := filepath.Join(mos, "specifications", "active", "SPEC-GAT")
	os.MkdirAll(specDir, 0o755)
	os.WriteFile(filepath.Join(specDir, "specification.mos"), []byte(`
specification "SPEC-GAT" {
  title = "Gate test spec"
  status = "draft"
}
`), 0o644)

	conDir := filepath.Join(mos, "contracts", "active", "CON-GAT")
	os.MkdirAll(conDir, 0o755)
	conSrc := `contract "CON-GAT" {
  title = "Gate test contract"
  status = "active"
  justifies = "SPEC-GAT"

  coverage {
    unit { applies = true }
    integration { applies = true }
  }
}
`
	os.WriteFile(filepath.Join(conDir, "contract.mos"), []byte(conSrc), 0o644)

	f, _ := dsl.Parse(conSrc, nil)
	ab := f.Artifact.(*dsl.ArtifactBlock)
	err := evaluateTestMatrixGate(root, "CON-GAT", ab)
	if err == nil {
		t.Fatal("expected gate to block transition when spec has no test_matrix")
	}
	if !strings.Contains(err.Error(), "unit") || !strings.Contains(err.Error(), "integration") {
		t.Errorf("expected error to mention missing layers, got: %s", err)
	}
}

func TestTestMatrixGatePasses(t *testing.T) {
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")
	writeTestConfig(t, mos)

	specDir := filepath.Join(mos, "specifications", "active", "SPEC-GP")
	os.MkdirAll(specDir, 0o755)
	os.WriteFile(filepath.Join(specDir, "specification.mos"), []byte(`
specification "SPEC-GP" {
  title = "Gate pass spec"
  status = "draft"

  test_matrix {
    unit {
      symbol = "TestUnit"
    }
    integration {
      symbol = "TestIntegration"
    }
  }
}
`), 0o644)

	conDir := filepath.Join(mos, "contracts", "active", "CON-GP")
	os.MkdirAll(conDir, 0o755)
	conSrc := `contract "CON-GP" {
  title = "Gate pass contract"
  status = "active"
  justifies = "SPEC-GP"

  coverage {
    unit { applies = true }
    integration { applies = true }
  }
}
`
	os.WriteFile(filepath.Join(conDir, "contract.mos"), []byte(conSrc), 0o644)

	f, _ := dsl.Parse(conSrc, nil)
	ab := f.Artifact.(*dsl.ArtifactBlock)
	err := evaluateTestMatrixGate(root, "CON-GP", ab)
	if err != nil {
		t.Fatalf("expected gate to pass, got: %v", err)
	}
}

func TestTestMatrixGateNoJustifies(t *testing.T) {
	conSrc := `contract "CON-NJ" {
  title = "No justifies"
  status = "active"

  coverage {
    unit { applies = true }
  }
}
`
	f, _ := dsl.Parse(conSrc, nil)
	ab := f.Artifact.(*dsl.ArtifactBlock)
	err := evaluateTestMatrixGate(t.TempDir(), "CON-NJ", ab)
	if err != nil {
		t.Fatalf("expected gate to pass when no justifies, got: %v", err)
	}
}
