package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/dsl"
)

func setupTemplateTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mosDir := filepath.Join(dir, MosDir)
	tmplDir := filepath.Join(mosDir, templatesDir)
	if err := os.MkdirAll(tmplDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeTemplate(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, MosDir, templatesDir, name+".mos")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadTemplate_Found(t *testing.T) {
	root := setupTemplateTestDir(t)
	writeTemplate(t, root, "bug-fix", `contract "TEMPLATE" {
  kind = "bug"
  priority = "high"

  section "Steps to Reproduce" {
    text = "Describe steps here"
  }
}
`)

	ab, err := LoadTemplate(root, "bug-fix")
	if err != nil {
		t.Fatalf("LoadTemplate: %v", err)
	}
	if ab.Kind != "contract" {
		t.Errorf("expected kind=contract, got %s", ab.Kind)
	}
	kind, ok := dsl.FieldString(ab.Items, "kind")
	if !ok || kind != "bug" {
		t.Errorf("expected field kind=bug, got %q ok=%v", kind, ok)
	}
}

func TestLoadTemplate_NotFound(t *testing.T) {
	root := setupTemplateTestDir(t)
	_, err := LoadTemplate(root, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing template")
	}
}

func TestMergeTemplate_FieldsAndBlocks(t *testing.T) {
	target := &dsl.ArtifactBlock{
		Kind: "contract",
		Name: "CON-1",
		Items: []dsl.Node{
			&dsl.Field{Key: "title", Value: &dsl.StringVal{Text: "My Bug"}},
			&dsl.Field{Key: "status", Value: &dsl.StringVal{Text: "draft"}},
		},
	}
	tmpl := &dsl.ArtifactBlock{
		Kind: "contract",
		Name: "TEMPLATE",
		Items: []dsl.Node{
			&dsl.Field{Key: "priority", Value: &dsl.StringVal{Text: "high"}},
			&dsl.Field{Key: "title", Value: &dsl.StringVal{Text: "SHOULD NOT OVERWRITE"}},
			&dsl.Block{
				Name:  "section",
				Title: "Steps to Reproduce",
				Items: []dsl.Node{
					&dsl.Field{Key: "text", Value: &dsl.StringVal{Text: "Describe steps here"}},
				},
			},
		},
	}

	MergeTemplate(target, tmpl)

	title, _ := dsl.FieldString(target.Items, "title")
	if title != "My Bug" {
		t.Errorf("title should not be overwritten, got %q", title)
	}

	prio, ok := dsl.FieldString(target.Items, "priority")
	if !ok || prio != "high" {
		t.Errorf("expected priority=high from template, got %q ok=%v", prio, ok)
	}

	var foundSection bool
	for _, item := range target.Items {
		if b, ok := item.(*dsl.Block); ok && b.Name == "section" && b.Title == "Steps to Reproduce" {
			foundSection = true
		}
	}
	if !foundSection {
		t.Error("expected section block from template")
	}
}

func TestGenericCreateWithTemplate(t *testing.T) {
	root := setupTemplateTestDir(t)
	writeTemplate(t, root, "my-template", `contract "TEMPLATE" {
  priority = "medium"

  section "Notes" {
    text = "Template notes"
  }
}
`)

	// Set up registry for GenericCreate
	regContent := `config "mos" {
  artifact_type "testitem" {
    directory = "testitems"
    prefix = "TI"
    lifecycle {
      active_states = ["draft", "active"]
      archive_states = ["complete"]
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(root, MosDir, "config.mos"), []byte(regContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, MosDir, "testitems", "active"), 0o755); err != nil {
		t.Fatal(err)
	}

	td := ArtifactTypeDef{
		Kind:      "testitem",
		Directory: "testitems",
		Lifecycle: LifecycleDef{
			ActiveStates:  []string{"draft", "active"},
			ArchiveStates: []string{"complete"},
		},
	}

	path, err := GenericCreateWithTemplate(root, td, "TI-001", map[string]string{
		"title": "Test Item",
	}, "my-template")
	if err != nil {
		t.Fatalf("GenericCreateWithTemplate: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "Test Item") {
		t.Error("expected title in output")
	}
	if !strings.Contains(content, "priority") {
		t.Error("expected priority from template")
	}
	if !strings.Contains(content, "Notes") {
		t.Error("expected section from template")
	}
}
