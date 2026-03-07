package history

import (
	"fmt"
	"os"
	"strings"

	"github.com/dpopsuev/mos/moslib/vcs"
)

func HashObject(store vcs.ObjectStore, path string) (vcs.Hash, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return vcs.ZeroHash, fmt.Errorf("read %s: %w", path, err)
	}
	return store.StoreBlob(data)
}

func CatFile(store vcs.ObjectStore, h vcs.Hash) (vcs.ObjectType, []byte, error) {
	typ, err := store.TypeOf(h)
	if err != nil {
		return 0, nil, err
	}
	switch typ {
	case vcs.ObjectBlob:
		data, err := store.ReadBlob(h)
		return vcs.ObjectBlob, data, err
	case vcs.ObjectTree:
		entries, err := store.ReadTree(h)
		if err != nil {
			return vcs.ObjectTree, nil, err
		}
		var b strings.Builder
		for _, e := range entries {
			kind := "blob"
			if e.Mode == vcs.ModeDir {
				kind = "tree"
			}
			fmt.Fprintf(&b, "%06o %s %s\t%s\n", e.Mode, kind, e.Hash.Short(), e.Name)
		}
		return vcs.ObjectTree, []byte(b.String()), nil
	case vcs.ObjectCommit:
		cd, err := store.ReadCommit(h)
		if err != nil {
			return vcs.ObjectCommit, nil, err
		}
		var b strings.Builder
		fmt.Fprintf(&b, "tree %s\n", cd.Tree)
		for _, p := range cd.Parents {
			fmt.Fprintf(&b, "parent %s\n", p)
		}
		fmt.Fprintf(&b, "author %s <%s> %d\n", cd.Author, cd.Email, cd.Time.Unix())
		fmt.Fprintf(&b, "\n%s\n", cd.Message)
		return vcs.ObjectCommit, []byte(b.String()), nil
	default:
		return 0, nil, fmt.Errorf("unknown object type")
	}
}

func LsTree(store vcs.ObjectStore, h vcs.Hash, recursive bool) ([]LsTreeEntry, error) {
	typ, err := store.TypeOf(h)
	if err != nil {
		return nil, err
	}
	var treeHash vcs.Hash
	if typ == vcs.ObjectCommit {
		cd, err := store.ReadCommit(h)
		if err != nil {
			return nil, err
		}
		treeHash = cd.Tree
	} else if typ == vcs.ObjectTree {
		treeHash = h
	} else {
		return nil, fmt.Errorf("object %s is a %s, not a tree or commit", h.Short(), typ)
	}
	return lsTreeWalk(store, treeHash, "", recursive)
}

type LsTreeEntry struct {
	Mode uint32
	Type string
	Hash vcs.Hash
	Path string
}

func lsTreeWalk(store vcs.ObjectStore, h vcs.Hash, prefix string, recursive bool) ([]LsTreeEntry, error) {
	entries, err := store.ReadTree(h)
	if err != nil {
		return nil, err
	}
	var result []LsTreeEntry
	for _, e := range entries {
		path := e.Name
		if prefix != "" {
			path = prefix + "/" + e.Name
		}
		kind := "blob"
		if e.Mode == vcs.ModeDir {
			kind = "tree"
		}
		result = append(result, LsTreeEntry{Mode: e.Mode, Type: kind, Hash: e.Hash, Path: path})
		if recursive && e.Mode == vcs.ModeDir {
			sub, err := lsTreeWalk(store, e.Hash, path, true)
			if err != nil {
				return nil, err
			}
			result = append(result, sub...)
		}
	}
	return result, nil
}

func RevParse(store vcs.ObjectStore, root string, name string) (vcs.Hash, error) {
	if h, err := vcs.ParseHash(name); err == nil {
		if store.HasObject(h) {
			return h, nil
		}
	}

	if name == "HEAD" {
		branch, detached := vcs.ReadSymbolicHead(root)
		if detached {
			return vcs.ParseHash(branch)
		}
		return store.ResolveRef("heads/" + branch)
	}

	if h, err := store.ResolveRef("heads/" + name); err == nil {
		return h, nil
	}
	if h, err := store.ResolveRef("tags/" + name); err == nil {
		return h, nil
	}
	if h, err := store.ResolveRef(name); err == nil {
		return h, nil
	}

	return vcs.ZeroHash, fmt.Errorf("cannot resolve %q", name)
}
