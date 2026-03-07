package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func seedProjectConfig(t *testing.T, root string) {
	t.Helper()
	configContent := `config {
  mos {
    version = 1
  }
  backend {
    type = "git"
  }
  governance {
    model = "bdfl"
    scope = "cabinet"
  }
  project "contracts" {
    prefix = "CON"
    sequence = 0
    default = true
  }
  project "bugs" {
    prefix = "BUG"
    sequence = 0
  }
  project "features" {
    prefix = "FEAT"
    sequence = 0
  }
}
`
	if err := os.WriteFile(filepath.Join(root, ".mos", "config.mos"), []byte(configContent), 0644); err != nil {
		t.Fatalf("writing config.mos: %v", err)
	}
}

// Feature 1: Project Namespace

func TestLoadProjects(t *testing.T) {
	root := setupScaffold(t)
	seedProjectConfig(t, root)

	projects, err := LoadProjects(root)
	if err != nil {
		t.Fatalf("LoadProjects: %v", err)
	}
	if len(projects) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(projects))
	}

	byName := map[string]ProjectDef{}
	for _, p := range projects {
		byName[p.Name] = p
	}

	con := byName["contracts"]
	if con.Prefix != "CON" {
		t.Errorf("contracts.Prefix = %q, want CON", con.Prefix)
	}
	if !con.Default {
		t.Error("contracts should be default")
	}

	bugs := byName["bugs"]
	if bugs.Prefix != "BUG" {
		t.Errorf("bugs.Prefix = %q, want BUG", bugs.Prefix)
	}
	if bugs.Default {
		t.Error("bugs should not be default")
	}
}

func TestNextID(t *testing.T) {
	root := setupScaffold(t)
	seedProjectConfig(t, root)

	year := time.Now().UTC().Format("2006")

	id1, err := NextID(root, "bugs")
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	expected1 := fmt.Sprintf("BUG-%s-001", year)
	if id1 != expected1 {
		t.Errorf("first ID = %q, want %q", id1, expected1)
	}

	id2, err := NextID(root, "bugs")
	if err != nil {
		t.Fatalf("NextID second call: %v", err)
	}
	expected2 := fmt.Sprintf("BUG-%s-002", year)
	if id2 != expected2 {
		t.Errorf("second ID = %q, want %q", id2, expected2)
	}

	projects, _ := LoadProjects(root)
	for _, p := range projects {
		if p.Name == "bugs" && p.Sequence != 2 {
			t.Errorf("bugs sequence = %d, want 2", p.Sequence)
		}
	}
}

func TestNextID_SkipsArchivedIDs(t *testing.T) {
	root := setupScaffold(t)
	seedProjectConfig(t, root)

	year := time.Now().UTC().Format("2006")

	id1, err := NextID(root, "contracts")
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	expected1 := fmt.Sprintf("CON-%s-001", year)
	if id1 != expected1 {
		t.Fatalf("first ID = %q, want %q", id1, expected1)
	}

	archiveDir := filepath.Join(root, ".mos", "contracts", "archive", fmt.Sprintf("CON-%s-002", year))
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		t.Fatalf("creating archive dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(archiveDir, "contract.mos"), []byte(`contract "CON-`+year+`-002" { title = "archived" status = "complete" }`), 0644); err != nil {
		t.Fatalf("writing archived contract: %v", err)
	}

	id2, err := NextID(root, "contracts")
	if err != nil {
		t.Fatalf("NextID after archive: %v", err)
	}
	expected2 := fmt.Sprintf("CON-%s-003", year)
	if id2 != expected2 {
		t.Errorf("second ID = %q, want %q (should skip 002 in archive)", id2, expected2)
	}

	projects, _ := LoadProjects(root)
	for _, p := range projects {
		if p.Name == "contracts" && p.Sequence != 3 {
			t.Errorf("contracts sequence = %d, want 3", p.Sequence)
		}
	}
}

func TestDefaultProjectFallback(t *testing.T) {
	root := setupScaffold(t)
	seedProjectConfig(t, root)

	year := time.Now().UTC().Format("2006")

	path, err := CreateContract(root, "", ContractOpts{
		Title:   "Default project test",
		Project: "",
	})
	if err != nil {
		t.Fatalf("CreateContract with default project: %v", err)
	}
	expectedPrefix := fmt.Sprintf("CON-%s-001", year)
	if !strings.Contains(path, expectedPrefix) {
		t.Errorf("expected path to contain %q, got %q", expectedPrefix, path)
	}
}

func TestConcurrentProjectCreation(t *testing.T) {
	root := setupScaffold(t)
	seedProjectConfig(t, root)

	const n = 10
	var wg sync.WaitGroup
	ids := make([]string, n)
	errs := make([]error, n)

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			id, err := NextID(root, "bugs")
			ids[idx] = id
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	seen := map[string]bool{}
	for i, id := range ids {
		if errs[i] != nil {
			t.Errorf("goroutine %d: NextID failed: %v", i, errs[i])
			continue
		}
		if seen[id] {
			t.Errorf("duplicate ID: %s", id)
		}
		seen[id] = true
	}

	if len(seen) != n {
		t.Errorf("expected %d unique IDs, got %d", n, len(seen))
	}
}

// Feature 6: Composable Filters

func TestComposableFilters(t *testing.T) {
	root := setupScaffold(t)
	seedProjectConfig(t, root)

	CreateContract(root, "", ContractOpts{Title: "Match all", Project: "bugs", Kind: "bug", Labels: []string{"harness"}, Priority: "p1"})
	CreateContract(root, "", ContractOpts{Title: "Wrong kind", Project: "bugs", Kind: "feature", Labels: []string{"harness"}, Priority: "p1"})
	CreateContract(root, "", ContractOpts{Title: "Wrong project", Project: "features", Kind: "bug", Labels: []string{"harness"}, Priority: "p1"})
	CreateContract(root, "", ContractOpts{Title: "Wrong label", Project: "bugs", Kind: "bug", Labels: []string{"security"}, Priority: "p1"})
	CreateContract(root, "", ContractOpts{Title: "Wrong priority", Project: "bugs", Kind: "bug", Labels: []string{"harness"}, Priority: "p5"})

	results, err := ListContracts(root, ListOpts{
		Project:  "bugs",
		Kind:     "bug",
		Label:    "harness",
		Priority: "p1",
	})
	if err != nil {
		t.Fatalf("ListContracts composable: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 match, got %d", len(results))
	}
	if len(results) > 0 && results[0].Title != "Match all" {
		t.Errorf("expected 'Match all', got %q", results[0].Title)
	}
}

// --- CON-2026-022: Contract Hierarchy & Subgraphs ---

// Feature: Parent Field

func TestFindProjectByPrefix(t *testing.T) {
	projects := []ProjectDef{
		{Name: "contracts", Prefix: "CON", Default: true},
		{Name: "bugs", Prefix: "BUG"},
		{Name: "specs", Prefix: "SPEC"},
	}

	p := FindProjectByPrefix(projects, "bug")
	if p == nil || p.Name != "bugs" {
		t.Fatalf("expected bugs project, got %v", p)
	}

	p = FindProjectByPrefix(projects, "BUG")
	if p == nil || p.Name != "bugs" {
		t.Fatalf("expected bugs project for uppercase, got %v", p)
	}

	p = FindProjectByPrefix(projects, "nonexistent")
	if p != nil {
		t.Fatalf("expected nil for nonexistent prefix, got %v", p)
	}
}
