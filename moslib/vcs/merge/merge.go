package merge

import (
	"fmt"
	"time"

	"github.com/dpopsuev/mos/moslib/vcs"
	"github.com/dpopsuev/mos/moslib/vcs/staging"
)

type MergeConflict struct {
	Path      string
	BaseHash  vcs.Hash
	OursHash  vcs.Hash
	TheirHash vcs.Hash
}

type MergeResult struct {
	CommitHash  vcs.Hash
	FastForward bool
	Conflicts   []MergeConflict
}

func Merge(repo *vcs.Repository, sourceBranch, author, email string) (*MergeResult, error) {
	ourHead, err := vcs.ResolveHead(repo)
	if err != nil {
		return nil, fmt.Errorf("resolve HEAD: %w", err)
	}

	theirHead, err := repo.Store.ResolveRef(vcs.HeadsPrefix + sourceBranch)
	if err != nil {
		return nil, fmt.Errorf("resolve branch %q: %w", sourceBranch, err)
	}

	if ourHead == theirHead {
		return &MergeResult{CommitHash: ourHead}, nil
	}

	base, err := FindMergeBase(repo.Store, ourHead, theirHead)
	if err != nil {
		return nil, fmt.Errorf("find merge base: %w", err)
	}

	if base == ourHead {
		branch := vcs.CurrentBranch(repo.Root)
		if branch != "" {
			if err := repo.Store.UpdateRef(vcs.HeadsPrefix+branch, theirHead); err != nil {
				return nil, err
			}
		} else {
			if err := vcs.WriteDetachedHead(repo.Root, theirHead); err != nil {
				return nil, err
			}
		}
		if err := staging.ResetIndexToCommit(repo, theirHead); err != nil {
			return nil, err
		}
		return &MergeResult{CommitHash: theirHead, FastForward: true}, nil
	}

	baseMap, err := commitTreeMap(repo.Store, base)
	if err != nil {
		return nil, err
	}
	ourMap, err := commitTreeMap(repo.Store, ourHead)
	if err != nil {
		return nil, err
	}
	theirMap, err := commitTreeMap(repo.Store, theirHead)
	if err != nil {
		return nil, err
	}

	merged, conflicts := threeWayMerge(baseMap, ourMap, theirMap)
	if len(conflicts) > 0 {
		return &MergeResult{Conflicts: conflicts}, nil
	}

	idx, err := staging.LoadIndex(repo.Root)
	if err != nil {
		return nil, err
	}
	idx.Entries = nil
	for path, hash := range merged {
		idx.Set(path, hash, vcs.ModeRegular)
	}
	if err := idx.Save(); err != nil {
		return nil, err
	}

	treeHash, err := idx.BuildTree(repo.Store)
	if err != nil {
		return nil, err
	}

	cd := vcs.CommitData{
		Tree:    treeHash,
		Parents: []vcs.Hash{ourHead, theirHead},
		Author:  author,
		Email:   email,
		Time:    time.Now().UTC(),
		Message: fmt.Sprintf("Merge branch '%s'", sourceBranch),
	}
	commitHash, err := repo.Store.StoreCommit(cd)
	if err != nil {
		return nil, err
	}

	branch := vcs.CurrentBranch(repo.Root)
	if branch != "" {
		if err := repo.Store.UpdateRef(vcs.HeadsPrefix+branch, commitHash); err != nil {
			return nil, err
		}
	} else {
		if err := vcs.WriteDetachedHead(repo.Root, commitHash); err != nil {
			return nil, err
		}
	}

	return &MergeResult{CommitHash: commitHash}, nil
}

