package cliutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// IsAgentMode returns true when MOS_AGENT=1 is set.
func IsAgentMode() bool {
	return os.Getenv("MOS_AGENT") == "1"
}

// AgentEnvelope is the standard JSON wrapper for agent protocol output.
type AgentEnvelope struct {
	Status   string `json:"status"`
	Output   string `json:"output,omitempty"`
	Message  string `json:"message,omitempty"`
	ExitCode int    `json:"exit_code"`
}

// CaptureStdout redirects os.Stdout to a buffer, runs fn, then restores.
// Returns captured output and any error from fn.
func CaptureStdout(fn func() error) (string, error) {
	origOut := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", fn()
	}
	os.Stdout = w

	fnErr := fn()

	w.Close()
	os.Stdout = origOut

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	return buf.String(), fnErr
}

// EmitAgentEnvelope writes a JSON agent envelope to stdout.
func EmitAgentEnvelope(output string, fnErr error) {
	env := AgentEnvelope{
		Status:   "ok",
		Output:   output,
		ExitCode: 0,
	}
	if fnErr != nil {
		env.Status = "error"
		msg := fnErr.Error()
		if msg == "" {
			env.ExitCode = 1
		} else {
			env.Message = msg
			env.ExitCode = 1
		}
	}
	data, _ := json.MarshalIndent(env, "", "  ")
	fmt.Fprintln(os.Stdout, string(data))
}
