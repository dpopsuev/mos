// Package topology implements multi-repo governance operations:
// checkpoint/rollback, pre-flight delta, promote/demote, split/merge,
// and batch union/secede.
package topology

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dpopsuev/mos/moslib/names"
)

// Checkpoint tags the current state of the repo for potential rollback.
func Checkpoint(root, label string) (string, error) {
	tag := fmt.Sprintf("mos-checkpoint/%s/%d", label, time.Now().UnixNano())
	if err := gitCmd(root, "tag", tag); err != nil {
		return "", fmt.Errorf("creating checkpoint: %w", err)
	}
	return tag, nil
}

// Rollback restores the repo to the given checkpoint tag.
func Rollback(root, tag string) error {
	if err := gitCmd(root, "reset", "--hard", tag); err != nil {
		return fmt.Errorf("rolling back to %s: %w", tag, err)
	}
	return nil
}

// DeltaReport describes what changed between two repos or states.
type DeltaReport struct {
	Added    []string
	Modified []string
	Removed  []string
}

// IsEmpty returns true when there are no changes.
func (d *DeltaReport) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Modified) == 0 && len(d.Removed) == 0
}

// PreFlightDelta compares the .mos directory between two paths and reports changes.
func PreFlightDelta(source, target string) (*DeltaReport, error) {
	sourceMos := filepath.Join(source, names.MosDir)
	targetMos := filepath.Join(target, names.MosDir)

	report := &DeltaReport{}

	sourceFiles, err := walkMosFiles(sourceMos)
	if err != nil {
		return nil, fmt.Errorf("walking source: %w", err)
	}
	targetFiles, err := walkMosFiles(targetMos)
	if err != nil {
		return nil, fmt.Errorf("walking target: %w", err)
	}

	for relPath, srcContent := range sourceFiles {
		if tgtContent, ok := targetFiles[relPath]; !ok {
			report.Added = append(report.Added, relPath)
		} else if srcContent != tgtContent {
			report.Modified = append(report.Modified, relPath)
		}
	}

	for relPath := range targetFiles {
		if _, ok := sourceFiles[relPath]; !ok {
			report.Removed = append(report.Removed, relPath)
		}
	}

	return report, nil
}

// FormatDelta returns a human-readable summary.
func FormatDelta(d *DeltaReport) string {
	if d.IsEmpty() {
		return "No differences found.\n"
	}
	var b strings.Builder
	for _, f := range d.Added {
		fmt.Fprintf(&b, "+ %s\n", f)
	}
	for _, f := range d.Modified {
		fmt.Fprintf(&b, "~ %s\n", f)
	}
	for _, f := range d.Removed {
		fmt.Fprintf(&b, "- %s\n", f)
	}
	return b.String()
}

// Promote copies .mos governance artifacts from upstream into the target repo.
func Promote(upstreamRoot, targetRoot string) error {
	srcMos := filepath.Join(upstreamRoot, names.MosDir)
	tgtMos := filepath.Join(targetRoot, names.MosDir)

	for _, sub := range []string{"rules", "lexicon"} {
		src := filepath.Join(srcMos, sub)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		if err := copyDir(src, filepath.Join(tgtMos, sub)); err != nil {
			return fmt.Errorf("promoting %s: %w", sub, err)
		}
	}

	upstreamField := filepath.Join(tgtMos, "upstream.mos")
	content := fmt.Sprintf("upstream {\n  source = %q\n}\n", upstreamRoot)
	return os.WriteFile(upstreamField, []byte(content), names.FilePerm)
}

// Demote removes upstream metadata and inherited rules from a repo,
// making it standalone. Only removes artifacts that match the upstream.
func Demote(root string) error {
	mosDir := filepath.Join(root, names.MosDir)
	upstreamPath := filepath.Join(mosDir, "upstream.mos")
	if _, err := os.Stat(upstreamPath); err != nil {
		return fmt.Errorf("no upstream found; repo is already standalone")
	}
	return os.Remove(upstreamPath)
}

// SplitPlan describes how to split a repo into two.
type SplitPlan struct {
	SourceRoot string
	TargetRoot string
	Artifacts  []string // artifact IDs to move
}

// Split moves specified artifacts from the source repo to a new target.
func Split(plan SplitPlan) error {
	srcMos := filepath.Join(plan.SourceRoot, names.MosDir)
	tgtMos := filepath.Join(plan.TargetRoot, names.MosDir)

	if err := os.MkdirAll(tgtMos, names.DirPerm); err != nil {
		return err
	}

	for _, id := range plan.Artifacts {
		for _, scope := range []string{"contracts" + "/" + names.ActiveDir, "contracts" + "/" + names.ArchiveDir} {
			src := filepath.Join(srcMos, scope, id)
			if _, err := os.Stat(src); err != nil {
				continue
			}
			dst := filepath.Join(tgtMos, scope, id)
			if err := os.MkdirAll(filepath.Dir(dst), names.DirPerm); err != nil {
				return err
			}
			if err := os.Rename(src, dst); err != nil {
				return fmt.Errorf("moving %s: %w", id, err)
			}
		}
	}
	return nil
}

// Merge copies all .mos artifacts from source into target, skipping conflicts.
func Merge(sourceRoot, targetRoot string) error {
	srcMos := filepath.Join(sourceRoot, names.MosDir)
	tgtMos := filepath.Join(targetRoot, names.MosDir)

	files, err := walkMosFiles(srcMos)
	if err != nil {
		return fmt.Errorf("walking source: %w", err)
	}

	for relPath, content := range files {
		dst := filepath.Join(tgtMos, relPath)
		if _, err := os.Stat(dst); err == nil {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), names.DirPerm); err != nil {
			return err
		}
		if err := os.WriteFile(dst, []byte(content), names.FilePerm); err != nil {
			return err
		}
	}
	return nil
}

// Union batch-promotes multiple repos into a federation.
func Union(upstreamRoot string, targets []string) []error {
	var errs []error
	for _, t := range targets {
		if err := Promote(upstreamRoot, t); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", t, err))
		}
	}
	return errs
}

// Secede batch-demotes multiple repos from a federation.
func Secede(roots []string) []error {
	var errs []error
	for _, r := range roots {
		if err := Demote(r); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", r, err))
		}
	}
	return errs
}

func gitCmd(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func walkMosFiles(root string) (map[string]string, error) {
	files := make(map[string]string)
	if _, err := os.Stat(root); err != nil {
		return files, nil
	}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[rel] = string(data)
		return nil
	})
	return files, err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, names.DirPerm)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
