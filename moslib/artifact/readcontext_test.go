package artifact

import (
	"os"
	"path/filepath"
	"testing"
)

func setupReadContextRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")
	for _, dir := range []string{
		"lexicon",
		"resolution",
		"templates",
		"rules/mechanical",
		"rules/interpretive",
		"contracts/active/CON-001",
		"contracts/archive/CON-002",
		"architectures/active/ARCH-desired",
		"specifications/active/SPEC-001",
	} {
		os.MkdirAll(filepath.Join(mosDir, dir), 0755)
	}
	return root
}

func TestReadConfig(t *testing.T) {
	root := setupReadContextRoot(t)
	os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(`config {
  mos { version = 1 }
}`), 0644)

	f, err := ReadConfig(root)
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil File")
	}
}

func TestReadConfig_NotExist(t *testing.T) {
	root := t.TempDir()
	_, err := ReadConfig(root)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestReadLexiconFile(t *testing.T) {
	root := setupReadContextRoot(t)
	os.WriteFile(filepath.Join(root, ".mos", "lexicon", "default.mos"), []byte(`lexicon {
  terms {
    foo = "bar"
  }
}`), 0644)

	f, err := ReadLexiconFile(root, "default.mos")
	if err != nil {
		t.Fatalf("ReadLexiconFile: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil File")
	}
}

func TestReadLayers(t *testing.T) {
	root := setupReadContextRoot(t)
	os.WriteFile(filepath.Join(root, ".mos", "resolution", "layers.mos"), []byte(`layers {
  layer "L1" { level = 1 }
}`), 0644)

	f, err := ReadLayers(root)
	if err != nil {
		t.Fatalf("ReadLayers: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil File")
	}
}

func TestReadTemplate(t *testing.T) {
	root := setupReadContextRoot(t)
	os.WriteFile(filepath.Join(root, ".mos", "templates", "contract.mos"), []byte(`template {
  scope {}
}`), 0644)

	f, err := ReadTemplate(root)
	if err != nil {
		t.Fatalf("ReadTemplate: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil File")
	}
}

func TestReadRuleInventory(t *testing.T) {
	root := setupReadContextRoot(t)
	os.WriteFile(filepath.Join(root, ".mos", "rules", "mechanical", "r1.mos"), []byte(`rule "R-001" {
  status = "active"
}`), 0644)
	os.WriteFile(filepath.Join(root, ".mos", "rules", "interpretive", "r2.mos"), []byte(`rule "R-002" {
  status = "active"
}`), 0644)

	rules := ReadRuleInventory(root, nil)
	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}
	if _, ok := rules["R-001"]; !ok {
		t.Error("missing R-001")
	}
	if _, ok := rules["R-002"]; !ok {
		t.Error("missing R-002")
	}
}

func TestReadContractInventory(t *testing.T) {
	root := setupReadContextRoot(t)
	os.WriteFile(filepath.Join(root, ".mos", "contracts", "active", "CON-001", "contract.mos"),
		[]byte(`contract "CON-001" { status = "draft" }`), 0644)
	os.WriteFile(filepath.Join(root, ".mos", "contracts", "archive", "CON-002", "contract.mos"),
		[]byte(`contract "CON-002" { status = "complete" }`), 0644)

	contracts := ReadContractInventory(root, nil)
	if len(contracts) != 2 {
		t.Errorf("expected 2 contracts, got %d", len(contracts))
	}
}

func TestReadArchitecture(t *testing.T) {
	root := setupReadContextRoot(t)
	os.WriteFile(filepath.Join(root, ".mos", "architectures", "active", "ARCH-desired", "architecture.mos"),
		[]byte(`architecture "ARCH-desired" { description = "test" }`), 0644)

	ab, err := ReadArchitecture(root, "ARCH-desired")
	if err != nil {
		t.Fatalf("ReadArchitecture: %v", err)
	}
	if ab.Name != "ARCH-desired" {
		t.Errorf("expected name ARCH-desired, got %s", ab.Name)
	}
}

func TestReadArchitecture_NotExist(t *testing.T) {
	root := t.TempDir()
	_, err := ReadArchitecture(root, "ARCH-missing")
	if err == nil {
		t.Fatal("expected error for missing architecture")
	}
}

func TestReadArtifactInventory(t *testing.T) {
	root := setupReadContextRoot(t)
	dir := filepath.Join(root, ".mos", "sprints", "active", "SPR-001")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "sprint.mos"), []byte(`sprint "SPR-001" { status = "planned" }`), 0644)

	ids := ReadArtifactInventory(root, "sprints", "sprint")
	if len(ids) != 1 {
		t.Errorf("expected 1, got %d", len(ids))
	}
	if _, ok := ids["SPR-001"]; !ok {
		t.Error("missing SPR-001")
	}
}

func TestReadDSLFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.mos")
	os.WriteFile(path, []byte(`rule "R-001" { status = "active" }`), 0644)

	f, err := ReadDSLFile(path, nil)
	if err != nil {
		t.Fatalf("ReadDSLFile: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil File")
	}
}

func TestReadConfigBlock(t *testing.T) {
	root := setupReadContextRoot(t)
	os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(`config {
  mos { version = 1 }
}`), 0644)

	ab, err := ReadConfigBlock(root)
	if err != nil {
		t.Fatalf("ReadConfigBlock: %v", err)
	}
	if ab == nil {
		t.Fatal("expected non-nil ArtifactBlock")
	}
}

func TestReadMosFile(t *testing.T) {
	root := setupReadContextRoot(t)
	os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(`test`), 0644)

	data, err := ReadMosFile(root, "config.mos")
	if err != nil {
		t.Fatalf("ReadMosFile: %v", err)
	}
	if string(data) != "test" {
		t.Errorf("unexpected content: %s", data)
	}
}

func TestReadMosDirEntries(t *testing.T) {
	root := setupReadContextRoot(t)
	os.WriteFile(filepath.Join(root, ".mos", "lexicon", "default.mos"), []byte(`x`), 0644)

	entries, err := ReadMosDirEntries(root, "lexicon")
	if err != nil {
		t.Fatalf("ReadMosDirEntries: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected at least one entry")
	}
}
