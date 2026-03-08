package lsp

import (
	"path/filepath"
	"strings"
)

// Location represents an LSP Location (file + range).
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Range is a zero-based line/character range.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Position is a zero-based line/character position.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Definition returns go-to-definition results for the given position.
func Definition(path, content string, line, character int, ctx *ProjectContext) *Location {
	if ctx == nil {
		return nil
	}

	lines := strings.Split(content, "\n")
	if line >= len(lines) {
		return nil
	}

	lineText := lines[line]
	word := extractWordAt(lineText, character)
	if word == "" {
		return nil
	}

	if rulePath, ok := ctx.RuleIDs[word]; ok {
		return &Location{
			URI:   PathToURI(rulePath),
			Range: Range{Start: Position{0, 0}, End: Position{0, 0}},
		}
	}

	ref := extractDSLInclude(lineText)
	if ref != "" {
		target := filepath.Join(filepath.Dir(path), ref)
		return &Location{
			URI:   PathToURI(target),
			Range: Range{Start: Position{0, 0}, End: Position{0, 0}},
		}
	}

	return nil
}

func extractDSLInclude(line string) string {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "include ") {
		return ""
	}
	ref := strings.TrimSpace(strings.TrimPrefix(trimmed, "include "))
	ref = strings.Trim(ref, `"`)
	return ref
}
