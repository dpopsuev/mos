package artifact

import (
	"fmt"
	"sort"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// SprintProposal is a suggested sprint built from unassigned backlog contracts.
type SprintProposal struct {
	Title       string   `json:"title"`
	ContractIDs []string `json:"contract_ids"`
	Reasoning   string   `json:"reasoning"`
}

// PlanSprint scans draft/active contracts without a sprint assignment,
// ranks by dependency order, and proposes a sprint of up to max contracts.
func PlanSprint(root string, max int) (*SprintProposal, error) {
	if max <= 0 {
		max = 8
	}

	candidates, err := QueryArtifacts(root, QueryOpts{Kind: KindContract})
	if err != nil {
		return nil, fmt.Errorf("PlanSprint: querying contracts: %w", err)
	}

	var unassigned []QueryResult
	for _, c := range candidates {
		if c.Status != "draft" && c.Status != "active" {
			continue
		}
		if c.Sprint != "" {
			continue
		}
		unassigned = append(unassigned, c)
	}

	if len(unassigned) == 0 {
		return nil, fmt.Errorf("PlanSprint: no unassigned draft/active contracts in backlog")
	}

	sort.Slice(unassigned, func(i, j int) bool {
		pi := planPriority(unassigned[i])
		pj := planPriority(unassigned[j])
		if pi != pj {
			return pi < pj
		}
		return unassigned[i].ID < unassigned[j].ID
	})

	if len(unassigned) > max {
		unassigned = unassigned[:max]
	}

	ids := make([]string, len(unassigned))
	for i, c := range unassigned {
		ids[i] = c.ID
	}

	return &SprintProposal{
		Title:       "Backlog Sprint",
		ContractIDs: ids,
		Reasoning:   fmt.Sprintf("Selected %d of %d unassigned contracts, prioritized by status (active first) and ID order.", len(ids), len(unassigned)),
	}, nil
}

func planPriority(q QueryResult) int {
	score := 0
	if q.Status == "active" {
		score -= 10
	}
	if q.Path != "" {
		ab, err := dsl.ReadArtifact(q.Path)
		if err == nil {
			if deps, ok := dsl.FieldString(ab.Items, "depends_on"); ok && deps != "" {
				score += 5
			}
		}
	}
	return score
}
