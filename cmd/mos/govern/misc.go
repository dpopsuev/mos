package govern

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/governance/audit"
	"github.com/dpopsuev/mos/moslib/harness"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/registry"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project orientation (sprint, contracts, lint, harness)",
	RunE:  runStatus,
}

var statusFormat string

func init() {
	StatusCmd.Flags().StringVar(&statusFormat, "format", "text", "Output format: text or json")
}

func runStatus(cmd *cobra.Command, args []string) error {
	report, err := audit.RunAudit(".", audit.AuditOpts{})
	if err != nil {
		return err
	}

	if statusFormat == names.FormatJSON {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Println("=== Project Status ===")

	reg, err := registry.LoadRegistry(".")
	if err == nil {
		if td, ok := reg.Types["directive"]; ok {
			directives, _ := artifact.GenericList(".", td, "")
			for _, d := range directives {
				if d.Status == "active" || d.Status == "declared" {
					text := directiveText(d.Path)
					if text != "" {
						fmt.Printf("\n  DIRECTIVE: %s\n", text)
					} else {
						fmt.Printf("\n  DIRECTIVE: %s\n", d.Title)
					}
					break
				}
			}
		}
	}

	if len(report.SprintStatus) > 0 {
		fmt.Println()
		for _, s := range report.SprintStatus {
			fmt.Printf("Sprint: %s", s.ID)
			if s.Title != "" {
				fmt.Printf(" — %s", s.Title)
			}
			fmt.Println()
			fmt.Printf("  Progress: %d/%d complete\n", s.Complete, s.Total)
		}
	} else {
		fmt.Println("\nNo active sprint.")
	}

	contracts, err := artifact.ListContracts(".", artifact.ListOpts{Status: names.StatusDraft})
	if err == nil && len(contracts) > 0 {
		fmt.Printf("\nOpen contracts: %d\n", len(contracts))
		for _, c := range contracts {
			fmt.Printf("  %-20s %-12s %s\n", c.ID, c.Status, c.Title)
		}
	}

	fmt.Printf("\nLint: %d errors, %d warnings, %d info\n",
		report.LintErrors, report.LintWarnings, report.LintInfos)

	mosDir := filepath.Join(".", names.MosDir)
	specs, hErr := harness.Discover(mosDir)
	if hErr == nil && len(specs) > 0 {
		fmt.Printf("Harness: %d rule(s) defined (run `mgate harness run` to execute)\n", len(specs))
	}

	if reg != nil {
		if td, ok := reg.Types["watch"]; ok {
			watches, _ := artifact.GenericList(".", td, "active")
			if len(watches) > 0 {
				fmt.Printf("\nWatches: %d active\n", len(watches))
				for _, w := range watches {
					fmt.Printf("  %-20s %s\n", w.ID, w.Title)
				}
			}
		}

		triggers, _ := artifact.EvaluateWatchTriggers(".", time.Now())
		for _, tr := range triggers {
			fmt.Printf("  ! %s [%s]\n", tr.Message, tr.Action)
		}
	}

	if len(report.OrphanContracts) > 0 {
		fmt.Printf("\nOrphan contracts: %d\n", len(report.OrphanContracts))
	}
	return nil
}

var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Compute and apply schema migrations",
	RunE:  runMigrate,
}

var migrateApply bool

func init() {
	MigrateCmd.Flags().BoolVar(&migrateApply, "apply", false, "Apply the computed migration")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	reg, err := registry.LoadRegistry(".")
	if err != nil {
		return err
	}
	diffs, err := artifact.ComputeMigration(".", reg)
	if err != nil {
		return err
	}
	if !migrateApply {
		fmt.Print(artifact.FormatMigrationPlan(diffs))
		return nil
	}
	count, err := artifact.ApplyMigration(diffs)
	if err != nil {
		return err
	}
	fmt.Printf("Migrated %d instance(s).\n", count)
	return nil
}

var InitProjectCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new .mos project",
	RunE:  runInitProject,
}

var (
	initModel   string
	initScope   string
	initName    string
	initPurpose string
)

func init() {
	InitProjectCmd.Flags().StringVar(&initModel, "model", "", "Governance model (e.g. bdfl, council)")
	InitProjectCmd.Flags().StringVar(&initScope, "scope", "", "Project scope")
	InitProjectCmd.Flags().StringVar(&initName, "name", "", "Project name")
	InitProjectCmd.Flags().StringVar(&initPurpose, "purpose", "", "Project purpose statement")
}

func runInitProject(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}
	opts := artifact.InitOpts{
		Model:   initModel,
		Scope:   initScope,
		Name:    initName,
		Purpose: initPurpose,
	}
	if err := artifact.Init(path, opts); err != nil {
		return err
	}
	fmt.Printf("Initialized %s/ in %s\n", names.MosDir, path)
	return nil
}

var FmtCmd = &cobra.Command{
	Use:   "fmt [path]",
	Short: "Format .mos files",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		var files []string
		if info.IsDir() {
			err = filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !fi.IsDir() && strings.HasSuffix(fi.Name(), names.MosDir) {
					files = append(files, p)
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			files = []string{path}
		}

		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "reading %s: %v\n", f, err)
				continue
			}
			parsed, err := dsl.Parse(string(data), nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "parsing %s: %v\n", f, err)
				continue
			}
			formatted := dsl.Format(parsed, nil)
			if string(data) != formatted {
				if err := os.WriteFile(f, []byte(formatted), names.FilePerm); err != nil {
					fmt.Fprintf(os.Stderr, "writing %s: %v\n", f, err)
					continue
				}
				fmt.Printf("formatted %s\n", f)
			}
		}

		relocations, err := artifact.RelocateMisplacedArtifacts(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "auto-archive: %v\n", err)
		}
		for _, r := range relocations {
			fmt.Printf("relocated %s %s: %s → %s\n", r.Kind, r.ID, r.From, r.To)
		}
		return nil
	},
}

var reclassifyTo string
var reclassifyFormat string

var ReclassifyCmd = &cobra.Command{
	Use:   "reclassify <id>",
	Short: "Change an artifact's kind (e.g. contract -> specification)",
	Long:  "Reclassify an artifact from one kind to another, preserving all content.\nThe original location receives a tombstone pointing to the new ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if reclassifyTo == "" {
			return fmt.Errorf("--to is required")
		}
		result, err := artifact.Reclassify(".", args[0], reclassifyTo)
		if err != nil {
			return err
		}
		if reclassifyFormat == names.FormatJSON {
			data, _ := json.MarshalIndent(map[string]string{
				"old_id":   result.OldID,
				"new_id":   result.NewID,
				"old_kind": result.OldKind,
				"new_kind": result.NewKind,
				"path":     result.NewPath,
			}, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Reclassified %s %s → %s %s\n", result.OldKind, result.OldID, result.NewKind, result.NewID)
			fmt.Printf("  New path: %s\n", result.NewPath)
		}
		return nil
	},
}

func init() {
	ReclassifyCmd.Flags().StringVar(&reclassifyTo, "to", "", "Target artifact kind")
	ReclassifyCmd.Flags().StringVar(&reclassifyFormat, "format", "text", "Output format: text or json")
}

func directiveText(path string) string {
	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return ""
	}
	text, _ := dsl.FieldString(ab.Items, "text")
	return text
}
