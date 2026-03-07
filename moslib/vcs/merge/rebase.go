package merge

import (
	"fmt"
	"time"

	"github.com/dpopsuev/mos/moslib/vcs"
	"github.com/dpopsuev/mos/moslib/vcs/history"
	"github.com/dpopsuev/mos/moslib/vcs/staging"
)

type RebaseResult struct {
	NewTip        vcs.Hash
	ReplayedCount int
}

func Rebase(repo *vcs.Repository, ontoBranch, author, email string) (*RebaseResult, error) {
	ourHead, err := vcs.ResolveHead(repo)
	if err != nil {
		return nil, fmt.Errorf("resolve HEAD: %w", err)
	}

	ontoHash, err := repo.Store.ResolveRef(vcs.HeadsPrefix + ontoBranch)
	if err != nil {
		return nil, fmt.Errorf("resolve branch %q: %w", ontoBranch, err)
	}

	base, err := FindMergeBase(repo.Store, ourHead, ontoHash)
	if err != nil {
		return nil, fmt.Errorf("find merge base: %w", err)
	}

	if base == ourHead {
		return &RebaseResult{NewTip: ontoHash}, nil
	}

	commits, err := collectCommits(repo.Store, ourHead, base)
	if err != nil {
		return nil, err
	}

	tip := ontoHash
	for _, entry := range commits {
		newTip, err := cherryPick(repo, entry, tip, author, email)
		if err != nil {
			return nil, fmt.Errorf("cherry-pick %s: %w", entry.Hash.Short(), err)
		}
		tip = newTip
	}

	branch := vcs.CurrentBranch(repo.Root)
	if branch != "" {
		if err := repo.Store.UpdateRef(vcs.HeadsPrefix+branch, tip); err != nil {
			return nil, err
		}
	} else {
		if err := vcs.WriteDetachedHead(repo.Root, tip); err != nil {
			return nil, err
		}
	}

	if err := staging.ResetIndexToCommit(repo, tip); err != nil {
		return nil, err
	}

	return &RebaseResult{NewTip: tip, ReplayedCount: len(commits)}, nil
}

func collectCommits(store vcs.ObjectStore, head, stop vcs.Hash) ([]history.LogEntry, error) {
	var result []history.LogEntry
	current := head
	for current != stop && !current.IsZero() {
		cd, err := store.ReadCommit(current)
		if err != nil {
			return nil, err
		}
		result = append(result, history.LogEntry{Hash: current, Commit: *cd})
		if len(cd.Parents) == 0 {
			break
		}
		current = cd.Parents[0]
	}
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result, nil
}

func cherryPick(repo *vcs.Repository, entry history.LogEntry, onto vcs.Hash, author, email string) (vcs.Hash, error) {
	cd := entry.Commit

	var parentTree map[string]vcs.Hash
	if len(cd.Parents) > 0 {
		var err error
		parentTree, err = commitTreeMap(repo.Store, cd.Parents[0])
		if err != nil {
			return vcs.ZeroHash, err
		}
	} else {
		parentTree = map[string]vcs.Hash{}
	}

	commitTree, err := commitTreeMap(repo.Store, entry.Hash)
	if err != nil {
		return vcs.ZeroHash, err
	}

	diffs := staging.DiffTrees(parentTree, commitTree)

	ontoTree, err := commitTreeMap(repo.Store, onto)
	if err != nil {
		return vcs.ZeroHash, err
	}

	for _, d := range diffs {
		switch d.Kind {
		case staging.DiffAdded, staging.DiffModified:
			ontoTree[d.Path] = d.NewHash
		case staging.DiffDeleted:
			delete(ontoTree, d.Path)
		}
	}

	idx := &staging.Index{}
	for path, hash := range ontoTree {
		idx.Set(path, hash, vcs.ModeRegular)
	}

	treeHash, err := idx.BuildTree(repo.Store)
	if err != nil {
		return vcs.ZeroHash, err
	}

	newCommit := vcs.CommitData{
		Tree:    treeHash,
		Parents: []vcs.Hash{onto},
		Author:  author,
		Email:   email,
		Time:    time.Now().UTC(),
		Message: cd.Message,
	}

	return repo.Store.StoreCommit(newCommit)
}
