package linter

import (
	"strings"

	"github.com/dpopsuev/mos/moslib/vcs"
	"github.com/dpopsuev/mos/moslib/vcs/staging"
)

// DetectChangedFiles returns .mos file paths that have been added or modified
// since the last VCS commit. Returns (nil, nil) if VCS is unavailable,
// signaling the caller to fall back to a full lint.
func DetectChangedFiles(root string) ([]string, error) {
	repo, err := vcs.OpenRepo(root)
	if err != nil {
		return nil, nil
	}

	workEntries, err := staging.SnapshotWorkingTree(root, repo.Store)
	if err != nil {
		return nil, nil
	}
	workMap := make(map[string]vcs.Hash, len(workEntries))
	for _, e := range workEntries {
		workMap[e.Path] = e.Hash
	}

	headHash, err := vcs.ResolveHead(repo)
	var headMap map[string]vcs.Hash
	if err == nil && !headHash.IsZero() {
		cd, err := repo.Store.ReadCommit(headHash)
		if err == nil {
			headMap, _ = staging.FlattenTree(repo.Store, cd.Tree, "")
		}
	}
	if headMap == nil {
		headMap = map[string]vcs.Hash{}
	}

	diffs := staging.DiffTrees(headMap, workMap)
	var changed []string
	for _, d := range diffs {
		if d.Kind == staging.DiffDeleted {
			continue
		}
		if strings.HasSuffix(d.Path, ".mos") {
			changed = append(changed, d.Path)
		}
	}
	return changed, nil
}
