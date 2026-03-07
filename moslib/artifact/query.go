package artifact

import (
	"cmp"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// QueryOpts defines the filters for querying artifacts.
type QueryOpts struct {
	Kind       string // filter by artifact kind ("contract", "need", etc.), empty = all
	Status     string
	Labels     []string
	References string // reverse lookup: find artifacts referencing this ID
	Count      bool
	GroupBy    string // "status", "kind", "sprint"
	Format     string // "text" or "json"
	Rich       bool   // return full artifact content in JSON output
	Verbose    bool   // show key metadata fields in text output
	Justifies  string // filter: justifies field matches this ID
	SprintEq   string // filter: sprint field matches this value
	DependsOn  string // filter: depends_on field matches this ID
	Priority   string // filter: priority field matches
	Unlinked   bool   // filter: no sprint assignment
	State      string // filter: state field matches (current, desired, both)
	Group      string // filter: group field matches
	Where      []string // generic field=value filters (AND logic)
}

// QueryResult represents a single artifact match.
type QueryResult struct {
	ID     string
	Kind   string
	Status string
	Sprint string
	Title  string
	Path   string
}

// QueryArtifacts scans all artifact types and returns matches.
func QueryArtifacts(root string, opts QueryOpts) ([]QueryResult, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, fmt.Errorf("loading registry: %w", err)
	}

	linkFields := reg.AllLinkFields()

	var results []QueryResult

	if opts.Kind == "" || opts.Kind == "all" || opts.Kind == KindContract {
		r, err := queryContracts(root, opts, linkFields)
		if err != nil {
			return nil, fmt.Errorf("QueryArtifacts: %w", err)
		}
		results = append(results, r...)
	}

	for _, coreKind := range []string{KindSpecification, KindRule, KindBinder} {
		if opts.Kind != "" && opts.Kind != "all" && opts.Kind != coreKind {
			continue
		}
		td, ok := reg.Types[coreKind]
		if !ok {
			continue
		}
		r, err := queryCustomType(root, coreKind, td.Directory, opts, linkFields)
		if err != nil {
			return nil, fmt.Errorf("QueryArtifacts: %w", err)
		}
		results = append(results, r...)
	}

	for kind, td := range reg.Types {
		if td.Core {
			continue
		}
		if opts.Kind != "" && opts.Kind != "all" && opts.Kind != kind {
			continue
		}
		r, err := queryCustomType(root, kind, td.Directory, opts, linkFields)
		if err != nil {
			return nil, fmt.Errorf("QueryArtifacts: %w", err)
		}
		results = append(results, r...)
	}

	slices.SortFunc(results, func(a, b QueryResult) int { return cmp.Compare(a.ID, b.ID) })
	return results, nil
}

func queryContracts(root string, opts QueryOpts, linkFields []string) ([]QueryResult, error) {
	var results []QueryResult
	mosDir := filepath.Join(root, MosDir)

	for _, sub := range []string{ActiveDir, ArchiveDir} {
		dir := filepath.Join(mosDir, DirContracts, sub)
		entries, err := storeReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(dir, e.Name(), "contract.mos")
			r, ok := matchArtifact(path, KindContract, e.Name(), opts, linkFields)
			if ok {
				results = append(results, r)
			}
		}
	}
	return results, nil
}

func queryCustomType(root, kind, directory string, opts QueryOpts, linkFields []string) ([]QueryResult, error) {
	var results []QueryResult
	mosDir := filepath.Join(root, MosDir)

	for _, sub := range []string{ActiveDir, ArchiveDir} {
		dir := filepath.Join(mosDir, directory, sub)
		entries, err := storeReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(dir, e.Name(), kind+".mos")
			r, ok := matchArtifact(path, kind, e.Name(), opts, linkFields)
			if ok {
				results = append(results, r)
			}
		}
	}
	return results, nil
}

