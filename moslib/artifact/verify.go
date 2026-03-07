package artifact

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/harness"
)

// VerifyResult holds the outcome of harness verification for one scenario.
type VerifyResult struct {
	Scenario string
	Pass     bool
	RuleID   string
	Evidence string
}

// VerifyContract discovers harness specs, runs them, and transitions
// matching "implemented" scenarios to "verified" on success.
func VerifyContract(root, id string) ([]VerifyResult, error) {
	scenarios, err := ListScenarios(root, id)
	if err != nil {
		return nil, fmt.Errorf("listing scenarios: %w", err)
	}

	var implemented []ScenarioInfo
	for _, s := range scenarios {
		if s.Status == "implemented" {
			implemented = append(implemented, s)
		}
	}
	if len(implemented) == 0 {
		return nil, nil
	}

	mosDir := filepath.Join(root, MosDir)
	specs, err := harness.Discover(mosDir)
	if err != nil {
		return nil, fmt.Errorf("discovering harness: %w", err)
	}
	if len(specs) == 0 {
		return nil, fmt.Errorf("no harness specs found; cannot verify")
	}

	evidence := harness.Run(root, specs)
	allPass := true
	for _, ev := range evidence {
		if !ev.Pass {
			allPass = false
			break
		}
	}

	var results []VerifyResult
	for _, s := range implemented {
		vr := VerifyResult{
			Scenario: s.Name,
			Pass:     allPass,
		}
		if len(evidence) > 0 {
			vr.RuleID = evidence[0].RuleID
			vr.Evidence = evidence[0].Stdout
		}

		if allPass {
			now := time.Now().UTC().Format(time.RFC3339)
			setScenarioField(root, id, s.Name, "status", "verified")
			setScenarioField(root, id, s.Name, "verified_at", now)
		}
		results = append(results, vr)
	}

	if allPass {
		reg, err := LoadRegistry(root)
		if err == nil {
			EvaluateHooks(root, id, reg)
		}
	}

	return results, nil
}

func setScenarioField(root, id, scenarioName, fieldKey, fieldValue string) {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return
	}
	_ = dsl.WithArtifact(contractPath, func(ab *dsl.ArtifactBlock) error {
		setFieldOnMatchedScenario(ab.Items, scenarioName, fieldKey, fieldValue)
		return nil
	})
}

func setFieldOnMatchedScenario(items []dsl.Node, name, key, value string) bool {
	needle := strings.ToLower(name)
	for _, item := range items {
		switch n := item.(type) {
		case *dsl.Block:
			if setFieldOnMatchedScenario(n.Items, name, key, value) {
				return true
			}
		case *dsl.FeatureBlock:
			for _, group := range n.Groups {
				switch g := group.(type) {
				case *dsl.Scenario:
					if strings.Contains(strings.ToLower(g.Name), needle) {
						setFieldOnScenario(g, key, value)
						return true
					}
				case *dsl.Group:
					for _, sc := range g.Scenarios {
						if strings.Contains(strings.ToLower(sc.Name), needle) {
							setFieldOnScenario(sc, key, value)
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func setFieldOnScenario(s *dsl.Scenario, key, value string) {
	for _, f := range s.Fields {
		if f.Key == key {
			f.Value = &dsl.StringVal{Text: value}
			return
		}
	}
	s.Fields = append(s.Fields, &dsl.Field{
		Key:   key,
		Value: &dsl.StringVal{Text: value},
	})
}
