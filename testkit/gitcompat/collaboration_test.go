package gitcompat

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/vcs"
	"github.com/dpopsuev/mos/moslib/vcs/history"
	"github.com/dpopsuev/mos/moslib/vcs/merge"
	"github.com/dpopsuev/mos/moslib/vcs/staging"
	"github.com/dpopsuev/mos/moslib/vcs/transport"
	"github.com/dpopsuev/mos/testkit/forge"
)

func initDevRepo(t *testing.T, name string) (string, *vcs.Repository) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	os.MkdirAll(filepath.Join(dir, ".mos"), 0755)
	repo, err := vcs.InitRepo(dir, "git")
	if err != nil {
		t.Fatalf("InitRepo %s: %v", name, err)
	}
	return dir, repo
}

func writeGovFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	full := filepath.Join(root, relPath)
	os.MkdirAll(filepath.Dir(full), 0755)
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
}

// CON-2026-251: Concurrent Governance Push-Pull-Merge

func TestConcurrentNonConflictingMerge(t *testing.T) {
	f := forge.InProcess(t)
	remoteURL, err := f.CreateRepo("collab-nonconflict")
	if err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	aliceDir, alice := initDevRepo(t, "alice")
	transport.AddRemote(alice, "origin", remoteURL)

	writeGovFile(t, aliceDir, ".mos/config.mos", "config { version = 1 }")
	staging.Add(alice, []string{"."})
	staging.Commit(alice, "alice", "alice@dev", "initial governance")
	if err := transport.Push(alice, "origin", transport.PushOpts{}); err != nil {
		t.Fatalf("alice initial push: %v", err)
	}

	bobDir, bob := initDevRepo(t, "bob")
	transport.AddRemote(bob, "origin", remoteURL)
	_, err = transport.Pull(bob, "origin", "main", "bob", "bob@dev")
	if err != nil {
		t.Fatalf("bob initial pull: %v", err)
	}

	// Alice adds contract-A and pushes.
	writeGovFile(t, aliceDir, ".mos/contracts/contract-A.mos", `contract "A" { title = "Alice work" }`)
	staging.Add(alice, []string{".mos/contracts/contract-A.mos"})
	staging.Commit(alice, "alice", "alice@dev", "add contract A")
	if err := transport.Push(alice, "origin", transport.PushOpts{}); err != nil {
		t.Fatalf("alice push contract-A: %v", err)
	}

	// Bob adds contract-B locally (doesn't know about Alice's push).
	writeGovFile(t, bobDir, ".mos/contracts/contract-B.mos", `contract "B" { title = "Bob work" }`)
	staging.Add(bob, []string{".mos/contracts/contract-B.mos"})
	staging.Commit(bob, "bob", "bob@dev", "add contract B")

	// Bob tries to push -- should fail (non-fast-forward).
	pushErr := transport.Push(bob, "origin", transport.PushOpts{})
	if pushErr == nil {
		t.Fatal("expected push to fail (non-fast-forward), but it succeeded")
	}
	t.Logf("Bob push rejected as expected: %v", pushErr)

	// Bob pulls to reconcile.
	mergeResult, err := transport.Pull(bob, "origin", "main", "bob", "bob@dev")
	if err != nil {
		t.Fatalf("bob pull: %v", err)
	}
	if mergeResult.CommitHash.IsZero() {
		t.Fatal("expected non-zero merge commit")
	}
	if len(mergeResult.Conflicts) > 0 {
		t.Fatalf("expected no conflicts, got %v", mergeResult.Conflicts)
	}

	// Verify the merged tree contains both contracts.
	bobHead, _ := vcs.ResolveHead(bob)
	cd, _ := bob.Store.ReadCommit(bobHead)
	mergedTree, _ := staging.FlattenTree(bob.Store, cd.Tree, "")

	foundA, foundB, foundConfig := false, false, false
	for path := range mergedTree {
		if strings.Contains(path, "contract-A") {
			foundA = true
		}
		if strings.Contains(path, "contract-B") {
			foundB = true
		}
		if strings.Contains(path, "config.mos") {
			foundConfig = true
		}
	}
	if !foundA || !foundB || !foundConfig {
		t.Errorf("merged tree should contain both contracts and config: A=%v, B=%v, config=%v, paths=%v",
			foundA, foundB, foundConfig, treeKeys(mergedTree))
	}

	// Bob pushes the merge result.
	if err := transport.Push(bob, "origin", transport.PushOpts{}); err != nil {
		t.Fatalf("bob push after merge: %v", err)
	}

	// Alice pulls and verifies she sees both contracts.
	alicePull, err := transport.Pull(alice, "origin", "main", "alice", "alice@dev")
	if err != nil {
		t.Fatalf("alice pull merged: %v", err)
	}
	if alicePull.CommitHash.IsZero() {
		t.Fatal("alice pull returned zero commit")
	}

	aliceHead, _ := vcs.ResolveHead(alice)
	acd, _ := alice.Store.ReadCommit(aliceHead)
	aliceTree, _ := staging.FlattenTree(alice.Store, acd.Tree, "")
	foundA, foundB = false, false
	for path := range aliceTree {
		if strings.Contains(path, "contract-A") {
			foundA = true
		}
		if strings.Contains(path, "contract-B") {
			foundB = true
		}
	}
	if !foundA || !foundB {
		t.Errorf("alice should see both contracts after pull: foundA=%v, foundB=%v", foundA, foundB)
	}

	t.Logf("Non-conflicting merge: success, %d artifacts in merged tree", len(mergedTree))
}

