//go:build llm

package simulation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dpopsuev/mos/moslib/testkit/mockide"
)

type LLMAgent struct {
	ide    *mockide.MockIDE
	apiKey string
	model  string
	tools  []ToolDef
}

type ToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

func NewLLMAgent(ide *mockide.MockIDE, tools []ToolDef) *LLMAgent {
	apiKey := os.Getenv("LLM_API_KEY")
	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	return &LLMAgent{ide: ide, apiKey: apiKey, model: model, tools: tools}
}

func (a *LLMAgent) RunPrompt(systemPrompt, userPrompt string, maxTurns int) ([]StepResult, error) {
	if a.apiKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY not set")
	}

	var results []StepResult
	messages := []map[string]string{
		{"role": "user", "content": userPrompt},
	}

	for turn := 0; turn < maxTurns; turn++ {
		resp, err := a.callLLM(systemPrompt, messages)
		if err != nil {
			return results, fmt.Errorf("llm turn %d: %w", turn, err)
		}

		toolCalls := extractToolCalls(resp)
		if len(toolCalls) == 0 {
			break
		}

		for _, tc := range toolCalls {
			raw, err := a.ide.Call(tc.Name, tc.Params)
			result := StepResult{Tool: tc.Name, Output: raw, Passed: err == nil}
			if err != nil {
				result.Error = err.Error()
			}
			results = append(results, result)
		}
	}

	return results, nil
}

type toolCall struct {
	Name   string         `json:"name"`
	Params map[string]any `json:"input"`
}

func (a *LLMAgent) callLLM(system string, messages []map[string]string) (json.RawMessage, error) {
	body := map[string]any{
		"model":      a.model,
		"max_tokens": 4096,
		"system":     system,
		"messages":   messages,
		"tools":      a.tools,
	}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return raw, nil
}

func extractToolCalls(resp json.RawMessage) []toolCall {
	var parsed struct {
		Content []struct {
			Type  string         `json:"type"`
			Name  string         `json:"name"`
			Input map[string]any `json:"input"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp, &parsed); err != nil {
		return nil
	}
	var calls []toolCall
	for _, c := range parsed.Content {
		if c.Type == "tool_use" {
			calls = append(calls, toolCall{Name: c.Name, Params: c.Input})
		}
	}
	return calls
}
