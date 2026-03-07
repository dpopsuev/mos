package staging

import (
	"fmt"

	"github.com/dpopsuev/mos/moslib/vcs"
)

// ResetIndexToCommit loads the tree from a commit and replaces the index entries.
func ResetIndexToCommit(repo *vcs.Repository, commitHash vcs.Hash) error {
	cd, err := repo.Store.ReadCommit(commitHash)
	if err != nil {
		return fmt.Errorf("read commit: %w", err)
	}
	flatMap, err := FlattenTree(repo.Store, cd.Tree, "")
	if err != nil {
		return fmt.Errorf("flatten tree: %w", err)
	}
	idx, err := LoadIndex(repo.Root)
	if err != nil {
		return err
	}
	idx.Entries = nil
	for path, hash := range flatMap {
		idx.Set(path, hash, vcs.ModeRegular)
	}
	return idx.Save()
}
