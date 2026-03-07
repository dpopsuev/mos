package artifact

import (
	"os"
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/dsl"
)

func TestCON037_NeedCADSchemaIncludesAcceptance(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["need"]

	hasAcceptance := false
	for _, f := range td.Fields {
		if f.Name == "acceptance" {
			hasAcceptance = true
		}
	}
	if !hasAcceptance {
		t.Error("need CAD should include acceptance in its fields schema")
	}
}

func TestCON037_CreateNeedWithAcceptanceCriteria(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["need"]

	path, err := GenericCreate(root, td, "NEED-AC-001", map[string]string{
		"title": "Faster CI", "sensation": "CI is slow", "status": "identified",
	})
	if err != nil {
		t.Fatalf("GenericCreate: %v", err)
	}

	data, _ := os.ReadFile(path)
	f, _ := dsl.Parse(string(data), nil)
	ab := f.Artifact.(*dsl.ArtifactBlock)
	ab.Items = append(ab.Items, &dsl.Block{
		Name: "acceptance",
		Items: []dsl.Node{
			&dsl.Block{Name: "criterion", Title: "sub-10-min", Items: []dsl.Node{
				&dsl.Field{Key: "description", Value: &dsl.StringVal{Text: "P95 pipeline duration under 10 minutes"}},
				&dsl.Field{Key: "verified_by", Value: &dsl.StringVal{Text: "harness"}},
			}},
			&dsl.Block{Name: "criterion", Title: "no-regression", Items: []dsl.Node{
				&dsl.Field{Key: "description", Value: &dsl.StringVal{Text: "Test coverage does not decrease"}},
				&dsl.Field{Key: "verified_by", Value: &dsl.StringVal{Text: "harness"}},
			}},
		},
	})
	writeArtifact(path, f)

	data2, _ := os.ReadFile(path)
	if !strings.Contains(string(data2), "acceptance") {
		t.Error("need file should contain acceptance block")
	}
	if !strings.Contains(string(data2), "sub-10-min") {
		t.Error("need file should contain sub-10-min criterion")
	}
}

func TestCON037_NeedWithoutCriteriaIsValid(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["need"]

	_, err := GenericCreate(root, td, "NEED-NOAC-001", map[string]string{
		"title": "Simple need", "sensation": "Something hurts", "status": "identified",
	})
	if err != nil {
		t.Fatalf("GenericCreate without acceptance should succeed: %v", err)
	}
}

func TestCON037_ParseAcceptanceCriteria(t *testing.T) {
	ab := &dsl.ArtifactBlock{
		Kind: "need",
		Name: "NEED-TEST",
		Items: []dsl.Node{
			&dsl.Field{Key: "title", Value: &dsl.StringVal{Text: "Test need"}},
			&dsl.Block{
				Name: "acceptance",
				Items: []dsl.Node{
					&dsl.Block{Name: "criterion", Title: "fast", Items: []dsl.Node{
						&dsl.Field{Key: "description", Value: &dsl.StringVal{Text: "Must be fast"}},
						&dsl.Field{Key: "verified_by", Value: &dsl.StringVal{Text: "harness"}},
					}},
					&dsl.Block{Name: "criterion", Title: "safe", Items: []dsl.Node{
						&dsl.Field{Key: "description", Value: &dsl.StringVal{Text: "Must be safe"}},
						&dsl.Field{Key: "verified_by", Value: &dsl.StringVal{Text: "human"}},
					}},
					&dsl.Block{Name: "criterion", Title: "smart", Items: []dsl.Node{
						&dsl.Field{Key: "description", Value: &dsl.StringVal{Text: "Must be smart"}},
						&dsl.Field{Key: "verified_by", Value: &dsl.StringVal{Text: "agent"}},
					}},
				},
			},
		},
	}

	criteria := ParseAcceptanceCriteria(ab)
	if len(criteria) != 3 {
		t.Fatalf("expected 3 criteria, got %d", len(criteria))
	}
	if criteria[0].Name != "fast" || criteria[0].VerifiedBy != "harness" {
		t.Errorf("criteria[0] = %+v, want fast/harness", criteria[0])
	}
	if criteria[1].Name != "safe" || criteria[1].VerifiedBy != "human" {
		t.Errorf("criteria[1] = %+v, want safe/human", criteria[1])
	}
	if criteria[2].Name != "smart" || criteria[2].VerifiedBy != "agent" {
		t.Errorf("criteria[2] = %+v, want smart/agent", criteria[2])
	}
}

func TestCON037_ParseAcceptanceCriteriaNone(t *testing.T) {
	ab := &dsl.ArtifactBlock{
		Kind: "need",
		Name: "NEED-EMPTY",
		Items: []dsl.Node{
			&dsl.Field{Key: "title", Value: &dsl.StringVal{Text: "No criteria"}},
		},
	}
	criteria := ParseAcceptanceCriteria(ab)
	if len(criteria) != 0 {
		t.Errorf("expected 0 criteria, got %d", len(criteria))
	}
}

func TestCON037_FormatCriteria(t *testing.T) {
	criteria := []Criterion{
		{Name: "fast", Description: "Must be fast", VerifiedBy: "harness"},
		{Name: "safe", Description: "Must be safe", VerifiedBy: "human"},
	}
	output := FormatCriteria(criteria, nil)
	if !strings.Contains(output, "fast") {
		t.Error("output should contain criterion name 'fast'")
	}
	if !strings.Contains(output, "harness") {
		t.Error("output should contain verified_by 'harness'")
	}
	if !strings.Contains(output, "Must be fast") {
		t.Error("output should contain description")
	}
}

func TestCON037_FormatCriteriaWithCoverage(t *testing.T) {
	criteria := []Criterion{
		{Name: "fast", Description: "Must be fast", VerifiedBy: "harness"},
		{Name: "safe", Description: "Must be safe", VerifiedBy: "human"},
	}
	coverage := map[string]CriterionCoverage{
		"fast": {Criterion: criteria[0], AddressedBy: "SPEC-001", Verified: true},
	}
	output := FormatCriteria(criteria, coverage)
	if !strings.Contains(output, "SPEC-001") {
		t.Error("output should show addressing spec")
	}
	if !strings.Contains(output, "verified") {
		t.Error("output should show verified status")
	}
	if !strings.Contains(output, "uncovered") {
		t.Error("output should show uncovered for 'safe'")
	}
}

func TestCON037_FormatCriteriaEmpty(t *testing.T) {
	output := FormatCriteria(nil, nil)
	if !strings.Contains(output, "no acceptance criteria") {
		t.Error("output should indicate no criteria")
	}
}

// --- CON-2026-038: Spec-to-Criterion Linkage ---
