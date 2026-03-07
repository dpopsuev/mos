package vcs

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	HeadsPrefix   = "heads/"
	TagsPrefix    = "tags/"
	SymHeadFile   = ".mos/vcs/HEAD"
	DefaultBranch = "main"
)

func InitHead(root string) error {
	return WriteSymbolicHead(root, DefaultBranch)
}

func WriteSymbolicHead(root, branch string) error {
	p := filepath.Join(root, SymHeadFile)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte("ref: "+branch+"\n"), 0644)
}

func WriteDetachedHead(root string, h Hash) error {
	p := filepath.Join(root, SymHeadFile)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(h.String()+"\n"), 0644)
}

func ReadSymbolicHead(root string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(root, SymHeadFile))
	if err != nil {
		return DefaultBranch, false
	}
	content := strings.TrimSpace(string(data))
	if strings.HasPrefix(content, "ref: ") {
		return content[5:], false
	}
	return content, true
}

func ResolveHead(repo *Repository) (Hash, error) {
	branch, detached := ReadSymbolicHead(repo.Root)
	if detached {
		return ParseHash(branch)
	}
	return repo.Store.ResolveRef(HeadsPrefix + branch)
}

func CurrentBranch(root string) string {
	branch, detached := ReadSymbolicHead(root)
	if detached {
		return ""
	}
	return branch
}
