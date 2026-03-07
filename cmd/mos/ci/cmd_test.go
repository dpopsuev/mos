package ci

import (
	"bytes"
	"encoding/json"
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
	// Minimal main.go so go build ./... succeeds
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("writing main.go: %v", err)
	}
	if err := artifact.Init(root, artifact.InitOpts{Name: "test", Model: "bdfl", Scope: "cabinet"}); err != nil {
		t.Fatalf("artifact.Init: %v", err)
	}
	return root
}

func runCICmd(t *testing.T, root string, args []string) (stdout, stderr string, err error) {
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
	cmd.SetArgs(append([]string{"ci"}, args...))

	outDone := make(chan struct{})
	errDone := make(chan struct{})
	var outBuf, errBuf bytes.Buffer
	go func() {
		_, _ = outBuf.ReadFrom(rOut)
		close(outDone)
	}()
	go func() {
		_, _ = errBuf.ReadFrom(rErr)
		close(errDone)
	}()

	runErr := cmd.Execute()
	wOut.Close()
	wErr.Close()
	<-outDone
	<-errDone

	return outBuf.String(), errBuf.String(), runErr
}

func TestCIFastSkipsHeavyStages(t *testing.T) {
	root := setupWorkspace(t)
	_, stderr, err := runCICmd(t, root, []string{"--fast"})
	if err != nil {
		t.Fatalf("ci --fast: %v", err)
	}
	// --fast runs build, vet, lint only; should not run test, audit, harness
	if !strings.Contains(stderr, "build") {
		t.Errorf("expected build stage in output; got %q", stderr)
	}
	if !strings.Contains(stderr, "vet") {
		t.Errorf("expected vet stage in output; got %q", stderr)
	}
	if !strings.Contains(stderr, "lint") {
		t.Errorf("expected lint stage in output; got %q", stderr)
	}
	if strings.Contains(stderr, "test") {
		t.Errorf("--fast should skip test stage; got %q", stderr)
	}
}

func TestCIFormatJSON(t *testing.T) {
	root := setupWorkspace(t)
	stdout, _, err := runCICmd(t, root, []string{"--fast", "--format", "json"})
	if err != nil {
		t.Fatalf("ci --fast --format json: %v", err)
	}
	var results []struct {
		Stage string `json:"stage"`
		Pass  bool   `json:"pass"`
	}
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("expected valid JSON output: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one stage in JSON output")
	}
}

func TestCIRequiresWorkspace(t *testing.T) {
	root := t.TempDir()
	// No go.mod, no governance - CI should fail at build or earlier
	origDir, origErr := os.Getwd()
	if origErr != nil {
		t.Fatalf("os.Getwd: %v", origErr)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("os.Chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	cmd := &cobra.Command{Use: "mos"}
	cmd.AddCommand(Cmd)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"ci", "--fast"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected CI to fail without valid workspace")
	}
}
