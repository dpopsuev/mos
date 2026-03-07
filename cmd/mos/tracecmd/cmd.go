package tracecmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/moslib/arch"
	"github.com/dpopsuev/mos/moslib/artifact"
	mclone "github.com/dpopsuev/mos/moslib/clone"
	"github.com/dpopsuev/mos/moslib/mesh"
	"github.com/dpopsuev/mos/moslib/model"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/survey"
)

// --- Mesh ---

var meshFlags struct {
	format string
	all    bool
}

var MeshCmd = &cobra.Command{
	Use:   "mesh [path]",
	Short: "Show the context mesh for a code file, package, or entire project",
	Long: `Resolve a Go source path or package to its governance context mesh.

Shows which specs, contracts, needs, and sprints are linked to the given
code location, plus its architectural import edges.

Use --all for a project-wide view covering every package.

Output formats: text (default), json, mermaid`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMesh,
}

func init() {
	MeshCmd.Flags().StringVar(&meshFlags.format, "format", "text", "Output format: text, json, mermaid")
	MeshCmd.Flags().BoolVar(&meshFlags.all, "all", false, "Project-wide mesh: all packages and governance links")
}

func runMesh(cmd *cobra.Command, args []string) error {
	if meshFlags.all {
		g, err := mesh.ResolveAll(".")
		if err != nil {
			return err
		}
		switch meshFlags.format {
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(g)
		case "mermaid":
			fmt.Print(mesh.RenderMermaid(g))
		default:
			fmt.Print(mesh.RenderTreeAll(g))
		}
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("requires a path argument (or use --all for project-wide)")
	}

	target := args[0]
	pkgPath := mesh.ResolvePackagePath(target)

	g, err := mesh.Resolve(".", pkgPath)
	if err != nil {
		return err
	}

	switch meshFlags.format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(g)
	case "mermaid":
		fmt.Print(mesh.RenderMermaid(g))
	default:
		fmt.Print(mesh.RenderTree(g, pkgPath))
	}
	return nil
}

// --- Clone ---

var cloneFlags struct {
	from  string
	kinds string
	state string
	group string
}

var CloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone artifacts from another Mos project",
	Long: `Reads artifacts from a source .mos/ tree and copies them into the current
project with new IDs. Maintains internal cross-references by remapping IDs.

Examples:
  mos clone --from ~/Workspace/git --kinds specification,architecture
  mos clone --from ~/Workspace/git --kinds specification --state current
  mos clone --from ~/Workspace/git --kinds specification --group builtin`,
	RunE: runClone,
}

func init() {
	CloneCmd.Flags().StringVar(&cloneFlags.from, "from", "", "Source project root (required)")
	CloneCmd.Flags().StringVar(&cloneFlags.kinds, "kinds", "", "Comma-separated artifact kinds to clone")
	CloneCmd.Flags().StringVar(&cloneFlags.state, "state", "", "Only clone specs with this state")
	CloneCmd.Flags().StringVar(&cloneFlags.group, "group", "", "Only clone artifacts with this group value")
	_ = CloneCmd.MarkFlagRequired("from")
}

func runClone(cmd *cobra.Command, args []string) error {
	var kinds []string
	if cloneFlags.kinds != "" {
		kinds = strings.Split(cloneFlags.kinds, ",")
	}

	opts := mclone.Opts{
		From:  cloneFlags.from,
		Kinds: kinds,
		State: cloneFlags.state,
		Group: cloneFlags.group,
	}

	results, idMap, err := mclone.Run(".", opts, artifact.Svc.Create, artifact.Svc.FindPath, artifact.Svc.FieldStr)
	if err != nil {
		return err
	}

	for _, r := range results {
		fmt.Printf("cloned %s -> %s (%s)\n", r.OldID, r.NewID, r.Title)
	}

	if len(idMap) > 0 {
		mapJSON, _ := json.MarshalIndent(idMap, "", "  ")
		fmt.Printf("\nID mapping:\n%s\n", string(mapJSON))
	}
	fmt.Printf("\nCloned %d artifact(s)\n", len(results))
	return nil
}

// --- Spec Generate ---

var specGenFlags struct {
	state   string
	status  string
	exclude string
}

var SpecGenCmd = &cobra.Command{
	Use:   "generate",
	Short: "Auto-generate specifications from architecture",
	Long: `Reads the architecture artifact (ARCH-auto) and creates a specification
for each component/service that does not already have a corresponding spec.
Generated specs include the source directory path and summary statistics.`,
	RunE: runSpecGen,
}

