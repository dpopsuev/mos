package binder

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/artifact"
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

func runBinderCmd(t *testing.T, root string, args []string) (stdout, stderr string, err error) {
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
	cmd.SetArgs(append([]string{"binder"}, args...))

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

func TestBinderCreateWithArgs(t *testing.T) {
	root := setupWorkspace(t)
	stdout, _, err := runBinderCmd(t, root, []string{"create", "BND-001", "--title", "Test Binder"})
	if err != nil {
		t.Fatalf("binder create: %v", err)
	}
	if !strings.Contains(stdout, "Created binder:") {
		t.Errorf("expected 'Created binder:' in output; got %q", stdout)
	}

	bindersDir := filepath.Join(root, ".mos", "binders", "active")
	entries, err := os.ReadDir(bindersDir)
	if err != nil {
		t.Fatalf("ReadDir binders: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one binder directory under .mos/binders/active/")
	}
	found := false
	for _, e := range entries {
		if e.IsDir() && e.Name() == "BND-001" {
			binderPath := filepath.Join(bindersDir, e.Name(), "binder.mos")
			if _, statErr := os.Stat(binderPath); statErr == nil {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("expected binder.mos under .mos/binders/active/BND-001/; entries: %v", entries)
	}
}

func TestBinderListEmpty(t *testing.T) {
	root := setupWorkspace(t)
	stdout, _, err := runBinderCmd(t, root, []string{"list"})
	if err != nil {
		t.Fatalf("binder list: %v", err)
	}
	if !strings.Contains(stdout, "(no binders found)") {
		t.Errorf("expected '(no binders found)' in output; got %q", stdout)
	}
}

func TestBinderShowWithoutID(t *testing.T) {
	root := setupWorkspace(t)
	_, _, err := runBinderCmd(t, root, []string{"show"})
	if err == nil {
		t.Fatal("expected error when running binder show without id")
	}
	if !strings.Contains(err.Error(), "usage") && !strings.Contains(err.Error(), "show") {
		t.Errorf("expected usage error; got %v", err)
	}
}
