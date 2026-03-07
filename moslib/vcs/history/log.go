package history

import (
	"github.com/dpopsuev/mos/moslib/vcs"
	"github.com/dpopsuev/mos/moslib/vcs/staging"
)

type LogEntry struct {
	Hash   vcs.Hash
	Commit vcs.CommitData
}

func Log(store vcs.ObjectStore, from vcs.Hash, maxCount int) ([]LogEntry, error) {
	var entries []LogEntry
	seen := map[vcs.Hash]bool{}
	queue := []vcs.Hash{from}

	for len(queue) > 0 {
		if maxCount > 0 && len(entries) >= maxCount {
			break
		}

		h := queue[0]
		queue = queue[1:]

		if seen[h] || h.IsZero() {
			continue
		}
		seen[h] = true

		cd, err := store.ReadCommit(h)
		if err != nil {
			return entries, err
		}
		entries = append(entries, LogEntry{Hash: h, Commit: *cd})

		for _, p := range cd.Parents {
			if !seen[p] {
				queue = append(queue, p)
			}
		}
	}

	return entries, nil
}

func CommitStat(store vcs.ObjectStore, entry LogEntry) ([]staging.DiffEntry, error) {
	var parentTree vcs.Hash
	if len(entry.Commit.Parents) > 0 {
		pc, err := store.ReadCommit(entry.Commit.Parents[0])
		if err != nil {
			return nil, err
		}
		parentTree = pc.Tree
	}

	oldMap, err := staging.FlattenTree(store, parentTree, "")
	if err != nil {
		return nil, err
	}
	newMap, err := staging.FlattenTree(store, entry.Commit.Tree, "")
	if err != nil {
		return nil, err
	}
	return staging.DiffTrees(oldMap, newMap), nil
}
