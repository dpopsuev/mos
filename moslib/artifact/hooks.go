package artifact

import (
	"fmt"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// EvaluateHooks loads a contract's scenarios, evaluates lifecycle hooks,
// and auto-transitions the contract's fields when thresholds are met.
func EvaluateHooks(root, contractID string, reg *Registry) error {
	td, ok := reg.Types[KindContract]
	if !ok || len(td.Lifecycle.Hooks) == 0 {
		return nil
	}

	scenarios, err := ListScenarios(root, contractID)
	if err != nil {
		return nil
	}
	if len(scenarios) == 0 {
		return nil
	}

	contractPath, err := FindContractPath(root, contractID)
	if err != nil {
		return nil
	}
	if err := dsl.WithArtifact(contractPath, func(ab *dsl.ArtifactBlock) error {
		currentFields := make(map[string]string)
		for k, v := range dsl.ToMap(ab) {
			if s, ok := v.(string); ok {
				currentFields[k] = s
			}
		}

		for _, hook := range td.Lifecycle.Hooks {
			if currentFields[hook.SetField] == hook.SetValue {
				continue
			}

			var fires bool
			switch hook.Trigger {
			case "on_any":
				fires = evaluateOnAny(scenarios, hook)
			case "on_all":
				fires = evaluateOnAll(scenarios, hook)
			}

			if fires {
				oldVal := currentFields[hook.SetField]
				dsl.SetField(&ab.Items, hook.SetField, &dsl.StringVal{Text: hook.SetValue})
				currentFields[hook.SetField] = hook.SetValue

				AppendContractLedger(root, contractID, LedgerEntry{
					Event:    "hook_triggered",
					Field:    hook.SetField,
					OldValue: oldVal,
					NewValue: hook.SetValue,
				})
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("EvaluateHooks: %w", err)
	}
	return nil
}

func evaluateOnAny(scenarios []ScenarioInfo, hook HookDef) bool {
	for _, s := range scenarios {
		if s.Status == hook.Threshold {
			return true
		}
	}
	return false
}

func evaluateOnAll(scenarios []ScenarioInfo, hook HookDef) bool {
	for _, s := range scenarios {
		if s.Status != hook.Threshold {
			return false
		}
	}
	return true
}

// GetContractStatus reads the current status of a contract.
func GetContractStatus(root, id string) (string, error) {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return "", fmt.Errorf("GetContractStatus: %w", err)
	}
	ab, err := dsl.ReadArtifact(contractPath)
	if err != nil {
		return "", fmt.Errorf("GetContractStatus: %w", err)
	}
	s, _ := dsl.FieldString(ab.Items, "status")
	return s, nil
}
