package mesh

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/registry"
	"golang.org/x/mod/modfile"
)

// Resolve builds a context mesh graph for the given Go package path.
// It links the package to specs (via include directives), contracts (via justifies),
// needs (via contract.justifies), sprints (via contract.sprint), and architecture edges.
func Resolve(root, packagePath string) (*Graph, error) {
	reg, err := registry.LoadRegistry(root)
	if err != nil {
		return nil, fmt.Errorf("mesh.Resolve: %w", err)
	}

	modPath := detectModulePath(root)

	g := &Graph{}
	g.AddNode(Node{ID: packagePath, Kind: "package", Label: shortPath(packagePath, modPath)})

	specTD, ok := reg.Types["specification"]
	if !ok {
		return nil, fmt.Errorf("mesh.Resolve: specification type not found in registry")
	}
	specs, err := artifact.GenericList(root, specTD, "")
	if err != nil {
		return nil, fmt.Errorf("mesh.Resolve: listing specs: %w", err)
	}

	// Build reverse index: package path -> spec IDs
	// Supports both full import paths and module-relative paths
	specIndex := map[string][]string{}
	for _, si := range specs {
		ab, err := dsl.ReadArtifact(si.Path)
		if err != nil {
			continue
		}
		for _, item := range ab.Items {
			sb, ok := item.(*dsl.SpecBlock)
			if !ok {
				continue
			}
			for _, inc := range sb.Includes {
				specIndex[inc.Path] = append(specIndex[inc.Path], si.ID)
				if modPath != "" && !strings.HasPrefix(inc.Path, modPath) {
					specIndex[modPath+"/"+inc.Path] = append(specIndex[modPath+"/"+inc.Path], si.ID)
				}
			}
		}
	}

	matchedSpecs := specIndex[packagePath]
	for _, specID := range matchedSpecs {
		g.AddNode(Node{ID: specID, Kind: "spec", Label: specID})
		g.AddEdge(Edge{From: specID, To: packagePath, Relation: "includes"})
	}

	contractTD, ok := reg.Types[names.KindContract]
	if !ok {
		return nil, fmt.Errorf("mesh.Resolve: contract type not found in registry")
	}
	contracts, err := artifact.GenericList(root, contractTD, "")
	if err != nil {
		return nil, fmt.Errorf("mesh.Resolve: listing contracts: %w", err)
	}

	needIDs := map[string]bool{}
	sprintIDs := map[string]bool{}

	for _, ci := range contracts {
		ab, err := dsl.ReadArtifact(ci.Path)
		if err != nil {
			continue
		}
		justifies, _ := dsl.FieldString(ab.Items, "justifies")
		sprint, _ := dsl.FieldString(ab.Items, "sprint")
		title, _ := dsl.FieldString(ab.Items, "title")

		linked := false
		for _, specID := range matchedSpecs {
			for _, j := range splitCSV(justifies) {
				if j == specID {
					linked = true
					break
				}
			}
			if linked {
				break
			}
		}
		if !linked {
			continue
		}

		g.AddNode(Node{ID: ci.ID, Kind: "contract", Label: title, Meta: map[string]string{"status": ci.Status}})
		for _, j := range splitCSV(justifies) {
			for _, specID := range matchedSpecs {
				if j == specID {
					g.AddEdge(Edge{From: ci.ID, To: specID, Relation: "justifies"})
				}
			}
		}

		for _, j := range splitCSV(justifies) {
			if strings.HasPrefix(j, "NEED-") {
				needIDs[j] = true
				g.AddEdge(Edge{From: ci.ID, To: j, Relation: "addresses"})
			}
		}

		if sprint != "" {
			sprintIDs[sprint] = true
			g.AddEdge(Edge{From: ci.ID, To: sprint, Relation: "scheduled_in"})
		}
	}

	for needID := range needIDs {
		info, err := loadArtifactInfo(root, reg, "need", needID)
		if err == nil {
			g.AddNode(Node{ID: needID, Kind: "need", Label: info.title, Meta: map[string]string{"status": info.status}})
		} else {
			g.AddNode(Node{ID: needID, Kind: "need", Label: needID})
		}
	}

	for sprintID := range sprintIDs {
		info, err := loadArtifactInfo(root, reg, names.KindSprint, sprintID)
		if err == nil {
			g.AddNode(Node{ID: sprintID, Kind: "sprint", Label: info.title, Meta: map[string]string{"status": info.status}})
		} else {
			g.AddNode(Node{ID: sprintID, Kind: "sprint", Label: sprintID})
		}
	}

	addArchEdges(root, reg, g, packagePath)
	return g, nil
}

