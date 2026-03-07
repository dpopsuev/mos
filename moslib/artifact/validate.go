package artifact

import "github.com/dpopsuev/mos/moslib/model"

// ValidateArtifactFunc validates an artifact file against the project context.
// Returns non-nil error if validation finds severity-error diagnostics.
type ValidateArtifactFunc func(path, mosDir string) error

// LintDiagnostic is a simplified diagnostic for governance audit use.
type LintDiagnostic struct {
	File            string
	Line            int
	Severity        string // "error", "warning", "info"
	Message         string
	Rule            string
	ArtifactID      string
	SuggestedAction string
}

// LintFunc runs full lint on a project root, returning diagnostics.
type LintFunc func(root string) ([]LintDiagnostic, error)

// LexiconLoaderFunc loads merged lexicon terms from a .mos directory.
type LexiconLoaderFunc func(mosDir string) (map[string]string, error)

// ScanProjectFunc scans source code and returns the project model
// with dependency graph. Used by audit to check forbidden edges.
type ScanProjectFunc func(root string) (*model.Project, error)

// Package-level validation hooks. Must be set by the binary entry point
// (cmd/mos) before calling governance functions that require validation.
// Tests should set these in TestMain or init().
var (
	ValidateContract ValidateArtifactFunc
	ValidateRule     ValidateArtifactFunc
	LintAll          LintFunc
	LoadLexicon      LexiconLoaderFunc
	ScanProject      ScanProjectFunc
)
