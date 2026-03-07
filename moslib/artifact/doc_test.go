package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCON035_DocTypeInDefaultConfig(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)

	doc, ok := reg.Types["doc"]
	if !ok {
		t.Fatal("doc type not in registry")
	}
	if doc.Directory != "docs" {
		t.Errorf("doc.Directory = %q, want docs", doc.Directory)
	}

	hasKind := false
	for _, f := range doc.Fields {
		if f.Name == "kind" && len(f.Enum) == 5 {
			hasKind = true
		}
	}
	if !hasKind {
		t.Error("doc missing kind enum field")
	}
}

func TestCON035_CreateDocArtifact(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["doc"]

	path, err := GenericCreate(root, td, "DOC-2026-001", map[string]string{
		"title":     "API Reference",
		"kind":      "api-reference",
		"status":    "draft",
		"documents": "CON-2026-015",
		"source":    "docs/api-reference.md",
	})
	if err != nil {
		t.Fatalf("GenericCreate doc: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("doc file not found: %v", err)
	}
	assertParses(t, path)
}

func TestCON035_ListDocsByKind(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["doc"]

	GenericCreate(root, td, "DOC-001", map[string]string{
		"title": "API Ref", "kind": "api-reference", "status": "draft",
	})
	GenericCreate(root, td, "DOC-002", map[string]string{
		"title": "Runbook", "kind": "runbook", "status": "published",
	})
	GenericCreate(root, td, "DOC-003", map[string]string{
		"title": "ADR-1", "kind": "adr", "status": "published",
	})

	all, err := GenericList(root, td, "")
	if err != nil {
		t.Fatalf("GenericList: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 docs, got %d", len(all))
	}

	published, err := GenericList(root, td, "published")
	if err != nil {
		t.Fatalf("GenericList published: %v", err)
	}
	if len(published) != 2 {
		t.Errorf("expected 2 published docs, got %d", len(published))
	}
}

func TestCON035_DocLifecycleTransitions(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["doc"]

	GenericCreate(root, td, "DOC-LC-001", map[string]string{
		"title": "Lifecycle test", "kind": "runbook", "status": "draft",
	})

	if err := GenericUpdateStatus(root, td, "DOC-LC-001", "published"); err != nil {
		t.Fatalf("transition to published: %v", err)
	}
	if err := GenericUpdateStatus(root, td, "DOC-LC-001", "stale"); err != nil {
		t.Fatalf("transition to stale: %v", err)
	}
	if err := GenericUpdateStatus(root, td, "DOC-LC-001", "retired"); err != nil {
		t.Fatalf("transition to retired: %v", err)
	}

	items, _ := GenericList(root, td, "retired")
	if len(items) != 1 {
		t.Error("expected retired doc in archive")
	}
}

func TestCON035_DocDirectoriesCreatedByInit(t *testing.T) {
	root := setupScaffold(t)
	for _, sub := range []string{"active", "archive"} {
		dir := filepath.Join(root, ".mos", "docs", sub)
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("docs/%s directory not created by init: %v", sub, err)
		}
	}
}

func TestCON035_DocDocumentsFieldParseable(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["doc"]

	GenericCreate(root, td, "DOC-REF-001", map[string]string{
		"title":     "Doc with reference",
		"kind":      "adr",
		"status":    "draft",
		"documents": "ARCH-2026-001",
	})

	path, _ := FindGenericPath(root, td, "DOC-REF-001")
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `documents = "ARCH-2026-001"`) {
		t.Error("doc should contain documents field referencing ARCH-2026-001")
	}
}

// --- CON-2026-036: Lifecycle Chain -- Cross-Type Traceability ---