func TestConcurrentConflictDetected(t *testing.T) {
	f := forge.InProcess(t)
	remoteURL, err := f.CreateRepo("collab-conflict")
	if err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	aliceDir, alice := initDevRepo(t, "alice")
	transport.AddRemote(alice, "origin", remoteURL)

	writeGovFile(t, aliceDir, ".mos/rules/rule-R1.mos", `rule "R1" { name = "Original rule" }`)
	writeGovFile(t, aliceDir, ".mos/config.mos", "config { version = 1 }")
	staging.Add(alice, []string{"."})
	staging.Commit(alice, "alice", "alice@dev", "initial governance")
	transport.Push(alice, "origin", transport.PushOpts{})

	bobDir, bob := initDevRepo(t, "bob")
	transport.AddRemote(bob, "origin", remoteURL)
	transport.Pull(bob, "origin", "main", "bob", "bob@dev")

	// Alice modifies rule-R1 and pushes.
	writeGovFile(t, aliceDir, ".mos/rules/rule-R1.mos", `rule "R1" { name = "Alice version" enforcement = "error" }`)
	staging.Add(alice, []string{".mos/rules/rule-R1.mos"})
	staging.Commit(alice, "alice", "alice@dev", "alice modifies rule R1")
	transport.Push(alice, "origin", transport.PushOpts{})

	// Bob modifies the SAME rule differently.
	writeGovFile(t, bobDir, ".mos/rules/rule-R1.mos", `rule "R1" { name = "Bob version" enforcement = "warning" }`)
	staging.Add(bob, []string{".mos/rules/rule-R1.mos"})
	staging.Commit(bob, "bob", "bob@dev", "bob modifies rule R1")

	// Bob pulls -- should detect conflict.
	mergeResult, err := transport.Pull(bob, "origin", "main", "bob", "bob@dev")
	if err != nil {
		t.Fatalf("bob pull: %v", err)
	}
	if len(mergeResult.Conflicts) == 0 {
		t.Fatal("expected conflict on rule-R1, but Pull reported no conflicts")
	}

	foundConflict := false
	for _, c := range mergeResult.Conflicts {
		if strings.Contains(c.Path, "rule-R1") {
			foundConflict = true
			t.Logf("Conflict detected: path=%s ours=%s theirs=%s base=%s",
				c.Path, c.OursHash.Short(), c.TheirHash.Short(), c.BaseHash.Short())
		}
	}
	if !foundConflict {
		t.Fatalf("expected conflict on rule-R1, got conflicts on: %v", mergeResult.Conflicts)
	}
}

// CON-2026-253: Forge-Based Governance Branch Workflow

