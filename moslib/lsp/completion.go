package lsp

import (
	"strings"

	"github.com/dpopsuev/mos/moslib/names"
)

// CompletionItem represents a single completion suggestion.
type CompletionItem struct {
	Label  string `json:"label"`
	Kind   int    `json:"kind"`
	Detail string `json:"detail,omitempty"`
}

const (
	CompletionKindField   = 5
	CompletionKindKeyword = 14
	CompletionKindBlock   = 15
)

var configBlocks = []CompletionItem{
	{Label: "mos", Kind: CompletionKindBlock, Detail: "Mos version and identity"},
	{Label: "backend", Kind: CompletionKindBlock, Detail: "Storage backend configuration"},
	{Label: "governance", Kind: CompletionKindBlock, Detail: "Governance model configuration"},
	{Label: "identity", Kind: CompletionKindBlock, Detail: "Identity configuration"},
	{Label: "sync", Kind: CompletionKindBlock, Detail: "Sync tick configuration"},
	{Label: "lexicon", Kind: CompletionKindBlock, Detail: "Lexicon base reference"},
}

var configFields = map[string][]CompletionItem{
	"mos": {
		{Label: "version", Kind: CompletionKindField, Detail: "Schema version (integer, required)"},
	},
	"backend": {
		{Label: "type", Kind: CompletionKindField, Detail: "Backend type: \"git\" or \"native\""},
	},
	"governance": {
		{Label: "model", Kind: CompletionKindField, Detail: "Governance model: bdfl|committee|consensus|custom"},
		{Label: "ratification_authority", Kind: CompletionKindField, Detail: "List of signing key fingerprints"},
	},
}

var ruleBlocks = []CompletionItem{
	{Label: "harness", Kind: CompletionKindBlock, Detail: "Mechanical harness command"},
	{Label: "identity", Kind: CompletionKindBlock, Detail: "Creator and amendment tracking"},
	{Label: "feature", Kind: CompletionKindKeyword, Detail: "Behavioral specification (Gherkin)"},
	{Label: "spec", Kind: CompletionKindBlock, Detail: "Multi-file specification with includes"},
}

var ruleFields = map[string][]CompletionItem{
	"": {
		{Label: "name", Kind: CompletionKindField, Detail: "Human-readable rule name (required)"},
		{Label: "type", Kind: CompletionKindField, Detail: "\"mechanical\" or \"interpretive\" (required)"},
		{Label: "scope", Kind: CompletionKindField, Detail: "Resolution layer scope (required)"},
		{Label: "enforcement", Kind: CompletionKindField, Detail: "\"error\" or \"warning\" (required)"},
		{Label: "tags", Kind: CompletionKindField, Detail: "Classification tags (list of strings)"},
	},
}

var contractBlocks = []CompletionItem{
	{Label: "bill", Kind: CompletionKindBlock, Detail: "Bill introduction metadata"},
	{Label: "feature", Kind: CompletionKindKeyword, Detail: "Behavioral specification (Gherkin)"},
	{Label: "spec", Kind: CompletionKindBlock, Detail: "Multi-file specification with includes"},
	{Label: "execution", Kind: CompletionKindBlock, Detail: "Execution rules and overrides"},
	{Label: "coverage", Kind: CompletionKindBlock, Detail: "Test coverage requirements"},
	{Label: "tasks", Kind: CompletionKindBlock, Detail: "Work items"},
	{Label: "history", Kind: CompletionKindBlock, Detail: "Lifecycle event log"},
	{Label: "evidence", Kind: CompletionKindBlock, Detail: "Lab results and drift assessment"},
	{Label: "security", Kind: CompletionKindBlock, Detail: "OWASP spot-check"},
}

var contractFields = map[string][]CompletionItem{
	"": {
		{Label: "title", Kind: CompletionKindField, Detail: "Contract title (required)"},
		{Label: "status", Kind: CompletionKindField, Detail: "draft|active|complete|abandoned (required)"},
		{Label: "serves", Kind: CompletionKindField, Detail: "Parent goal ID"},
	},
	"bill": {
		{Label: "introduced_by", Kind: CompletionKindField, Detail: "Author identity (required)"},
		{Label: "introduced_at", Kind: CompletionKindField, Detail: "Introduction date (required)"},
		{Label: "intent", Kind: CompletionKindField, Detail: "Bill intent statement (required)"},
	},
	"execution": {
		{Label: "rules_override", Kind: CompletionKindField, Detail: "Rule IDs to override during contract"},
		{Label: "rules_suspended", Kind: CompletionKindField, Detail: "Rule IDs suspended during contract"},
	},
}

