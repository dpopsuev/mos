package rule

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

func runRuleCmd(t *testing.T, root string, args []string) (stdout, stderr string, err error) {
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
	cmd.SetArgs(append([]string{"rule"}, args...))

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

func TestRuleCreateRequiresNameAndType(t *testing.T) {
	root := setupWorkspace(t)
	_, stderr, err := runRuleCmd(t, root, []string{"create", "myrule"})
	if err == nil {
		t.Fatal("expected error when running rule create without --name and --type")
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "required") && !strings.Contains(stderr, "required") &&
		!strings.Contains(errStr, "name") && !strings.Contains(errStr, "type") {
		t.Errorf("expected error about required name/type; got err=%v stderr=%q", err, stderr)
	}
}

func TestRuleListEmpty(t *testing.T) {
	root := setupWorkspace(t)
	stdout, _, err := runRuleCmd(t, root, []string{"list"})
	if err != nil {
		t.Fatalf("rule list: %v", err)
	}
	if !strings.Contains(stdout, "(no rules found)") {
		t.Errorf("expected '(no rules found)' in output; got %q", stdout)
	}
}

func TestRuleShowRequiresID(t *testing.T) {
	root := setupWorkspace(t)
	_, _, err := runRuleCmd(t, root, []string{"show"})
	if err == nil {
		t.Fatal("expected error when running rule show without args")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg") && !strings.Contains(err.Error(), "show") {
		t.Errorf("expected args error; got %v", err)
	}
}

func TestRuleCreateWithRequiredFlags(t *testing.T) {
	root := setupWorkspace(t)
	stdout, _, err := runRuleCmd(t, root, []string{
		"create", "test-rule",
		"--name", "Test Rule",
		"--type", "mechanical",
		"--scope", "project",
		"--enforcement", "error",
	})
	if err != nil {
		t.Fatalf("rule create: %v", err)
	}
	if !strings.Contains(stdout, "Created rule:") {
		t.Errorf("expected 'Created rule:' in output; got %q", stdout)
	}

	rulePath := filepath.Join(root, ".mos", "rules", "mechanical", "test-rule.mos")
	if _, statErr := os.Stat(rulePath); statErr != nil {
		t.Errorf("expected rule file at %s: %v", rulePath, statErr)
	}
}
