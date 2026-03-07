package spec

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

func runSpecCmd(t *testing.T, root string, args []string) (stdout, stderr string, err error) {
	t.Helper()
	origDir, dirErr := os.Getwd()
	if dirErr != nil {
		t.Fatalf("os.Getwd: %v", dirErr)
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
	cmd.SetArgs(append([]string{"spec"}, args...))

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

func TestSpecCreateRequiresNothing(t *testing.T) {
	root := setupWorkspace(t)
	stdout, _, err := runSpecCmd(t, root, []string{"create", "--title", "Test"})
	if err != nil {
		t.Fatalf("spec create --title Test: %v", err)
	}
	if !strings.Contains(stdout, "Created specification:") {
		t.Errorf("expected 'Created specification:' in output; got %q", stdout)
	}
}

func TestSpecListEmpty(t *testing.T) {
	root := setupWorkspace(t)
	stdout, _, err := runSpecCmd(t, root, []string{"list"})
	if err != nil {
		t.Fatalf("spec list: %v", err)
	}
	if !strings.Contains(stdout, "(no specifications found)") {
		t.Errorf("expected '(no specifications found)' in output; got %q", stdout)
	}
}

func TestSpecShowRequiresID(t *testing.T) {
	root := setupWorkspace(t)
	_, _, err := runSpecCmd(t, root, []string{"show"})
	if err == nil {
		t.Fatal("expected error when running spec show without args")
	}
	if !strings.Contains(err.Error(), "usage") && !strings.Contains(err.Error(), "show") {
		t.Errorf("expected usage error; got %v", err)
	}
}
