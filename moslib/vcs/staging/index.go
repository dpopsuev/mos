package staging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dpopsuev/mos/moslib/vcs"
)

type IndexEntry struct {
	Path string   `json:"path"`
	Hash vcs.Hash `json:"hash"`
	Mode uint32   `json:"mode"`
}

type Index struct {
	Entries   []IndexEntry         `json:"entries"`
	TreeCache map[string]vcs.Hash  `json:"tree_cache,omitempty"`
	path      string
	dirty     map[string]bool
}

const indexFile = ".mos/vcs/index.json"

func LoadIndex(root string) (*Index, error) {
	idx := &Index{path: filepath.Join(root, indexFile)}
	data, err := os.ReadFile(idx.path)
	if err != nil {
		if os.IsNotExist(err) {
			return idx, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, idx); err != nil {
		return nil, fmt.Errorf("corrupt index: %w", err)
	}
	return idx, nil
}

func (idx *Index) Save() error {
	if err := os.MkdirAll(filepath.Dir(idx.path), 0755); err != nil {
		return err
	}
	sort.Slice(idx.Entries, func(i, j int) bool {
		return idx.Entries[i].Path < idx.Entries[j].Path
	})
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(idx.path, data, 0644)
}

func (idx *Index) Set(path string, h vcs.Hash, mode uint32) {
	idx.invalidateAncestors(path)
	for i, e := range idx.Entries {
		if e.Path == path {
			idx.Entries[i].Hash = h
			idx.Entries[i].Mode = mode
			return
		}
	}
	idx.Entries = append(idx.Entries, IndexEntry{Path: path, Hash: h, Mode: mode})
}

func (idx *Index) Remove(path string) {
	idx.invalidateAncestors(path)
	for i, e := range idx.Entries {
		if e.Path == path {
			idx.Entries = append(idx.Entries[:i], idx.Entries[i+1:]...)
			return
		}
	}
}

func (idx *Index) invalidateAncestors(path string) {
	if idx.dirty == nil {
		idx.dirty = make(map[string]bool)
	}
	idx.dirty[""] = true
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			idx.dirty[path[:i]] = true
		}
	}
}

func (idx *Index) Lookup(path string) (IndexEntry, bool) {
	for _, e := range idx.Entries {
		if e.Path == path {
			return e, true
		}
	}
	return IndexEntry{}, false
}

func (idx *Index) BuildTree(store vcs.ObjectStore) (vcs.Hash, error) {
	if idx.TreeCache == nil {
		idx.TreeCache = make(map[string]vcs.Hash)
	}
	h, err := idx.buildCached(store, "", idx.Entries)
	if err != nil {
		return vcs.ZeroHash, err
	}
	idx.dirty = nil
	return h, nil
}

func (idx *Index) isDirty(prefix string) bool {
	if idx.dirty == nil {
		return true
	}
	return idx.dirty[prefix]
}

func (idx *Index) buildCached(store vcs.ObjectStore, prefix string, entries []IndexEntry) (vcs.Hash, error) {
	if !idx.isDirty(prefix) {
		if cached, ok := idx.TreeCache[prefix]; ok {
			return cached, nil
		}
	}

	dirs := map[string][]IndexEntry{}
	var files []vcs.TreeEntry

	for _, e := range entries {
		dir, base := splitFirst(e.Path)
		if dir == "" {
			files = append(files, vcs.TreeEntry{Name: base, Hash: e.Hash, Mode: e.Mode})
		} else {
			sub := IndexEntry{Path: base, Hash: e.Hash, Mode: e.Mode}
			dirs[dir] = append(dirs[dir], sub)
		}
	}

	var treeEntries []vcs.TreeEntry
	treeEntries = append(treeEntries, files...)

	dirNames := make([]string, 0, len(dirs))
	for d := range dirs {
		dirNames = append(dirNames, d)
	}
	sort.Strings(dirNames)

	for _, d := range dirNames {
		childPrefix := d
		if prefix != "" {
			childPrefix = prefix + "/" + d
		}
		subtreeHash, err := idx.buildCached(store, childPrefix, dirs[d])
		if err != nil {
			return vcs.ZeroHash, err
		}
		treeEntries = append(treeEntries, vcs.TreeEntry{Name: d, Hash: subtreeHash, Mode: vcs.ModeDir})
	}

	sort.Slice(treeEntries, func(i, j int) bool {
		return treeEntries[i].Name < treeEntries[j].Name
	})

	h, err := store.StoreTree(treeEntries)
	if err != nil {
		return vcs.ZeroHash, err
	}
	idx.TreeCache[prefix] = h
	return h, nil
}

func splitFirst(path string) (string, string) {
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			return path[:i], path[i+1:]
		}
	}
	return "", path
}
