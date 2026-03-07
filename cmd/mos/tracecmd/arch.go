package tracecmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/model"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/survey"
)

var archSyncFlags struct {
	scanPath        string
	scannerName     string
	writeMarkdown   bool
	excludeTests    bool
	includeTests    bool
	grouped         bool
	depth           int
	outputDir       string
	includeExternal bool
	churnDays       int
}

var ArchSyncCmd = &cobra.Command{
	Use:   "architecture",
	Short: "Derive architecture from live import graph",
	Long: `Scan source code, extract the package dependency graph, and write
a .mos architecture artifact. Optionally generates ARCHITECTURE.md.

Component groups are read from .mos/config.mos (component_group blocks)
or auto-inferred with --grouped.

Subcommands:
  sync    Scan and update the architecture model`,
}

var archSyncSubCmd = &cobra.Command{
	Use:   "sync",
	Short: "Scan source code and write the architecture artifact",
	RunE: func(c *cobra.Command, args []string) error {
		excludeTests := archSyncFlags.excludeTests
		if archSyncFlags.includeTests {
			excludeTests = false
		}
		depthSet := c.Flags().Changed("depth")
		return runArchSync(syncParams{
			scanPath:        archSyncFlags.scanPath,
			scannerName:     archSyncFlags.scannerName,
			writeMarkdown:   archSyncFlags.writeMarkdown,
			excludeTests:    excludeTests,
			grouped:         archSyncFlags.grouped,
			depth:           archSyncFlags.depth,
			depthSet:        depthSet,
			outputDir:       archSyncFlags.outputDir,
			includeExternal: archSyncFlags.includeExternal,
			churnDays:       archSyncFlags.churnDays,
		})
	},
}

func init() {
	archSyncSubCmd.Flags().StringVar(&archSyncFlags.scanPath, "path", ".", "Path to scan")
	archSyncSubCmd.Flags().StringVar(&archSyncFlags.scannerName, "scanner", "auto", "Scanner: auto, go, packages, lsp, ctags, rust, typescript, composite")
	archSyncSubCmd.Flags().BoolVar(&archSyncFlags.writeMarkdown, "markdown", true, "Write ARCHITECTURE.md")
	archSyncSubCmd.Flags().BoolVar(&archSyncFlags.excludeTests, "exclude-tests", true, "Exclude testkit/ packages")
	archSyncSubCmd.Flags().BoolVar(&archSyncFlags.includeTests, "include-tests", false, "Include test packages in the architecture (overrides --exclude-tests)")
	archSyncSubCmd.Flags().BoolVar(&archSyncFlags.grouped, "grouped", false, "Use component_group blocks from config")
	archSyncSubCmd.Flags().IntVar(&archSyncFlags.depth, "depth", 0, "Group namespaces by first N directory segments (implies --grouped)")
	archSyncSubCmd.Flags().StringVar(&archSyncFlags.outputDir, "output", "", "Output directory for artifacts (default: scanPath or temp dir for non-Mos projects)")
	archSyncSubCmd.Flags().BoolVar(&archSyncFlags.includeExternal, "include-external", false, "Include external (third-party) dependencies in the graph")
	archSyncSubCmd.Flags().IntVar(&archSyncFlags.churnDays, "churn-days", 0, "Overlay file churn from last N days of git history (0 = disabled)")
	ArchSyncCmd.AddCommand(archSyncSubCmd)
}

type syncParams struct {
	scanPath        string
	scannerName     string
	writeMarkdown   bool
	excludeTests    bool
	grouped         bool
	depth           int
	depthSet        bool
	outputDir       string
	includeExternal bool
	churnDays       int
}