var declarationFields = map[string][]CompletionItem{
	"": {
		{Label: "name", Kind: CompletionKindField, Detail: "Project name (required)"},
		{Label: "created", Kind: CompletionKindField, Detail: "Creation date (required)"},
		{Label: "authors", Kind: CompletionKindField, Detail: "Founding authors (list of strings)"},
	},
}

var declarationBlocks = []CompletionItem{
	{Label: "sensation", Kind: CompletionKindBlock, Detail: "Business need description"},
	{Label: "contextualization", Kind: CompletionKindBlock, Detail: "Domain boundary"},
	{Label: "principles", Kind: CompletionKindBlock, Detail: "Founding convictions"},
	{Label: "lexicon", Kind: CompletionKindBlock, Detail: "Initial terms"},
}

var dslKeywords = []CompletionItem{
	{Label: "feature", Kind: CompletionKindKeyword, Detail: "Feature block (specification)"},
	{Label: "scenario", Kind: CompletionKindKeyword, Detail: "Scenario block (test case)"},
	{Label: "group", Kind: CompletionKindKeyword, Detail: "Group block (scenario grouping)"},
	{Label: "background", Kind: CompletionKindKeyword, Detail: "Background block (shared preconditions)"},
	{Label: "given", Kind: CompletionKindKeyword, Detail: "Precondition step block"},
	{Label: "when", Kind: CompletionKindKeyword, Detail: "Action step block"},
	{Label: "then", Kind: CompletionKindKeyword, Detail: "Assertion step block"},
	{Label: "include", Kind: CompletionKindKeyword, Detail: "Include external file"},
}

// ArtifactKind identifies the type of .mos/ artifact from its file path.
type ArtifactKind int

const (
	ArtifactUnknown ArtifactKind = iota
	ArtifactConfig
	ArtifactDeclaration
	ArtifactRule
	ArtifactContract
)

// DetectArtifactKind determines the artifact type from a file path.
func DetectArtifactKind(path string) ArtifactKind {
	if strings.HasSuffix(path, names.ConfigFile) && strings.Contains(path, names.MosDir) {
		return ArtifactConfig
	}
	if strings.HasSuffix(path, "declaration.mos") && strings.Contains(path, names.MosDir) {
		return ArtifactDeclaration
	}
	if strings.Contains(path, "/rules/") && strings.HasSuffix(path, names.MosDir) {
		return ArtifactRule
	}
	if strings.Contains(path, "/contracts/") && strings.HasSuffix(path, "contract.mos") {
		return ArtifactContract
	}
	return ArtifactUnknown
}

// Complete returns completion items for the given document content and cursor position.
func Complete(path, content string, line, character int) []CompletionItem {
	kind := DetectArtifactKind(path)
	currentBlock := detectCurrentBlock(content, line)

	if isInsideStepBlock(content, line) {
		return dslKeywords
	}

	switch kind {
	case ArtifactConfig:
		return completeForArtifact(configBlocks, configFields, currentBlock)
	case ArtifactRule:
		return completeForArtifact(ruleBlocks, ruleFields, currentBlock)
	case ArtifactContract:
		return completeForArtifact(contractBlocks, contractFields, currentBlock)
	case ArtifactDeclaration:
		return completeForArtifact(declarationBlocks, declarationFields, currentBlock)
	}

	return nil
}

func completeForArtifact(blocks []CompletionItem, fields map[string][]CompletionItem, currentBlock string) []CompletionItem {
	if items, ok := fields[currentBlock]; ok {
		return items
	}
	return blocks
}

func detectCurrentBlock(content string, targetLine int) string {
	lines := strings.Split(content, "\n")
	depth := 0
	for i := targetLine; i >= 0; i-- {
		if i >= len(lines) {
			continue
		}
		line := strings.TrimSpace(lines[i])
		depth += strings.Count(line, "}")
		depth -= strings.Count(line, "{")
		if depth < 0 {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				name := strings.Trim(parts[0], `"`)
				return name
			}
		}
	}
	return ""
}

func isInsideStepBlock(content string, targetLine int) bool {
	lines := strings.Split(content, "\n")
	depth := 0
	for i := targetLine; i >= 0; i-- {
		if i >= len(lines) {
			continue
		}
		line := strings.TrimSpace(lines[i])
		depth += strings.Count(line, "}")
		depth -= strings.Count(line, "{")
		if depth < 0 {
			for _, kw := range []string{"feature", "scenario", "group"} {
				if strings.HasPrefix(line, kw+" ") || strings.HasPrefix(line, kw+"{") || line == kw {
					return true
				}
			}
			return false
		}
	}
	return false
}