func init() {
	SpecGenCmd.Flags().StringVar(&specGenFlags.state, "state", "current", "Specification state: current or desired")
	SpecGenCmd.Flags().StringVar(&specGenFlags.status, "status", "", "Override initial status (default: candidate)")
	SpecGenCmd.Flags().StringVar(&specGenFlags.exclude, "exclude", "", "Comma-separated directory patterns to skip (e.g. t,Documentation)")
}

func runSpecGen(cmd *cobra.Command, args []string) error {
	root := "."

	var excludePatterns []string
	if specGenFlags.exclude != "" {
		for _, p := range strings.Split(specGenFlags.exclude, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				excludePatterns = append(excludePatterns, p)
			}
		}
	}

	opts := arch.SpecGenOpts{
		State:           specGenFlags.state,
		StatusOverride:  specGenFlags.status,
		ExcludePatterns: excludePatterns,
	}

	results, err := arch.GenerateSpecs(root, opts, artifact.Svc.Create, artifact.Svc.FindPath)
	if err != nil {
		return err
	}

	for _, r := range results {
		fmt.Printf("created %s -> %s (%s) [group=%s]\n", r.ID, r.Title, r.Pkg, r.Group)
	}
	fmt.Printf("\nGenerated %d specification(s)\n", len(results))
	return nil
}

// --- Survey ---

var (
	surveyFormat  string
	surveyScanner string
	surveyLSPCmd  string
)

var SurveyCmd = &cobra.Command{
	Use:   "survey <path>",
	Short: "Run governance surveys",
	Args:  cobra.ExactArgs(1),
	RunE:  runSurvey,
}

func init() {
	SurveyCmd.Flags().StringVar(&surveyFormat, "format", "tree", "Output format: tree or json")
	SurveyCmd.Flags().StringVar(&surveyScanner, "scanner", "auto", "Scanner backend: auto, go, packages, lsp")
	SurveyCmd.Flags().StringVar(&surveyLSPCmd, "lsp-cmd", "", "LSP server command")
}

func runSurvey(cmd *cobra.Command, args []string) error {
	sc := &survey.AutoScanner{Override: surveyScanner, LSPCmd: surveyLSPCmd}
	mod, err := sc.Scan(args[0])
	if err != nil {
		return err
	}

	switch surveyFormat {
	case names.FormatJSON:
		data, err := json.MarshalIndent(mod, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "tree":
		printTree(mod)
	default:
		return fmt.Errorf("unknown format %q", surveyFormat)
	}
	return nil
}

func printTree(mod *model.Project) {
	fmt.Printf("%s\n", mod.Path)
	for i, pkg := range mod.Namespaces {
		last := i == len(mod.Namespaces)-1
		prefix := "├── "
		childPrefix := "│   "
		if last && (mod.DependencyGraph == nil || len(mod.DependencyGraph.Edges) == 0) {
			prefix = "└── "
			childPrefix = "    "
		}

		fmt.Printf("%s%s\n", prefix, shortPath(mod.Path, pkg.ImportPath))

		items := make([]string, 0, len(pkg.Files)+len(pkg.Symbols))
		for _, f := range pkg.Files {
			items = append(items, f.Path)
		}
		for _, s := range pkg.Symbols {
			vis := "+"
			if !s.Exported {
				vis = "-"
			}
			items = append(items, fmt.Sprintf("[%s] %s%s", s.Kind, vis, s.Name))
		}

		for j, item := range items {
			connector := "├── "
			if j == len(items)-1 {
				connector = "└── "
			}
			fmt.Printf("%s%s%s\n", childPrefix, connector, item)
		}
	}

	if mod.DependencyGraph != nil && len(mod.DependencyGraph.Edges) > 0 {
		fmt.Println()
		fmt.Println("Dependency Graph:")
		for _, e := range mod.DependencyGraph.Edges {
			kind := "internal"
			if e.External {
				kind = "external"
			}
			fmt.Printf("  %s -> %s (%s)\n",
				shortPath(mod.Path, e.From),
				shortPath(mod.Path, e.To),
				kind)
		}
	}
}

func shortPath(modPath, importPath string) string {
	if importPath == modPath {
		return "."
	}
	if strings.HasPrefix(importPath, modPath+"/") {
		return strings.TrimPrefix(importPath, modPath+"/")
	}
	return importPath
}
