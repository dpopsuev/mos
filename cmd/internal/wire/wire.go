package wire

import (
	"fmt"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/linter"
	"github.com/dpopsuev/mos/moslib/model"
	"github.com/dpopsuev/mos/moslib/registry"
	"github.com/dpopsuev/mos/moslib/schema"
	"github.com/dpopsuev/mos/moslib/survey"
)

// Init wires dependency-injection callbacks between moslib packages.
// Every binary (mos, mgov, mvcs, mgate, mtrace, mstore) must call this
// before executing any command.
func Init() {
	artifact.ValidateContract = func(path, mosDir string) error {
		ctx, err := linter.LoadContext(mosDir)
		if err != nil {
			return fmt.Errorf("loading context for validation: %w", err)
		}
		for _, d := range linter.ValidateContractFile(path, ctx) {
			if d.Severity == linter.SeverityError {
				return fmt.Errorf("validation failed: %s", d.Message)
			}
		}
		return nil
	}

	artifact.ValidateRule = func(path, mosDir string) error {
		ctx, err := linter.LoadContext(mosDir)
		if err != nil {
			return fmt.Errorf("loading context for validation: %w", err)
		}
		for _, d := range linter.ValidateRuleFile(path, ctx) {
			if d.Severity == linter.SeverityError {
				return fmt.Errorf("validation failed: %s", d.Message)
			}
		}
		return nil
	}

	artifact.LintAll = func(root string) ([]artifact.LintDiagnostic, error) {
		l := &linter.Linter{}
		diags, err := l.Lint(root)
		if err != nil {
			return nil, err
		}
		result := make([]artifact.LintDiagnostic, len(diags))
		for i, d := range diags {
			result[i] = artifact.LintDiagnostic{
				File:            d.File,
				Line:            d.Line,
				Severity:        d.Severity.String(),
				Message:         d.Message,
				Rule:            d.Rule,
				ArtifactID:      d.ArtifactID,
				SuggestedAction: d.SuggestedAction,
			}
		}
		return result, nil
	}

	artifact.ScanProject = func(root string) (*model.Project, error) {
		sc := &survey.AutoScanner{}
		return sc.Scan(root)
	}

	linter.LoadArtifactSchemas = func(root string) ([]schema.ArtifactSchema, error) {
		reg, err := registry.LoadRegistry(root)
		if err != nil {
			return nil, err
		}
		var schemas []schema.ArtifactSchema
		for _, td := range reg.Types {
			schemas = append(schemas, td.ToSchema())
		}
		return schemas, nil
	}

	artifact.LoadLexicon = func(mosDir string) (map[string]string, error) {
		ctx, err := linter.LoadContext(mosDir)
		if err != nil {
			return nil, err
		}
		if ctx.Lexicon == nil {
			return map[string]string{}, nil
		}
		return ctx.Lexicon.Terms, nil
	}
}
