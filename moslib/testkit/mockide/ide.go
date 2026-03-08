package mockide

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"sync/atomic"
)

type ToolCall struct {
	Name   string         `json:"name"`
	Params map[string]any `json:"params,omitempty"`
}

type Result struct {
	Content json.RawMessage `json:"content,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type MockIDE struct {
	cmd    *exec.Cmd
	stdin  *json.Encoder
	stdout *bufio.Scanner
	mu     sync.Mutex
	nextID atomic.Int64
}

func New(binary string, args ...string) (*MockIDE, error) {
	cmd := exec.Command(binary, args...)
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}

	ide := &MockIDE{
		cmd:    cmd,
		stdin:  json.NewEncoder(stdinPipe),
		stdout: bufio.NewScanner(stdoutPipe),
	}
	return ide, nil
}

type jsonrpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (ide *MockIDE) Initialize() (json.RawMessage, error) {
	return ide.rawCall("initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "mockide",
			"version": "0.1.0",
		},
	})
}

func (ide *MockIDE) Call(toolName string, params map[string]any) (json.RawMessage, error) {
	return ide.rawCall("tools/call", map[string]any{
		"name":      toolName,
		"arguments": params,
	})
}

func (ide *MockIDE) CallSequence(calls []ToolCall) []Result {
	results := make([]Result, len(calls))
	for i, c := range calls {
		raw, err := ide.Call(c.Name, c.Params)
		if err != nil {
			results[i] = Result{Error: err.Error()}
		} else {
			results[i] = Result{Content: raw}
		}
	}
	return results
}

func (ide *MockIDE) Close() error {
	if ide.cmd.Process != nil {
		ide.cmd.Process.Kill()
	}
	return ide.cmd.Wait()
}

func (ide *MockIDE) rawCall(method string, params any) (json.RawMessage, error) {
	ide.mu.Lock()
	defer ide.mu.Unlock()

	id := ide.nextID.Add(1)
	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := ide.stdin.Encode(req); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	for ide.stdout.Scan() {
		line := ide.stdout.Bytes()
		var resp jsonrpcResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}
		if resp.ID == id {
			if resp.Error != nil {
				return nil, fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
			}
			return resp.Result, nil
		}
	}

	if err := ide.stdout.Err(); err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	return nil, fmt.Errorf("no response for id %d", id)
}
