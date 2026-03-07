package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/dsl"
)

func TestInit(t *testing.T) {
	root := t.TempDir()
	writeGoMod(t, root, "github.com/test/myproject")

	err := Init(root, InitOpts{
		Model:   "bdfl",
		Scope:   "cabinet",
		Name:    "myproject",
		Purpose: "A test project",
	})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	mosDir := filepath.Join(root, ".mos")
	for _, path := range []string{
		filepath.Join(mosDir, "config.mos"),
		filepath.Join(mosDir, "lexicon", "default.mos"),
		filepath.Join(mosDir, "resolution", "layers.mos"),
		filepath.Join(mosDir, "declaration.mos"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist", path)
		}
		assertParses(t, path)
	}

	for _, dir := range []string{
		filepath.Join(mosDir, "rules", "mechanical"),
		filepath.Join(mosDir, "rules", "interpretive"),
		filepath.Join(mosDir, "contracts", "active"),
		filepath.Join(mosDir, "contracts", "archive"),
	} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("expected directory %s to exist", dir)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", dir)
		}
	}

	assertConfigField(t, filepath.Join(mosDir, "config.mos"), "governance", "model", "bdfl")
	assertConfigField(t, filepath.Join(mosDir, "config.mos"), "governance", "scope", "cabinet")
	assertDeclField(t, filepath.Join(mosDir, "declaration.mos"), "name", "myproject")

	vocabData, err := os.ReadFile(filepath.Join(mosDir, "lexicon", "default.mos"))
	if err != nil {
		t.Fatalf("reading lexicon: %v", err)
	}
	vocabContent := string(vocabData)
	for _, term := range []string{"pillar", "component", "priority", "automation_status"} {
		if !strings.Contains(vocabContent, term) {
			t.Errorf("expected commented ALM term %q in lexicon template", term)
		}
	}
}

func TestInitAlreadyExists(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".mos"), 0755)

	err := Init(root, InitOpts{Name: "test"})
	if err == nil {
		t.Fatal("expected error for existing .mos/, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestInitAutoName(t *testing.T) {
	root := t.TempDir()
	writeGoMod(t, root, "github.com/someone/awesome-project")

	err := Init(root, InitOpts{})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	assertDeclField(t, filepath.Join(root, ".mos", "declaration.mos"), "name", "awesome-project")
}

func assertConfigField(t *testing.T, path, block, key, want string) {
	t.Helper()
	data, _ := os.ReadFile(path)
	f, _ := dsl.Parse(string(data), nil)
	ab := f.Artifact.(*dsl.ArtifactBlock)
	for _, item := range ab.Items {
		b, ok := item.(*dsl.Block)
		if !ok || b.Name != block {
			continue
		}
		for _, bi := range b.Items {
			fld, ok := bi.(*dsl.Field)
			if !ok || fld.Key != key {
				continue
			}
			sv, ok := fld.Value.(*dsl.StringVal)
			if !ok {
				t.Errorf("%s.%s is not a string", block, key)
				return
			}
			if sv.Text != want {
				t.Errorf("%s.%s = %q, want %q", block, key, sv.Text, want)
			}
			return
		}
	}
	t.Errorf("%s.%s not found", block, key)
}

func assertDeclField(t *testing.T, path, key, want string) {
	t.Helper()
	data, _ := os.ReadFile(path)
	f, _ := dsl.Parse(string(data), nil)
	ab := f.Artifact.(*dsl.ArtifactBlock)
	for _, item := range ab.Items {
		fld, ok := item.(*dsl.Field)
		if !ok || fld.Key != key {
			continue
		}
		sv, ok := fld.Value.(*dsl.StringVal)
		if !ok {
			t.Errorf("declaration.%s is not a string", key)
			return
		}
		if sv.Text != want {
			t.Errorf("declaration.%s = %q, want %q", key, sv.Text, want)
		}
		return
	}
	t.Errorf("declaration.%s not found", key)
}

// --- CON-2026-021: Project Primitives ---
