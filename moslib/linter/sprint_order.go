package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
)

type sprintInfo struct {
	ID     string
	Seq    int
	Status string
	File   string
}

// validateSprintOrdering checks that no higher-numbered sprint is closed while
// a lower-numbered non-cancelled sprint is still planned. This enforces linear
// sprint execution order.
func validateSprintOrdering(ctx *ProjectContext) []Diagnostic {
	sprints := discoverSprints(ctx)
	if len(sprints) < 2 {
		return nil
	}

	sort.Slice(sprints, func(i, j int) bool {
		return sprints[i].Seq < sprints[j].Seq
	})

	terminalStates := make(map[string]bool)
	if sch := findCustomSchema(ctx, "sprint"); sch != nil {
		for _, s := range sch.ArchiveStates {
			terminalStates[s] = true
		}
	}
	if len(terminalStates) == 0 {
		for _, s := range []string{"complete", "closed", "cancelled", "abandoned", "duplicate"} {
			terminalStates[s] = true
		}
	}

	var diags []Diagnostic
	for i, s := range sprints {
		if !terminalStates[s.Status] {
			continue
		}
		for j := 0; j < i; j++ {
			lower := sprints[j]
			if terminalStates[lower.Status] {
				continue
			}
			diags = append(diags, Diagnostic{
				File:     s.File,
				Severity: SeverityError,
				Rule:     "sprint-order",
				Message:  fmt.Sprintf("%s is closed but lower-numbered %s has status %q — sprints must close in sequential order", s.ID, lower.ID, lower.Status),
			})
		}
	}
	return diags
}

func discoverSprints(ctx *ProjectContext) []sprintInfo {
	var sprints []sprintInfo

	for _, sub := range []string{"active", "archive"} {
		dir := filepath.Join(ctx.Root, "sprints", sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(dir, e.Name(), "sprint.mos")
			f, err := parseDSLFile(path, ctx.Keywords)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			status, _ := dsl.FieldString(ab.Items, "status")
			seq := parseSprintSeq(ab.Name)
			if seq < 0 {
				continue
			}
			sprints = append(sprints, sprintInfo{
				ID:     ab.Name,
				Seq:    seq,
				Status: status,
				File:   path,
			})
		}
	}
	return sprints
}

// parseSprintSeq extracts the trailing numeric sequence from a sprint ID
// like "SPR-2026-033" -> 2026033. Uses the full numeric suffix to ensure
// cross-year ordering works correctly.
func parseSprintSeq(id string) int {
	parts := strings.Split(id, "-")
	if len(parts) < 3 || parts[0] != "SPR" {
		return -1
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1
	}
	seq, err := strconv.Atoi(parts[2])
	if err != nil {
		return -1
	}
	return year*1000 + seq
}
