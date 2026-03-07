package arch

import (
	"fmt"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/registry"
)

// SpecGenOpts controls spec generation from an architecture artifact.
type SpecGenOpts struct {
	State           string   // "current" or "desired"
	StatusOverride  string   // override initial status (empty = auto-detect)
	ExcludePatterns []string // directory patterns to skip
}

// SpecGenResult describes a single generated specification.
type SpecGenResult struct {
	ID    string
	Title string
	Pkg   string
	Group string
}

// GenerateSpecsFn is the type for the GenericCreate function injected from artifact.
type GenericCreateFn func(root string, td registry.ArtifactTypeDef, id string, fields map[string]string) (string, error)

// FindGenericArtifactPathFn locates an artifact's filesystem path.
type FindGenericArtifactPathFn func(root string, td registry.ArtifactTypeDef, id string) (string, error)

// GenerateSpecs reads ARCH-auto and creates a specification for each component
// that does not already have a corresponding spec. Returns the list of created specs.
func GenerateSpecs(root string, opts SpecGenOpts, create GenericCreateFn, findPath FindGenericArtifactPathFn) ([]SpecGenResult, error) {
	if ReadArchitectureFn == nil {
		return nil, fmt.Errorf("ReadArchitectureFn not initialized")
	}
	ab, err := ReadArchitectureFn(root, "ARCH-auto")
	if err != nil {
		return nil, fmt.Errorf("read architecture: %w (run 'mos architecture sync' first)", err)
	}

	archModel := ParseArchModel(ab)
	existing := FindExistingSpecIncludes(root)

	reg, err := registry.LoadRegistry(root)
	if err != nil {
		return nil, fmt.Errorf("load registry: %w", err)
	}
	specTD, ok := reg.Types["specification"]
	if !ok {
		return nil, fmt.Errorf("specification type not found in registry")
	}

	var results []SpecGenResult
	for _, svc := range archModel.Services {
		pkg := svc.Package
		if pkg == "" {
			pkg = svc.Name
		}

		if ShouldExclude(pkg, opts.ExcludePatterns) {
			continue
		}
		if existing[pkg] {
			continue
		}

		title := svc.Name
		if title == "" {
			title = pkg
		}

		initialStatus := ResolveInitialStatus(specTD, opts.StatusOverride)
		fields := map[string]string{
			"title":  title,
			"status": initialStatus,
		}
		if opts.State != "" {
			fields["state"] = opts.State
		}
		group := DeriveGroup(pkg)
		if group != "" {
			fields["group"] = group
		}

		newID, err := registry.NextIDForType(root, specTD.Prefix, specTD.Directory)
		if err != nil {
			continue
		}
		id, err := create(root, specTD, newID, fields)
		if err != nil {
			continue
		}

		specPath, err := findPath(root, specTD, id)
		if err == nil {
			AddSpecInclude(specPath, pkg)
			AddSpecSection(specPath, "Survey", BuildSurveySummary(svc))
		}

		results = append(results, SpecGenResult{ID: id, Title: title, Pkg: pkg, Group: group})
	}

	return results, nil
}

// ResolveInitialStatus determines the initial status for generated specs.
func ResolveInitialStatus(specTD registry.ArtifactTypeDef, override string) string {
	if override != "" {
		return override
	}
	for _, fs := range specTD.Fields {
		if fs.Name == "status" && len(fs.Enum) > 0 {
			for _, e := range fs.Enum {
				if e == "candidate" {
					return "candidate"
				}
			}
			return fs.Enum[0]
		}
	}
	return "candidate"
}

// DeriveGroup extracts the top-level directory from a package path.
func DeriveGroup(pkgPath string) string {
	pkgPath = strings.TrimPrefix(pkgPath, "./")
	if pkgPath == "" || pkgPath == "." {
		return "root"
	}
	parts := strings.SplitN(pkgPath, "/", 2)
	return parts[0]
}

// ShouldExclude checks if a package path matches any exclusion pattern.
func ShouldExclude(pkgPath string, patterns []string) bool {
	for _, pat := range patterns {
		top := DeriveGroup(pkgPath)
		if top == pat || strings.HasPrefix(pkgPath, pat+"/") || strings.HasPrefix(pkgPath, "./"+pat+"/") {
			return true
		}
	}
	return false
}

// ReadSpecIncludesFn reads specification includes. Injected to avoid import cycles.
var ReadSpecIncludesFn func(root string) map[string]bool

// FindExistingSpecIncludes scans all specifications for include directives.
func FindExistingSpecIncludes(root string) map[string]bool {
	if ReadSpecIncludesFn != nil {
		return ReadSpecIncludesFn(root)
	}
	return make(map[string]bool)
}

// AddSpecInclude appends an include directive to a spec artifact.
func AddSpecInclude(specPath, pkgPath string) {
	_ = dsl.WithArtifact(specPath, func(ab *dsl.ArtifactBlock) error {
		var specBlock *dsl.SpecBlock
		for _, item := range ab.Items {
			if sb, ok := item.(*dsl.SpecBlock); ok {
				specBlock = sb
				break
			}
		}
		if specBlock == nil {
			specBlock = &dsl.SpecBlock{}
			ab.Items = append(ab.Items, specBlock)
		}
		specBlock.Includes = append(specBlock.Includes, &dsl.IncludeDirective{Path: pkgPath})
		return nil
	})
}

// AddSpecSection appends a named section block to a spec artifact.
func AddSpecSection(specPath, name, text string) {
	_ = dsl.WithArtifact(specPath, func(ab *dsl.ArtifactBlock) error {
		section := &dsl.Block{
			Name:  "section",
			Title: name,
			Items: []dsl.Node{
				&dsl.Field{Key: "text", Value: &dsl.StringVal{Text: text}},
			},
		}
		ab.Items = append(ab.Items, section)
		return nil
	})
}

// BuildSurveySummary produces a textual summary for a service component.
func BuildSurveySummary(svc ArchService) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Package: %s", svc.Package))
	if len(svc.Symbols) > 0 {
		parts = append(parts, fmt.Sprintf("Exports: %s", strings.Join(svc.Symbols, ", ")))
	}
	if len(svc.Exposes) > 0 {
		parts = append(parts, fmt.Sprintf("Exposes: %s", strings.Join(svc.Exposes, ", ")))
	}
	return strings.Join(parts, "\n")
}
