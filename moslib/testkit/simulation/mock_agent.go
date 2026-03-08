package simulation

import (
	"encoding/json"
	"fmt"

	"github.com/dpopsuev/mos/moslib/testkit/mockide"
)

type MockAgent struct {
	ide *mockide.MockIDE
}

func NewMockAgent(ide *mockide.MockIDE) *MockAgent {
	return &MockAgent{ide: ide}
}

func (a *MockAgent) RunScenario(scenario *Scenario) ([]StepResult, error) {
	results := make([]StepResult, 0, len(scenario.Steps))

	for i, step := range scenario.Steps {
		raw, err := a.ide.Call(step.Tool, step.Params)

		result := StepResult{Tool: step.Tool}

		if step.AssertErr {
			if err != nil {
				result.Passed = true
				result.Details = "expected error received"
			} else {
				result.Passed = false
				result.Details = "expected error but got success"
			}
			result.Error = fmt.Sprintf("%v", err)
			results = append(results, result)
			continue
		}

		if err != nil {
			result.Error = err.Error()
			result.Passed = false
			result.Details = fmt.Sprintf("step %d failed: %v", i, err)
			results = append(results, result)
			return results, fmt.Errorf("step %d (%s): %w", i, step.Tool, err)
		}

		result.Output = raw
		result.Passed = true

		if len(step.AssertKeys) > 0 {
			var parsed map[string]any
			if err := json.Unmarshal(raw, &parsed); err == nil {
				for _, key := range step.AssertKeys {
					if _, ok := parsed[key]; !ok {
						result.Passed = false
						result.Details = fmt.Sprintf("missing key %q in response", key)
						break
					}
				}
			}
		}

		results = append(results, result)
	}

	return results, nil
}
