package artifact

import (
	"fmt"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// Criterion represents a single acceptance criterion on a need.
type Criterion struct {
	Name        string
	Description string
	VerifiedBy  string // "harness", "human", or "agent"
}

// ParseAcceptanceCriteria extracts acceptance criteria from a need artifact AST.
func ParseAcceptanceCriteria(ab *dsl.ArtifactBlock) []Criterion {
	if blk := dsl.FindBlock(ab.Items, "acceptance"); blk != nil {
		return parseCriteriaFromBlock(blk)
	}
	return nil
}

func parseCriteriaFromBlock(blk *dsl.Block) []Criterion {
	var criteria []Criterion
	for _, item := range blk.Items {
		sub, ok := item.(*dsl.Block)
		if !ok || sub.Name != "criterion" {
			continue
		}
		c := Criterion{Name: sub.Title}
		c.Description, _ = dsl.FieldString(sub.Items, "description")
		c.VerifiedBy, _ = dsl.FieldString(sub.Items, "verified_by")
		criteria = append(criteria, c)
	}
	return criteria
}

// CriterionCoverage describes the coverage status of a single criterion.
type CriterionCoverage struct {
	Criterion   Criterion
	AddressedBy string // spec ID, empty if uncovered
	Verified    bool
}

// FormatCriteria renders acceptance criteria with optional coverage info.
func FormatCriteria(criteria []Criterion, coverage map[string]CriterionCoverage) string {
	if len(criteria) == 0 {
		return "(no acceptance criteria defined)\n"
	}

	var b strings.Builder
	b.WriteString("Acceptance Criteria:\n")
	for _, c := range criteria {
		fmt.Fprintf(&b, "  criterion %q:\n", c.Name)
		fmt.Fprintf(&b, "    description: %s\n", c.Description)
		fmt.Fprintf(&b, "    verified_by: %s\n", c.VerifiedBy)
		if coverage != nil {
			if cov, ok := coverage[c.Name]; ok && cov.AddressedBy != "" {
				status := "unverified"
				if cov.Verified {
					status = "verified"
				}
				fmt.Fprintf(&b, "    addressed_by: %s (%s)\n", cov.AddressedBy, status)
			} else {
				b.WriteString("    addressed_by: (uncovered)\n")
			}
		}
	}
	return b.String()
}
