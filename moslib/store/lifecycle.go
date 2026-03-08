package store

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func EnsureDaemon(workspaceRoots []string) (Store, error) {
	sockPath := DefaultSocketPath()

	client, err := Dial(sockPath, workspaceRoots)
	if err == nil {
		return client, nil
	}

	cleanStalePID()
	cleanStaleSocket(sockPath)

	if err := startDaemon(sockPath); err != nil {
		return nil, fmt.Errorf("start daemon: %w", err)
	}

	for i := 0; i < 50; i++ {
		time.Sleep(100 * time.Millisecond)
		client, err = Dial(sockPath, workspaceRoots)
		if err == nil {
			return client, nil
		}
	}
	return nil, fmt.Errorf("daemon did not start within 5s: %w", err)
}

func startDaemon(sockPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	cmd := exec.Command(exe, "daemon", "--socket", sockPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("exec daemon: %w", err)
	}

	cmd.Process.Release()
	return nil
}

func cleanStalePID() {
	pidPath := DefaultPIDPath()
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		os.Remove(pidPath)
		return
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		os.Remove(pidPath)
		return
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		os.Remove(pidPath)
	}
}

func cleanStaleSocket(sockPath string) {
	conn, err := net.DialTimeout("unix", sockPath, 500*time.Millisecond)
	if err != nil {
		os.Remove(sockPath)
		return
	}
	conn.Close()
}
