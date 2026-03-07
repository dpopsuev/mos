package artifact

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dpopsuev/mos/moslib/dsl"
)

func setupScaffold(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeGoMod(t, root, "github.com/test/scaffold")
	if err := Init(root, InitOpts{Name: "scaffold", Model: "bdfl", Scope: "cabinet"}); err != nil {
		t.Fatalf("scaffold setup failed: %v", err)
	}
	return root
}

func loadTestRegistry(t *testing.T, root string) *Registry {
	t.Helper()
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	return reg
}

func writeGoMod(t *testing.T, root, module string) {
	t.Helper()
	content := "module " + module + "\n\ngo 1.25.7\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(content), 0644); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}
}

func assertParses(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if _, err := dsl.Parse(string(data), nil); err != nil {
		t.Fatalf("parse error in %s: %v", path, err)
	}
}

func assertLintClean(t *testing.T, root string) {
	t.Helper()
	if LintAll == nil {
		return
	}
	diags, err := LintAll(root)
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	for _, d := range diags {
		if d.Severity == "error" {
			t.Errorf("lint error: %s [%s] %s", d.File, d.Rule, d.Message)
		}
	}
}
