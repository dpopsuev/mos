package staging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dpopsuev/mos/moslib/vcs"
)

func Commit(repo *vcs.Repository, author, email, message string) (vcs.Hash, error) {
	idx, err := LoadIndex(repo.Root)
	if err != nil {
		return vcs.ZeroHash, fmt.Errorf("load index: %w", err)
	}
	if len(idx.Entries) == 0 {
		return vcs.ZeroHash, fmt.Errorf("nothing to commit (empty index)")
	}

	treeHash, err := idx.BuildTree(repo.Store)
	if err != nil {
		return vcs.ZeroHash, fmt.Errorf("build tree: %w", err)
	}

	cd := vcs.CommitData{
		Tree:    treeHash,
		Author:  author,
		Email:   email,
		Time:    time.Now().UTC(),
		Message: message,
	}

	parent, err := vcs.ResolveHead(repo)
	if err == nil && !parent.IsZero() {
		cd.Parents = []vcs.Hash{parent}
	}

	commitHash, err := repo.Store.StoreCommit(cd)
	if err != nil {
		return vcs.ZeroHash, fmt.Errorf("store commit: %w", err)
	}

	branch, detached := vcs.ReadSymbolicHead(repo.Root)
	if detached {
		if err := vcs.WriteDetachedHead(repo.Root, commitHash); err != nil {
			return vcs.ZeroHash, fmt.Errorf("update detached HEAD: %w", err)
		}
	} else {
		if err := repo.Store.UpdateRef(vcs.HeadsPrefix+branch, commitHash); err != nil {
			return vcs.ZeroHash, fmt.Errorf("update branch %s: %w", branch, err)
		}
	}

	return commitHash, nil
}

func Add(repo *vcs.Repository, paths []string) error {
	idx, err := LoadIndex(repo.Root)
	if err != nil {
		return err
	}

	if len(paths) == 0 || (len(paths) == 1 && paths[0] == ".") {
		entries, err := SnapshotWorkingTree(repo.Root, repo.Store)
		if err != nil {
			return err
		}
		idx.Entries = entries
	} else {
		for _, p := range paths {
			abs := filepath.Join(repo.Root, p)
			data, err := os.ReadFile(abs)
			if err != nil {
				return fmt.Errorf("read %s: %w", p, err)
			}
			h, err := repo.Store.StoreBlob(data)
			if err != nil {
				return err
			}
			idx.Set(p, h, vcs.ModeRegular)
		}
	}

	return idx.Save()
}
