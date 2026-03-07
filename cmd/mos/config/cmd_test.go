package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/registry"
	"github.com/spf13/cobra"
)

func setupWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	content := "module github.com/test/scaffold\n\ngo 1.25.7\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(content), 0644); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}
	if err := artifact.Init(root, artifact.InitOpts{Name: "test", Model: "bdfl", Scope: "cabinet"}); err != nil {
		t.Fatalf("artifact.Init: %v", err)
	}
	return root
}

func runConfigCmd(t *testing.T, root string, args []string) (stdout, stderr string, err error) {
	t.Helper()
	origDir, origErr := os.Getwd()
	if origErr != nil {
		t.Fatalf("os.Getwd: %v", origErr)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("os.Chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	origOut, origStderr := os.Stdout, os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr
	defer func() { os.Stdout = origOut; os.Stderr = origStderr }()

	cmd := &cobra.Command{Use: "mos"}
	cmd.AddCommand(Cmd)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(append([]string{"config"}, args...))

	done := make(chan struct{})
	var outBuf, errBuf bytes.Buffer
	go func() {
		_, _ = outBuf.ReadFrom(rOut)
		close(done)
	}()
	go func() { _, _ = errBuf.ReadFrom(rErr) }()

	runErr := cmd.Execute()
	wOut.Close()
	wErr.Close()
	<-done

	return outBuf.String(), errBuf.String(), runErr
}

func TestConfigAddProject(t *testing.T) {
	root := setupWorkspace(t)
	stdout, _, err := runConfigCmd(t, root, []string{"add-project", "myapp", "--prefix", "APP"})
	if err != nil {
		t.Fatalf("config add-project: %v", err)
	}
	if !strings.Contains(stdout, "Added project") {
		t.Errorf("expected 'Added project' in output; got %q", stdout)
	}
	if !strings.Contains(stdout, "myapp") || !strings.Contains(stdout, "APP") {
		t.Errorf("expected project name and prefix in output; got %q", stdout)
	}

	projects, err := registry.LoadProjects(root)
	if err != nil {
		t.Fatalf("LoadProjects: %v", err)
	}
	var found bool
	for _, p := range projects {
		if p.Name == "myapp" && p.Prefix == "APP" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected project myapp with prefix APP")
	}
}

func TestConfigAddType(t *testing.T) {
	root := setupWorkspace(t)
	stdout, _, err := runConfigCmd(t, root, []string{"add-type", "epic", "--directory", "epics"})
	if err != nil {
		t.Fatalf("config add-type: %v", err)
	}
	if !strings.Contains(stdout, "Added artifact_type") {
		t.Errorf("expected 'Added artifact_type' in output; got %q", stdout)
	}
	if !strings.Contains(stdout, "epic") {
		t.Errorf("expected type name in output; got %q", stdout)
	}

	reg, err := registry.LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	if td, ok := reg.Types["epic"]; !ok || td.Directory != "epics" {
		t.Errorf("expected epic type with directory epics; got %v", reg.Types["epic"])
	}
}

func TestConfigRemoveProject(t *testing.T) {
	root := setupWorkspace(t)
	// Add then remove
	_, _, err := runConfigCmd(t, root, []string{"add-project", "toremove", "--prefix", "TR"})
	if err != nil {
		t.Fatalf("config add-project: %v", err)
	}

	stdout, _, err := runConfigCmd(t, root, []string{"remove-project", "toremove"})
	if err != nil {
		t.Fatalf("config remove-project: %v", err)
	}
	if !strings.Contains(stdout, "Removed project") {
		t.Errorf("expected 'Removed project' in output; got %q", stdout)
	}

	projects, err := registry.LoadProjects(root)
	if err != nil {
		t.Fatalf("LoadProjects: %v", err)
	}
	for _, p := range projects {
		if p.Name == "toremove" {
			t.Error("expected project toremove to be removed")
			break
		}
	}
}