func logTree(t *testing.T, label string, repo *vcs.Repository) {
	t.Helper()
	head, err := vcs.ResolveHead(repo)
	if err != nil {
		t.Logf("[%s] no HEAD: %v", label, err)
		return
	}
	cd, err := repo.Store.ReadCommit(head)
	if err != nil {
		t.Logf("[%s] HEAD=%s but ReadCommit failed: %v", label, head.Short(), err)
		return
	}
	tree, err := staging.FlattenTree(repo.Store, cd.Tree, "")
	if err != nil {
		t.Logf("[%s] HEAD=%s msg=%q FlattenTree failed: %v", label, head.Short(), cd.Message, err)
		return
	}
	t.Logf("[%s] HEAD=%s msg=%q branch=%s parents=%d paths=%v",
		label, head.Short(), cd.Message, vcs.CurrentBranch(repo.Root), len(cd.Parents), treeKeys(tree))
}

func logRefs(t *testing.T, label string, repo *vcs.Repository) {
	t.Helper()
	branches, _ := history.ListBranches(repo.Store)
	names := make([]string, len(branches))
	for i, b := range branches {
		names[i] = b.Name + "=" + b.Hash.Short()
	}
	t.Logf("[%s] branches: %v current=%s", label, names, vcs.CurrentBranch(repo.Root))
}