// MergeHash merges a specific commit hash into the current branch.
func MergeHash(repo *vcs.Repository, theirHash vcs.Hash, author, email string) (*MergeResult, error) {
	ourHead, err := vcs.ResolveHead(repo)
	if err != nil {
		return nil, fmt.Errorf("resolve HEAD: %w", err)
	}

	base, err := FindMergeBase(repo.Store, ourHead, theirHash)
	if err != nil {
		return nil, fmt.Errorf("find merge base: %w", err)
	}

	if base == ourHead {
		branch := vcs.CurrentBranch(repo.Root)
		if branch != "" {
			if err := repo.Store.UpdateRef(vcs.HeadsPrefix+branch, theirHash); err != nil {
				return nil, err
			}
		} else {
			if err := vcs.WriteDetachedHead(repo.Root, theirHash); err != nil {
				return nil, err
			}
		}
		if err := staging.ResetIndexToCommit(repo, theirHash); err != nil {
			return nil, err
		}
		return &MergeResult{CommitHash: theirHash, FastForward: true}, nil
	}

	baseMap, err := commitTreeMap(repo.Store, base)
	if err != nil {
		return nil, err
	}
	ourMap, err := commitTreeMap(repo.Store, ourHead)
	if err != nil {
		return nil, err
	}
	theirMap, err := commitTreeMap(repo.Store, theirHash)
	if err != nil {
		return nil, err
	}

	merged, conflicts := threeWayMerge(baseMap, ourMap, theirMap)
	if len(conflicts) > 0 {
		return &MergeResult{Conflicts: conflicts}, nil
	}

	idx, err := staging.LoadIndex(repo.Root)
	if err != nil {
		return nil, err
	}
	idx.Entries = nil
	for path, hash := range merged {
		idx.Set(path, hash, vcs.ModeRegular)
	}
	if err := idx.Save(); err != nil {
		return nil, err
	}

	treeHash, err := idx.BuildTree(repo.Store)
	if err != nil {
		return nil, err
	}

	cd := vcs.CommitData{
		Tree:    treeHash,
		Parents: []vcs.Hash{ourHead, theirHash},
		Author:  author,
		Email:   email,
		Time:    time.Now().UTC(),
		Message: fmt.Sprintf("Merge remote branch '%s'", vcs.CurrentBranch(repo.Root)),
	}
	commitHash, err := repo.Store.StoreCommit(cd)
	if err != nil {
		return nil, err
	}

	branch := vcs.CurrentBranch(repo.Root)
	if branch != "" {
		if err := repo.Store.UpdateRef(vcs.HeadsPrefix+branch, commitHash); err != nil {
			return nil, err
		}
	} else {
		if err := vcs.WriteDetachedHead(repo.Root, commitHash); err != nil {
			return nil, err
		}
	}

	return &MergeResult{CommitHash: commitHash}, nil
}

func FindMergeBase(store vcs.ObjectStore, a, b vcs.Hash) (vcs.Hash, error) {
	if a == b {
		return a, nil
	}

	ancestorsA := map[vcs.Hash]bool{a: true}
	ancestorsB := map[vcs.Hash]bool{b: true}
	queueA := []vcs.Hash{a}
	queueB := []vcs.Hash{b}

	for len(queueA) > 0 || len(queueB) > 0 {
		if len(queueA) > 0 {
			h := queueA[0]
			queueA = queueA[1:]
			if ancestorsB[h] {
				return h, nil
			}
			cd, err := store.ReadCommit(h)
			if err != nil {
				continue
			}
			for _, p := range cd.Parents {
				if !ancestorsA[p] {
					ancestorsA[p] = true
					queueA = append(queueA, p)
				}
			}
		}
		if len(queueB) > 0 {
			h := queueB[0]
			queueB = queueB[1:]
			if ancestorsA[h] {
				return h, nil
			}
			cd, err := store.ReadCommit(h)
			if err != nil {
				continue
			}
			for _, p := range cd.Parents {
				if !ancestorsB[p] {
					ancestorsB[p] = true
					queueB = append(queueB, p)
				}
			}
		}
	}

	return vcs.ZeroHash, fmt.Errorf("no common ancestor found")
}

func commitTreeMap(store vcs.ObjectStore, commitHash vcs.Hash) (map[string]vcs.Hash, error) {
	cd, err := store.ReadCommit(commitHash)
	if err != nil {
		return nil, err
	}
	return staging.FlattenTree(store, cd.Tree, "")
}

func threeWayMerge(base, ours, theirs map[string]vcs.Hash) (map[string]vcs.Hash, []MergeConflict) {
	merged := map[string]vcs.Hash{}
	var conflicts []MergeConflict

	allPaths := map[string]bool{}
	for p := range base {
		allPaths[p] = true
	}
	for p := range ours {
		allPaths[p] = true
	}
	for p := range theirs {
		allPaths[p] = true
	}

	for path := range allPaths {
		bh := base[path]
		oh := ours[path]
		th := theirs[path]

		switch {
		case oh == th:
			if !oh.IsZero() {
				merged[path] = oh
			}
		case oh == bh:
			if !th.IsZero() {
				merged[path] = th
			}
		case th == bh:
			if !oh.IsZero() {
				merged[path] = oh
			}
		default:
			conflicts = append(conflicts, MergeConflict{
				Path:      path,
				BaseHash:  bh,
				OursHash:  oh,
				TheirHash: th,
			})
			if !oh.IsZero() {
				merged[path] = oh
			}
		}
	}

	return merged, conflicts
}
