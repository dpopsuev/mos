package tracecmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/moslib/model"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/survey"
)

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
