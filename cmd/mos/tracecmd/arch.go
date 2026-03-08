package tracecmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/moslib/arch"
	"github.com/dpopsuev/mos/moslib/names"
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
	report, err := arch.ScanAndBuild(p.scanPath, arch.ScanOpts{
		ScannerOverride: p.scannerName,
		ExcludeTests:    p.excludeTests,
		IncludeExternal: p.includeExternal,
		Grouped:         p.grouped,
		Depth:           p.depth,
		ChurnDays:       p.churnDays,
	})
	if err != nil {
		return err
	}

	archModel := report.Architecture

	outRoot := resolveOutputDir(p.scanPath, p.outputDir)
	mosContent := arch.RenderArchMos(archModel)
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
		mdContent := arch.RenderArchMarkdown(archModel)
		mdPath := filepath.Join(outRoot, "ARCHITECTURE.md")
		if err := os.WriteFile(mdPath, []byte(mdContent), 0o644); err != nil {
			return fmt.Errorf("write ARCHITECTURE.md: %w", err)
		}
		fmt.Printf("Wrote %s\n", mdPath)
	}

	fmt.Printf("Architecture: %d components, %d edges\n", len(archModel.Services), len(archModel.Edges))

	if !p.depthSet && !p.grouped && report.SuggestedDepth > 0 {
		fmt.Printf("Suggested grouping: --depth %d\n", report.SuggestedDepth)
	}

	return nil
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
