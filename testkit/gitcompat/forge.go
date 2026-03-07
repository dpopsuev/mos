package gitcompat

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	"github.com/dpopsuev/mos/moslib/vcs"
	"github.com/dpopsuev/mos/testkit/forge"
)

// AssertGovernanceRoundTrip verifies that governance objects written to a Git
// repository via GitStore and pushed to a remote Forge survive a full
// clone-push-clone cycle. This directly validates NEED-2026-004 criterion C7:
// "git push transports governance refs alongside code refs."
//
// The test flow:
//  1. Clone the repo ("Alice")
//  2. Write governance blob, tree, commit via GitStore
//  3. Update refs/mos/main
//  4. Push everything including refs/mos/* to the forge
//  5. Fresh clone ("Bob")
//  6. Open GitStore on Bob's clone
//  7. Assert refs/mos/main resolves to the same commit hash
//  8. Assert blob content is identical
func AssertGovernanceRoundTrip(t testing.TB, f forge.Forge) {
	t.Helper()

	repoURL, err := f.CreateRepo("governance-roundtrip")
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}

	// --- Alice: clone, write governance objects, push ---
	aliceDir := filepath.Join(t.TempDir(), "alice")
	if _, err := forge.GitExec("", "clone", repoURL, aliceDir); err != nil {
		t.Fatalf("alice clone: %v", err)
	}

	aliceStore, err := vcs.NewGitStore(aliceDir)
	if err != nil {
		t.Fatalf("alice gitstore: %v", err)
	}

	blobContent := []byte("governance payload v1")
	blobHash, err := aliceStore.StoreBlob(blobContent)
	if err != nil {
		t.Fatalf("store blob: %v", err)
	}

	treeHash, err := aliceStore.StoreTree([]vcs.TreeEntry{
		{Name: "artifact.json", Hash: blobHash, Mode: vcs.ModeRegular},
	})
	if err != nil {
		t.Fatalf("store tree: %v", err)
	}

	commitHash, err := aliceStore.StoreCommit(vcs.CommitData{
		Tree:    treeHash,
		Author:  "mos-test",
		Email:   "test@mos.dev",
		Time:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Message: "governance: initial snapshot",
	})
	if err != nil {
		t.Fatalf("store commit: %v", err)
	}

	if err := aliceStore.UpdateRef("main", commitHash); err != nil {
		t.Fatalf("update ref: %v", err)
	}

	// Push code refs and governance refs
	if _, err := forge.GitExec(aliceDir, "push", "origin", "main"); err != nil {
		t.Fatalf("push main: %v", err)
	}
	if _, err := forge.GitExec(aliceDir, "push", "origin", "refs/mos/main:refs/mos/main"); err != nil {
		t.Fatalf("push refs/mos/main: %v", err)
	}

	// --- Bob: fresh clone, verify governance objects ---
	bobDir := filepath.Join(t.TempDir(), "bob")
	if _, err := forge.GitExec("", "clone", repoURL, bobDir); err != nil {
		t.Fatalf("bob clone: %v", err)
	}
	// Fetch governance refs explicitly (git clone only fetches refs/heads/*)
	if _, err := forge.GitExec(bobDir, "fetch", "origin", "refs/mos/*:refs/mos/*"); err != nil {
		t.Fatalf("bob fetch refs/mos: %v", err)
	}

	bobStore, err := vcs.NewGitStore(bobDir)
	if err != nil {
		t.Fatalf("bob gitstore: %v", err)
	}

	gotCommitHash, err := bobStore.ResolveRef("main")
	if err != nil {
		t.Fatalf("bob resolve ref: %v", err)
	}
	if gotCommitHash != commitHash {
		t.Errorf("refs/mos/main: got %s, want %s", gotCommitHash.Short(), commitHash.Short())
	}

	gotCommit, err := bobStore.ReadCommit(gotCommitHash)
	if err != nil {
		t.Fatalf("bob read commit: %v", err)
	}
	if gotCommit.Message != "governance: initial snapshot" {
		t.Errorf("commit message: got %q, want %q", gotCommit.Message, "governance: initial snapshot")
	}

	gotEntries, err := bobStore.ReadTree(gotCommit.Tree)
	if err != nil {
		t.Fatalf("bob read tree: %v", err)
	}
	if len(gotEntries) != 1 || gotEntries[0].Name != "artifact.json" {
		t.Fatalf("tree entries: got %+v, want [{artifact.json ...}]", gotEntries)
	}

	gotBlob, err := bobStore.ReadBlob(gotEntries[0].Hash)
	if err != nil {
		t.Fatalf("bob read blob: %v", err)
	}
	if !bytes.Equal(gotBlob, blobContent) {
		t.Errorf("blob content: got %q, want %q", gotBlob, blobContent)
	}
}
