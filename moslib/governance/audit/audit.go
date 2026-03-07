package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dpopsuev/mos/moslib/arch"
	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/harness"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/registry"
)

// Report holds the aggregated results of a full project audit.
type Report struct {
	LintErrors   int            `json:"lint_errors"`
	LintWarnings int            `json:"lint_warnings"`
	LintInfos    int            `json:"lint_infos"`
	LintByRule   map[string]int `json:"lint_by_rule,omitempty"`

	Collisions []string `json:"collisions,omitempty"`

	MigrationsPending int `json:"migrations_pending"`

	UncoveredCriteria map[string][]string `json:"uncovered_criteria,omitempty"`

	SprintStatus []SprintSummary `json:"sprint_status,omitempty"`

	OrphanContracts []string `json:"orphan_contracts,omitempty"`

	ArchViolations  []string `json:"arch_violations,omitempty"`
	DriftViolations []string `json:"drift_violations,omitempty"`

	IntegrityScore float64                `json:"integrity_score"`
	VectorScores   []harness.VectorResult `json:"vector_scores,omitempty"`
}

// SprintSummary shows completion ratio for a sprint.
type SprintSummary struct {
	ID       string `json:"id"`
	Title    string `json:"title,omitempty"`
	Total    int    `json:"total"`
	Complete int    `json:"complete"`
}

// AuditOpts controls audit behavior.
type AuditOpts struct {
	Verbose   bool
	NoHarness bool
}

// RunAudit aggregates lint, collisions, migration, coverage, and sprint status.
func RunAudit(root string, opts AuditOpts) (*Report, error) {
	report := &Report{
		LintByRule:        make(map[string]int),
		UncoveredCriteria: make(map[string][]string),
	}

	if artifact.LintAll != nil {
		diags, err := artifact.LintAll(root)
		if err != nil {
			return nil, fmt.Errorf("running lint: %w", err)
		}
		for _, d := range diags {
			switch d.Severity {
			case "error":
				report.LintErrors++
			case "warning":
				report.LintWarnings++
			case "info":
				report.LintInfos++
			}
			report.LintByRule[d.Rule]++
			if d.Rule == "id-collision" {
				report.Collisions = append(report.Collisions, d.Message)
			}
			if d.Rule == "criterion-coverage" {
				report.UncoveredCriteria[d.File] = append(report.UncoveredCriteria[d.File], d.Message)
			}
		}
	}

	reg, err := registry.LoadRegistry(root)
	if err == nil {
		diffs, err := artifact.ComputeMigration(root, reg)
		if err == nil {
			report.MigrationsPending = len(diffs)
		}
	}

	report.SprintStatus = computeSprintStatus(root)

	report.OrphanContracts = findOrphanContracts(root)

	report.ArchViolations = checkArchViolations(root)

	report.DriftViolations = checkDriftViolations(root)

	mosDir := filepath.Join(root, names.MosDir)
	if opts.NoHarness {
		snapshots, err := harness.LoadSnapshots(mosDir)
		if err == nil && len(snapshots) > 0 {
			latest := snapshots[len(snapshots)-1]
			report.IntegrityScore = latest.IntegrityScore
			report.VectorScores = latest.Vectors
		} else {
			report.IntegrityScore = -1
		}
	} else {
		idx, err := harness.ComputeIntegrityIndex(root, mosDir)
		if err == nil {
			report.IntegrityScore = idx.Score
			report.VectorScores = idx.Vectors

			snap := harness.SnapshotFromIndex(idx, report.LintErrors, report.LintWarnings)
			_ = harness.StoreSnapshot(mosDir, snap)
		}
	}

	return report, nil
}

func checkArchViolations(root string) []string {
	if artifact.ScanProject == nil {
		return nil
	}

	declared := loadDeclaredArchModels(root)
	if len(declared) == 0 {
		return nil
	}

	hasForbidden := false
	for _, d := range declared {
		if len(d.Forbidden) > 0 {
			hasForbidden = true
			break
		}
	}
	if !hasForbidden {
		return nil
	}

	proj, err := artifact.ScanProject(root)
	if err != nil {
		return nil
	}

	modPath := proj.Path
	groups, _ := artifact.LoadComponentGroups(root)
	live := artifact.ProjectToArchModel(proj, artifact.SyncOptions{
		ModulePath:   modPath,
		ExcludeTests: true,
		Groups:       groups,
	})

	var allViolations []string
	for _, d := range declared {
		violations := artifact.CheckForbiddenEdges(live, d)
		allViolations = append(allViolations, violations...)
	}
	return allViolations
}

