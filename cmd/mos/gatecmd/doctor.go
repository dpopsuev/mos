package gatecmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/cmd/mos/cliutil"
	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/governance/audit"
	"github.com/dpopsuev/mos/moslib/harness"
	"github.com/dpopsuev/mos/moslib/linter"
	"github.com/dpopsuev/mos/moslib/names"
)

var DoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run a comprehensive health check on the project",
	Long: `Aggregates lint errors, harness failures, stale sprint refs, orphan
contracts, watch triggers, and actionable recommendations into one report.`,
	RunE: runDoctor,
}

var doctorFormat string

func init() {
	DoctorCmd.Flags().StringVar(&doctorFormat, "format", "text", "Output format: text or json")
}

type doctorReport struct {
	LintErrors   int              `json:"lint_errors"`
	LintWarnings int              `json:"lint_warnings"`
	LintInfos    int              `json:"lint_infos"`
	Orphans      []string         `json:"orphans,omitempty"`
	HarnessSpecs int              `json:"harness_specs"`
	HarnessFails int              `json:"harness_fails"`
	WatchAlerts  int              `json:"watch_alerts"`
	Sprints      []sprintHealth   `json:"sprints,omitempty"`
	Recs         []string         `json:"recommendations,omitempty"`
	Healthy      bool             `json:"healthy"`
}

type sprintHealth struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Complete int    `json:"complete"`
	Total    int    `json:"total"`
}

func runDoctor(cmd *cobra.Command, args []string) error {
	report := doctorReport{Healthy: true}

	// 1. Lint
	l := &linter.Linter{}
	diags, err := l.Lint(".")
	if err == nil {
		for _, d := range diags {
			switch d.Severity {
			case linter.SeverityError:
				report.LintErrors++
			case linter.SeverityWarning:
				report.LintWarnings++
			case linter.SeverityInfo:
				report.LintInfos++
			}
		}
	}

	// 2. Audit (orphans, sprints)
	auditReport, err := audit.RunAudit(".", audit.AuditOpts{})
	if err == nil {
		report.Orphans = auditReport.OrphanContracts
		for _, ss := range auditReport.SprintStatus {
			report.Sprints = append(report.Sprints, sprintHealth{
				ID: ss.ID, Title: ss.Title, Complete: ss.Complete, Total: ss.Total,
			})
		}
	}

	// 3. Harness (discover only, don't run — doctor is a quick check)
	mosDir := filepath.Join(".", names.MosDir)
	specs, err := harness.Discover(mosDir)
	if err == nil {
		report.HarnessSpecs = len(specs)
	}

	// 4. Watch triggers
	triggers, _ := artifact.EvaluateWatchTriggers(".", time.Now())
	report.WatchAlerts = len(triggers)

	// 5. Recommendations
	if report.LintErrors > 0 {
		report.Recs = append(report.Recs, fmt.Sprintf("Run `mos lint` to see %d error(s)", report.LintErrors))
		report.Healthy = false
	}
	if len(report.Orphans) > 0 {
		report.Recs = append(report.Recs, fmt.Sprintf("%d orphan contract(s) — consider linking or archiving", len(report.Orphans)))
	}
	if report.WatchAlerts > 0 {
		report.Recs = append(report.Recs, fmt.Sprintf("%d watch alert(s) — run `mos status` for details", report.WatchAlerts))
	}
	if report.HarnessSpecs > 0 {
		report.Recs = append(report.Recs, fmt.Sprintf("%d harness rule(s) defined — run `mos harness run` to verify", report.HarnessSpecs))
	}

	// Count completed artifacts in active dirs
	terminalCount := 0
	for _, d := range diags {
		if d.Rule == "terminal-in-active" {
			terminalCount++
		}
	}
	if terminalCount > 0 {
		report.Recs = append(report.Recs, fmt.Sprintf("Run `mos fmt` to archive %d completed artifact(s)", terminalCount))
	}

	if doctorFormat == "json" {
		data, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(data))
		if !report.Healthy {
			return cliutil.ErrNonZeroExit
		}
		return nil
	}

	// Text output
	fmt.Println("mos doctor")
	fmt.Println(strings.Repeat("─", 40))

	fmt.Printf("Lint:     %d error, %d warning, %d info\n", report.LintErrors, report.LintWarnings, report.LintInfos)
	fmt.Printf("Orphans:  %d\n", len(report.Orphans))
	fmt.Printf("Harness:  %d rule(s) defined\n", report.HarnessSpecs)
	fmt.Printf("Watches:  %d alert(s)\n", report.WatchAlerts)

	if len(report.Sprints) > 0 {
		fmt.Println()
		fmt.Println("Sprints:")
		for _, s := range report.Sprints {
			fmt.Printf("  %-20s %d/%d complete  %s\n", s.ID, s.Complete, s.Total, s.Title)
		}
	}

	if len(report.Recs) > 0 {
		fmt.Println()
		fmt.Println("Recommendations:")
		for _, r := range report.Recs {
			fmt.Printf("  - %s\n", r)
		}
	}

	verdict := "HEALTHY"
	if !report.Healthy {
		verdict = "NEEDS ATTENTION"
	}
	fmt.Printf("\nVerdict: %s\n", verdict)

	if !report.Healthy {
		return cliutil.ErrNonZeroExit
	}
	return nil
}