type artifactInfo struct {
	title  string
	status string
}

func loadArtifactInfo(root string, reg *registry.Registry, kind, id string) (artifactInfo, error) {
	td, ok := reg.Types[kind]
	if !ok {
		return artifactInfo{}, fmt.Errorf("kind %q not found", kind)
	}
	path, err := artifact.FindGenericPath(root, td, id)
	if err != nil {
		return artifactInfo{}, err
	}
	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return artifactInfo{}, err
	}
	title, _ := dsl.FieldString(ab.Items, "title")
	status, _ := dsl.FieldString(ab.Items, "status")
	return artifactInfo{title: title, status: status}, nil
}

func addArchEdges(root string, reg *registry.Registry, g *Graph, packagePath string) {
	td, ok := reg.Types["architecture"]
	if !ok {
		return
	}
	archs, err := artifact.GenericList(root, td, "")
	if err != nil || len(archs) == 0 {
		return
	}
	for _, ai := range archs {
		ab, err := dsl.ReadArtifact(ai.Path)
		if err != nil {
			continue
		}
		arch := artifact.ParseArchModel(ab)
		for _, e := range arch.Edges {
			if e.From == packagePath || e.To == packagePath {
				g.AddNode(Node{ID: e.From, Kind: "package", Label: e.From})
				g.AddNode(Node{ID: e.To, Kind: "package", Label: e.To})
				g.AddEdge(Edge{From: e.From, To: e.To, Relation: "imports"})
			}
		}
	}
}

