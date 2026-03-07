package vcs_test

import (
	"testing"
	"time"

	"github.com/dpopsuev/mos/moslib/vcs"
)

type storeFactory struct {
	name    string
	create  func(t *testing.T) vcs.ObjectStore
}

func factories(t *testing.T) []storeFactory {
	return []storeFactory{
		{
			name: "FSStore",
			create: func(t *testing.T) vcs.ObjectStore {
				dir := t.TempDir()
				s, err := vcs.NewFSStore(dir)
				if err != nil {
					t.Fatal(err)
				}
				return s
			},
		},
		{
			name: "GitStore",
			create: func(t *testing.T) vcs.ObjectStore {
				dir := t.TempDir()
				s, err := vcs.NewGitStore(dir)
				if err != nil {
					t.Fatal(err)
				}
				return s
			},
		},
	}
}

func TestBlobRoundTrip(t *testing.T) {
	for _, f := range factories(t) {
		t.Run(f.name, func(t *testing.T) {
			s := f.create(t)
			data := []byte("contract { title = \"hello\" }")

			h, err := s.StoreBlob(data)
			if err != nil {
				t.Fatalf("StoreBlob: %v", err)
			}
			if h.IsZero() {
				t.Fatal("expected non-zero hash")
			}

			got, err := s.ReadBlob(h)
			if err != nil {
				t.Fatalf("ReadBlob: %v", err)
			}
			if string(got) != string(data) {
				t.Fatalf("ReadBlob mismatch: got %q, want %q", got, data)
			}
		})
	}
}

func TestBlobDeterministic(t *testing.T) {
	for _, f := range factories(t) {
		t.Run(f.name, func(t *testing.T) {
			s := f.create(t)
			data := []byte("deterministic content")

			h1, _ := s.StoreBlob(data)
			h2, _ := s.StoreBlob(data)
			if h1 != h2 {
				t.Fatalf("same content produced different hashes: %s vs %s", h1, h2)
			}
		})
	}
}

func TestTreeRoundTrip(t *testing.T) {
	for _, f := range factories(t) {
		t.Run(f.name, func(t *testing.T) {
			s := f.create(t)

			bh1, _ := s.StoreBlob([]byte("file1 content"))
			bh2, _ := s.StoreBlob([]byte("file2 content"))

			entries := []vcs.TreeEntry{
				{Name: "contract.mos", Hash: bh1, Mode: vcs.ModeRegular},
				{Name: "spec.mos", Hash: bh2, Mode: vcs.ModeRegular},
			}
			th, err := s.StoreTree(entries)
			if err != nil {
				t.Fatalf("StoreTree: %v", err)
			}

			got, err := s.ReadTree(th)
			if err != nil {
				t.Fatalf("ReadTree: %v", err)
			}
			if len(got) != 2 {
				t.Fatalf("expected 2 entries, got %d", len(got))
			}

			byName := map[string]vcs.TreeEntry{}
			for _, e := range got {
				byName[e.Name] = e
			}
			if byName["contract.mos"].Hash != bh1 {
				t.Fatalf("contract.mos hash mismatch")
			}
			if byName["spec.mos"].Hash != bh2 {
				t.Fatalf("spec.mos hash mismatch")
			}
		})
	}
}

func TestCommitRoundTrip(t *testing.T) {
	for _, f := range factories(t) {
		t.Run(f.name, func(t *testing.T) {
			s := f.create(t)

			bh, _ := s.StoreBlob([]byte("initial"))
			th, _ := s.StoreTree([]vcs.TreeEntry{
				{Name: "root.mos", Hash: bh, Mode: vcs.ModeRegular},
			})

			ts := time.Date(2026, 3, 4, 12, 0, 0, 0, time.UTC)
			cd := vcs.CommitData{
				Tree:    th,
				Author:  "mos",
				Email:   "mos@test",
				Time:    ts,
				Message: "initial commit",
			}

			ch, err := s.StoreCommit(cd)
			if err != nil {
				t.Fatalf("StoreCommit: %v", err)
			}

			got, err := s.ReadCommit(ch)
			if err != nil {
				t.Fatalf("ReadCommit: %v", err)
			}
			if got.Tree != th {
				t.Fatalf("tree hash mismatch: %s vs %s", got.Tree, th)
			}
			if got.Author != "mos" {
				t.Fatalf("author mismatch: %q", got.Author)
			}
			if got.Message != "initial commit" {
				t.Fatalf("message mismatch: %q", got.Message)
			}
			if len(got.Parents) != 0 {
				t.Fatalf("expected 0 parents, got %d", len(got.Parents))
			}
		})
	}
}

