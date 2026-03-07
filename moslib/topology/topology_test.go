package topology

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	gitCmd(dir, "config", "user.email", "test@test.com")
	gitCmd(dir, "config", "user.name", "test")
}

func setupTopologyRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")
	os.MkdirAll(filepath.Join(mosDir, "rules", "mechanical"), 0755)
	os.MkdirAll(filepath.Join(mosDir, "contracts", "active"), 0755)
	os.MkdirAll(filepath.Join(mosDir, "lexicon"), 0755)

	os.WriteFile(filepath.Join(mosDir, "config.mos"), []byte("project {\n  name = \"test\"\n}\n"), 0644)
	os.WriteFile(filepath.Join(mosDir, "rules", "mechanical", "R-001.mos"),
		[]byte("rule \"R-001\" {\n  name = \"Test Rule\"\n}\n"), 0644)
	os.WriteFile(filepath.Join(mosDir, "lexicon", "default.mos"),
		[]byte("lexicon {\n  terms {\n    governance = \"Self-governance framework\"\n  }\n}\n"), 0644)

	os.MkdirAll(filepath.Join(mosDir, "contracts", "active", "CON-001"), 0755)
	os.WriteFile(filepath.Join(mosDir, "contracts", "active", "CON-001", "contract.mos"),
		[]byte("contract \"CON-001\" {\n  title = \"Test Contract\"\n  status = \"active\"\n}\n"), 0644)

	initGitRepo(t, root)
	gitCmd(root, "add", "-A")
	gitCmd(root, "commit", "-m", "initial")
	return root
}

// Scenario 1: Pre-flight reports delta without mutating
func TestCON020_PreFlightDelta(t *testing.T) {
	source := setupTopologyRepo(t)
	target := setupTopologyRepo(t)

	os.WriteFile(filepath.Join(source, ".mos", "rules", "mechanical", "R-NEW.mos"),
		[]byte("rule \"R-NEW\" {\n  name = \"New Rule\"\n}\n"), 0644)

	delta, err := PreFlightDelta(source, target)
	if err != nil {
		t.Fatalf("PreFlightDelta: %v", err)
	}

	if delta.IsEmpty() {
		t.Error("expected non-empty delta")
	}

	found := false
	for _, a := range delta.Added {
		if strings.Contains(a, "R-NEW") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected R-NEW in added files, got %v", delta.Added)
	}

	targetContent, _ := os.ReadFile(filepath.Join(target, ".mos", "config.mos"))
	if !strings.Contains(string(targetContent), "test") {
		t.Error("target should not be mutated by pre-flight")
	}
}

// Scenario 2: Checkpoint enables rollback after failed operation
func TestCON020_CheckpointRollback(t *testing.T) {
	root := setupTopologyRepo(t)

	tag, err := Checkpoint(root, "before-promote")
	if err != nil {
		t.Fatalf("Checkpoint: %v", err)
	}
	if tag == "" {
		t.Fatal("expected non-empty tag")
	}

	testFile := filepath.Join(root, ".mos", "new-file.mos")
	os.WriteFile(testFile, []byte("change"), 0644)
	gitCmd(root, "add", "-A")
	gitCmd(root, "commit", "-m", "simulated change")

	if _, err := os.Stat(testFile); err != nil {
		t.Fatal("expected new file to exist before rollback")
	}

	if err := Rollback(root, tag); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	if _, err := os.Stat(testFile); err == nil {
		t.Error("expected new file to be gone after rollback")
	}
}

// Scenario 3: Promote standalone repo into federation
func TestCON020_PromoteToFederation(t *testing.T) {
	upstream := setupTopologyRepo(t)
	target := setupTopologyRepo(t)

	if err := Promote(upstream, target); err != nil {
		t.Fatalf("Promote: %v", err)
	}

	upstreamFile := filepath.Join(target, ".mos", "upstream.mos")
	if _, err := os.Stat(upstreamFile); err != nil {
		t.Error("expected upstream.mos to exist after promote")
	}
	content, _ := os.ReadFile(upstreamFile)
	if !strings.Contains(string(content), upstream) {
		t.Error("upstream.mos should reference the upstream path")
	}

	ruleFile := filepath.Join(target, ".mos", "rules", "mechanical", "R-001.mos")
	if _, err := os.Stat(ruleFile); err != nil {
		t.Error("expected inherited rule to exist after promote")
	}
}

// Scenario 4: Promote repo from one parent to another
func TestCON020_PromoteToNewParent(t *testing.T) {
	oldParent := setupTopologyRepo(t)
	newParent := setupTopologyRepo(t)
	child := setupTopologyRepo(t)

	Promote(oldParent, child)
	Promote(newParent, child)

	content, _ := os.ReadFile(filepath.Join(child, ".mos", "upstream.mos"))
	if !strings.Contains(string(content), newParent) {
		t.Error("upstream should reference new parent after re-promote")
	}
}

// Scenario 5: Demote repo to standalone preserving inherited state
func TestCON020_DemoteToStandalone(t *testing.T) {
	upstream := setupTopologyRepo(t)
	target := setupTopologyRepo(t)

	Promote(upstream, target)

	if err := Demote(target); err != nil {
		t.Fatalf("Demote: %v", err)
	}

	upstreamFile := filepath.Join(target, ".mos", "upstream.mos")
	if _, err := os.Stat(upstreamFile); err == nil {
		t.Error("expected upstream.mos to be removed after demote")
	}

	ruleFile := filepath.Join(target, ".mos", "rules", "mechanical", "R-001.mos")
	if _, err := os.Stat(ruleFile); err != nil {
		t.Error("inherited rules should be preserved after demote")
	}
}

