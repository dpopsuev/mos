package simulation

import (
	"encoding/json"
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

type Scenario struct {
	Name  string `yaml:"name" json:"name"`
	Steps []Step `yaml:"steps" json:"steps"`
}

type Step struct {
	Tool       string         `yaml:"tool" json:"tool"`
	Params     map[string]any `yaml:"params" json:"params"`
	AssertKeys []string       `yaml:"assert_keys,omitempty" json:"assert_keys,omitempty"`
	AssertErr  bool           `yaml:"assert_err,omitempty" json:"assert_err,omitempty"`
}

type StepResult struct {
	Tool    string          `json:"tool"`
	Output  json.RawMessage `json:"output,omitempty"`
	Error   string          `json:"error,omitempty"`
	Passed  bool            `json:"passed"`
	Details string          `json:"details,omitempty"`
}

func LoadScenario(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenario: %w", err)
	}
	var s Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse scenario: %w", err)
	}
	return &s, nil
}
