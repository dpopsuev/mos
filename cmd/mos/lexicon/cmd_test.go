package lexicon

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/linter"
	"github.com/spf13/cobra"
)

func init() {
	// Wire LoadLexicon for tests (lexicon list uses it).
	artifact.LoadLexicon = func(mosDir string) (map[string]string, error) {
		ctx, err := linter.LoadContext(mosDir)
		if err != nil {
			return nil, err
		}
		if ctx.Lexicon == nil {
			return map[string]string{}, nil
		}
		return ctx.Lexicon.Terms, nil
	}
}

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

func runLexiconCmd(t *testing.T, root string, args []string) (stdout, stderr string, err error) {
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
	cmd.SetArgs(append([]string{"lexicon"}, args...))

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

func TestLexiconListEmpty(t *testing.T) {
	root := setupWorkspace(t)
	stdout, _, err := runLexiconCmd(t, root, []string{"list"})
	if err != nil {
		t.Fatalf("lexicon list: %v", err)
	}
	if !strings.Contains(stdout, "(no terms defined)") {
		t.Errorf("expected '(no terms defined)' in output; got %q", stdout)
	}
}

func TestLexiconAddWithArgs(t *testing.T) {
	root := setupWorkspace(t)
	stdout, _, err := runLexiconCmd(t, root, []string{"add", "pillar", "--description", "A test category"})
	if err != nil {
		t.Fatalf("lexicon add: %v", err)
	}
	if !strings.Contains(stdout, "Added term: pillar") {
		t.Errorf("expected 'Added term: pillar' in output; got %q", stdout)
	}

	lexiconPath := filepath.Join(root, ".mos", "lexicon", "default.mos")
	content, err := os.ReadFile(lexiconPath)
	if err != nil {
		t.Fatalf("ReadFile lexicon: %v", err)
	}
	if !strings.Contains(string(content), "pillar") {
		t.Errorf("expected 'pillar' in lexicon file; got %q", string(content))
	}
}

func TestLexiconRemoveWithoutKey(t *testing.T) {
	root := setupWorkspace(t)
	_, _, err := runLexiconCmd(t, root, []string{"remove"})
	if err == nil {
		t.Fatal("expected error when running lexicon remove without key")
	}
	if !strings.Contains(err.Error(), "usage") && !strings.Contains(err.Error(), "remove") {
		t.Errorf("expected usage error; got %v", err)
	}
}
