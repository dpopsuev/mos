package artifact

import (
	"fmt"
	"os"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// ContextResult holds the resolved context for a contract: its metadata
// plus all rules whose applies_to list includes the contract's kind.
type ContextResult struct {
	ID    string
	Title string
	Kind  string
	Goal  string
	Status string
	Rules  []ContextRule
}

// ContextRule holds a rule that matched the contract's kind.
type ContextRule struct {
	ID      string
	Name    string
	Content string
}

// ContractContext resolves a contract's metadata and all rules whose
// applies_to field includes the contract's kind.
func ContractContext(root, id string) (*ContextResult, error) {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return nil, fmt.Errorf("ContractContext: %w", err)
	}
	info, err := readContractInfo(id, contractPath)
	if err != nil {
		return nil, fmt.Errorf("reading contract: %w", err)
	}

	goal := readContractGoal(contractPath)

	result := &ContextResult{
		ID:     info.ID,
		Title:  info.Title,
		Kind:   info.Kind,
		Goal:   goal,
		Status: info.Status,
	}

	if result.Kind == "" {
		return result, nil
	}

	rules, err := ListRules(root, "")
	if err != nil {
		return result, nil
	}

	for _, r := range rules {
		for _, kind := range r.AppliesTo {
			if kind == result.Kind {
				content, _ := readRuleContent(r.Path)
				result.Rules = append(result.Rules, ContextRule{
					ID:      r.ID,
					Name:    r.Name,
					Content: content,
				})
				break
			}
		}
	}

	return result, nil
}

func readContractGoal(path string) string {
	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return ""
	}
	s, _ := dsl.FieldString(ab.Items, "goal")
	return s
}

func readRuleContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatContext returns a human-readable context output for agent consumption.
func FormatContext(ctx *ContextResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Contract: %s\n", ctx.ID)
	fmt.Fprintf(&b, "Title: %s\n", ctx.Title)
	fmt.Fprintf(&b, "Kind: %s\n", ctx.Kind)
	fmt.Fprintf(&b, "Status: %s\n", ctx.Status)
	if ctx.Goal != "" {
		fmt.Fprintf(&b, "Goal: %s\n", ctx.Goal)
	}

	if len(ctx.Rules) > 0 {
		fmt.Fprintf(&b, "\n# Applicable Rules (%d)\n", len(ctx.Rules))
		for _, r := range ctx.Rules {
			fmt.Fprintf(&b, "\n## Rule: %s (%s)\n", r.ID, r.Name)
			fmt.Fprintf(&b, "%s\n", r.Content)
		}
	} else {
		fmt.Fprintf(&b, "\nNo applicable rules for kind %q.\n", ctx.Kind)
	}

	return b.String()
}
