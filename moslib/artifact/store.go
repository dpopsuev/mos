package artifact

import (
	"io/fs"
	"path/filepath"

	"github.com/dpopsuev/mos/moslib/store"
)

// Re-export store types for backward compatibility.
type Store = store.Store
type FSStore = store.FSStore

var DefaultStore = store.DefaultStore

func storeReadFile(path string) ([]byte, error)              { return store.ReadFile(path) }
func storeReadDir(path string) ([]fs.DirEntry, error)        { return store.ReadDir(path) }
func storeStat(path string) (fs.FileInfo, error)             { return store.Stat(path) }
func storeMkdirAll(path string, perm fs.FileMode) error      { return store.MkdirAll(path, perm) }
func storeRemoveAll(path string) error                       { return store.RemoveAll(path) }
func storeRename(old, new string) error                      { return store.Rename(old, new) }

// storeWalkArtifactDirs iterates over active and archive subdirectories.
func storeWalkArtifactDirs(mosDir string, td ArtifactTypeDef, fn func(dirPath, sub, entryName string) error) error {
	for _, sub := range []string{ActiveDir, ArchiveDir} {
		dir := filepath.Join(mosDir, td.Directory, sub)
		entries, err := storeReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			if err := fn(dir, sub, e.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}
