package vcs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FSStore is an ObjectStore backed by flat files under a directory.
// Objects live in <root>/objects/<xx>/<rest-of-hash>.
// Refs live in <root>/refs/<name>.
type FSStore struct {
	root string
}

var _ ObjectStore = (*FSStore)(nil)

func NewFSStore(root string) (*FSStore, error) {
	for _, sub := range []string{"objects", "refs"} {
		if err := os.MkdirAll(filepath.Join(root, sub), 0755); err != nil {
			return nil, fmt.Errorf("fsstore init: %w", err)
		}
	}
	return &FSStore{root: root}, nil
}

func (s *FSStore) objectPath(h Hash) string {
	hex := h.String()
	return filepath.Join(s.root, "objects", hex[:2], hex[2:])
}

func (s *FSStore) storeRaw(data []byte) (Hash, error) {
	h := NewHash(data)
	p := s.objectPath(h)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return ZeroHash, err
	}
	return h, os.WriteFile(p, data, 0644)
}

func (s *FSStore) readRaw(h Hash) ([]byte, error) {
	data, err := os.ReadFile(s.objectPath(h))
	if err != nil {
		return nil, fmt.Errorf("object %s not found: %w", h.Short(), err)
	}
	return data, nil
}

func (s *FSStore) StoreBlob(data []byte) (Hash, error) {
	tagged := append([]byte("blob\n"), data...)
	return s.storeRaw(tagged)
}

func (s *FSStore) ReadBlob(h Hash) ([]byte, error) {
	data, err := s.readRaw(h)
	if err != nil {
		return nil, err
	}
	prefix := []byte("blob\n")
	if len(data) < len(prefix) || string(data[:len(prefix)]) != string(prefix) {
		return nil, fmt.Errorf("object %s is not a blob", h.Short())
	}
	return data[len(prefix):], nil
}

func (s *FSStore) StoreTree(entries []TreeEntry) (Hash, error) {
	data := serializeTree(entries)
	return s.storeRaw(data)
}

func (s *FSStore) ReadTree(h Hash) ([]TreeEntry, error) {
	data, err := s.readRaw(h)
	if err != nil {
		return nil, err
	}
	return deserializeTree(data)
}

func (s *FSStore) StoreCommit(c CommitData) (Hash, error) {
	data := serializeCommit(c)
	return s.storeRaw(data)
}

func (s *FSStore) ReadCommit(h Hash) (*CommitData, error) {
	data, err := s.readRaw(h)
	if err != nil {
		return nil, err
	}
	return deserializeCommit(data)
}

func (s *FSStore) HasObject(h Hash) bool {
	_, err := os.Stat(s.objectPath(h))
	return err == nil
}

func (s *FSStore) TypeOf(h Hash) (ObjectType, error) {
	data, err := s.readRaw(h)
	if err != nil {
		return 0, err
	}
	switch {
	case len(data) >= 5 && string(data[:5]) == "blob\n":
		return ObjectBlob, nil
	case len(data) >= 5 && string(data[:5]) == "tree\n":
		return ObjectTree, nil
	case len(data) >= 7 && string(data[:7]) == "commit\n":
		return ObjectCommit, nil
	default:
		return 0, fmt.Errorf("unknown object type for %s", h.Short())
	}
}

func (s *FSStore) refPath(name string) string {
	return filepath.Join(s.root, "refs", name)
}

func (s *FSStore) UpdateRef(name string, h Hash) error {
	p := s.refPath(name)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(h.String()+"\n"), 0644)
}

func (s *FSStore) ResolveRef(name string) (Hash, error) {
	data, err := os.ReadFile(s.refPath(name))
	if err != nil {
		return ZeroHash, fmt.Errorf("ref %q not found: %w", name, err)
	}
	return ParseHash(strings.TrimSpace(string(data)))
}

func (s *FSStore) ListRefs(prefix string) ([]Ref, error) {
	var refs []Ref
	refsDir := filepath.Join(s.root, "refs")
	err := filepath.Walk(refsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(refsDir, path)
		if prefix != "" && !strings.HasPrefix(rel, prefix) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		h, err := ParseHash(strings.TrimSpace(string(data)))
		if err != nil {
			return err
		}
		refs = append(refs, Ref{Name: rel, Hash: h})
		return nil
	})
	return refs, err
}

func (s *FSStore) DeleteRef(name string) error {
	return os.Remove(s.refPath(name))
}

func (s *FSStore) AllObjects() ([]Hash, error) {
	var hashes []Hash
	objDir := filepath.Join(s.root, "objects")
	err := filepath.Walk(objDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(objDir, path)
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) != 2 {
			return nil
		}
		hexStr := parts[0] + parts[1]
		h, err := ParseHash(hexStr)
		if err != nil {
			return nil // skip malformed
		}
		hashes = append(hashes, h)
		return nil
	})
	return hashes, err
}