func loadDeclaredArchModels(root string) []artifact.ArchModel {
	mosDir := filepath.Join(root, names.MosDir)
	var models []artifact.ArchModel

	for _, sub := range []string{names.ActiveDir} {
		dir := filepath.Join(mosDir, names.DirArchitectures, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(dir, e.Name(), "architecture.mos")
			ab, err := dsl.ReadArtifact(path)
			if err != nil {
				continue
			}
			models = append(models, artifact.ParseArchModel(ab))
		}
	}
	return models
}

func computeSprintStatus(root string) []SprintSummary {
	mosDir := filepath.Join(root, names.MosDir)
	var summaries []SprintSummary

	for _, sub := range []string{names.ActiveDir} {
		dir := filepath.Join(mosDir, names.DirSprints, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			sprintID := e.Name()
			path := filepath.Join(dir, sprintID, "sprint.mos")
			ab, err := dsl.ReadArtifact(path)
			if err != nil {
				continue
			}
			title := artifact.FieldStr(ab.Items, names.FieldTitle)

			members := collectSprintMembers(root, sprintID, ab)

			total := len(members)
			complete := 0
			for _, m := range members {
				if m.Status == names.StatusComplete {
					complete++
				}
			}
			summaries = append(summaries, SprintSummary{
				ID:       sprintID,
				Title:    title,
				Total:    total,
				Complete: complete,
			})
		}
	}
	return summaries
}

func collectSprintMembers(root, sprintID string, sprintAB *dsl.ArtifactBlock) []artifact.QueryResult {
	seen := make(map[string]artifact.QueryResult)

	results, err := artifact.QueryArtifacts(root, artifact.QueryOpts{References: sprintID})
	if err == nil {
		for _, r := range results {
			seen[r.ID] = r
		}
	}

	contractsList := artifact.FieldStr(sprintAB.Items, "contracts")
	if contractsList != "" {
		for _, cid := range strings.Split(contractsList, ",") {
			cid = strings.TrimSpace(cid)
			if cid == "" || seen[cid].ID != "" {
				continue
			}
			cpath, err := artifact.FindContractPath(root, cid)
			if err != nil {
				continue
			}
			cab, err := dsl.ReadArtifact(cpath)
			if err != nil {
				continue
			}
			seen[cid] = artifact.QueryResult{
				ID:     cid,
				Kind:   names.KindContract,
				Status: artifact.FieldStr(cab.Items, names.FieldStatus),
				Sprint: sprintID,
				Path:   cpath,
			}
		}
	}

	batches, err := artifact.QueryArtifacts(root, artifact.QueryOpts{Kind: names.KindBatch, References: sprintID})
	if err == nil {
		for _, batch := range batches {
			batchContracts, err := artifact.QueryArtifacts(root, artifact.QueryOpts{References: batch.ID})
			if err != nil {
				continue
			}
			for _, bc := range batchContracts {
				if seen[bc.ID].ID == "" {
					seen[bc.ID] = bc
				}
			}
		}
	}

	members := make([]artifact.QueryResult, 0, len(seen))
	for _, m := range seen {
		members = append(members, m)
	}
	return members
}

func checkDriftViolations(root string) []string {
	if artifact.ScanProject == nil {
		return nil
	}

	desired, err := arch.LoadDesiredArch(root)
	if err != nil {
		return nil
	}
	if len(desired.Layers) == 0 {
		return nil
	}

	proj, err := artifact.ScanProject(root)
	if err != nil {
		return nil
	}

	groups, _ := artifact.LoadComponentGroups(root)
	live := artifact.ProjectToArchModel(proj, artifact.SyncOptions{
		ModulePath:   proj.Path,
		ExcludeTests: true,
		Groups:       groups,
	})

	driftResults := arch.DetectDrift(live, *desired)
	var violations []string
	for _, v := range driftResults {
		violations = append(violations, fmt.Sprintf("[%s] %s -> %s: %s", v.Rule, v.From, v.To, v.Message))
	}
	return violations
}

