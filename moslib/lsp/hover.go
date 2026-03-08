package lsp

import (
	"fmt"
	"strings"
)

// HoverResult represents the content of a hover response.
type HoverResult struct {
	Contents string `json:"contents"`
}

var fieldDocs = map[string]map[string]string{
	"mos": {
		"version": "Schema version for the .mos/ directory format. Integer, required. Current version: 1.",
	},
	"backend": {
		"type": "Storage backend type. Values: \"git\" (hybrid .git/.mos coexistence) or \"native\" (pure .mos storage).",
	},
	"governance": {
		"model":                  "Governance model. Values: bdfl (single authority), committee (group decision), consensus (all participants), custom.",
		"ratification_authority": "List of SSH key fingerprints authorized to ratify Bills into law.",
	},
	"rule": {
		"name":        "Human-readable name for this rule.",
		"type":        "Rule classification: \"mechanical\" (deterministic, machine-verifiable) or \"interpretive\" (requires human judgment).",
		"scope":       "Resolution layer this rule applies to. Must be a declared layer in resolution/layers.mos.",
		"enforcement": "How violations are reported: \"error\" (blocks) or \"warning\" (advisory).",
		"tags":        "Classification tags for this rule. Validated against lexicon if defined.",
	},
	"contract": {
		"title":  "Human-readable contract title.",
		"status": "Lifecycle state: draft, active, complete, or abandoned.",
	},
	"bill": {
		"introduced_by": "Identity (SSH key fingerprint or name) of the bill's author.",
		"introduced_at": "Date the bill was introduced (ISO 8601 datetime).",
		"intent":        "Statement of intent: what this bill aims to achieve.",
	},
	"declaration": {
		"name":    "Project name as declared in the founding document.",
		"created": "Date the mos was established (ISO 8601 datetime).",
		"authors": "List of founding authors.",
	},
	"execution": {
		"rules_override":  "List of rule IDs whose enforcement is modified during this contract.",
		"rules_suspended": "List of rule IDs suspended for the duration of this contract.",
	},
}

// Hover returns hover information for the given position.
func Hover(path, content string, line, character int, ctx *ProjectContext) *HoverResult {
	lines := strings.Split(content, "\n")
	if line >= len(lines) {
		return nil
	}

	lineText := lines[line]
	word := extractWordAt(lineText, character)
	if word == "" {
		return nil
	}

	block := detectCurrentBlock(content, line)

	if docs, ok := fieldDocs[block]; ok {
		if doc, ok := docs[word]; ok {
			return &HoverResult{Contents: doc}
		}
	}

	if ctx != nil && ctx.Lexicon != nil {
		if def, ok := ctx.Lexicon.Terms[strings.ToLower(word)]; ok {
			return &HoverResult{Contents: fmt.Sprintf("**%s** -- %s", word, def)}
		}
	}

	if ctx != nil {
		if rulePath, ok := ctx.RuleIDs[word]; ok {
			return &HoverResult{Contents: fmt.Sprintf("Rule **%s** defined in `%s`", word, rulePath)}
		}
	}

	return nil
}

func extractWordAt(line string, col int) string {
	if col >= len(line) {
		col = len(line) - 1
	}
	if col < 0 {
		return ""
	}

	start := col
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}
	end := col
	for end < len(line) && isWordChar(line[end]) {
		end++
	}

	if start == end {
		return ""
	}
	return strings.Trim(line[start:end], `"'`)
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.'
}
