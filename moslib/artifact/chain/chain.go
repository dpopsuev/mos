package chain

import (
	"fmt"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/artifact"
)

// BlameRef is a source reference attached to a chain link.
type BlameRef struct {
	File   string `json:"file"`
	Lines  string `json:"lines,omitempty"`
	Symbol string `json:"symbol,omitempty"`
}

// ChainLink represents a single artifact in a traceability chain.
type ChainLink struct {
	Kind   string     `json:"kind"`
	ID     string     `json:"id"`
	Title  string     `json:"title,omitempty"`
	LinkTo string     `json:"link_to,omitempty"`
	Blame  []BlameRef `json:"blame,omitempty"`
}

// ChainResult holds the full traceability chain for an artifact.
type ChainResult struct {
	Root     ChainLink   `json:"root"`
	Upward   []ChainLink `json:"upward,omitempty"`
	Downward []ChainLink `json:"downward,omitempty"`
}

// NegativeSpaceEntry represents one level of the negative-space chain.
type NegativeSpaceEntry struct {
	Kind   string   `json:"kind"`
	ID     string   `json:"id"`
	Field  string   `json:"field"`
	Values []string `json:"values,omitempty"`
}

// NegativeChainResult holds the negative-space chain for an artifact.
type NegativeChainResult struct {
	Entries []NegativeSpaceEntry `json:"entries,omitempty"`
}

// WalkChain traverses cross-type links upward and downward from a starting artifact.
func WalkChain(root string, startKind, startID string) (*ChainResult, error) {
	reg, err := artifact.LoadRegistry(root)
	if err != nil {
		return nil, fmt.Errorf("loading registry: %w", err)
	}

	result := &ChainResult{}

	rootArt, rootPath, err := readArtifactFieldsAndPath(root, reg, startKind, startID)
	if err != nil {
		return nil, err
	}
	rootAB := readABFromPath(rootPath)
	rootLink := ChainLink{Kind: startKind, ID: startID, Title: rootArt["title"]}
	if rootAB != nil {
		rootLink.Blame = extractBlameRefs(rootAB)
	}
	result.Root = rootLink
	walkUpward(root, reg, startKind, startID, rootArt, rootAB, &result.Upward)
	walkDownward(root, reg, startKind, startID, &result.Downward)

	return result, nil
}

func walkUpward(root string, reg *artifact.Registry, kind, id string, fields map[string]string, ab *dsl.ArtifactBlock, chain *[]ChainLink) {
	td := reg.Types[kind]
	traceFields := td.TraceFields()
	if len(traceFields) == 0 {
		traceFields = []string{"justifies", "implements", "documents"}
	}
	for _, linkField := range traceFields {
		var targetIDs []string
		if ab != nil {
			targetIDs = dsl.FieldStringSlice(ab.Items, linkField)
		} else if v := fields[linkField]; v != "" {
			targetIDs = []string{v}
		}
		if len(targetIDs) == 0 {
			continue
		}
		for _, targetID := range targetIDs {
			for targetKind, td := range reg.Types {
				path, err := artifact.FindGenericPath(root, td, targetID)
				if err != nil {
					continue
				}
				parentFields := readFieldsFromPath(path)
				parentAB := readABFromPath(path)
				*chain = append(*chain, ChainLink{
					Kind:   targetKind,
					ID:     targetID,
					Title:  parentFields["title"],
					LinkTo: id,
				})
				walkUpward(root, reg, targetKind, targetID, parentFields, parentAB, chain)
				return
			}
		}
	}
}

func walkDownward(root string, reg *artifact.Registry, kind, id string, chain *[]ChainLink) {
	for childKind, td := range reg.Types {
		items, err := artifact.GenericList(root, td, "")
		if err != nil {
			continue
		}
		for _, item := range items {
			path, err := artifact.FindGenericPath(root, td, item.ID)
			if err != nil {
				continue
			}
			fields := readFieldsFromPath(path)
			childAB := readABFromPath(path)
			childTraceFields := td.TraceFields()
			if len(childTraceFields) == 0 {
				childTraceFields = []string{"justifies", "implements", "documents"}
			}
			for _, linkField := range childTraceFields {
				matched := false
				if childAB != nil {
					for _, v := range dsl.FieldStringSlice(childAB.Items, linkField) {
						if v == id {
							matched = true
							break
						}
					}
				} else {
					matched = fields[linkField] == id
				}
				if matched {
					link := ChainLink{
						Kind:   childKind,
						ID:     item.ID,
						Title:  fields["title"],
						LinkTo: id,
					}
					if childAB != nil {
						link.Blame = extractBlameRefs(childAB)
					}
					*chain = append(*chain, link)
					walkDownward(root, reg, childKind, item.ID, chain)
				}
			}
		}
	}
}

func readArtifactFieldsAndPath(root string, reg *artifact.Registry, kind, id string) (map[string]string, string, error) {
	td, ok := reg.Types[kind]
	if !ok {
		return nil, "", fmt.Errorf("unknown artifact kind %q", kind)
	}
	path, err := artifact.FindGenericPath(root, td, id)
	if err != nil {
		if kind == artifact.KindContract {
			path, err = artifact.FindContractPath(root, id)
			if err != nil {
				return nil, "", fmt.Errorf("artifact %s/%s not found", kind, id)
			}
		} else {
			return nil, "", fmt.Errorf("artifact %s/%s not found", kind, id)
		}
	}
	return readFieldsFromPath(path), path, nil
}