func findOrphanContracts(root string) []string {
	reg, err := registry.LoadRegistry(root)
	if err != nil {
		return nil
	}
	contractTD := reg.Types[names.KindContract]
	linkFields := contractTD.LinkFields()
	if len(linkFields) == 0 {
		linkFields = []string{"parent", "justifies", "depends_on", "batch", "sprint"}
	}

	mosDir := filepath.Join(root, names.MosDir)
	dir := filepath.Join(mosDir, names.DirContracts, names.ActiveDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var orphans []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(dir, e.Name(), "contract.mos")
		ab, err := dsl.ReadArtifact(path)
		if err != nil {
			continue
		}
		hasLink := false
		for _, lf := range linkFields {
			if artifact.FieldStr(ab.Items, lf) != "" {
				hasLink = true
				break
			}
		}
		if !hasLink {
			orphans = append(orphans, e.Name())
		}
	}
	sort.Strings(orphans)
	return orphans
}

// FormatReport formats the audit report for terminal output.
func FormatReport(r *Report, verbose bool) string {
	var sb strings.Builder

	sb.WriteString("=== Audit Report ===\n")

	fmt.Fprintf(&sb, "Lint: %d errors, %d warnings, %d info\n",
		r.LintErrors, r.LintWarnings, r.LintInfos)

	if verbose && len(r.LintByRule) > 0 {
		rules := make([]string, 0, len(r.LintByRule))
		for k := range r.LintByRule {
			rules = append(rules, k)
		}
		sort.Strings(rules)
		for _, rule := range rules {
			fmt.Fprintf(&sb, "  %s: %d\n", rule, r.LintByRule[rule])
		}
	}

	if len(r.Collisions) > 0 {
		fmt.Fprintf(&sb, "ID Collisions: %d\n", len(r.Collisions))
		if verbose {
			for _, c := range r.Collisions {
				fmt.Fprintf(&sb, "  %s\n", c)
			}
		}
	} else {
		sb.WriteString("ID Collisions: 0\n")
	}

	fmt.Fprintf(&sb, "Migrations: %d pending\n", r.MigrationsPending)

	if len(r.UncoveredCriteria) > 0 {
		total := 0
		for _, msgs := range r.UncoveredCriteria {
			total += len(msgs)
		}
		fmt.Fprintf(&sb, "Coverage: %d uncovered criteria\n", total)
		if verbose {
			for file, msgs := range r.UncoveredCriteria {
				fmt.Fprintf(&sb, "  %s:\n", file)
				for _, msg := range msgs {
					fmt.Fprintf(&sb, "    %s\n", msg)
				}
			}
		}
	} else {
		sb.WriteString("Coverage: all criteria covered\n")
	}

	for _, s := range r.SprintStatus {
		fmt.Fprintf(&sb, "Sprint %s: %d/%d complete", s.ID, s.Complete, s.Total)
		if s.Title != "" {
			fmt.Fprintf(&sb, " (%s)", s.Title)
		}
		sb.WriteString("\n")
	}

	if len(r.OrphanContracts) > 0 {
		fmt.Fprintf(&sb, "Orphan contracts: %d\n", len(r.OrphanContracts))
		if verbose {
			for _, id := range r.OrphanContracts {
				fmt.Fprintf(&sb, "  %s\n", id)
			}
		}
	}

	if len(r.ArchViolations) > 0 {
		fmt.Fprintf(&sb, "Architecture violations: %d\n", len(r.ArchViolations))
		for _, v := range r.ArchViolations {
			fmt.Fprintf(&sb, "  FORBIDDEN: %s\n", v)
		}
	} else {
		sb.WriteString("Architecture violations: 0\n")
	}

	if len(r.DriftViolations) > 0 {
		fmt.Fprintf(&sb, "Structural drift: %d violation(s)\n", len(r.DriftViolations))
		for _, v := range r.DriftViolations {
			fmt.Fprintf(&sb, "  DRIFT: %s\n", v)
		}
	} else {
		sb.WriteString("Structural drift: 0\n")
	}

	if r.IntegrityScore == -1 {
		sb.WriteString("Integrity Index: N/A (run full audit first)\n")
	} else if len(r.VectorScores) > 0 {
		idx := &harness.IntegrityIndex{Score: r.IntegrityScore, Vectors: r.VectorScores}
		sb.WriteString(harness.FormatIntegrityText(idx))
		sb.WriteString("\n")
	}

	return sb.String()
}
