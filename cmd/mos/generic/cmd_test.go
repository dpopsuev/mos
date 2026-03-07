package generic

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

func runNeedCmd(t *testing.T, root string, cmd *cobra.Command, args []string) (stdout, stderr string, err error) {
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

	rootCmd := &cobra.Command{Use: "mos"}
	rootCmd.AddCommand(cmd)
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.SetArgs(args)

	done := make(chan struct{})
	var outBuf, errBuf bytes.Buffer
	go func() {
		_, _ = outBuf.ReadFrom(rOut)
		close(done)
	}()
	go func() { _, _ = errBuf.ReadFrom(rErr) }()

	runErr := rootCmd.Execute()
	wOut.Close()
	wErr.Close()
	<-done

	return outBuf.String(), errBuf.String(), runErr
}

func TestNeedCreateWithArgs(t *testing.T) {
	td := registry.ArtifactTypeDef{
		Kind:      "need",
		Directory: "needs",
		Prefix:    "NEED",
		Lifecycle: registry.LifecycleDef{
			ActiveStates:  []string{"identified", "validated", "addressed"},
			ArchiveStates: []string{"retired"},
		},
	}
	cmd := NewCmd(td)
	root := setupWorkspace(t)

	stdout, _, err := runNeedCmd(t, root, cmd, []string{"need", "create", "NEED-001", "--title", "Test Need", "--status", "identified"})
	if err != nil {
		t.Fatalf("need create: %v", err)
	}
	if !strings.Contains(stdout, "Created need:") {
		t.Errorf("expected 'Created need:' in output; got %q", stdout)
	}

	needsDir := filepath.Join(root, ".mos", "needs", "active")
	needPath := filepath.Join(needsDir, "NEED-001", "need.mos")
	if _, statErr := os.Stat(needPath); statErr != nil {
		t.Fatalf("expected need.mos at %s: %v", needPath, statErr)
	}
}

func TestNeedListEmpty(t *testing.T) {
	td := registry.ArtifactTypeDef{Kind: "need", Directory: "needs", Prefix: "NEED"}
	cmd := NewCmd(td)
	root := setupWorkspace(t)

	stdout, _, err := runNeedCmd(t, root, cmd, []string{"need", "list"})
	if err != nil {
		t.Fatalf("need list: %v", err)
	}
	if !strings.Contains(stdout, "(no needs found)") {
		t.Errorf("expected '(no needs found)' in output; got %q", stdout)
	}
}

func TestNeedShowWithoutID(t *testing.T) {
	td := registry.ArtifactTypeDef{Kind: "need", Directory: "needs", Prefix: "NEED"}
	cmd := NewCmd(td)
	root := setupWorkspace(t)

	_, _, err := runNeedCmd(t, root, cmd, []string{"need", "show"})
	if err == nil {
		t.Fatal("expected error when running need show without id")
	}
	if !strings.Contains(err.Error(), "usage") && !strings.Contains(err.Error(), "show") {
		t.Errorf("expected usage error; got %v", err)
	}
}
