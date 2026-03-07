package cliutil

import (
	"os"
	"testing"
)

func TestIsAgentMode_Off(t *testing.T) {
	os.Unsetenv("MOS_AGENT")
	if IsAgentMode() {
		t.Error("expected false when MOS_AGENT not set")
	}
}

func TestIsAgentMode_On(t *testing.T) {
	os.Setenv("MOS_AGENT", "1")
	defer os.Unsetenv("MOS_AGENT")
	if !IsAgentMode() {
		t.Error("expected true when MOS_AGENT=1")
	}
}

func TestIsAgentMode_Other(t *testing.T) {
	os.Setenv("MOS_AGENT", "true")
	defer os.Unsetenv("MOS_AGENT")
	if IsAgentMode() {
		t.Error("expected false when MOS_AGENT=true (only '1' is valid)")
	}
}

func TestCaptureStdout(t *testing.T) {
	output, err := CaptureStdout(func() error {
		os.Stdout.WriteString("hello agent")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "hello agent" {
		t.Errorf("output = %q, want %q", output, "hello agent")
	}
}

func TestCaptureStdout_Error(t *testing.T) {
	output, err := CaptureStdout(func() error {
		os.Stdout.WriteString("partial")
		return ErrNonZeroExit
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if output != "partial" {
		t.Errorf("output = %q, want %q", output, "partial")
	}
}
