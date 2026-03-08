package lsp

// ProjectContext holds project-level context for LSP operations.
// With the linter package removed, this is a minimal stub.
type ProjectContext struct {
	RuleIDs map[string]string // rule ID -> file path
	Lexicon *Lexicon          // optional lexicon for term definitions
}

// Lexicon holds term definitions for hover.
type Lexicon struct {
	Terms map[string]string
}