// Scenario 6: Split overloaded repo into two
func TestCON020_Split(t *testing.T) {
	source := setupTopologyRepo(t)
	targetRoot := t.TempDir()

	os.MkdirAll(filepath.Join(source, ".mos", "contracts", "active", "CON-002"), 0755)
	os.WriteFile(filepath.Join(source, ".mos", "contracts", "active", "CON-002", "contract.mos"),
		[]byte("contract \"CON-002\" {\n  title = \"Second\"\n}\n"), 0644)

	plan := SplitPlan{
		SourceRoot: source,
		TargetRoot: targetRoot,
		Artifacts:  []string{"CON-002"},
	}

	if err := Split(plan); err != nil {
		t.Fatalf("Split: %v", err)
	}

	movedPath := filepath.Join(targetRoot, ".mos", "contracts", "active", "CON-002", "contract.mos")
	if _, err := os.Stat(movedPath); err != nil {
		t.Error("expected CON-002 to exist in target after split")
	}

	sourcePath := filepath.Join(source, ".mos", "contracts", "active", "CON-002")
	if _, err := os.Stat(sourcePath); err == nil {
		t.Error("expected CON-002 to be removed from source after split")
	}

	con001Path := filepath.Join(source, ".mos", "contracts", "active", "CON-001", "contract.mos")
	if _, err := os.Stat(con001Path); err != nil {
		t.Error("expected CON-001 to remain in source after split")
	}
}

// Scenario 7: Merge two tightly-coupled repos
func TestCON020_Merge(t *testing.T) {
	source := setupTopologyRepo(t)
	target := setupTopologyRepo(t)

	os.MkdirAll(filepath.Join(source, ".mos", "contracts", "active", "CON-SRC"), 0755)
	os.WriteFile(filepath.Join(source, ".mos", "contracts", "active", "CON-SRC", "contract.mos"),
		[]byte("contract \"CON-SRC\" {\n  title = \"Source Contract\"\n}\n"), 0644)

	if err := Merge(source, target); err != nil {
		t.Fatalf("Merge: %v", err)
	}

	mergedPath := filepath.Join(target, ".mos", "contracts", "active", "CON-SRC", "contract.mos")
	if _, err := os.Stat(mergedPath); err != nil {
		t.Error("expected CON-SRC to exist in target after merge")
	}

	existingPath := filepath.Join(target, ".mos", "contracts", "active", "CON-001", "contract.mos")
	if _, err := os.Stat(existingPath); err != nil {
		t.Error("expected existing CON-001 to remain after merge")
	}
}

// Scenario 8: Union: batch promote multiple repos into federation
func TestCON020_Union(t *testing.T) {
	upstream := setupTopologyRepo(t)
	target1 := setupTopologyRepo(t)
	target2 := setupTopologyRepo(t)

	errs := Union(upstream, []string{target1, target2})
	if len(errs) != 0 {
		t.Fatalf("Union errors: %v", errs)
	}

	for _, target := range []string{target1, target2} {
		upstreamFile := filepath.Join(target, ".mos", "upstream.mos")
		if _, err := os.Stat(upstreamFile); err != nil {
			t.Errorf("expected upstream.mos in %s after union", target)
		}
	}
}

// Scenario 9: Secede: batch demote repos from federation
func TestCON020_Secede(t *testing.T) {
	upstream := setupTopologyRepo(t)
	target1 := setupTopologyRepo(t)
	target2 := setupTopologyRepo(t)

	Promote(upstream, target1)
	Promote(upstream, target2)

	errs := Secede([]string{target1, target2})
	if len(errs) != 0 {
		t.Fatalf("Secede errors: %v", errs)
	}

	for _, target := range []string{target1, target2} {
		upstreamFile := filepath.Join(target, ".mos", "upstream.mos")
		if _, err := os.Stat(upstreamFile); err == nil {
			t.Errorf("expected upstream.mos to be removed in %s after secede", target)
		}
	}
}

// Scenario 10: History is never rewritten
func TestCON020_HistoryNotRewritten(t *testing.T) {
	root := setupTopologyRepo(t)

	out, err := exec.Command("git", "-C", root, "log", "--oneline").Output()
	if err != nil {
		t.Fatalf("git log: %v", err)
	}
	commitsBefore := strings.Count(string(out), "\n")

	Checkpoint(root, "test")

	testFile := filepath.Join(root, ".mos", "topo-test.mos")
	os.WriteFile(testFile, []byte("test"), 0644)
	gitCmd(root, "add", "-A")
	gitCmd(root, "commit", "-m", "topo change")

	out2, err := exec.Command("git", "-C", root, "log", "--oneline").Output()
	if err != nil {
		t.Fatalf("git log: %v", err)
	}
	commitsAfter := strings.Count(string(out2), "\n")

	if commitsAfter <= commitsBefore {
		t.Error("expected new commits, not rewritten history")
	}

	out3, err := exec.Command("git", "-C", root, "tag", "-l", "mos-checkpoint/*").Output()
	if err != nil {
		t.Fatalf("git tag: %v", err)
	}
	if !strings.Contains(string(out3), "mos-checkpoint/test") {
		t.Error("expected checkpoint tag to exist; history should use tags, not force push")
	}
}
