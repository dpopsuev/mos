package dsl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetField_CreatesNew(t *testing.T) {
	items := []Node{
		&Field{Key: "title", Value: &StringVal{Text: "hello"}},
	}
	SetField(&items, "status", &StringVal{Text: "active"})
	if len(items) != 2 {
		t.Fatalf("len = %d, want 2", len(items))
	}
	v, ok := FieldString(items, "status")
	if !ok || v != "active" {
		t.Errorf("status = %q, ok=%v", v, ok)
	}
}

func TestSetField_UpdatesExisting(t *testing.T) {
	items := []Node{
		&Field{Key: "status", Value: &StringVal{Text: "draft"}},
		&Field{Key: "title", Value: &StringVal{Text: "hello"}},
	}
	SetField(&items, "status", &StringVal{Text: "active"})
	if len(items) != 2 {
		t.Fatalf("len = %d, want 2 (should not duplicate)", len(items))
	}
	v, _ := FieldString(items, "status")
	if v != "active" {
		t.Errorf("status = %q, want active", v)
	}
}

func TestSetField_OnEmptySlice(t *testing.T) {
	var items []Node
	SetField(&items, "key", &StringVal{Text: "val"})
	if len(items) != 1 {
		t.Fatalf("len = %d, want 1", len(items))
	}
}

func TestRemoveBlock_Found(t *testing.T) {
	items := []Node{
		&Field{Key: "title", Value: &StringVal{Text: "x"}},
		&Block{Name: "scope", Items: []Node{
			&Field{Key: "depends_on", Value: &StringVal{Text: "CON-1"}},
		}},
		&Block{Name: "acceptance"},
	}
	ok := RemoveBlock(&items, "scope")
	if !ok {
		t.Error("RemoveBlock returned false, want true")
	}
	if len(items) != 2 {
		t.Fatalf("len = %d, want 2", len(items))
	}
	if FindBlock(items, "scope") != nil {
		t.Error("scope block still present after removal")
	}
}

func TestRemoveBlock_NotFound(t *testing.T) {
	items := []Node{
		&Field{Key: "title", Value: &StringVal{Text: "x"}},
	}
	ok := RemoveBlock(&items, "nonexistent")
	if ok {
		t.Error("RemoveBlock returned true for missing block")
	}
}

func TestRemoveNamedBlock_Found(t *testing.T) {
	items := []Node{
		&Block{Name: "criterion", Title: "test-a"},
		&Block{Name: "criterion", Title: "test-b"},
	}
	ok := RemoveNamedBlock(&items, "criterion", "test-a")
	if !ok {
		t.Error("RemoveNamedBlock returned false")
	}
	if len(items) != 1 {
		t.Fatalf("len = %d, want 1", len(items))
	}
	if items[0].(*Block).Title != "test-b" {
		t.Error("wrong block removed")
	}
}

func TestRemoveNamedBlock_NotFound(t *testing.T) {
	items := []Node{
		&Block{Name: "criterion", Title: "test-a"},
	}
	ok := RemoveNamedBlock(&items, "criterion", "nonexistent")
	if ok {
		t.Error("RemoveNamedBlock returned true for missing block")
	}
}

func TestWalkBlocks_RecursiveVisit(t *testing.T) {
	items := []Node{
		&Field{Key: "title", Value: &StringVal{Text: "x"}},
		&Block{Name: "outer", Items: []Node{
			&Block{Name: "inner", Items: []Node{
				&Field{Key: "deep", Value: &StringVal{Text: "y"}},
			}},
		}},
		&Block{Name: "sibling"},
	}
	var visited []string
	WalkBlocks(items, func(b *Block) bool {
		visited = append(visited, b.Name)
		return true
	})
	if len(visited) != 3 {
		t.Fatalf("visited %d blocks, want 3: %v", len(visited), visited)
	}
	want := []string{"outer", "inner", "sibling"}
	for i, name := range want {
		if visited[i] != name {
			t.Errorf("visited[%d] = %q, want %q", i, visited[i], name)
		}
	}
}

