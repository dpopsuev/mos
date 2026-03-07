package staging

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dpopsuev/mos/moslib/vcs"
)

func SnapshotWorkingTree(root string, store vcs.ObjectStore) ([]IndexEntry, error) {
	mosDir := filepath.Join(root, ".mos")
	var entries []IndexEntry

	err := filepath.Walk(mosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "vcs" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)

		if ShouldSkipFile(rel) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		h, err := store.StoreBlob(data)
		if err != nil {
			return err
		}
		entries = append(entries, IndexEntry{
			Path: rel,
			Hash: h,
			Mode: vcs.ModeRegular,
		})
		return nil
	})
	return entries, err
}

func ShouldSkipFile(rel string) bool {
	skips := []string{
		".mos/vcs/",
		".mos/vcs.json",
		".mos/.lock",
	}
	for _, s := range skips {
		if strings.HasPrefix(rel, s) || rel == s {
			return true
		}
	}
	return false
}
