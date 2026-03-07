package artifact

import (
	"os"
	"path/filepath"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// Predicate represents a single when-block: all fields must match (AND semantics).
type Predicate struct {
	Fields map[string]string
}

// HarnessDef describes the harness block on a policy rule.
type HarnessDef struct {
	Command string
	Timeout string
}

// Policy is a rule that has at least one when-block predicate.
// Rules without when-blocks are global rules, not policies.
type Policy struct {
	RuleID      string
	Name        string
	Type        string // mechanical | interpretive
	Enforcement string // error | warning | info
	Predicates  []Predicate
	Harness     *HarnessDef
	Path        string
}

// LoadPolicies walks .mos/rules/ and returns all rules that contain at least
// one when-block. Rules without when-blocks are excluded.
func LoadPolicies(root string) ([]Policy, error) {
	mosDir := filepath.Join(root, MosDir)
	var policies []Policy

	for _, ruleType := range []string{"mechanical", "interpretive"} {
		base := filepath.Join(mosDir, "rules", ruleType)
		entries, err := os.ReadDir(base)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".mos" {
				continue
			}
			rulePath := filepath.Join(base, entry.Name())
			p, ok, err := parsePolicy(rulePath)
			if err != nil || !ok {
				continue
			}
			policies = append(policies, p)
		}
	}
	return policies, nil
}

func parsePolicy(path string) (Policy, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, false, err
	}
	f, err := dsl.Parse(string(data), nil)
	if err != nil {
		return Policy{}, false, err
	}
	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return Policy{}, false, nil
	}

	var predicates []Predicate
	var harness *HarnessDef

	dsl.WalkBlocks(ab.Items, func(blk *dsl.Block) bool {
		if blk.Name == "when" {
			pred := extractPredicate(blk)
			if len(pred.Fields) > 0 {
				predicates = append(predicates, pred)
			}
			return false // don't descend into when
		}
		if blk.Name == "harness" {
			harness = extractHarness(blk)
			return false
		}
		return true // descend for nested blocks
	})

	if len(predicates) == 0 {
		return Policy{}, false, nil
	}

	name, _ := dsl.FieldString(ab.Items, "name")
	ruleType, _ := dsl.FieldString(ab.Items, "type")
	enforcement, _ := dsl.FieldString(ab.Items, "enforcement")

	return Policy{
		RuleID:      ab.Name,
		Name:        name,
		Type:        ruleType,
		Enforcement: enforcement,
		Predicates:  predicates,
		Harness:     harness,
		Path:        path,
	}, true, nil
}

func extractPredicate(blk *dsl.Block) Predicate {
	// Build map from block items; dsl has no generic "items to map" helper
	fields := make(map[string]string)
	for _, item := range blk.Items {
		field, ok := item.(*dsl.Field)
		if !ok {
			continue
		}
		if sv, ok := field.Value.(*dsl.StringVal); ok {
			fields[field.Key] = sv.Text
		}
	}
	return Predicate{Fields: fields}
}

func extractHarness(blk *dsl.Block) *HarnessDef {
	h := &HarnessDef{}
	h.Command, _ = dsl.FieldString(blk.Items, "command")
	h.Timeout, _ = dsl.FieldString(blk.Items, "timeout")
	return h
}

// MatchesArtifact returns true if any of the policy's predicates matches
// the given artifact fields. Each predicate uses AND semantics (all fields
// must match); multiple predicates use OR semantics (any may match).
func MatchesArtifact(p Policy, fields map[string]string) bool {
	for _, pred := range p.Predicates {
		if matchesPredicate(pred, fields) {
			return true
		}
	}
	return false
}

func matchesPredicate(pred Predicate, fields map[string]string) bool {
	for k, v := range pred.Fields {
		if fields[k] != v {
			return false
		}
	}
	return true
}

// MatchingPolicies returns the subset of policies whose predicates match
// the given artifact fields.
func MatchingPolicies(policies []Policy, fields map[string]string) []Policy {
	var matched []Policy
	for _, p := range policies {
		if MatchesArtifact(p, fields) {
			matched = append(matched, p)
		}
	}
	return matched
}

// AddRuleWhen adds a when-block to a rule artifact with the given field predicates.
func AddRuleWhen(root, ruleID string, fields map[string]string) error {
	path, err := findRulePath(root, ruleID)
	if err != nil {
		return err
	}

	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		var whenItems []dsl.Node
		keys := sortedKeys(fields)
		for _, k := range keys {
			whenItems = append(whenItems, &dsl.Field{
				Key:   k,
				Value: &dsl.StringVal{Text: fields[k]},
			})
		}
		ab.Items = append(ab.Items, &dsl.Block{
			Name:  "when",
			Items: whenItems,
		})
		return nil
	})
}
