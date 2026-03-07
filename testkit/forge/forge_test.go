package forge_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dpopsuev/mos/testkit/forge"
)

func TestInProcessCreateRepo(t *testing.T) {
	f := forge.InProcess(t)

	url, err := f.CreateRepo("test-project")
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	if url == "" {
		t.Fatal("url is empty")
	}
}

func TestCloneAndVerify(t *testing.T) {
	f := forge.InProcess(t)

	repoURL, err := f.CreateRepo("clone-test")
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}

	cloneDir := filepath.Join(t.TempDir(), "clone")
	if _, err := forge.GitExec("", "clone", repoURL, cloneDir); err != nil {
		t.Fatalf("clone: %v", err)
	}

	readme := filepath.Join(cloneDir, "README.md")
	if _, err := os.Stat(readme); err != nil {
		t.Fatalf("README.md not found: %v", err)
	}

	out, err := forge.GitExec(cloneDir, "log", "--oneline", "-1")
	if err != nil {
		t.Fatalf("log: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("empty log output")
	}
}

func TestPushAndFetch(t *testing.T) {
	f := forge.InProcess(t)

	repoURL, err := f.CreateRepo("push-test")
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}

	// Alice clones
	aliceDir := filepath.Join(t.TempDir(), "alice")
	if _, err := forge.GitExec("", "clone", repoURL, aliceDir); err != nil {
		t.Fatalf("alice clone: %v", err)
	}

	// Alice creates a file, commits, pushes
	testFile := filepath.Join(aliceDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello from alice"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := forge.GitExec(aliceDir, "add", "test.txt"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := forge.GitExec(aliceDir, "commit", "-m", "alice's commit"); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if _, err := forge.GitExec(aliceDir, "push"); err != nil {
		t.Fatalf("push: %v", err)
	}

	// Bob clones and should see Alice's commit
	bobDir := filepath.Join(t.TempDir(), "bob")
	if _, err := forge.GitExec("", "clone", repoURL, bobDir); err != nil {
		t.Fatalf("bob clone: %v", err)
	}

	bobFile := filepath.Join(bobDir, "test.txt")
	content, err := os.ReadFile(bobFile)
	if err != nil {
		t.Fatalf("read test.txt in bob's clone: %v", err)
	}
	if string(content) != "hello from alice" {
		t.Errorf("content = %q, want %q", content, "hello from alice")
	}
}

func TestTwoUserPush(t *testing.T) {
	f := forge.InProcess(t)

	repoURL, err := f.CreateRepo("two-user")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Alice clones
	aliceDir := filepath.Join(t.TempDir(), "alice")
	if _, err := forge.GitExec("", "clone", repoURL, aliceDir); err != nil {
		t.Fatalf("alice clone: %v", err)
	}

	// Bob clones
	bobDir := filepath.Join(t.TempDir(), "bob")
	if _, err := forge.GitExec("", "clone", repoURL, bobDir); err != nil {
		t.Fatalf("bob clone: %v", err)
	}

	// Alice pushes
	if err := os.WriteFile(filepath.Join(aliceDir, "alice.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	forge.GitExec(aliceDir, "add", "alice.txt")
	forge.GitExec(aliceDir, "commit", "-m", "alice")
	if _, err := forge.GitExec(aliceDir, "push"); err != nil {
		t.Fatalf("alice push: %v", err)
	}

	// Bob pulls, then pushes
	if _, err := forge.GitExec(bobDir, "pull"); err != nil {
		t.Fatalf("bob pull: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bobDir, "bob.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	forge.GitExec(bobDir, "add", "bob.txt")
	forge.GitExec(bobDir, "commit", "-m", "bob")
	if _, err := forge.GitExec(bobDir, "push"); err != nil {
		t.Fatalf("bob push: %v", err)
	}

	// Verify both files exist by cloning fresh
	verifyDir := filepath.Join(t.TempDir(), "verify")
	forge.GitExec("", "clone", repoURL, verifyDir)

	for _, name := range []string{"alice.txt", "bob.txt", "README.md"} {
		if _, err := os.Stat(filepath.Join(verifyDir, name)); err != nil {
			t.Errorf("file %s missing in final clone: %v", name, err)
		}
	}
}
