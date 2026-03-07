//go:build integration

package forge_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dpopsuev/mos/testkit/forge"
)

func TestGiteaCreateRepo(t *testing.T) {
	f := forge.Gitea(t)

	url, err := f.CreateRepo("smoke-test")
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	if url == "" {
		t.Fatal("url is empty")
	}
	t.Logf("repo URL: %s", url)
}

func TestGiteaCloneAndPush(t *testing.T) {
	f := forge.Gitea(t)

	repoURL, err := f.CreateRepo("push-test")
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}

	cloneDir := filepath.Join(t.TempDir(), "clone")
	if _, err := forge.GitExec("", "clone", repoURL, cloneDir); err != nil {
		t.Fatalf("clone: %v", err)
	}

	testFile := filepath.Join(cloneDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello gitea"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := forge.GitExec(cloneDir, "add", "test.txt"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := forge.GitExec(cloneDir, "commit", "-m", "test commit"); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if _, err := forge.GitExec(cloneDir, "push"); err != nil {
		t.Fatalf("push: %v", err)
	}

	verifyDir := filepath.Join(t.TempDir(), "verify")
	if _, err := forge.GitExec("", "clone", repoURL, verifyDir); err != nil {
		t.Fatalf("verify clone: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(verifyDir, "test.txt"))
	if err != nil {
		t.Fatalf("read test.txt: %v", err)
	}
	if string(content) != "hello gitea" {
		t.Errorf("content = %q, want %q", content, "hello gitea")
	}
}
