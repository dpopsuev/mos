package artifact

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/schema"
)

// ScenarioInfo describes a scenario found within a contract.
type ScenarioInfo struct {
	Name   string
	Status string   // "pending", "implemented", "verified", or "done" (legacy)
	Labels []string // classification labels (e.g. "happy_path", "sad_path", "adversarial")
	Actor  string   // persona performing the action (e.g. "contributor", "attacker")
}

// ValidateScenarioTransition checks whether moving from oldStatus to newStatus
// is allowed by the ordered enum constraints. Returns nil if allowed.
func ValidateScenarioTransition(reg *Registry, oldStatus, newStatus string) error {
	td, ok := reg.Types[KindContract]
	if !ok {
		return nil
	}
	var fd *schema.FieldSchema
	for i := range td.ScenarioFields {
		if td.ScenarioFields[i].Name == FieldStatus {
			fd = &td.ScenarioFields[i]
			break
		}
	}
	if fd == nil || !fd.Ordered || len(fd.Enum) == 0 {
		return nil
	}

	oldIdx := enumIndex(fd.Enum, oldStatus)
	newIdx := enumIndex(fd.Enum, newStatus)
	if oldIdx < 0 || newIdx < 0 {
		return nil
	}
	if newIdx < oldIdx {
		return fmt.Errorf("backward transition from %q to %q is not allowed (ordered enum)", oldStatus, newStatus)
	}

	for _, tr := range fd.Transitions {
		if tr.From == oldStatus && tr.To == newStatus && tr.VerifiedBy != "" {
			return fmt.Errorf("transition from %q to %q requires verification by %s", oldStatus, newStatus, tr.VerifiedBy)
		}
	}
	return nil
}

func enumIndex(enum []string, val string) int {
	return slices.Index(enum, val)
}

// ListScenarios returns all scenarios in the given contract, with their status.
// Scenarios without an explicit status field are reported as "pending".
func ListScenarios(root, id string) ([]ScenarioInfo, error) {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return nil, fmt.Errorf("ListScenarios: %w", err)
	}
	ab, err := dsl.ReadArtifact(contractPath)
	if err != nil {
		return nil, fmt.Errorf("ListScenarios: %w", err)
	}
	var infos []ScenarioInfo
	collectScenarios(ab.Items, &infos)
	return infos, nil
}

func collectScenarios(items []dsl.Node, infos *[]ScenarioInfo) {
	for _, item := range items {
		switch n := item.(type) {
		case *dsl.Block:
			collectScenarios(n.Items, infos)
		case *dsl.FeatureBlock:
			collectFeatureScenarios(n, infos)
		}
	}
}

func collectFeatureScenarios(fb *dsl.FeatureBlock, infos *[]ScenarioInfo) {
	for _, group := range fb.Groups {
		switch g := group.(type) {
		case *dsl.Scenario:
			status := scenarioStatus(g)
			if status == "" {
				status = "pending"
			}
			*infos = append(*infos, ScenarioInfo{Name: g.Name, Status: status, Labels: g.Labels(), Actor: g.Actor()})
		case *dsl.Group:
			for _, sc := range g.Scenarios {
				status := scenarioStatus(sc)
				if status == "" {
					status = "pending"
				}
				*infos = append(*infos, ScenarioInfo{Name: sc.Name, Status: status, Labels: sc.Labels(), Actor: sc.Actor()})
			}
		}
	}
}