func readABFromPath(path string) *dsl.ArtifactBlock {
	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return nil
	}
	return ab
}

func readFieldsFromPath(path string) map[string]string {
	fields := make(map[string]string)
	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return fields
	}
	m := dsl.ToMap(ab)
	for k, v := range m {
		switch x := v.(type) {
		case string:
			fields[k] = x
		case []any:
			if len(x) > 0 {
				if s, ok := x[0].(string); ok {
					fields[k] = s
				}
			}
		}
	}
	return fields
}

func readListFieldFromPath(path, key string) []string {
	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return nil
	}
	result := dsl.FieldStringSlice(ab.Items, key)
	if len(result) == 0 {
		if blk := dsl.FindBlock(ab.Items, "scope"); blk != nil {
			result = dsl.FieldStringSlice(blk.Items, key)
		}
	}
	return result
}

func extractBlameRefs(ab *dsl.ArtifactBlock) []BlameRef {
	entries := artifact.ParseBlameEntries(ab)
	if len(entries) == 0 {
		return nil
	}
	refs := make([]BlameRef, len(entries))
	for i, e := range entries {
		refs[i] = BlameRef{File: e.File, Lines: e.Lines, Symbol: e.Symbol}
	}
	return refs
}

// WalkNegativeChain traces the negative-space chain for an artifact.
func WalkNegativeChain(root string, startKind, startID string) (*NegativeChainResult, error) {
	reg, err := artifact.LoadRegistry(root)
	if err != nil {
		return nil, fmt.Errorf("loading registry: %w", err)
	}

	result := &NegativeChainResult{}

	c, err := WalkChain(root, startKind, startID)
	if err != nil {
		return nil, err
	}

	allLinks := append([]ChainLink{c.Root}, c.Upward...)
	allLinks = append(allLinks, c.Downward...)

	for _, link := range allLinks {
		td, ok := reg.Types[link.Kind]
		if !ok {
			continue
		}
		path, _ := artifact.FindGenericPath(root, td, link.ID)
		if path == "" && link.Kind == artifact.KindContract {
			path, _ = artifact.FindContractPath(root, link.ID)
		}
		if path == "" {
			continue
		}

		switch link.Kind {
		case artifact.KindNeed:
			excludes := readListFieldFromPath(path, "excludes")
			if len(excludes) > 0 {
				result.Entries = append(result.Entries, NegativeSpaceEntry{
					Kind: artifact.KindNeed, ID: link.ID, Field: "scope.excludes", Values: excludes,
				})
			}
		case artifact.KindSpecification:
			nonGoals := readListFieldFromPath(path, "non_goals")
			if len(nonGoals) > 0 {
				result.Entries = append(result.Entries, NegativeSpaceEntry{
					Kind: artifact.KindSpecification, ID: link.ID, Field: "non_goals", Values: nonGoals,
				})
			}
		}
	}

	return result, nil
}

// FormatChain renders a chain result as human-readable text.
func FormatChain(c *ChainResult) string {
	var b strings.Builder

	if len(c.Upward) > 0 {
		for i := len(c.Upward) - 1; i >= 0; i-- {
			link := c.Upward[i]
			fmt.Fprintf(&b, "[%s] %s: %s\n", link.Kind, link.ID, link.Title)
			formatBlameRefs(&b, link.Blame)
			b.WriteString("  |\n")
		}
	}

	fmt.Fprintf(&b, "[%s] %s: %s  <-- you are here\n", c.Root.Kind, c.Root.ID, c.Root.Title)
	formatBlameRefs(&b, c.Root.Blame)

	if len(c.Downward) > 0 {
		for _, link := range c.Downward {
			b.WriteString("  |\n")
			fmt.Fprintf(&b, "[%s] %s: %s\n", link.Kind, link.ID, link.Title)
			formatBlameRefs(&b, link.Blame)
		}
	}

	return b.String()
}

func formatBlameRefs(b *strings.Builder, refs []BlameRef) {
	for _, r := range refs {
		loc := r.File
		if r.Lines != "" {
			loc += ":" + r.Lines
		}
		if r.Symbol != "" {
			fmt.Fprintf(b, "    blame: %s (%s)\n", loc, r.Symbol)
		} else {
			fmt.Fprintf(b, "    blame: %s\n", loc)
		}
	}
}

// FormatNegativeChain renders the negative-space chain as human-readable text.
func FormatNegativeChain(nc *NegativeChainResult) string {
	if len(nc.Entries) == 0 {
		return "(no negative-space declarations found)\n"
	}

	var b strings.Builder
	b.WriteString("Negative-Space Chain:\n")
	for _, e := range nc.Entries {
		fmt.Fprintf(&b, "  [%s] %s.%s:\n", e.Kind, e.ID, e.Field)
		for _, v := range e.Values {
			fmt.Fprintf(&b, "    - %s\n", v)
		}
	}
	return b.String()
}
