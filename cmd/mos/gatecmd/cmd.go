package gatecmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/cmd/mos/cliutil"
	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/governance/audit"
	"github.com/dpopsuev/mos/moslib/harness"
	"github.com/dpopsuev/mos/moslib/linter"
	"github.com/dpopsuev/mos/moslib/names"
)

var AuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Run audit checks on governance artifacts",
	RunE:  runAudit,
}

var (
	auditVerbose    bool
	auditTrend      bool
	auditVelocity   bool
	auditTrajectory bool
	auditSprint     string
	auditNoHarness  bool
	auditFormat     string
)

func init() {
	AuditCmd.Flags().BoolVarP(&auditVerbose, "verbose", "v", false, "Show detailed output")
	AuditCmd.Flags().BoolVar(&auditTrend, "trend", false, "Show integrity index trend sparkline")
	AuditCmd.Flags().BoolVar(&auditVelocity, "velocity", false, "Show velocity report")
	AuditCmd.Flags().BoolVar(&auditTrajectory, "trajectory", false, "Show per-axis trajectory analysis with sparklines")
	AuditCmd.Flags().StringVar(&auditSprint, "sprint", "", "Scope velocity to a sprint time window (sprint ID)")
	AuditCmd.Flags().BoolVar(&auditNoHarness, "no-harness", false, "Skip harness execution, use cached snapshots for integrity")
	AuditCmd.Flags().StringVar(&auditFormat, "format", "text", "Output format: text or json")
}

func runAudit(cmd *cobra.Command, args []string) error {
	report, err := audit.RunAudit(".", audit.AuditOpts{Verbose: auditVerbose, NoHarness: auditNoHarness})
	if err != nil {
		return err
	}

	switch auditFormat {
	case names.FormatJSON:
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case names.FormatText:
		fmt.Print(audit.FormatReport(report, auditVerbose))

		mosDir := filepath.Join(".", names.MosDir)

		if auditTrend || auditVelocity || auditTrajectory {
			snapshots, err := harness.LoadAllSnapshots(mosDir)
			if err != nil {
				return fmt.Errorf("loading snapshots: %w", err)
			}

			if auditSprint != "" && auditVelocity {
				from, to, ok := resolveSprintWindow(".", auditSprint)
				if ok {
					var filtered []harness.StateSnapshot
					for _, s := range snapshots {
						if (s.Timestamp.Equal(from) || s.Timestamp.After(from)) &&
							(s.Timestamp.Equal(to) || s.Timestamp.Before(to)) {
							filtered = append(filtered, s)
						}
					}
					snapshots = filtered
				}
			}

			if auditTrend {
				points := harness.ComputeTrend(snapshots)
				fmt.Print(harness.FormatTrendText(points))
			}

			if auditVelocity {
				vel := harness.ComputeVelocity(snapshots)
				fmt.Print(harness.FormatVelocityText(vel))
			}

			if auditTrajectory {
				trajReport := harness.AnalyzeTrajectory(snapshots, 0)
				fmt.Print(harness.FormatTrajectoryText(trajReport))
			}
		}
	default:
		return fmt.Errorf("unknown format %q", auditFormat)
	}

	if report.LintErrors > 0 || len(report.Collisions) > 0 {
		return cliutil.ErrNonZeroExit
	}
	return nil
}

func resolveSprintWindow(root, sprintID string) (from, to time.Time, ok bool) {
	mosDir := filepath.Join(root, names.MosDir)
	path := filepath.Join(mosDir, names.DirSprints, names.ActiveDir, sprintID, "sprint.mos")
	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}

	startStr := artifact.FieldStr(ab.Items, "start_date")
	endStr := artifact.FieldStr(ab.Items, "end_date")
	if startStr == "" || endStr == "" {
		return time.Time{}, time.Time{}, false
	}

	start, err1 := time.Parse("2006-01-02", startStr)
	end, err2 := time.Parse("2006-01-02", endStr)
	if err1 != nil || err2 != nil {
		return time.Time{}, time.Time{}, false
	}
	return start, end.Add(24*time.Hour - time.Second), true
}

var LintCmd = &cobra.Command{
	Use:   "lint [path]",
	Short: "Run linter checks",
	RunE:  runLint,
}

var (
	lintFormat   string
	lintSeverity string
	lintSummary  bool
	lintVerbose  bool
	lintChanged  bool
	lintFiles    string
	lintNewOnly  bool
	lintAgent    bool
)

func init() {
	LintCmd.Flags().StringVar(&lintFormat, "format", "text", "Output format: text or json")
	LintCmd.Flags().StringVar(&lintSeverity, "severity", "", "Filter by severity: error, warning, or info")
	LintCmd.Flags().BoolVar(&lintSummary, "summary", false, "Show aggregated counts by severity and rule")
	LintCmd.Flags().BoolVarP(&lintVerbose, "verbose", "v", false, "Show sample diagnostics per rule (with --summary)")
	LintCmd.Flags().BoolVar(&lintChanged, "changed", false, "Lint only VCS-modified files")
	LintCmd.Flags().StringVar(&lintFiles, "files", "", "Comma-separated list of files to lint")
	LintCmd.Flags().BoolVar(&lintNewOnly, "new-only", false, "Suppress pre-existing diagnostics (use with --changed)")
	LintCmd.Flags().BoolVar(&lintAgent, "agent", false, "Agent-friendly output (implies --summary -v --format json --severity error)")
}

