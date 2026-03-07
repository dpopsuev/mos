package forge

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Forge provides Git repository hosting for tests.
type Forge interface {
	CreateRepo(name string) (cloneURL string, err error)
	RepoURL(name string) string
	Close() error
}

// inProcess is a Forge backed by bare go-git repositories on the local filesystem.
// Uses file:// URLs for clone/push/fetch via the real git binary,
// ensuring full protocol compatibility.
type inProcess struct {
	t       testing.TB
	dataDir string
	repos   map[string]string
}

// InProcess creates an in-process Git forge backed by bare repositories.
func InProcess(t testing.TB) Forge {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not found, skipping forge tests")
	}

	dataDir := t.TempDir()
	f := &inProcess{
		t:       t,
		dataDir: dataDir,
		repos:   make(map[string]string),
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func (f *inProcess) CreateRepo(name string) (string, error) {
	repoPath := filepath.Join(f.dataDir, name+".git")

	repo, err := git.PlainInit(repoPath, true)
	if err != nil {
		return "", fmt.Errorf("init bare repo %s: %w", name, err)
	}

	if err := createInitialCommit(repo); err != nil {
		return "", fmt.Errorf("initial commit for %s: %w", name, err)
	}

	f.repos[name] = repoPath
	return f.RepoURL(name), nil
}

func (f *inProcess) RepoURL(name string) string {
	return filepath.Join(f.dataDir, name+".git")
}

func (f *inProcess) Close() error {
	return nil
}

func createInitialCommit(repo *git.Repository) error {
	blobHash, err := storeBlob(repo, []byte("# Mos\n"))
	if err != nil {
		return err
	}

	tree := &object.Tree{Entries: []object.TreeEntry{
		{Name: "README.md", Mode: 0o100644, Hash: blobHash},
	}}
	treeObj := repo.Storer.NewEncodedObject()
	if err := tree.Encode(treeObj); err != nil {
		return err
	}
	treeHash, err := repo.Storer.SetEncodedObject(treeObj)
	if err != nil {
		return err
	}

	sig := object.Signature{
		Name:  "mos-test",
		Email: "test@mos.dev",
	}
	commit := &object.Commit{
		Author:    sig,
		Committer: sig,
		Message:   "initial commit",
		TreeHash:  treeHash,
	}
	commitObj := repo.Storer.NewEncodedObject()
	if err := commit.Encode(commitObj); err != nil {
		return err
	}
	commitHash, err := repo.Storer.SetEncodedObject(commitObj)
	if err != nil {
		return err
	}

	ref := plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/main"), commitHash)
	if err := repo.Storer.SetReference(ref); err != nil {
		return err
	}
	head := plumbing.NewSymbolicReference(plumbing.HEAD, "refs/heads/main")
	return repo.Storer.SetReference(head)
}

func storeBlob(repo *git.Repository, data []byte) (plumbing.Hash, error) {
	obj := repo.Storer.NewEncodedObject()
	obj.SetType(plumbing.BlobObject)
	obj.SetSize(int64(len(data)))
	w, err := obj.Writer()
	if err != nil {
		return plumbing.ZeroHash, err
	}
	if _, err := w.Write(data); err != nil {
		return plumbing.ZeroHash, err
	}
	if err := w.Close(); err != nil {
		return plumbing.ZeroHash, err
	}
	return repo.Storer.SetEncodedObject(obj)
}

// GitExec runs a git command in the given directory.
func GitExec(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@mos.dev",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@mos.dev",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("git %v: %w\n%s", args, err, out)
	}
	return out, nil
}
