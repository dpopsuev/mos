package vcs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dpopsuev/mos/moslib/vcs"
	"github.com/dpopsuev/mos/moslib/vcs/history"
	"github.com/dpopsuev/mos/moslib/vcs/staging"
)

func TestInitAndOpenRepo(t *testing.T) {
	for _, backend := range []string{"fs", "git"} {
		t.Run(backend, func(t *testing.T) {
			dir := t.TempDir()
			repo, err := vcs.InitRepo(dir, backend)
			if err != nil {
				t.Fatal(err)
			}
			if repo.Config.Backend != backend {
				t.Fatalf("expected backend %q, got %q", backend, repo.Config.Backend)
			}

			repo2, err := vcs.OpenRepo(dir)
			if err != nil {
				t.Fatal(err)
			}
			if repo2.Config.Backend != backend {
				t.Fatalf("reopen: expected backend %q, got %q", backend, repo2.Config.Backend)
			}
		})
	}
}

func TestAddAndCommit(t *testing.T) {
	for _, backend := range []string{"fs", "git"} {
		t.Run(backend, func(t *testing.T) {
			dir := t.TempDir()

			mosDir := filepath.Join(dir, ".mos", "contracts", "active", "CON-001")
			os.MkdirAll(mosDir, 0755)
			os.WriteFile(filepath.Join(mosDir, "contract.mos"), []byte(`contract "CON-001" { title = "test" }`), 0644)
			os.WriteFile(filepath.Join(dir, ".mos", "config.mos"), []byte(`mos { version = 1 }`), 0644)

			repo, err := vcs.InitRepo(dir, backend)
			if err != nil {
				t.Fatal(err)
			}

			if err := staging.Add(repo, nil); err != nil {
				t.Fatal(err)
			}

			idx, err := staging.LoadIndex(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(idx.Entries) < 2 {
				t.Fatalf("expected at least 2 index entries, got %d", len(idx.Entries))
			}

			ch, err := staging.Commit(repo, "test-agent", "agent@mos", "initial governance snapshot")
			if err != nil {
				t.Fatal(err)
			}
			if ch.IsZero() {
				t.Fatal("commit hash is zero")
			}

			head, err := vcs.ResolveHead(repo)
			if err != nil {
				t.Fatal(err)
			}
			if head != ch {
				t.Fatalf("HEAD mismatch: %s vs %s", head, ch)
			}

			cd, err := repo.Store.ReadCommit(ch)
			if err != nil {
				t.Fatal(err)
			}
			if cd.Message != "initial governance snapshot" {
				t.Fatalf("message mismatch: %q", cd.Message)
			}
			if len(cd.Parents) != 0 {
				t.Fatalf("expected no parents for initial commit, got %d", len(cd.Parents))
			}
		})
	}
}

func TestSecondCommitHasParent(t *testing.T) {
	dir := t.TempDir()

	mosDir := filepath.Join(dir, ".mos", "contracts", "active", "CON-001")
	os.MkdirAll(mosDir, 0755)
	os.WriteFile(filepath.Join(mosDir, "contract.mos"), []byte(`contract "CON-001" { title = "v1" }`), 0644)

	repo, _ := vcs.InitRepo(dir, "fs")
	staging.Add(repo, nil)
	c1, _ := staging.Commit(repo, "a", "a@b", "first")

	os.WriteFile(filepath.Join(mosDir, "contract.mos"), []byte(`contract "CON-001" { title = "v2" }`), 0644)
	staging.Add(repo, nil)
	c2, err := staging.Commit(repo, "a", "a@b", "second")
	if err != nil {
		t.Fatal(err)
	}

	cd, _ := repo.Store.ReadCommit(c2)
	if len(cd.Parents) != 1 || cd.Parents[0] != c1 {
		t.Fatalf("expected parent %s, got %v", c1, cd.Parents)
	}
}

func TestLog(t *testing.T) {
	dir := t.TempDir()
	mosDir := filepath.Join(dir, ".mos")
	os.MkdirAll(mosDir, 0755)
	os.WriteFile(filepath.Join(mosDir, "config.mos"), []byte("v1"), 0644)

	repo, _ := vcs.InitRepo(dir, "fs")
	staging.Add(repo, nil)
	staging.Commit(repo, "a", "a@b", "first")

	os.WriteFile(filepath.Join(mosDir, "config.mos"), []byte("v2"), 0644)
	staging.Add(repo, nil)
	staging.Commit(repo, "a", "a@b", "second")

	os.WriteFile(filepath.Join(mosDir, "config.mos"), []byte("v3"), 0644)
	staging.Add(repo, nil)
	staging.Commit(repo, "a", "a@b", "third")

	head, _ := vcs.ResolveHead(repo)
	entries, err := history.Log(repo.Store, head, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 log entries, got %d", len(entries))
	}
	if entries[0].Commit.Message != "third" {
		t.Fatalf("expected newest first, got %q", entries[0].Commit.Message)
	}
	if entries[2].Commit.Message != "first" {
		t.Fatalf("expected oldest last, got %q", entries[2].Commit.Message)
	}
}

func TestDiffWorkingVsIndex(t *testing.T) {
	dir := t.TempDir()
	mosDir := filepath.Join(dir, ".mos")
	os.MkdirAll(mosDir, 0755)
	os.WriteFile(filepath.Join(mosDir, "a.mos"), []byte("old"), 0644)

	repo, _ := vcs.InitRepo(dir, "fs")
	staging.Add(repo, nil)

	os.WriteFile(filepath.Join(mosDir, "a.mos"), []byte("new"), 0644)
	os.WriteFile(filepath.Join(mosDir, "b.mos"), []byte("added"), 0644)

	idx, _ := staging.LoadIndex(dir)
	indexMap := staging.IndexToMap(idx)

	workEntries, _ := staging.SnapshotWorkingTree(dir, repo.Store)
	workMap := map[string]vcs.Hash{}
	for _, e := range workEntries {
		workMap[e.Path] = e.Hash
	}

	diffs := staging.DiffTrees(indexMap, workMap)
	if len(diffs) != 2 {
		t.Fatalf("expected 2 diffs, got %d: %v", len(diffs), diffs)
	}

	byPath := map[string]staging.DiffEntry{}
	for _, d := range diffs {
		byPath[d.Path] = d
	}
	if byPath[".mos/a.mos"].Kind != staging.DiffModified {
		t.Fatalf("expected a.mos modified, got %s", byPath[".mos/a.mos"].Kind)
	}
	if byPath[".mos/b.mos"].Kind != staging.DiffAdded {
		t.Fatalf("expected b.mos added, got %s", byPath[".mos/b.mos"].Kind)
	}
}

func TestMigrateFSToGitAndBack(t *testing.T) {
	dir := t.TempDir()
	mosDir := filepath.Join(dir, ".mos")
	os.MkdirAll(mosDir, 0755)
	os.WriteFile(filepath.Join(mosDir, "config.mos"), []byte("content"), 0644)

	repo, _ := vcs.InitRepo(dir, "fs")
	staging.Add(repo, nil)
	_, err := staging.Commit(repo, "a", "a@b", "pre-migrate")
	if err != nil {
		t.Fatal(err)
	}

	if err := vcs.Migrate(repo, "git"); err != nil {
		t.Fatal(err)
	}
	if repo.Config.Backend != "git" {
		t.Fatalf("expected git backend after migrate, got %q", repo.Config.Backend)
	}

	head, err := vcs.ResolveHead(repo)
	if err != nil {
		t.Fatal(err)
	}
	cd, err := repo.Store.ReadCommit(head)
	if err != nil {
		t.Fatal(err)
	}
	if cd.Message != "pre-migrate" {
		t.Fatalf("commit message lost in migration: %q", cd.Message)
	}
}
