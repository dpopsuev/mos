//go:build container

package containers

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestContainerSimulation(t *testing.T) {
	if _, err := exec.LookPath("podman"); err != nil {
		t.Skip("podman not found, skipping container test")
	}

	_, thisFile, _, _ := runtime.Caller(0)
	composeDir := filepath.Dir(thisFile)
	composeFile := filepath.Join(composeDir, "compose.yaml")

	t.Cleanup(func() {
		cmd := exec.Command("podman", "compose", "-f", composeFile, "down", "-v")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	})

	build := exec.Command("podman", "compose", "-f", composeFile, "build")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("compose build: %v", err)
	}

	up := exec.Command("podman", "compose", "-f", composeFile, "up", "--abort-on-container-exit")
	up.Stdout = os.Stdout
	up.Stderr = os.Stderr
	if err := up.Run(); err != nil {
		t.Logf("compose up exited: %v (expected for short-lived containers)", err)
	}

	logsDir := t.TempDir()
	for _, svc := range []string{"ide-a1", "ide-a2", "ide-b1", "ide-b2"} {
		out, err := exec.Command("podman", "compose", "-f", composeFile, "logs", svc).Output()
		if err != nil {
			t.Logf("logs %s: %v", svc, err)
		}
		os.WriteFile(filepath.Join(logsDir, svc+".log"), out, 0o644)
	}

	verify := exec.Command("bash", filepath.Join(composeDir, "verify.sh"), logsDir)
	verify.Stdout = os.Stdout
	verify.Stderr = os.Stderr
	if err := verify.Run(); err != nil {
		t.Fatalf("verification failed: %v", err)
	}
}