func matchArtifact(path, kind, dirName string, opts QueryOpts, linkFields []string) (QueryResult, bool) {
	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return QueryResult{}, false
	}

	id := ab.Name
	if id == "" {
		id = dirName
	}

	status := FieldStr(ab.Items, FieldStatus)
	sprint := FieldStr(ab.Items, "sprint")
	labels := dsl.FieldStringSlice(ab.Items, FieldLabels)

	if opts.Status != "" && status != opts.Status {
		return QueryResult{}, false
	}

	if len(opts.Labels) > 0 {
		labelSet := make(map[string]bool)
		for _, l := range labels {
			labelSet[l] = true
		}
		matched := false
		for _, want := range opts.Labels {
			if labelSet[want] {
				matched = true
				break
			}
		}
		if !matched {
			return QueryResult{}, false
		}
	}

	if opts.References != "" {
		found := false
		for _, linkField := range linkFields {
			val := FieldStr(ab.Items, linkField)
			if val == opts.References {
				found = true
				break
			}
			for _, sv := range dsl.FieldStringSlice(ab.Items, linkField) {
				if sv == opts.References {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return QueryResult{}, false
		}
	}

	if opts.Justifies != "" {
		// Specification uses "satisfies" (spec→need); contract/batch use "justifies"
		fieldName := "justifies"
		if kind == KindSpecification {
			fieldName = "satisfies"
		}
		j := FieldStr(ab.Items, fieldName)
		if !csvContains(j, opts.Justifies) {
			return QueryResult{}, false
		}
	}
	if opts.SprintEq != "" && sprint != opts.SprintEq {
		return QueryResult{}, false
	}
	if opts.DependsOn != "" {
		d := FieldStr(ab.Items, "depends_on")
		if !csvContains(d, opts.DependsOn) {
			return QueryResult{}, false
		}
	}
	if opts.Priority != "" {
		p := FieldStr(ab.Items, "priority")
		if p != opts.Priority {
			return QueryResult{}, false
		}
	}
	if opts.Unlinked && sprint != "" {
		return QueryResult{}, false
	}

	if opts.State != "" {
		state := FieldStr(ab.Items, "state")
		if state != opts.State {
			return QueryResult{}, false
		}
	}

	if opts.Group != "" {
		group := FieldStr(ab.Items, "group")
		if group != opts.Group {
			return QueryResult{}, false
		}
	}

	for _, w := range opts.Where {
		k, v, ok := strings.Cut(w, "=")
		if !ok {
			continue
		}
		actual := FieldStr(ab.Items, k)
		if actual == v {
			continue
		}
		if slices.Contains(dsl.FieldStringSlice(ab.Items, k), v) {
			continue
		}
		return QueryResult{}, false
	}

	title := FieldStr(ab.Items, "title")

	return QueryResult{
		ID:     id,
		Kind:   kind,
		Status: status,
		Sprint: sprint,
		Title:  title,
		Path:   path,
	}, true
}

// RichQueryResults converts query results to full artifact maps using dsl.ToMap.
func RichQueryResults(results []QueryResult) []map[string]any {
	out := make([]map[string]any, 0, len(results))
	for _, r := range results {
		ab, err := dsl.ReadArtifact(r.Path)
		if err != nil {
			continue
		}
		m := dsl.ToMap(ab)
		m["_path"] = r.Path
		out = append(out, m)
	}
	return out
}

func csvContains(csv, target string) bool {
	for _, part := range strings.Split(csv, ",") {
		if strings.TrimSpace(part) == target {
			return true
		}
	}
	return false
}

// FieldStr extracts a string value from a DSL node list by key.
// Handles StringVal, DateTimeVal, and ListVal (first element).
func FieldStr(items []dsl.Node, key string) string {
	for _, item := range items {
		f, ok := item.(*dsl.Field)
		if !ok {
			continue
		}
		if f.Key == key {
			switch v := f.Value.(type) {
			case *dsl.StringVal:
				return v.Text
			case *dsl.DateTimeVal:
				return v.Raw
			case *dsl.ListVal:
				if len(v.Items) > 0 {
					if sv, ok := v.Items[0].(*dsl.StringVal); ok {
						return sv.Text
					}
				}
			}
		}
	}
	return ""
}

// GroupResults groups query results by the given field.
func GroupResults(results []QueryResult, field string) map[string]int {
	counts := make(map[string]int)
	for _, r := range results {
		var key string
		switch field {
		case FieldStatus:
			key = r.Status
		case FieldKind:
			key = r.Kind
		case "sprint":
			key = r.Sprint
		case "group":
			ab, err := dsl.ReadArtifact(r.Path)
			if err == nil {
				key = FieldStr(ab.Items, "group")
			}
		}
		if key == "" {
			key = "(none)"
		}
		counts[key]++
	}
	return counts
}

// FormatQueryResults formats query results for output.
func FormatQueryResults(results []QueryResult, opts QueryOpts) string {
	if opts.Count {
		if opts.GroupBy != "" {
			groups := GroupResults(results, opts.GroupBy)
			keys := slices.Sorted(maps.Keys(groups))
			var sb strings.Builder
			for _, k := range keys {
				fmt.Fprintf(&sb, "%s: %d\n", k, groups[k])
			}
			return sb.String()
		}
		return fmt.Sprintf("%d\n", len(results))
	}

	var sb strings.Builder
	for _, r := range results {
		fmt.Fprintf(&sb, "%-20s %-12s %s\n", r.ID, r.Status, r.Title)
		if opts.Verbose {
			writeVerboseFields(&sb, r)
		}
	}
	return sb.String()
}

func writeVerboseFields(sb *strings.Builder, r QueryResult) {
	ab, err := dsl.ReadArtifact(r.Path)
	if err != nil {
		return
	}

	var keys []string
	switch r.Kind {
	case KindSpecification:
		keys = []string{"satisfies", "enforcement", "verification_method", "group"}
	case KindContract:
		keys = []string{"justifies", "kind", "priority", "severity", "sprint"}
	case "need":
		keys = []string{"originating", "derives_from", "urgency"}
	default:
		keys = []string{"justifies", "sprint"}
	}

	for _, k := range keys {
		v := FieldStr(ab.Items, k)
		if v == "" {
			continue
		}
		fmt.Fprintf(sb, "  %s: %s\n", k, v)
	}
}
