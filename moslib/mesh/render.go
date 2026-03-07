package mesh

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/mod/modfile"
)

// RenderTree produces a human-readable text representation of a mesh graph.
func RenderTree(g *Graph, pkgPath string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Package: %s\n", pkgPath)

	specs := g.NodesOfKind("spec")
	if len(specs) > 0 {
		fmt.Fprintf(&b, "\nSpecs (%d):\n", len(specs))
		for _, s := range specs {
			fmt.Fprintf(&b, "  %s  %s\n", s.ID, s.Label)

			for _, e := range g.EdgesTo(s.ID) {
				if e.Relation == "justifies" {
					node := FindNode(g, e.From)
					statusTag := ""
					if node != nil && node.Meta["status"] != "" {
						statusTag = " [" + node.Meta["status"] + "]"
					}
					fmt.Fprintf(&b, "    <- %s  %s%s\n", e.From, NodeLabel(node), statusTag)
				}
			}
		}
	}

	contracts := g.NodesOfKind("contract")
	if len(contracts) > 0 {
		fmt.Fprintf(&b, "\nContracts (%d):\n", len(contracts))
		for _, c := range contracts {
			statusTag := ""
			if c.Meta["status"] != "" {
				statusTag = " [" + c.Meta["status"] + "]"
			}
			fmt.Fprintf(&b, "  %s  %s%s\n", c.ID, c.Label, statusTag)
		}
	}

	needs := g.NodesOfKind("need")
	if len(needs) > 0 {
		fmt.Fprintf(&b, "\nNeeds (%d):\n", len(needs))
		for _, n := range needs {
			fmt.Fprintf(&b, "  %s  %s\n", n.ID, n.Label)
		}
	}

	sprints := g.NodesOfKind("sprint")
	if len(sprints) > 0 {
		fmt.Fprintf(&b, "\nSprints (%d):\n", len(sprints))
		for _, s := range sprints {
			statusTag := ""
			if s.Meta["status"] != "" {
				statusTag = " [" + s.Meta["status"] + "]"
			}
			fmt.Fprintf(&b, "  %s  %s%s\n", s.ID, s.Label, statusTag)
		}
	}

	imports := CollectImports(g, pkgPath)
	if len(imports) > 0 {
		fmt.Fprintf(&b, "\nImports (%d):\n", len(imports))
		for _, imp := range imports {
			fmt.Fprintf(&b, "  -> %s\n", imp)
		}
	}

	importedBy := CollectImportedBy(g, pkgPath)
	if len(importedBy) > 0 {
		fmt.Fprintf(&b, "\nImported by (%d):\n", len(importedBy))
		for _, imp := range importedBy {
			fmt.Fprintf(&b, "  <- %s\n", imp)
		}
	}

	return b.String()
}

// RenderMermaid produces a Mermaid graph diagram of a mesh graph.
func RenderMermaid(g *Graph) string {
	var b strings.Builder
	b.WriteString("graph LR\n")

	kindOrder := []string{"package", "spec", "contract", "need", "sprint"}
	kindLabel := map[string]string{
		"package":  "Code",
		"spec":     "Specifications",
		"contract": "Contracts",
		"need":     "Needs",
		"sprint":   "Sprints",
	}

	for _, kind := range kindOrder {
		nodes := g.NodesOfKind(kind)
		if len(nodes) == 0 {
			continue
		}
		fmt.Fprintf(&b, "  subgraph %s\n", kindLabel[kind])
		for _, n := range nodes {
			safeID := MermaidID(n.ID)
			label := n.Label
			if label == "" {
				label = n.ID
			}
			fmt.Fprintf(&b, "    %s[\"%s\"]\n", safeID, MermaidEsc(label))
		}
		b.WriteString("  end\n")
	}

	for _, e := range g.Edges {
		from := MermaidID(e.From)
		to := MermaidID(e.To)
		fmt.Fprintf(&b, "  %s -->|%s| %s\n", from, e.Relation, to)
	}
	return b.String()
}

// MermaidID sanitizes an identifier for Mermaid diagrams.
func MermaidID(id string) string {
	r := strings.NewReplacer("/", "_", "-", "_", ".", "_", " ", "_")
	return r.Replace(id)
}

// MermaidEsc escapes strings for Mermaid labels.
func MermaidEsc(s string) string {
	return strings.ReplaceAll(s, "\"", "'")
}

// FindNode looks up a node by ID.
func FindNode(g *Graph, id string) *Node {
	for i := range g.Nodes {
		if g.Nodes[i].ID == id {
			return &g.Nodes[i]
		}
	}
	return nil
}

// NodeLabel returns the label of a node, or empty if nil.
func NodeLabel(n *Node) string {
	if n == nil {
		return ""
	}
	return n.Label
}

// CollectImports returns sorted import targets from a package.
func CollectImports(g *Graph, pkgPath string) []string {
	var out []string
	for _, e := range g.EdgesFrom(pkgPath) {
		if e.Relation == "imports" {
			out = append(out, e.To)
		}
	}
	sort.Strings(out)
	return out
}

// CollectImportedBy returns sorted importers of a package.
func CollectImportedBy(g *Graph, pkgPath string) []string {
	var out []string
	for _, e := range g.EdgesTo(pkgPath) {
		if e.Relation == "imports" {
			out = append(out, e.From)
		}
	}
	sort.Strings(out)
	return out
}

// ResolvePackagePath converts a filesystem path to a Go package import path.
func ResolvePackagePath(target string) string {
	if !strings.HasPrefix(target, ".") && !strings.HasPrefix(target, "/") && !strings.Contains(target, string(filepath.Separator)) {
		return target
	}

	abs, err := filepath.Abs(target)
	if err != nil {
		return target
	}

	cwd, err := os.Getwd()
	if err != nil {
		return target
	}

	modPath := DetectModPath(cwd)
	if modPath == "" {
		return target
	}

	rel, err := filepath.Rel(cwd, abs)
	if err != nil {
		return target
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		return modPath
	}
	return modPath + "/" + rel
}

// DetectModPath reads the module path from go.mod, falling back to directory name.
func DetectModPath(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		absRoot, _ := filepath.Abs(root)
		return filepath.Base(absRoot)
	}
	f, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		absRoot, _ := filepath.Abs(root)
		return filepath.Base(absRoot)
	}
	return f.Module.Mod.Path
}