func detectModulePath(root string) string {
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

func shortPath(pkgPath, modPath string) string {
	if modPath != "" && strings.HasPrefix(pkgPath, modPath+"/") {
		return strings.TrimPrefix(pkgPath, modPath+"/")
	}
	return pkgPath
}

// ResolveAll builds a project-wide context mesh graph covering all packages.
// It produces the same node/edge types as Resolve but for every spec-included
// package, plus unlinked packages as orphan nodes.
func ResolveAll(root string) (*Graph, error) {
	reg, err := registry.LoadRegistry(root)
	if err != nil {
		return nil, fmt.Errorf("mesh.ResolveAll: %w", err)
	}

	modPath := detectModulePath(root)
	g := &Graph{}

	specTD, ok := reg.Types["specification"]
	if !ok {
		return nil, fmt.Errorf("mesh.ResolveAll: specification type not found in registry")
	}
	specs, err := artifact.GenericList(root, specTD, "")
	if err != nil {
		return nil, fmt.Errorf("mesh.ResolveAll: listing specs: %w", err)
	}

	specIndex := map[string][]string{}
	for _, si := range specs {
		ab, err := dsl.ReadArtifact(si.Path)
		if err != nil {
			continue
		}
		for _, item := range ab.Items {
			sb, ok := item.(*dsl.SpecBlock)
			if !ok {
				continue
			}
			for _, inc := range sb.Includes {
				full := inc.Path
				if modPath != "" && !strings.HasPrefix(full, modPath) {
					full = modPath + "/" + inc.Path
				}
				specIndex[full] = append(specIndex[full], si.ID)
				if inc.Path != full {
					specIndex[inc.Path] = append(specIndex[inc.Path], si.ID)
				}
			}
		}
	}

	allLinkedPkgs := map[string]bool{}
	for pkg := range specIndex {
		allLinkedPkgs[pkg] = true
	}

	for pkg := range allLinkedPkgs {
		g.AddNode(Node{ID: pkg, Kind: "package", Label: shortPath(pkg, modPath)})
		for _, specID := range specIndex[pkg] {
			g.AddNode(Node{ID: specID, Kind: "spec", Label: specID})
			g.AddEdge(Edge{From: specID, To: pkg, Relation: "includes"})
		}
	}

	contractTD, ok := reg.Types[names.KindContract]
	if !ok {
		return nil, fmt.Errorf("mesh.ResolveAll: contract type not found in registry")
	}
	contracts, err := artifact.GenericList(root, contractTD, "")
	if err != nil {
		return nil, fmt.Errorf("mesh.ResolveAll: listing contracts: %w", err)
	}

	allSpecIDs := map[string]bool{}
	for _, ids := range specIndex {
		for _, id := range ids {
			allSpecIDs[id] = true
		}
	}

	needIDs := map[string]bool{}
	sprintIDs := map[string]bool{}

	for _, ci := range contracts {
		ab, err := dsl.ReadArtifact(ci.Path)
		if err != nil {
			continue
		}
		justifies, _ := dsl.FieldString(ab.Items, "justifies")
		sprint, _ := dsl.FieldString(ab.Items, "sprint")
		title, _ := dsl.FieldString(ab.Items, "title")

		linked := false
		for _, j := range splitCSV(justifies) {
			if allSpecIDs[j] {
				linked = true
				break
			}
		}
		if !linked {
			continue
		}

		g.AddNode(Node{ID: ci.ID, Kind: "contract", Label: title, Meta: map[string]string{"status": ci.Status}})
		for _, j := range splitCSV(justifies) {
			if allSpecIDs[j] {
				g.AddEdge(Edge{From: ci.ID, To: j, Relation: "justifies"})
			}
			if strings.HasPrefix(j, "NEED-") {
				needIDs[j] = true
				g.AddEdge(Edge{From: ci.ID, To: j, Relation: "addresses"})
			}
		}
		if sprint != "" {
			sprintIDs[sprint] = true
			g.AddEdge(Edge{From: ci.ID, To: sprint, Relation: "scheduled_in"})
		}
	}

	for needID := range needIDs {
		info, err := loadArtifactInfo(root, reg, "need", needID)
		if err == nil {
			g.AddNode(Node{ID: needID, Kind: "need", Label: info.title, Meta: map[string]string{"status": info.status}})
		} else {
			g.AddNode(Node{ID: needID, Kind: "need", Label: needID})
		}
	}

	for sprintID := range sprintIDs {
		info, err := loadArtifactInfo(root, reg, names.KindSprint, sprintID)
		if err == nil {
			g.AddNode(Node{ID: sprintID, Kind: "sprint", Label: info.title, Meta: map[string]string{"status": info.status}})
		} else {
			g.AddNode(Node{ID: sprintID, Kind: "sprint", Label: sprintID})
		}
	}

	addAllArchEdges(root, reg, g, allLinkedPkgs)
	return g, nil
}

func addAllArchEdges(root string, reg *registry.Registry, g *Graph, linkedPkgs map[string]bool) {
	td, ok := reg.Types["architecture"]
	if !ok {
		return
	}
	archs, err := artifact.GenericList(root, td, "")
	if err != nil || len(archs) == 0 {
		return
	}
	for _, ai := range archs {
		ab, err := dsl.ReadArtifact(ai.Path)
		if err != nil {
			continue
		}
		arch := artifact.ParseArchModel(ab)
		for _, svc := range arch.Services {
			g.AddNode(Node{ID: svc.Name, Kind: "package", Label: svc.Name})
		}
		for _, e := range arch.Edges {
			g.AddNode(Node{ID: e.From, Kind: "package", Label: e.From})
			g.AddNode(Node{ID: e.To, Kind: "package", Label: e.To})
			g.AddEdge(Edge{From: e.From, To: e.To, Relation: "imports"})
		}
	}
}

// RenderTreeAll produces a project-wide text rendering grouped by package.
func RenderTreeAll(g *Graph) string {
	var b strings.Builder

	pkgs := g.NodesOfKind("package")
	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].ID < pkgs[j].ID })

	linked := []Node{}
	unlinked := []Node{}
	for _, pkg := range pkgs {
		hasSpec := false
		for _, e := range g.EdgesTo(pkg.ID) {
			if e.Relation == "includes" {
				hasSpec = true
				break
			}
		}
		if hasSpec {
			linked = append(linked, pkg)
		} else {
			unlinked = append(unlinked, pkg)
		}
	}

	fmt.Fprintf(&b, "Linked packages (%d):\n", len(linked))
	for _, pkg := range linked {
		fmt.Fprintf(&b, "\n  %s\n", pkg.Label)
		for _, e := range g.EdgesTo(pkg.ID) {
			if e.Relation == "includes" {
				fmt.Fprintf(&b, "    spec: %s\n", e.From)
			}
		}
	}

	if len(unlinked) > 0 {
		fmt.Fprintf(&b, "\nUnlinked packages (%d):\n", len(unlinked))
		for _, pkg := range unlinked {
			fmt.Fprintf(&b, "  %s\n", pkg.Label)
		}
	}

	specs := g.NodesOfKind("spec")
	contracts := g.NodesOfKind("contract")
	needs := g.NodesOfKind("need")
	sprints := g.NodesOfKind("sprint")

	if len(specs) > 0 {
		fmt.Fprintf(&b, "\nSpecs (%d):\n", len(specs))
		for _, s := range specs {
			fmt.Fprintf(&b, "  %s\n", s.ID)
		}
	}
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
	if len(needs) > 0 {
		fmt.Fprintf(&b, "\nNeeds (%d):\n", len(needs))
		for _, n := range needs {
			fmt.Fprintf(&b, "  %s  %s\n", n.ID, n.Label)
		}
	}
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

	return b.String()
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
