package artifact

import (
	"fmt"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// BinderOpts configures the mos binder create command.
type BinderOpts struct {
	Title   string
	Project string // project name for auto-ID
}

// BinderInfo holds metadata about a discovered binder.
type BinderInfo struct {
	ID    string
	Title string
	Specs []string
	Path  string
}

// CreateBinder creates a binder artifact under .mos/binders/.
func CreateBinder(root, id string, opts BinderOpts) (string, error) {
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
				if p.Name == "binders" {
					generated, err := NextID(root, p.Name)
					if err != nil {
						return "", fmt.Errorf("auto-ID from binders project: %w", err)
					}
					id = generated
					break
				}
			}
		}
		if id == "" {
			return "", fmt.Errorf("binder id is required (no binders project configured)")
		}
	}

	reg, err := LoadRegistry(root)
	if err != nil {
		return "", fmt.Errorf("loading registry: %w", err)
	}
	td := reg.Types["binder"]

	fields := map[string]string{
		"title":  opts.Title,
		"status": "active",
	}

	return GenericCreate(root, td, id, fields)
}

// ListBinders returns all binders.
func ListBinders(root string) ([]BinderInfo, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, fmt.Errorf("loading registry: %w", err)
	}
	td := reg.Types["binder"]

	items, err := GenericList(root, td, "")
	if err != nil {
		return nil, err
	}

	var binders []BinderInfo
	for _, item := range items {
		info := readBinderInfo(item.Path, item.ID)
		binders = append(binders, info)
	}
	return binders, nil
}

// ShowBinder returns a formatted display of a binder with enforcement rollup.
func ShowBinder(root, id string) (string, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return "", fmt.Errorf("loading registry: %w", err)
	}
	td := reg.Types["binder"]
	specTD := reg.Types["specification"]

	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return "", err
	}

	info := readBinderInfo(path, id)

	var b strings.Builder
	fmt.Fprintf(&b, "%s: %s\n", info.ID, info.Title)
	fmt.Fprintf(&b, "Specifications: %d\n", len(info.Specs))

	if len(info.Specs) > 0 {
		counts := map[string]int{"disabled": 0, "warn": 0, "enforced": 0}
		for _, specID := range info.Specs {
			si := readSpecInfoByID(root, specTD, specID)
			if si.Enforcement != "" {
				counts[si.Enforcement]++
			}
			fmt.Fprintf(&b, "  %s: %s [%s]\n", specID, si.Title, si.Enforcement)
		}
		fmt.Fprintf(&b, "Enforcement: disabled=%d, warn=%d, enforced=%d\n",
			counts["disabled"], counts["warn"], counts["enforced"])
	}

	return b.String(), nil
}

// BinderBind adds a specification to a binder's specs list.
func BinderBind(root, binderID, specID string) error {
	reg, err := LoadRegistry(root)
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}
	td := reg.Types["binder"]
	specTD := reg.Types["specification"]

	if _, err := FindGenericPath(root, specTD, specID); err != nil {
		return fmt.Errorf("specification %q not found", specID)
	}

	path, err := FindGenericPath(root, td, binderID)
	if err != nil {
		return err
	}

	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		specs := dsl.FieldStringSlice(ab.Items, "specs")
		for _, v := range specs {
			if v == specID {
				return nil
			}
		}
		specs = append(specs, specID)
		items := make([]dsl.Value, 0, len(specs))
		for _, s := range specs {
			items = append(items, &dsl.StringVal{Text: s})
		}
		dsl.SetField(&ab.Items, "specs", &dsl.ListVal{Items: items})
		return nil
	})
}

// BinderUnbind removes a specification from a binder's specs list.
func BinderUnbind(root, binderID, specID string) error {
	reg, err := LoadRegistry(root)
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}
	td := reg.Types["binder"]

	path, err := FindGenericPath(root, td, binderID)
	if err != nil {
		return err
	}

	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		specs := dsl.FieldStringSlice(ab.Items, "specs")
		filtered := make([]string, 0, len(specs))
		for _, s := range specs {
			if s != specID {
				filtered = append(filtered, s)
			}
		}
		if len(filtered) == len(specs) {
			return nil
		}
		items := make([]dsl.Value, 0, len(filtered))
		for _, s := range filtered {
			items = append(items, &dsl.StringVal{Text: s})
		}
		dsl.SetField(&ab.Items, "specs", &dsl.ListVal{Items: items})
		return nil
	})
}

// BinderTrace returns a traceability report for all specifications in a binder.
func BinderTrace(root, id string) (string, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return "", fmt.Errorf("loading registry: %w", err)
	}
	td := reg.Types["binder"]
	specTD := reg.Types["specification"]

	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return "", err
	}

	info := readBinderInfo(path, id)

	var b strings.Builder
	fmt.Fprintf(&b, "Traceability Report: %s (%s)\n", info.ID, info.Title)
	fmt.Fprintf(&b, "%-16s %-20s %-20s %s\n", "Spec", "Symbol", "Harness", "Status")
	fmt.Fprintf(&b, "%s\n", strings.Repeat("-", 72))

	for _, specID := range info.Specs {
		si := readSpecInfoByID(root, specTD, specID)
		symbol := si.Symbol
		if symbol == "" {
			symbol = "(unbound)"
		}
		harn := si.Harness
		if harn == "" {
			harn = "(unbound)"
		}
		status := "complete"
		if si.Symbol == "" || si.Harness == "" {
			status = "INCOMPLETE"
		}
		fmt.Fprintf(&b, "%-16s %-20s %-20s %s\n", specID, symbol, harn, status)
	}

	return b.String(), nil
}

func readBinderInfo(path, id string) BinderInfo {
	info := BinderInfo{ID: id, Path: path}

	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return info
	}

	info.Title, _ = dsl.FieldString(ab.Items, "title")
	info.Specs = dsl.FieldStringSlice(ab.Items, "specs")

	return info
}