func TestForgeBranchWorkflow(t *testing.T) {
	f := forge.InProcess(t)
	remoteURL, err := f.CreateRepo("branch-workflow")
	if err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	// Developer 1 (maintainer): establish baseline on main.
	d1Dir, d1 := initDevRepo(t, "maintainer")
	transport.AddRemote(d1, "origin", remoteURL)

	writeGovFile(t, d1Dir, ".mos/config.mos", "config { version = 1 }")
	writeGovFile(t, d1Dir, ".mos/rules/rule-base.mos", `rule "base" { name = "Baseline rule" }`)
	staging.Add(d1, []string{"."})
	baseCommit, _ := staging.Commit(d1, "maintainer", "m@dev", "baseline governance")
	t.Logf("maintainer baseline commit: %s", baseCommit.Short())
	logTree(t, "maintainer after baseline", d1)

	if err := transport.Push(d1, "origin", transport.PushOpts{}); err != nil {
		t.Fatalf("maintainer push baseline: %v", err)
	}
	t.Log("maintainer pushed baseline to forge")

	// Developer 2 (contributor): clone, create feature branch, add a contract.
	d2Dir, d2 := initDevRepo(t, "contributor")
	transport.AddRemote(d2, "origin", remoteURL)
	pullResult, err := transport.Pull(d2, "origin", "main", "contributor", "c@dev")
	if err != nil {
		t.Fatalf("contributor initial pull: %v", err)
	}
	t.Logf("contributor initial pull: commit=%s ff=%v", pullResult.CommitHash.Short(), pullResult.FastForward)
	logTree(t, "contributor after initial pull", d2)
	logRefs(t, "contributor after initial pull", d2)

	contribHead, _ := vcs.ResolveHead(d2)
	if err := history.CreateBranch(d2.Store, "feature", contribHead); err != nil {
		t.Fatalf("create feature branch: %v", err)
	}
	if _, err := history.Checkout(d2, "feature"); err != nil {
		t.Fatalf("checkout feature: %v", err)
	}
	logRefs(t, "contributor after checkout feature", d2)

	writeGovFile(t, d2Dir, ".mos/contracts/new-feature.mos", `contract "new-feature" { title = "Contributor feature" }`)
	staging.Add(d2, []string{".mos/contracts/new-feature.mos"})
	featCommit, _ := staging.Commit(d2, "contributor", "c@dev", "add feature contract")
	t.Logf("contributor feature commit: %s", featCommit.Short())
	logTree(t, "contributor feature branch", d2)

	// Push feature branch to forge.
	if err := transport.Push(d2, "origin", transport.PushOpts{Branch: "feature"}); err != nil {
		t.Fatalf("push feature branch: %v", err)
	}
	t.Log("contributor pushed feature branch to forge")

	// Meanwhile, maintainer advances main with a new rule.
	writeGovFile(t, d1Dir, ".mos/rules/rule-new.mos", `rule "new" { name = "Maintainer added rule" }`)
	staging.Add(d1, []string{".mos/rules/rule-new.mos"})
	mainAdvance, _ := staging.Commit(d1, "maintainer", "m@dev", "add new rule")
	t.Logf("maintainer main advance commit: %s", mainAdvance.Short())
	logTree(t, "maintainer after advance", d1)
	if err := transport.Push(d1, "origin", transport.PushOpts{}); err != nil {
		t.Fatalf("maintainer push advance: %v", err)
	}
	t.Log("maintainer pushed advance to forge")

	// Contributor pulls main into feature branch to reconcile.
	t.Logf("contributor about to pull main while on branch=%s", vcs.CurrentBranch(d2.Root))
	logRefs(t, "contributor before pull", d2)

	pullResult, err = transport.Pull(d2, "origin", "main", "contributor", "c@dev")
	if err != nil {
		t.Fatalf("contributor pull main: %v", err)
	}
	t.Logf("contributor pull result: commit=%s ff=%v conflicts=%d",
		pullResult.CommitHash.Short(), pullResult.FastForward, len(pullResult.Conflicts))
	logTree(t, "contributor after pull main into feature", d2)
	logRefs(t, "contributor after pull", d2)

	if len(pullResult.Conflicts) > 0 {
		t.Fatalf("expected no conflicts, got %v", pullResult.Conflicts)
	}

	// Verify feature branch now has both the new rule and the new contract.
	featHead, _ := vcs.ResolveHead(d2)
	featCD, _ := d2.Store.ReadCommit(featHead)
	featTree, _ := staging.FlattenTree(d2.Store, featCD.Tree, "")
	t.Logf("feature tree after reconcile: %v", treeKeys(featTree))

	foundRule, foundContract := false, false
	for path := range featTree {
		if strings.Contains(path, "rule-new") {
			foundRule = true
		}
		if strings.Contains(path, "new-feature") {
			foundContract = true
		}
	}
	if !foundRule || !foundContract {
		t.Errorf("feature branch should have both rule-new and new-feature: rule=%v, contract=%v",
			foundRule, foundContract)
	}

	// Contributor merges feature into main and pushes.
	if _, err := history.Checkout(d2, "main"); err != nil {
		t.Fatalf("checkout main: %v", err)
	}
	logTree(t, "contributor main before pull", d2)

	d2Pull, err := transport.Pull(d2, "origin", "main", "contributor", "c@dev")
	if err != nil {
		t.Fatalf("contributor pull main before merge: %v", err)
	}
	t.Logf("contributor pull main: commit=%s ff=%v", d2Pull.CommitHash.Short(), d2Pull.FastForward)
	logTree(t, "contributor main after pull", d2)

	mergeResult, err := merge.Merge(d2, "feature", "contributor", "c@dev")
	if err != nil {
		t.Fatalf("merge feature into main: %v", err)
	}
	t.Logf("merge result: commit=%s ff=%v conflicts=%d",
		mergeResult.CommitHash.Short(), mergeResult.FastForward, len(mergeResult.Conflicts))
	logTree(t, "contributor main after merge", d2)

	if len(mergeResult.Conflicts) > 0 {
		t.Fatalf("merge conflicts: %v", mergeResult.Conflicts)
	}
	if err := transport.Push(d2, "origin", transport.PushOpts{}); err != nil {
		t.Fatalf("push merged main: %v", err)
	}
	t.Log("contributor pushed merged main")

	// Maintainer pulls and verifies everything.
	d1Pull, err := transport.Pull(d1, "origin", "main", "maintainer", "m@dev")
	if err != nil {
		t.Fatalf("maintainer final pull: %v", err)
	}
	t.Logf("maintainer final pull: commit=%s ff=%v", d1Pull.CommitHash.Short(), d1Pull.FastForward)
	logTree(t, "maintainer final state", d1)

	d1Head, _ := vcs.ResolveHead(d1)
	d1CD, _ := d1.Store.ReadCommit(d1Head)
	finalTree, _ := staging.FlattenTree(d1.Store, d1CD.Tree, "")
	t.Logf("final tree paths: %v", treeKeys(finalTree))

	foundRule, foundContract, foundBase := false, false, false
	for path := range finalTree {
		if strings.Contains(path, "rule-new") {
			foundRule = true
		}
		if strings.Contains(path, "new-feature") {
			foundContract = true
		}
		if strings.Contains(path, "rule-base") {
			foundBase = true
		}
	}
	if !foundRule || !foundContract || !foundBase {
		t.Errorf("final main should have all artifacts: rule-new=%v, new-feature=%v, rule-base=%v",
			foundRule, foundContract, foundBase)
	}

	_ = baseCommit
	t.Logf("Branch workflow: %d artifacts in final tree", len(finalTree))
}

func treeKeys(m map[string]vcs.Hash) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