// SetScenarioStatus sets the status of a named scenario within the contract.
// The name is matched case-insensitively as a substring.
func SetScenarioStatus(root, id, name, status string) error {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return fmt.Errorf("SetScenarioStatus: %w", err)
	}
	var oldStatus string
	if err := dsl.WithArtifact(contractPath, func(ab *dsl.ArtifactBlock) error {
		oldStatus = findScenarioStatus(ab.Items, name)
		if !setScenarioStatusInItems(ab.Items, name, status) {
			return fmt.Errorf("scenario matching %q not found in %s", name, id)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("SetScenarioStatus: %w", err)
	}

	AppendContractLedger(root, id, LedgerEntry{
		Event:        "scenario_status_changed",
		Field:        FieldStatus,
		OldValue:     oldStatus,
		NewValue:     status,
		ScenarioName: name,
	})

	reg, err := LoadRegistry(root)
	if err == nil {
		EvaluateHooks(root, id, reg)
	}
	return nil
}

func findScenarioStatus(items []dsl.Node, name string) string {
	needle := strings.ToLower(name)
	for _, item := range items {
		switch n := item.(type) {
		case *dsl.Block:
			if s := findScenarioStatus(n.Items, name); s != "" {
				return s
			}
		case *dsl.FeatureBlock:
			for _, group := range n.Groups {
				switch g := group.(type) {
				case *dsl.Scenario:
					if strings.Contains(strings.ToLower(g.Name), needle) {
						s := scenarioStatus(g)
						if s == "" {
							return "pending"
						}
						return s
					}
				case *dsl.Group:
					for _, sc := range g.Scenarios {
						if strings.Contains(strings.ToLower(sc.Name), needle) {
							s := scenarioStatus(sc)
							if s == "" {
								return "pending"
							}
							return s
						}
					}
				}
			}
		}
	}
	return "pending"
}

func setScenarioStatusInItems(items []dsl.Node, name, status string) bool {
	for _, item := range items {
		switch n := item.(type) {
		case *dsl.Block:
			if setScenarioStatusInItems(n.Items, name, status) {
				return true
			}
		case *dsl.FeatureBlock:
			if setScenarioStatusInFeature(n, name, status) {
				return true
			}
		}
	}
	return false
}

func setScenarioStatusInFeature(fb *dsl.FeatureBlock, name, status string) bool {
	needle := strings.ToLower(name)
	for _, group := range fb.Groups {
		switch g := group.(type) {
		case *dsl.Scenario:
			if strings.ToLower(g.Name) == needle || strings.Contains(strings.ToLower(g.Name), needle) {
				setStatusOnScenario(g, status)
				return true
			}
		case *dsl.Group:
			for _, sc := range g.Scenarios {
				if strings.ToLower(sc.Name) == needle || strings.Contains(strings.ToLower(sc.Name), needle) {
					setStatusOnScenario(sc, status)
					return true
				}
			}
		}
	}
	return false
}

func setStatusOnScenario(s *dsl.Scenario, status string) {
	for _, f := range s.Fields {
		if f.Key == FieldStatus {
			f.Value = &dsl.StringVal{Text: status}
			return
		}
	}
	s.Fields = append([]*dsl.Field{{
		Key:   FieldStatus,
		Value: &dsl.StringVal{Text: status},
	}}, s.Fields...)
}

// SetAllScenariosStatus sets the status of every scenario in the contract.
func SetAllScenariosStatus(root, id, status string) (int, error) {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return 0, fmt.Errorf("SetAllScenariosStatus: %w", err)
	}
	var count int
	if err := dsl.WithArtifact(contractPath, func(ab *dsl.ArtifactBlock) error {
		count = setAllStatusInItems(ab.Items, status)
		return nil
	}); err != nil {
		return 0, fmt.Errorf("SetAllScenariosStatus: %w", err)
	}
	if count == 0 {
		return 0, nil
	}

	reg, err := LoadRegistry(root)
	if err == nil {
		EvaluateHooks(root, id, reg)
	}
	return count, nil
}

func setAllStatusInItems(items []dsl.Node, status string) int {
	count := 0
	for _, item := range items {
		switch n := item.(type) {
		case *dsl.Block:
			count += setAllStatusInItems(n.Items, status)
		case *dsl.FeatureBlock:
			count += setAllStatusInFeature(n, status)
		}
	}
	return count
}

func setAllStatusInFeature(fb *dsl.FeatureBlock, status string) int {
	count := 0
	for _, group := range fb.Groups {
		switch g := group.(type) {
		case *dsl.Scenario:
			setStatusOnScenario(g, status)
			count++
		case *dsl.Group:
			for _, sc := range g.Scenarios {
				setStatusOnScenario(sc, status)
				count++
			}
		}
	}
	return count
}