func runArchSync(p syncParams) error {
	sc := &survey.AutoScanner{Override: p.scannerName}
	proj, err := sc.Scan(p.scanPath)
	if err != nil {
		return fmt.Errorf("survey scan: %w", err)
	}

	modPath := detectProjectPath(p.scanPath)
	if modPath == "" {
		modPath = proj.Path
	}

	opts := artifact.SyncOptions{
		ModulePath:      modPath,
		ExcludeTests:    p.excludeTests,
		IncludeExternal: p.includeExternal,
	}

	if p.depth > 0 {
		p.grouped = true
	}

	if p.grouped {
		groups, err := artifact.LoadComponentGroups(p.scanPath)
		if err != nil {
			return fmt.Errorf("load component groups: %w", err)
		}
		if len(groups) == 0 {
			d := p.depth
			if d == 0 {
				d = 2
			}
			groups = inferDefaultGroups(proj, modPath, d)
		}
		opts.Groups = groups
	}

	if p.churnDays > 0 {
		opts.ChurnData = artifact.ComputeChurn(p.scanPath, p.churnDays, modPath)
	}

	archModel := artifact.ProjectToArchModel(proj, opts)

	outRoot := resolveOutputDir(p.scanPath, p.outputDir)
	mosContent := artifact.RenderArchMos(archModel)
	archDir := filepath.Join(outRoot, names.MosDir, names.DirArchitectures, names.ActiveDir, "ARCH-auto")
	if err := os.MkdirAll(archDir, 0o755); err != nil {
		return fmt.Errorf("create architecture dir: %w", err)
	}
	archPath := filepath.Join(archDir, "architecture.mos")
	if err := os.WriteFile(archPath, []byte(mosContent), 0o644); err != nil {
		return fmt.Errorf("write architecture artifact: %w", err)
	}
	fmt.Printf("Wrote %s\n", archPath)

	if p.writeMarkdown {
		mdContent := artifact.RenderArchMarkdown(archModel)
		mdPath := filepath.Join(outRoot, "ARCHITECTURE.md")
		if err := os.WriteFile(mdPath, []byte(mdContent), 0o644); err != nil {
			return fmt.Errorf("write ARCHITECTURE.md: %w", err)
		}
		fmt.Printf("Wrote %s\n", mdPath)
	}

	fmt.Printf("Architecture: %d components, %d edges\n", len(archModel.Services), len(archModel.Edges))

	if !p.depthSet && !p.grouped {
		suggestDepth(proj, modPath, len(archModel.Services))
	}

	return nil
}

func suggestDepth(proj *model.Project, modPath string, flatCount int) {
	if flatCount <= 3 {
		return
	}
	bestDepth := 0
	bestCount := flatCount
	for d := 1; d <= 5; d++ {
		groups := inferDefaultGroups(proj, modPath, d)
		grouped := make(map[string]bool)
		ungrouped := 0
		for _, g := range groups {
			grouped[g.Name] = true
		}
		for _, ns := range proj.Namespaces {
			rel := ns.ImportPath
			if strings.HasPrefix(rel, modPath+"/") {
				rel = strings.TrimPrefix(rel, modPath+"/")
			}
			parts := strings.SplitN(rel, "/", d+1)
			var prefix string
			if len(parts) > d {
				prefix = strings.Join(parts[:d], "/")
			} else {
				prefix = strings.Join(parts, "/")
			}
			if !grouped[prefix] {
				ungrouped++
			}
		}
		count := len(groups) + ungrouped
		if count >= flatCount {
			break
		}
		if count < bestCount {
			bestCount = count
			bestDepth = d
		}
	}
	if bestDepth > 0 && bestCount < flatCount {
		fmt.Printf("Suggested grouping: --depth %d (%d components vs %d flat)\n", bestDepth, bestCount, flatCount)
	}
}

func detectProjectPath(root string) string {
	absRoot, _ := filepath.Abs(root)
	fallback := filepath.Base(absRoot)

	if data, err := os.ReadFile(filepath.Join(root, "go.mod")); err == nil {
		if f, err := modfile.Parse("go.mod", data, nil); err == nil {
			return f.Module.Mod.Path
		}
	}

	if data, err := os.ReadFile(filepath.Join(root, "Cargo.toml")); err == nil {
		if name := parseCargoProjectName(data); name != "" {
			return name
		}
	}

	if data, err := os.ReadFile(filepath.Join(root, "package.json")); err == nil {
		if name := parsePackageJSONName(data); name != "" {
			return name
		}
	}

	return fallback
}

func parseCargoProjectName(data []byte) string {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			}
		}
	}
	return ""
}

func parsePackageJSONName(data []byte) string {
	var pkg struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(data, &pkg) == nil {
		return pkg.Name
	}
	return ""
}

func resolveOutputDir(scanPath, outputDir string) string {
	if outputDir != "" {
		return outputDir
	}
	mosDir := filepath.Join(scanPath, names.MosDir)
	if _, err := os.Stat(mosDir); err == nil {
		return scanPath
	}
	absRoot, _ := filepath.Abs(scanPath)
	return filepath.Join(os.TempDir(), "mos-scan-"+filepath.Base(absRoot))
}

func inferDefaultGroups(proj *model.Project, modPath string, depth int) []artifact.ComponentGroup {
	prefixMap := make(map[string][]string)
	for _, ns := range proj.Namespaces {
		rel := ns.ImportPath
		if strings.HasPrefix(rel, modPath+"/") {
			rel = strings.TrimPrefix(rel, modPath+"/")
		}
		parts := strings.SplitN(rel, "/", depth+1)
		var prefix string
		if len(parts) > depth {
			prefix = strings.Join(parts[:depth], "/")
		} else {
			prefix = strings.Join(parts, "/")
		}
		prefixMap[prefix] = append(prefixMap[prefix], rel)
	}

	var groups []artifact.ComponentGroup
	for prefix, pkgs := range prefixMap {
		if len(pkgs) > 1 {
			groups = append(groups, artifact.ComponentGroup{Name: prefix, Packages: pkgs})
		}
	}
	return groups
}