func TestWalkBlocks_SkipDescent(t *testing.T) {
	items := []Node{
		&Block{Name: "outer", Items: []Node{
			&Block{Name: "inner"},
		}},
	}
	var visited []string
	WalkBlocks(items, func(b *Block) bool {
		visited = append(visited, b.Name)
		return false
	})
	if len(visited) != 1 {
		t.Fatalf("visited %d blocks, want 1 (should skip inner)", len(visited))
	}
}

func TestAppendToBlock_ExistingBlock(t *testing.T) {
	items := []Node{
		&Block{Name: "acceptance", Items: []Node{
			&Block{Name: "criterion", Title: "test-a"},
		}},
	}
	AppendToBlock(&items, "acceptance",
		&Block{Name: "criterion", Title: "test-b"},
	)
	b := FindBlock(items, "acceptance")
	if len(b.Items) != 2 {
		t.Fatalf("acceptance has %d items, want 2", len(b.Items))
	}
}

func TestAppendToBlock_CreatesBlock(t *testing.T) {
	items := []Node{
		&Field{Key: "title", Value: &StringVal{Text: "x"}},
	}
	AppendToBlock(&items, "scope",
		&Field{Key: "depends_on", Value: &StringVal{Text: "CON-1"}},
	)
	b := FindBlock(items, "scope")
	if b == nil {
		t.Fatal("scope block not created")
	}
	v, ok := FieldString(b.Items, "depends_on")
	if !ok || v != "CON-1" {
		t.Errorf("depends_on = %q, ok=%v", v, ok)
	}
}

func TestFindField_Found(t *testing.T) {
	items := []Node{
		&Field{Key: "title", Value: &StringVal{Text: "hello"}},
		&Field{Key: "status", Value: &StringVal{Text: "active"}},
	}
	f := FindField(items, "status")
	if f == nil {
		t.Fatal("FindField returned nil")
	}
	if f.Key != "status" {
		t.Errorf("Key = %q", f.Key)
	}
}

func TestFindField_NotFound(t *testing.T) {
	items := []Node{
		&Field{Key: "title", Value: &StringVal{Text: "hello"}},
	}
	f := FindField(items, "missing")
	if f != nil {
		t.Error("FindField should return nil for missing key")
	}
}

func TestWithArtifact_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.mos")
	src := `contract "CON-001" {
  title = "Test"
  status = "draft"
}
`
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	err := WithArtifact(path, func(ab *ArtifactBlock) error {
		SetField(&ab.Items, "status", &StringVal{Text: "active"})
		return nil
	})
	if err != nil {
		t.Fatalf("WithArtifact: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, `"active"`) {
		t.Errorf("file does not contain updated status:\n%s", content)
	}
	if !strings.Contains(content, `contract "CON-001"`) {
		t.Errorf("file lost artifact header:\n%s", content)
	}
}

func TestWithArtifact_ErrorDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.mos")
	src := `contract "CON-001" {
  title = "Test"
  status = "draft"
}
`
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	err := WithArtifact(path, func(ab *ArtifactBlock) error {
		SetField(&ab.Items, "status", &StringVal{Text: "active"})
		return os.ErrPermission
	})
	if err == nil {
		t.Fatal("expected error")
	}

	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "active") {
		t.Error("file was written despite error return")
	}
}

func TestWithArtifact_MissingFile(t *testing.T) {
	err := WithArtifact("/nonexistent/path.mos", func(ab *ArtifactBlock) error {
		return nil
	})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestWithArtifact_InvalidSyntax(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.mos")
	if err := os.WriteFile(path, []byte("not valid {{{"), 0644); err != nil {
		t.Fatal(err)
	}
	err := WithArtifact(path, func(ab *ArtifactBlock) error {
		return nil
	})
	if err == nil {
		t.Error("expected parse error")
	}
}
