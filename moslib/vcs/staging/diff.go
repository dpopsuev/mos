package staging

import (
	"sort"

	"github.com/dpopsuev/mos/moslib/vcs"
)

type DiffKind int

const (
	DiffAdded    DiffKind = iota
	DiffModified
	DiffDeleted
)

func (k DiffKind) String() string {
	switch k {
	case DiffAdded:
		return "added"
	case DiffModified:
		return "modified"
	case DiffDeleted:
		return "deleted"
	default:
		return "unknown"
	}
}

type DiffEntry struct {
	Path    string
	Kind    DiffKind
	OldHash vcs.Hash
	NewHash vcs.Hash
}

func DiffTrees(old, new map[string]vcs.Hash) []DiffEntry {
	var diffs []DiffEntry

	for path, newHash := range new {
		oldHash, exists := old[path]
		if !exists {
			diffs = append(diffs, DiffEntry{Path: path, Kind: DiffAdded, NewHash: newHash})
		} else if oldHash != newHash {
			diffs = append(diffs, DiffEntry{Path: path, Kind: DiffModified, OldHash: oldHash, NewHash: newHash})
		}
	}

	for path, oldHash := range old {
		if _, exists := new[path]; !exists {
			diffs = append(diffs, DiffEntry{Path: path, Kind: DiffDeleted, OldHash: oldHash})
		}
	}

	sort.Slice(diffs, func(i, j int) bool { return diffs[i].Path < diffs[j].Path })
	return diffs
}

func FlattenTree(store vcs.ObjectStore, root vcs.Hash, prefix string) (map[string]vcs.Hash, error) {
	if root.IsZero() {
		return map[string]vcs.Hash{}, nil
	}
	entries, err := store.ReadTree(root)
	if err != nil {
		return nil, err
	}
	result := map[string]vcs.Hash{}
	for _, e := range entries {
		path := e.Name
		if prefix != "" {
			path = prefix + "/" + e.Name
		}
		if e.Mode == vcs.ModeDir {
			sub, err := FlattenTree(store, e.Hash, path)
			if err != nil {
				return nil, err
			}
			for k, v := range sub {
				result[k] = v
			}
		} else {
			result[path] = e.Hash
		}
	}
	return result, nil
}

func IndexToMap(idx *Index) map[string]vcs.Hash {
	m := map[string]vcs.Hash{}
	for _, e := range idx.Entries {
		m[e.Path] = e.Hash
	}
	return m
}