func runLint(cmd *cobra.Command, args []string) error {
	if lintAgent {
		lintSummary = true
		lintVerbose = true
		lintFormat = "json"
		lintSeverity = "error"
	}

	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	l := &linter.Linter{}

	var diags []linter.Diagnostic
	var err error

	switch {
	case lintFiles != "":
		files := strings.Split(lintFiles, ",")
		for i := range files {
			files[i] = strings.TrimSpace(files[i])
		}
		diags, err = l.LintFiles(path, files)

	case lintChanged:
		changed, detectErr := linter.DetectChangedFiles(path)
		if detectErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "mos lint: warning: VCS detection failed: %v\n", detectErr)
		}
		if changed == nil {
			if detectErr == nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "mos lint: warning: VCS unavailable, falling back to full lint\n")
			}
			diags, err = l.Lint(path)
		} else {
			diags, err = l.LintFiles(path, changed)
		}

	default:
		diags, err = l.Lint(path)
	}

	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "mos lint: internal error: %v\n", err)
		return cliutil.ErrInternalLint
	}

	if lintNewOnly && (lintChanged || lintFiles != "") {
		baseline, baseErr := l.Lint(path)
		if baseErr == nil {
			diags = linter.FilterNewOnly(diags, baseline)
		}
	}

	if lintSeverity != "" {
		diags = filterBySeverity(diags, lintSeverity)
	}

	if lintSummary {
		return renderSummary(diags, lintFormat, lintVerbose)
	}

	switch lintFormat {
	case names.FormatJSON:
		data, err := json.MarshalIndent(diags, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case names.FormatText:
		for _, d := range diags {
			fmt.Printf("%s: %s [%s] %s\n", d.File, d.Severity, d.Rule, d.Message)
		}
	default:
		return fmt.Errorf("unknown format %q", lintFormat)
	}

	for _, d := range diags {
		if d.Severity == linter.SeverityError {
			return cliutil.ErrNonZeroExit
		}
	}
	return nil
}

func filterBySeverity(diags []linter.Diagnostic, severity string) []linter.Diagnostic {
	target := strings.ToLower(severity)
	var filtered []linter.Diagnostic
	for _, d := range diags {
		if d.Severity.String() == target {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

type summarySample struct {
	File            string `json:"file"`
	Message         string `json:"message"`
	ArtifactID      string `json:"artifact_id,omitempty"`
	SuggestedAction string `json:"suggested_action,omitempty"`
}

type summaryEntry struct {
	Severity string          `json:"severity"`
	Rule     string          `json:"rule"`
	Count    int             `json:"count"`
	Samples  []summarySample `json:"samples,omitempty"`
}

const maxSamples = 3

func renderSummary(diags []linter.Diagnostic, format string, verbose bool) error {
	type key struct{ sev, rule string }
	counts := map[key]int{}
	samples := map[key][]summarySample{}
	for _, d := range diags {
		k := key{d.Severity.String(), d.Rule}
		counts[k]++
		if verbose && len(samples[k]) < maxSamples {
			samples[k] = append(samples[k], summarySample{
				File:            d.File,
				Message:         d.Message,
				ArtifactID:      d.ArtifactID,
				SuggestedAction: d.SuggestedAction,
			})
		}
	}

	var entries []summaryEntry
	for k, c := range counts {
		e := summaryEntry{Severity: k.sev, Rule: k.rule, Count: c}
		if verbose {
			e.Samples = samples[k]
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Severity != entries[j].Severity {
			return severityRank(entries[i].Severity) < severityRank(entries[j].Severity)
		}
		return entries[i].Rule < entries[j].Rule
	})

	switch format {
	case names.FormatJSON:
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case names.FormatText:
		if len(entries) == 0 {
			fmt.Println("No diagnostics.")
			return nil
		}
		for _, e := range entries {
			fmt.Printf("%-8s %-30s %d\n", e.Severity, e.Rule, e.Count)
			for _, s := range e.Samples {
				fmt.Printf("           %s: %s\n", s.ArtifactID, s.Message)
				if s.SuggestedAction != "" {
					fmt.Printf("           -> %s\n", s.SuggestedAction)
				}
			}
		}
	default:
		return fmt.Errorf("unknown format %q", format)
	}

	for _, d := range diags {
		if d.Severity == linter.SeverityError {
			return cliutil.ErrNonZeroExit
		}
	}
	return nil
}

func severityRank(s string) int {
	switch s {
	case "error":
		return 0
	case "warning":
		return 1
	case "info":
		return 2
	default:
		return 3
	}
}

var HarnessCmd = &cobra.Command{
	Use:   "harness",
	Short: "Run test harnesses",
}

var HarnessRunCmd = &cobra.Command{
	Use:   "run [path]",
	Short: "Run test harnesses",
	RunE:  runHarness,
}

var (
	harnessFormat      string
	harnessRuleIDs     []string
	harnessEnforcement string
)

func init() {
	HarnessRunCmd.Flags().StringVar(&harnessFormat, "format", "text", "Output format: text or json")
	HarnessRunCmd.Flags().StringSliceVar(&harnessRuleIDs, "rule", nil, "Run only the specified rule(s)")
	HarnessRunCmd.Flags().StringVar(&harnessEnforcement, "enforcement", "", "Filter by enforcement level")
	HarnessCmd.AddCommand(HarnessRunCmd)
}

func runHarness(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	mosDir := filepath.Join(path, names.MosDir)
	if filepath.Base(path) == names.MosDir {
		mosDir = path
		path = filepath.Dir(path)
	}

	specs, err := harness.Discover(mosDir)
	if err != nil {
		return err
	}
	specs = harness.Filter(specs, harnessRuleIDs, harnessEnforcement)
	results := harness.Run(path, specs)

	switch harnessFormat {
	case names.FormatJSON:
		data, err := harness.FormatJSON(results)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case names.FormatText:
		fmt.Print(harness.FormatText(results))
	default:
		return fmt.Errorf("unknown format %q", harnessFormat)
	}

	for _, ev := range results {
		if !ev.Pass {
			return cliutil.ErrNonZeroExit
		}
	}
	return nil
}
