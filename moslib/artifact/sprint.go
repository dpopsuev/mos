package artifact

import (
	"fmt"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// SprintAddContracts appends contract IDs to a sprint's contracts field
// and sets the sprint back-reference on each contract.
func SprintAddContracts(root, sprintID string, contractIDs []string) error {
	reg, err := LoadRegistry(root)
	if err != nil {
		return fmt.Errorf("SprintAddContracts: %w", err)
	}
	td, ok := reg.Types[KindSprint]
	if !ok {
		return fmt.Errorf("SprintAddContracts: sprint type not found in registry")
	}

	sprintPath, err := FindGenericPath(root, td, sprintID)
	if err != nil {
		return fmt.Errorf("SprintAddContracts: %w", err)
	}

	if err := dsl.WithArtifact(sprintPath, func(ab *dsl.ArtifactBlock) error {
		current, _ := dsl.FieldString(ab.Items, "contracts")
		existing := parseCSV(current)
		set := make(map[string]bool, len(existing))
		for _, id := range existing {
			set[id] = true
		}
		for _, id := range contractIDs {
			if !set[id] {
				existing = append(existing, id)
				set[id] = true
			}
		}
		dsl.SetField(&ab.Items, "contracts", &dsl.StringVal{Text: strings.Join(existing, ",")})
		return nil
	}); err != nil {
		return err
	}

	for _, cid := range contractIDs {
		if err := SetArtifactField(root, cid, "sprint", sprintID); err != nil {
			return fmt.Errorf("SprintAddContracts: setting back-ref on %s: %w", cid, err)
		}
	}
	return nil
}

// SprintRemoveContracts removes contract IDs from a sprint's contracts field
// and clears the sprint back-reference on each contract.
func SprintRemoveContracts(root, sprintID string, contractIDs []string) error {
	reg, err := LoadRegistry(root)
	if err != nil {
		return fmt.Errorf("SprintRemoveContracts: %w", err)
	}
	td, ok := reg.Types[KindSprint]
	if !ok {
		return fmt.Errorf("SprintRemoveContracts: sprint type not found in registry")
	}

	sprintPath, err := FindGenericPath(root, td, sprintID)
	if err != nil {
		return fmt.Errorf("SprintRemoveContracts: %w", err)
	}

	removeSet := make(map[string]bool, len(contractIDs))
	for _, id := range contractIDs {
		removeSet[id] = true
	}

	if err := dsl.WithArtifact(sprintPath, func(ab *dsl.ArtifactBlock) error {
		current, _ := dsl.FieldString(ab.Items, "contracts")
		existing := parseCSV(current)
		var kept []string
		for _, id := range existing {
			if !removeSet[id] {
				kept = append(kept, id)
			}
		}
		if len(kept) == 0 {
			dsl.SetField(&ab.Items, "contracts", &dsl.StringVal{Text: ""})
		} else {
			dsl.SetField(&ab.Items, "contracts", &dsl.StringVal{Text: strings.Join(kept, ",")})
		}
		return nil
	}); err != nil {
		return err
	}

	for _, cid := range contractIDs {
		_ = SetArtifactField(root, cid, "sprint", "")
	}
	return nil
}

// SprintCloseResult describes the outcome of closing a sprint.
type SprintCloseResult struct {
	SprintID    string   `json:"sprint_id"`
	Contracts   []string `json:"contracts"`
	Closed      int      `json:"closed"`
	AlreadyDone int      `json:"already_done"`
}

// SprintClose marks all contracts in a sprint as complete and sets the sprint
// status to complete. Returns the list of affected contracts.
func SprintClose(root, sprintID string) (*SprintCloseResult, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, fmt.Errorf("SprintClose: %w", err)
	}
	td, ok := reg.Types[KindSprint]
	if !ok {
		return nil, fmt.Errorf("SprintClose: sprint type not found in registry")
	}

	sprintPath, err := FindGenericPath(root, td, sprintID)
	if err != nil {
		return nil, fmt.Errorf("SprintClose: %w", err)
	}

	ab, err := dsl.ReadArtifact(sprintPath)
	if err != nil {
		return nil, fmt.Errorf("SprintClose: reading sprint: %w", err)
	}
	contractsCSV, _ := dsl.FieldString(ab.Items, "contracts")
	contractIDs := parseCSV(contractsCSV)

	result := &SprintCloseResult{
		SprintID:  sprintID,
		Contracts: contractIDs,
	}

	for _, cid := range contractIDs {
		status, err := GetContractStatus(root, cid)
		if err != nil {
			continue
		}
		if status == StatusComplete {
			result.AlreadyDone++
			continue
		}
		if err := UpdateContractStatus(root, cid, StatusComplete); err != nil {
			return nil, fmt.Errorf("SprintClose: marking %s complete: %w", cid, err)
		}
		result.Closed++
	}

	if err := GenericUpdateStatus(root, td, sprintID, StatusComplete); err != nil {
		return nil, fmt.Errorf("SprintClose: setting sprint complete: %w", err)
	}

	return result, nil
}

// SprintCloseDryRun returns what SprintClose would do without mutating anything.
func SprintCloseDryRun(root, sprintID string) (*SprintCloseResult, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, fmt.Errorf("SprintCloseDryRun: %w", err)
	}
	td, ok := reg.Types[KindSprint]
	if !ok {
		return nil, fmt.Errorf("SprintCloseDryRun: sprint type not found in registry")
	}

	sprintPath, err := FindGenericPath(root, td, sprintID)
	if err != nil {
		return nil, fmt.Errorf("SprintCloseDryRun: %w", err)
	}

	ab, err := dsl.ReadArtifact(sprintPath)
	if err != nil {
		return nil, fmt.Errorf("SprintCloseDryRun: reading sprint: %w", err)
	}
	contractsCSV, _ := dsl.FieldString(ab.Items, "contracts")
	contractIDs := parseCSV(contractsCSV)

	result := &SprintCloseResult{
		SprintID:  sprintID,
		Contracts: contractIDs,
	}

	for _, cid := range contractIDs {
		status, err := GetContractStatus(root, cid)
		if err != nil {
			continue
		}
		if status == StatusComplete {
			result.AlreadyDone++
		} else {
			result.Closed++
		}
	}

	return result, nil
}

func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
