package staging

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dpopsuev/mos/moslib/vcs"
)

type CleanOpts struct {
	DryRun bool
	Force  bool
	Dirs   bool
}

func Clean(repo *vcs.Repository, opts CleanOpts) ([]string, error) {
	if !opts.DryRun && !opts.Force {
		return nil, fmt.Errorf("clean requires --force (or --dry-run to preview)")
	}

	idx, err := LoadIndex(repo.Root)
	if err != nil {
		return nil, fmt.Errorf("load index: %w", err)
	}
	indexMap := IndexToMap(idx)

	workEntries, err := SnapshotWorkingTree(repo.Root, repo.Store)
	if err != nil {
		return nil, fmt.Errorf("snapshot working tree: %w", err)
	}

	var untracked []string
	for _, e := range workEntries {
		if _, tracked := indexMap[e.Path]; !tracked {
			untracked = append(untracked, e.Path)
		}
	}
	sort.Strings(untracked)

	if opts.DryRun {
		return untracked, nil
	}

	var removed []string
	dirs := map[string]bool{}
	for _, rel := range untracked {
		abs := filepath.Join(repo.Root, rel)
		if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
			return removed, fmt.Errorf("remove %s: %w", rel, err)
		}
		removed = append(removed, rel)
		dirs[filepath.Dir(abs)] = true
	}

	if opts.Dirs {
		removeEmptyDirs(dirs, filepath.Join(repo.Root, ".mos"))
	}

	return removed, nil
}

func removeEmptyDirs(candidates map[string]bool, boundary string) {
	absBoundary, _ := filepath.Abs(boundary)
	for dir := range candidates {
		for {
			absDir, _ := filepath.Abs(dir)
			if absDir == absBoundary || len(absDir) <= len(absBoundary) {
				break
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) > 0 {
				break
			}
			os.Remove(dir)
			dir = filepath.Dir(dir)
		}
	}
}
