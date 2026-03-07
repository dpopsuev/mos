package linter

import (
	"fmt"

	"github.com/dpopsuev/mos/moslib/dsl"
)

type lintPolicy struct {
	ruleID      string
	name        string
	enforcement Severity
	predicates  []map[string]string // each map is AND; multiple maps are OR
}

func validatePolicies(ctx *ProjectContext) []Diagnostic {
	policies := loadLintPolicies(ctx)
	if len(policies) == 0 {
		return nil
	}

	var diags []Diagnostic

	allArtifacts := make(map[string]map[string]string)
	for kind, ids := range ctx.ArtifactIDs {
		allArtifacts[kind] = ids
	}
	for id, path := range ctx.ContractIDs {
		if allArtifacts["contract"] == nil {
			allArtifacts["contract"] = make(map[string]string)
		}
		allArtifacts["contract"][id] = path
	}

	for kind, ids := range allArtifacts {
		for id, path := range ids {
			f, err := parseDSLFile(path, ctx.Keywords)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}

			fields := artifactFieldMap(kind, id, ab)

			for _, pol := range policies {
				if !policyMatches(pol, fields) {
					continue
				}
				diags = append(diags, Diagnostic{
					File:     path,
					Severity: pol.enforcement,
					Rule:     "policy-match",
					Message:  fmt.Sprintf("policy %q (%s) applies to %s %q", pol.ruleID, pol.name, kind, id),
				})
			}
		}
	}

	return diags
}

func artifactFieldMap(kind, id string, ab *dsl.ArtifactBlock) map[string]string {
	fields := map[string]string{
		"artifact_kind": kind,
		"id":            id,
	}
	for _, item := range ab.Items {
		field, ok := item.(*dsl.Field)
		if !ok {
			continue
		}
		if sv, ok := field.Value.(*dsl.StringVal); ok {
			fields[field.Key] = sv.Text
		}
	}
	if _, hasKind := fields["kind"]; !hasKind {
		fields["kind"] = kind
	}
	return fields
}

func loadLintPolicies(ctx *ProjectContext) []lintPolicy {
	var policies []lintPolicy
	for ruleID, path := range ctx.RuleIDs {
		f, err := parseDSLFile(path, ctx.Keywords)
		if err != nil {
			continue
		}
		ab, ok := f.Artifact.(*dsl.ArtifactBlock)
		if !ok {
			continue
		}

		var predicates []map[string]string
		for _, item := range ab.Items {
			blk, ok := item.(*dsl.Block)
			if !ok || blk.Name != "when" {
				continue
			}
			pred := make(map[string]string)
			for _, bi := range blk.Items {
				field, ok := bi.(*dsl.Field)
				if !ok {
					continue
				}
				if sv, ok := field.Value.(*dsl.StringVal); ok {
					pred[field.Key] = sv.Text
				}
			}
			if len(pred) > 0 {
				predicates = append(predicates, pred)
			}
		}

		if len(predicates) == 0 {
			continue
		}

		name, _ := astFieldString(ab.Items, "name")
		enforcement := enforcementToSeverity(ab.Items)

		policies = append(policies, lintPolicy{
			ruleID:      ruleID,
			name:        name,
			enforcement: enforcement,
			predicates:  predicates,
		})
	}
	return policies
}

func enforcementToSeverity(items []dsl.Node) Severity {
	enf, _ := astFieldString(items, "enforcement")
	switch enf {
	case "error":
		return SeverityError
	case "warning":
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

func policyMatches(pol lintPolicy, fields map[string]string) bool {
	for _, pred := range pol.predicates {
		if predicateMatches(pred, fields) {
			return true
		}
	}
	return false
}

func predicateMatches(pred map[string]string, fields map[string]string) bool {
	for k, v := range pred {
		if fields[k] != v {
			return false
		}
	}
	return true
}
