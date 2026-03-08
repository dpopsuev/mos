package ingest

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dpopsuev/mos/moslib/store"
)

func tempStore(t *testing.T) store.Store {
	t.Helper()
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestIngestPlainFile(t *testing.T) {
	ctx := context.Background()
	s := tempStore(t)
	dir := t.TempDir()

	path := writeFile(t, dir, "readme.md", "# Hello World")

	res, err := Ingest(ctx, s, Request{
		Path:        path,
		Kind:        "doc",
		Description: "Project readme",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != "doc" {
		t.Fatalf("kind = %q", res.Kind)
	}
	if res.Size != 13 {
		t.Fatalf("size = %d", res.Size)
	}
	if res.EdgesAdded != 0 {
		t.Fatalf("edges = %d", res.EdgesAdded)
	}

	data, err := s.Get(ctx, filesBucket, string(res.NodeID))
	if err != nil || data == nil {
		t.Fatal("content not stored")
	}
	if string(data) != "# Hello World" {
		t.Fatalf("content = %q", data)
	}

	metaData, err := s.Get(ctx, metaBucket, string(res.NodeID))
	if err != nil || metaData == nil {
		t.Fatal("meta not stored")
	}
	var meta fileMeta
	json.Unmarshal(metaData, &meta)
	if meta.Kind != "doc" {
		t.Fatalf("meta kind = %q", meta.Kind)
	}
}

func TestIngestContractWithDependsOn(t *testing.T) {
	ctx := context.Background()
	s := tempStore(t)
	dir := t.TempDir()

	content := `contract "CON-2026-289" {
  title = "MCP server"
  depends_on = ["CON-2026-292", "CON-2026-288"]
}`
	path := writeFile(t, dir, "contract.mos", content)

	res, err := Ingest(ctx, s, Request{
		Path: path,
		Kind: "contract",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.EdgesAdded != 2 {
		t.Fatalf("edges = %d, want 2", res.EdgesAdded)
	}

	neighbors, err := s.Neighbors(ctx, res.NodeID, "depends_on", store.Outgoing)
	if err != nil {
		t.Fatal(err)
	}
	if len(neighbors) != 2 {
		t.Fatalf("depends_on neighbors = %d, want 2", len(neighbors))
	}

	targets := map[string]bool{}
	for _, e := range neighbors {
		targets[string(e.To)] = true
	}
	if !targets["CON-2026-292"] || !targets["CON-2026-288"] {
		t.Fatalf("unexpected targets: %v", targets)
	}
}

func TestIngestWithRelatesTo(t *testing.T) {
	ctx := context.Background()
	s := tempStore(t)
	dir := t.TempDir()

	path := writeFile(t, dir, "spec.md", "spec content")

	res, err := Ingest(ctx, s, Request{
		Path:      path,
		Kind:      "spec",
		RelatesTo: []string{"CON-2026-287", "CON-2026-290"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.EdgesAdded != 2 {
		t.Fatalf("edges = %d, want 2", res.EdgesAdded)
	}
}

func TestIngestUpsert(t *testing.T) {
	ctx := context.Background()
	s := tempStore(t)
	dir := t.TempDir()

	path := writeFile(t, dir, "doc.md", "v1")

	Ingest(ctx, s, Request{Path: path, Kind: "doc"})

	os.WriteFile(path, []byte("v2 updated"), 0o644)
	res, err := Ingest(ctx, s, Request{Path: path, Kind: "doc", Description: "updated"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Size != 10 {
		t.Fatalf("size = %d, want 10", res.Size)
	}

	data, _ := s.Get(ctx, filesBucket, string(res.NodeID))
	if string(data) != "v2 updated" {
		t.Fatalf("content = %q", data)
	}
}
