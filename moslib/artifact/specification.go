package artifact

import (
	"fmt"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// SpecOpts configures the mos spec create command.
type SpecOpts struct {
	Title       string
	Enforcement string // disabled | warn | enforced
	Symbol      string // implementation binding (e.g. "pkg/api.Handler")
	Harness     string // test binding (e.g. "tests/api_test.go:TestBackwardCompat")
	Project     string // project name for auto-ID
	Satisfies   string // need IDs this spec satisfies (spec→need link)
	Addresses   string
}

// SpecInfo holds metadata about a discovered specification.
type SpecInfo struct {
	ID          string
	Title       string
	Status      string
	Enforcement string
	Symbol      string
	Harness     string
	Path        string
}

// SpecUpdateOpts configures partial updates to a specification.
type SpecUpdateOpts struct {
	Title       *string
	Enforcement *string
	Symbol      *string
	Harness     *string
	Satisfies   *string
	Addresses   *string
}

// ValidEnforcements is the set of valid specification enforcement levels.
var ValidEnforcements = map[string]bool{"disabled": true, "warn": true, "enforced": true}

// CreateSpec creates a specification artifact under .mos/specifications/.
func CreateSpec(root, id string, opts SpecOpts) (string, error) {
	if id == "" && opts.Project != "" {
		generated, err := NextID(root, opts.Project)
		if err != nil {
			return "", fmt.Errorf("auto-ID from project %q: %w", opts.Project, err)
		}
		id = generated
	} else if id == "" {
		projects, err := LoadProjects(root)
		if err == nil {
			for _, p := range projects {
				if p.Name == "specifications" {
					generated, err := NextID(root, p.Name)
					if err != nil {
						return "", fmt.Errorf("auto-ID from specifications project: %w", err)
					}
					id = generated
					break
				}
			}
		}
		if id == "" {
			return "", fmt.Errorf("specification id is required (no specifications project configured)")
		}
	}

	if opts.Enforcement == "" {
		opts.Enforcement = "disabled"
	}
	if !ValidEnforcements[opts.Enforcement] {
		return "", fmt.Errorf("enforcement must be disabled, warn, or enforced; got %q", opts.Enforcement)
	}

	reg, err := LoadRegistry(root)
	if err != nil {
		return "", fmt.Errorf("loading registry: %w", err)
	}
	td := reg.Types["specification"]

	fields := map[string]string{
		"title":       opts.Title,
		"enforcement": opts.Enforcement,
		"status":      StatusActive,
	}
	if opts.Symbol != "" {
		fields["symbol"] = opts.Symbol
	}
	if opts.Harness != "" {
		fields["harness"] = opts.Harness
	}
	if opts.Satisfies != "" {
		fields["satisfies"] = opts.Satisfies
	}
	if opts.Addresses != "" {
		fields["addresses"] = opts.Addresses
	}

	return GenericCreate(root, td, id, fields)
}

// ListSpecs returns all specifications.
func ListSpecs(root string, enforcementFilter string) ([]SpecInfo, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, fmt.Errorf("loading registry: %w", err)
	}
	td := reg.Types["specification"]

	items, err := GenericList(root, td, "")
	if err != nil {
		return nil, err
	}

	var specs []SpecInfo
	for _, item := range items {
		info := readSpecInfo(root, td, item.ID, item.Path)
		if enforcementFilter != "" && info.Enforcement != enforcementFilter {
			continue
		}
		specs = append(specs, info)
	}
	return specs, nil
}

// ShowSpec returns a formatted display of a specification with traceability info.
func ShowSpec(root, id string) (string, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return "", fmt.Errorf("loading registry: %w", err)
	}
	td := reg.Types["specification"]

	info := readSpecInfoByID(root, td, id)
	if info.ID == "" {
		return "", fmt.Errorf("specification %q not found", id)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s: %s\n", info.ID, info.Title)
	fmt.Fprintf(&b, "Status: %s\n", info.Status)
	fmt.Fprintf(&b, "Enforcement: %s\n", info.Enforcement)
	if info.Symbol != "" {
		fmt.Fprintf(&b, "Symbol: %s\n", info.Symbol)
	} else {
		fmt.Fprintf(&b, "Symbol: (unbound)\n")
	}
	if info.Harness != "" {
		fmt.Fprintf(&b, "Harness: %s\n", info.Harness)
	} else {
		fmt.Fprintf(&b, "Harness: (unbound)\n")
	}

	traceComplete := info.Symbol != "" && info.Harness != ""
	if traceComplete {
		fmt.Fprintf(&b, "Traceability: complete\n")
	} else {
		fmt.Fprintf(&b, "Traceability: incomplete\n")
	}

	return b.String(), nil
}

// UpdateSpec applies partial updates to a specification.
func UpdateSpec(root, id string, opts SpecUpdateOpts) error {
	reg, err := LoadRegistry(root)
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}
	td := reg.Types["specification"]

	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return err
	}

	if opts.Enforcement != nil {
		if !ValidEnforcements[*opts.Enforcement] {
			return fmt.Errorf("enforcement must be disabled, warn, or enforced; got %q", *opts.Enforcement)
		}
	}

	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		if opts.Title != nil {
			dsl.SetField(&ab.Items, "title", &dsl.StringVal{Text: *opts.Title})
		}
		if opts.Enforcement != nil {
			dsl.SetField(&ab.Items, "enforcement", &dsl.StringVal{Text: *opts.Enforcement})
		}
		if opts.Symbol != nil {
			dsl.SetField(&ab.Items, "symbol", &dsl.StringVal{Text: *opts.Symbol})
		}
		if opts.Harness != nil {
			dsl.SetField(&ab.Items, "harness", &dsl.StringVal{Text: *opts.Harness})
		}
		if opts.Satisfies != nil {
			dsl.SetField(&ab.Items, "satisfies", &dsl.StringVal{Text: *opts.Satisfies})
		}
		if opts.Addresses != nil {
			dsl.SetField(&ab.Items, "addresses", &dsl.StringVal{Text: *opts.Addresses})
		}
		return nil
	})
}

// FindSpecPath returns the file path for a specification.
func FindSpecPath(root, id string) (string, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return "", fmt.Errorf("loading registry: %w", err)
	}
	td := reg.Types["specification"]
	return FindGenericPath(root, td, id)
}

func readSpecInfo(root string, td ArtifactTypeDef, id, path string) SpecInfo {
	info := SpecInfo{ID: id, Path: path}

	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return info
	}

	info.Title, _ = dsl.FieldString(ab.Items, "title")
	info.Status, _ = dsl.FieldString(ab.Items, "status")
	info.Enforcement, _ = dsl.FieldString(ab.Items, "enforcement")
	info.Symbol, _ = dsl.FieldString(ab.Items, "symbol")
	info.Harness, _ = dsl.FieldString(ab.Items, "harness")
	return info
}

func readSpecInfoByID(root string, td ArtifactTypeDef, id string) SpecInfo {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return SpecInfo{}
	}
	return readSpecInfo(root, td, id, path)
}
