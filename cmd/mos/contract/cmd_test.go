package contract

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

func runContractCmd(t *testing.T, root string, args []string) (stdout, stderr string, err error) {
	t.Helper()
	origDir, origErr := os.Getwd()
	if origErr != nil {
		t.Fatalf("os.Getwd: %v", origErr)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("os.Chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// CLI uses fmt.Printf which writes to os.Stdout; capture it
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
	cmd.SetArgs(append([]string{"contract"}, args...))

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

func TestContractCreateRequiresTitle(t *testing.T) {
	root := setupWorkspace(t)
	_, stderr, err := runContractCmd(t, root, []string{"create"})
	if err == nil {
		t.Fatal("expected error when running contract create without --title")
	}
	if stderr == "" && err != nil && err.Error() == "" {
		t.Log("cobra may report error via Execute() return; checking err")
	}
	if !strings.Contains(err.Error(), "required") && !strings.Contains(stderr, "required") && !strings.Contains(err.Error(), "title") {
		t.Errorf("expected error about required title; got err=%v stderr=%q", err, stderr)
	}
}

func TestContractCreateWithTitle(t *testing.T) {
	root := setupWorkspace(t)
	stdout, _, err := runContractCmd(t, root, []string{"create", "--title", "Test"})
	if err != nil {
		t.Fatalf("contract create --title Test: %v", err)
	}
	if !strings.Contains(stdout, "Created contract:") {
		t.Errorf("expected 'Created contract:' in output; got %q", stdout)
	}

	// Assert contract file on disk
	contractsDir := filepath.Join(root, ".mos", "contracts", "active")
	entries, err := os.ReadDir(contractsDir)
	if err != nil {
		t.Fatalf("ReadDir contracts: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one contract directory under .mos/contracts/active/")
	}
	found := false
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "CON-") {
			contractPath := filepath.Join(contractsDir, e.Name(), "contract.mos")
			if _, statErr := os.Stat(contractPath); statErr == nil {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("expected contract.mos under .mos/contracts/active/; entries: %v", entries)
	}
}

func TestContractListEmpty(t *testing.T) {
	root := setupWorkspace(t)
	stdout, _, err := runContractCmd(t, root, []string{"list"})
	if err != nil {
		t.Fatalf("contract list: %v", err)
	}
	if !strings.Contains(stdout, "(no contracts found)") {
		t.Errorf("expected '(no contracts found)' in output; got %q", stdout)
	}
}
