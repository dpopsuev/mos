package tracecmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/model"
)

func makeProject(modPath string, packages []string) *model.Project {
	p := model.NewProject(modPath)
	for _, pkg := range packages {
		p.AddNamespace(model.NewNamespace(pkg, modPath+"/"+pkg))
	}
	return p
}

func TestInferDefaultGroupsDepth1(t *testing.T) {
	proj := makeProject("example.com/test", []string{
		"cmd/server",
		"cmd/client",
		"moslib/survey",
		"moslib/model",
		"moslib/arch",
	})

	groups := inferDefaultGroups(proj, "example.com/test", 1)

	groupMap := make(map[string]artifact.ComponentGroup)
	for _, g := range groups {
		groupMap[g.Name] = g
	}

	cmd, ok := groupMap["cmd"]
	if !ok {
		t.Fatal("missing group 'cmd'")
	}
	if len(cmd.Packages) != 2 {
		t.Errorf("cmd group has %d packages, want 2", len(cmd.Packages))
	}

	moslib, ok := groupMap["moslib"]
	if !ok {
		t.Fatal("missing group 'moslib'")
	}
	if len(moslib.Packages) != 3 {
		t.Errorf("moslib group has %d packages, want 3", len(moslib.Packages))
	}
}

func TestInferDefaultGroupsDepth2(t *testing.T) {
	proj := makeProject("example.com/test", []string{
		"moslib/vcs/staging",
		"moslib/vcs/history",
		"moslib/vcs/merge",
		"moslib/survey",
		"moslib/model",
		"cmd/server",
	})

	groups := inferDefaultGroups(proj, "example.com/test", 2)

	groupMap := make(map[string]artifact.ComponentGroup)
	for _, g := range groups {
		groupMap[g.Name] = g
	}

	vcs, ok := groupMap["moslib/vcs"]
	if !ok {
		t.Fatal("missing group 'moslib/vcs'")
	}
	if len(vcs.Packages) != 3 {
		t.Errorf("moslib/vcs group has %d packages, want 3", len(vcs.Packages))
	}

	if _, ok := groupMap["moslib/survey"]; ok {
		t.Error("moslib/survey should not be a group (only 1 package)")
	}
}

func TestInferDefaultGroupsDepth3(t *testing.T) {
	proj := makeProject("example.com/test", []string{
		"a/b/c/d",
		"a/b/c/e",
		"a/b/f",
		"x/y",
	})

	groups := inferDefaultGroups(proj, "example.com/test", 3)

	groupMap := make(map[string]artifact.ComponentGroup)
	for _, g := range groups {
		groupMap[g.Name] = g
	}

	abc, ok := groupMap["a/b/c"]
	if !ok {
		t.Fatal("missing group 'a/b/c'")
	}
	if len(abc.Packages) != 2 {
		t.Errorf("a/b/c group has %d packages, want 2", len(abc.Packages))
	}

	if _, ok := groupMap["a/b/f"]; ok {
		t.Error("a/b/f should not be a group (only 1 package)")
	}
}

func TestDetectProjectPathGo(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/mymod\n\ngo 1.21\n")

	path := detectProjectPath(dir)
	if path != "example.com/mymod" {
		t.Errorf("detectProjectPath = %q, want example.com/mymod", path)
	}
}

func TestDetectProjectPathCargo(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Cargo.toml", `[package]
name = "my-rust-app"
version = "0.1.0"
`)

	path := detectProjectPath(dir)
	if path != "my-rust-app" {
		t.Errorf("detectProjectPath = %q, want my-rust-app", path)
	}
}

func TestDetectProjectPathPackageJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"name": "@scope/my-lib", "version": "1.0.0"}`)

	path := detectProjectPath(dir)
	if path != "@scope/my-lib" {
		t.Errorf("detectProjectPath = %q, want @scope/my-lib", path)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	p := dir + "/" + name
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
