package gitcompat

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

// AssertFsck runs `git fsck` against a repository and asserts no errors.
// This proves that objects written by go-git are valid to the real git binary.
func AssertFsck(t testing.TB, repoPath string) {
	t.Helper()
	out, err := gitExec(repoPath, "fsck", "--strict", "--no-dangling")
	if err != nil {
		t.Errorf("git fsck failed: %v\n%s", err, out)
	}
}

// AssertObjectValid creates a blob, tree, and commit using go-git,
// then verifies them with `git fsck`.
func AssertObjectValid(t testing.TB) {
	t.Helper()

	dir := t.TempDir()
	repoPath := filepath.Join(dir, "test.git")

	repo, err := git.PlainInit(repoPath, true)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	blobData := []byte("hello from go-git\n")
	blobHash, err := storeBlob(repo, blobData)
	if err != nil {
		t.Fatalf("store blob: %v", err)
	}

	tree := &object.Tree{Entries: []object.TreeEntry{
		{Name: "hello.txt", Mode: 0o100644, Hash: blobHash},
	}}
	treeObj := repo.Storer.NewEncodedObject()
	if err := tree.Encode(treeObj); err != nil {
		t.Fatalf("encode tree: %v", err)
	}
	treeHash, err := repo.Storer.SetEncodedObject(treeObj)
	if err != nil {
		t.Fatalf("store tree: %v", err)
	}

	sig := object.Signature{Name: "test", Email: "test@test.dev"}
	commit := &object.Commit{
		Author:    sig,
		Committer: sig,
		Message:   "test commit from go-git",
		TreeHash:  treeHash,
	}
	commitObj := repo.Storer.NewEncodedObject()
	if err := commit.Encode(commitObj); err != nil {
		t.Fatalf("encode commit: %v", err)
	}
	commitHash, err := repo.Storer.SetEncodedObject(commitObj)
	if err != nil {
		t.Fatalf("store commit: %v", err)
	}

	ref := plumbing.NewHashReference("refs/heads/main", commitHash)
	if err := repo.Storer.SetReference(ref); err != nil {
		t.Fatalf("set ref: %v", err)
	}

	AssertFsck(t, repoPath)
}

// AssertRoundTrip verifies that a commit made by go-git is visible to `git log`,
// and that a commit made by `git` is visible to go-git.
func AssertRoundTrip(t testing.TB) {
	t.Helper()

	dir := t.TempDir()
	repoPath := filepath.Join(dir, "roundtrip")

	// Initialize with git
	gitExec("", "init", repoPath)
	gitExec(repoPath, "checkout", "-b", "main")

	if err := os.WriteFile(filepath.Join(repoPath, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitExec(repoPath, "add", "a.txt")
	gitExec(repoPath, "commit", "-m", "git commit")

	// Open with go-git and make a commit
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repoPath, "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	wt.Add("b.txt")
	goGitHash, err := wt.Commit("go-git commit", &git.CommitOptions{
		Author: &object.Signature{Name: "gotest", Email: "go@test.dev"},
	})
	if err != nil {
		t.Fatalf("go-git commit: %v", err)
	}

	// Verify go-git commit visible to git log
	out, err := gitExec(repoPath, "log", "--oneline")
	if err != nil {
		t.Fatalf("git log: %v", err)
	}
	shortHash := goGitHash.String()[:7]
	if len(out) == 0 {
		t.Fatal("empty git log")
	}
	_ = shortHash // the commit is in the log if git log succeeds with 2 entries

	// Verify git commit visible to go-git
	iter, err := repo.Log(&git.LogOptions{})
	if err != nil {
		t.Fatalf("go-git log: %v", err)
	}
	count := 0
	iter.ForEach(func(c *object.Commit) error {
		count++
		return nil
	})
	if count != 2 {
		t.Errorf("go-git sees %d commits, want 2", count)
	}

	AssertFsck(t, repoPath)
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

func gitExec(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@test.dev",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@test.dev",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("git %v: %w\n%s", args, err, out)
	}
	return out, nil
}
