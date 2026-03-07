package chain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/artifact"
)

func setupScaffold(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	content := "module github.com/test/scaffold\n\ngo 1.25.7\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(content), 0644); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}
	if err := artifact.Init(root, artifact.InitOpts{Name: "scaffold", Model: "bdfl", Scope: "cabinet"}); err != nil {
		t.Fatalf("scaffold setup failed: %v", err)
	}
	return root
}

func TestCON036_WalkChainUpward(t *testing.T) {
	root := setupScaffold(t)
	reg, err := artifact.LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	needTD := reg.Types["need"]
	artifact.GenericCreate(root, needTD, "NEED-001", map[string]string{
		"title": "Faster CI", "sensation": "CI is slow", "status": "identified",
	})

	specTD := reg.Types["specification"]
	artifact.GenericCreate(root, specTD, "SPEC-001", map[string]string{
		"title": "Pipeline optimization", "enforcement": "warn", "status": "active",
		"satisfies": "NEED-001",
	})

	ch, err := WalkChain(root, "specification", "SPEC-001")
	if err != nil {
		t.Fatalf("WalkChain: %v", err)
	}
	if ch.Root.ID != "SPEC-001" {
		t.Errorf("root = %q, want SPEC-001", ch.Root.ID)
	}
	if len(ch.Upward) != 1 {
		t.Fatalf("expected 1 upward link, got %d", len(ch.Upward))
	}
	if ch.Upward[0].ID != "NEED-001" {
		t.Errorf("upward[0] = %q, want NEED-001", ch.Upward[0].ID)
	}
}

func TestCON036_WalkChainDownward(t *testing.T) {
	root := setupScaffold(t)
	reg, err := artifact.LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	needTD := reg.Types["need"]
	artifact.GenericCreate(root, needTD, "NEED-001", map[string]string{
		"title": "Faster CI", "sensation": "CI is slow", "status": "identified",
	})

	specTD := reg.Types["specification"]
	artifact.GenericCreate(root, specTD, "SPEC-001", map[string]string{
		"title": "Pipeline spec", "enforcement": "warn", "status": "active",
		"satisfies": "NEED-001",
	})

	ch, err := WalkChain(root, "need", "NEED-001")
	if err != nil {
		t.Fatalf("WalkChain: %v", err)
	}
	if len(ch.Downward) < 1 {
		t.Fatal("expected at least 1 downward link from need to spec")
	}
	found := false
	for _, link := range ch.Downward {
		if link.ID == "SPEC-001" {
			found = true
		}
	}
	if !found {
		t.Error("SPEC-001 not found in downward chain from NEED-001")
	}
}

func TestCON036_FormatChainOutput(t *testing.T) {
	ch := &ChainResult{
		Root: ChainLink{Kind: "specification", ID: "SPEC-001", Title: "Pipeline spec"},
		Upward: []ChainLink{
			{Kind: "need", ID: "NEED-001", Title: "Faster CI"},
		},
		Downward: []ChainLink{
			{Kind: "doc", ID: "DOC-001", Title: "Pipeline docs"},
		},
	}

	output := FormatChain(ch)
	if !strings.Contains(output, "NEED-001") {
		t.Error("chain output should contain NEED-001")
	}
	if !strings.Contains(output, "SPEC-001") {
		t.Error("chain output should contain SPEC-001")
	}
	if !strings.Contains(output, "DOC-001") {
		t.Error("chain output should contain DOC-001")
	}
	if !strings.Contains(output, "you are here") {
		t.Error("chain output should mark the root")
	}
}

func TestCON036_NegativeSpaceChain(t *testing.T) {
	root := setupScaffold(t)
	reg, err := artifact.LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	needTD := reg.Types["need"]
	needPath, _ := artifact.GenericCreate(root, needTD, "NEED-NEG-001", map[string]string{
		"title": "Scoped need", "sensation": "Pain", "status": "identified",
	})

	data, _ := os.ReadFile(needPath)
	f, _ := dsl.Parse(string(data), nil)
	ab := f.Artifact.(*dsl.ArtifactBlock)
	ab.Items = append(ab.Items, &dsl.Block{
		Name: "scope",
		Items: []dsl.Node{
			&dsl.Field{Key: "includes", Value: &dsl.ListVal{Items: []dsl.Value{
				&dsl.StringVal{Text: "caching"},
			}}},
			&dsl.Field{Key: "excludes", Value: &dsl.ListVal{Items: []dsl.Value{
				&dsl.StringVal{Text: "rewriting tests"},
				&dsl.StringVal{Text: "migration"},
			}}},
		},
	})
	formatted := dsl.Format(f, nil)
	os.WriteFile(needPath, []byte(formatted), 0644)

	nc, err := WalkNegativeChain(root, "need", "NEED-NEG-001")
	if err != nil {
		t.Fatalf("WalkNegativeChain: %v", err)
	}

	if len(nc.Entries) == 0 {
		t.Fatal("expected negative-space entries")
	}

	found := false
	for _, e := range nc.Entries {
		if e.Kind == "need" && e.Field == "scope.excludes" {
			found = true
			if len(e.Values) != 2 {
				t.Errorf("expected 2 excludes, got %d", len(e.Values))
			}
		}
	}
	if !found {
		t.Error("need scope.excludes not in negative chain")
	}
}

func TestCON036_FormatNegativeChainOutput(t *testing.T) {
	nc := &NegativeChainResult{
		Entries: []NegativeSpaceEntry{
			{Kind: "need", ID: "NEED-001", Field: "scope.excludes", Values: []string{"migration"}},
			{Kind: "specification", ID: "SPEC-001", Field: "non_goals", Values: []string{"rewrite"}},
		},
	}

	output := FormatNegativeChain(nc)
	if !strings.Contains(output, "scope.excludes") {
		t.Error("should contain scope.excludes")
	}
	if !strings.Contains(output, "non_goals") {
		t.Error("should contain non_goals")
	}
	if !strings.Contains(output, "migration") {
		t.Error("should contain migration")
	}
}

func TestCON036_EmptyNegativeChain(t *testing.T) {
	nc := &NegativeChainResult{}
	output := FormatNegativeChain(nc)
	if !strings.Contains(output, "no negative-space") {
		t.Error("should indicate no negative-space found")
	}
}
