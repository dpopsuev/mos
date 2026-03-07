package store

import (
	"io/fs"
	"os"
)

// Store abstracts filesystem I/O so governance logic can be tested
// with in-memory fakes and decoupled from direct os calls.
type Store interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	Stat(path string) (fs.FileInfo, error)
	ReadDir(path string) ([]fs.DirEntry, error)
	MkdirAll(path string, perm fs.FileMode) error
	RemoveAll(path string) error
	Rename(old, new string) error
}

// DefaultStore is the production filesystem-backed store.
var DefaultStore Store = &FSStore{}

// FSStore delegates all operations to the os package.
type FSStore struct{}

func (s *FSStore) ReadFile(path string) ([]byte, error)                        { return os.ReadFile(path) }
func (s *FSStore) WriteFile(path string, data []byte, perm fs.FileMode) error  { return os.WriteFile(path, data, perm) }
func (s *FSStore) Stat(path string) (fs.FileInfo, error)                       { return os.Stat(path) }
func (s *FSStore) ReadDir(path string) ([]fs.DirEntry, error)                  { return os.ReadDir(path) }
func (s *FSStore) MkdirAll(path string, perm fs.FileMode) error                { return os.MkdirAll(path, perm) }
func (s *FSStore) RemoveAll(path string) error                                 { return os.RemoveAll(path) }
func (s *FSStore) Rename(old, new string) error                                { return os.Rename(old, new) }

// ObjectStore extends Store with pack/unpack/verify for future backends.
type ObjectStore interface {
	Store

	Pack(root string) error
	Unpack(root string) error
	Verify(root string) ([]IntegrityError, error)
}

// IntegrityError describes a single integrity violation found during Verify.
type IntegrityError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

func (s *FSStore) Pack(string) error                          { return nil }
func (s *FSStore) Unpack(string) error                        { return nil }
func (s *FSStore) Verify(string) ([]IntegrityError, error)    { return nil, nil }

// DefaultObjectStore is the production ObjectStore.
var DefaultObjectStore ObjectStore = &FSStore{}

// ReadFile reads a file using DefaultStore.
func ReadFile(path string) ([]byte, error) {
	return DefaultStore.ReadFile(path)
}

// ReadDir reads a directory using DefaultStore.
func ReadDir(path string) ([]fs.DirEntry, error) {
	return DefaultStore.ReadDir(path)
}

// Stat stats a path using DefaultStore.
func Stat(path string) (fs.FileInfo, error) {
	return DefaultStore.Stat(path)
}

// MkdirAll creates directories using DefaultStore.
func MkdirAll(path string, perm fs.FileMode) error {
	return DefaultStore.MkdirAll(path, perm)
}

// RemoveAll removes a path using DefaultStore.
func RemoveAll(path string) error {
	return DefaultStore.RemoveAll(path)
}

// Rename renames a path using DefaultStore.
func Rename(old, new string) error {
	return DefaultStore.Rename(old, new)
}

// WriteFile writes a file using DefaultStore.
func WriteFile(path string, data []byte, perm fs.FileMode) error {
	return DefaultStore.WriteFile(path, data, perm)
}