func TestCommitWithParent(t *testing.T) {
	for _, f := range factories(t) {
		t.Run(f.name, func(t *testing.T) {
			s := f.create(t)
			ts := time.Date(2026, 3, 4, 12, 0, 0, 0, time.UTC)

			bh1, _ := s.StoreBlob([]byte("v1"))
			th1, _ := s.StoreTree([]vcs.TreeEntry{{Name: "a.mos", Hash: bh1, Mode: vcs.ModeRegular}})
			ch1, _ := s.StoreCommit(vcs.CommitData{Tree: th1, Author: "mos", Email: "mos@test", Time: ts, Message: "first"})

			bh2, _ := s.StoreBlob([]byte("v2"))
			th2, _ := s.StoreTree([]vcs.TreeEntry{{Name: "a.mos", Hash: bh2, Mode: vcs.ModeRegular}})
			ch2, err := s.StoreCommit(vcs.CommitData{Tree: th2, Parents: []vcs.Hash{ch1}, Author: "mos", Email: "mos@test", Time: ts, Message: "second"})
			if err != nil {
				t.Fatalf("StoreCommit: %v", err)
			}

			got, _ := s.ReadCommit(ch2)
			if len(got.Parents) != 1 || got.Parents[0] != ch1 {
				t.Fatalf("parent mismatch: got %v, want [%s]", got.Parents, ch1)
			}
		})
	}
}

func TestRefManagement(t *testing.T) {
	for _, f := range factories(t) {
		t.Run(f.name, func(t *testing.T) {
			s := f.create(t)

			bh, _ := s.StoreBlob([]byte("ref test"))
			th, _ := s.StoreTree([]vcs.TreeEntry{{Name: "x.mos", Hash: bh, Mode: vcs.ModeRegular}})
			ch, _ := s.StoreCommit(vcs.CommitData{
				Tree: th, Author: "mos", Email: "mos@test",
				Time: time.Now(), Message: "ref test",
			})

			if err := s.UpdateRef("heads/main", ch); err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			resolved, err := s.ResolveRef("heads/main")
			if err != nil {
				t.Fatalf("ResolveRef: %v", err)
			}
			if resolved != ch {
				t.Fatalf("ResolveRef mismatch: %s vs %s", resolved, ch)
			}

			refs, err := s.ListRefs("heads/")
			if err != nil {
				t.Fatalf("ListRefs: %v", err)
			}
			if len(refs) != 1 || refs[0].Name != "heads/main" {
				t.Fatalf("ListRefs unexpected: %v", refs)
			}

			if err := s.DeleteRef("heads/main"); err != nil {
				t.Fatalf("DeleteRef: %v", err)
			}
			_, err = s.ResolveRef("heads/main")
			if err == nil {
				t.Fatal("expected error after DeleteRef")
			}
		})
	}
}

func TestHasObject(t *testing.T) {
	for _, f := range factories(t) {
		t.Run(f.name, func(t *testing.T) {
			s := f.create(t)

			if s.HasObject(vcs.ZeroHash) {
				t.Fatal("HasObject should return false for ZeroHash")
			}

			h, _ := s.StoreBlob([]byte("exists"))
			if !s.HasObject(h) {
				t.Fatal("HasObject should return true after store")
			}
		})
	}
}

func TestTypeOf(t *testing.T) {
	for _, f := range factories(t) {
		t.Run(f.name, func(t *testing.T) {
			s := f.create(t)

			bh, _ := s.StoreBlob([]byte("typed"))
			typ, err := s.TypeOf(bh)
			if err != nil {
				t.Fatal(err)
			}
			if typ != vcs.ObjectBlob {
				t.Fatalf("expected blob, got %s", typ)
			}

			th, _ := s.StoreTree([]vcs.TreeEntry{{Name: "f", Hash: bh, Mode: vcs.ModeRegular}})
			typ, _ = s.TypeOf(th)
			if typ != vcs.ObjectTree {
				t.Fatalf("expected tree, got %s", typ)
			}

			ch, _ := s.StoreCommit(vcs.CommitData{
				Tree: th, Author: "a", Email: "a@b",
				Time: time.Now(), Message: "m",
			})
			typ, _ = s.TypeOf(ch)
			if typ != vcs.ObjectCommit {
				t.Fatalf("expected commit, got %s", typ)
			}
		})
	}
}

func TestAllObjects(t *testing.T) {
	for _, f := range factories(t) {
		t.Run(f.name, func(t *testing.T) {
			s := f.create(t)

			h1, _ := s.StoreBlob([]byte("obj1"))
			h2, _ := s.StoreBlob([]byte("obj2"))

			all, err := s.AllObjects()
			if err != nil {
				t.Fatal(err)
			}
			found := map[vcs.Hash]bool{}
			for _, h := range all {
				found[h] = true
			}
			if !found[h1] || !found[h2] {
				t.Fatalf("AllObjects missing expected hashes")
			}
		})
	}
}
