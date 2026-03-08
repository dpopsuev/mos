package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) Store {
	t.Helper()
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestKVRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := tempStore(t)

	if err := s.Put(ctx, "b", "k1", []byte("v1")); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(ctx, "b", "k1")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "v1" {
		t.Fatalf("got %q, want %q", got, "v1")
	}

	if err := s.Delete(ctx, "b", "k1"); err != nil {
		t.Fatal(err)
	}
	got, err = s.Get(ctx, "b", "k1")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil after delete, got %q", got)
	}
}

func TestKVList(t *testing.T) {
	ctx := context.Background()
	s := tempStore(t)

	for _, k := range []string{"a:1", "a:2", "a:3", "b:1"} {
		if err := s.Put(ctx, "b", k, []byte(k)); err != nil {
			t.Fatal(err)
		}
	}

	items, err := s.List(ctx, "b", "a:")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}
}

func TestEdgeAddAndNeighbors(t *testing.T) {
	ctx := context.Background()
	s := tempStore(t)

	s.AddEdge(ctx, "A", "B", "depends_on", nil)
	s.AddEdge(ctx, "A", "C", "depends_on", nil)
	s.AddEdge(ctx, "B", "D", "depends_on", nil)

	out, err := s.Neighbors(ctx, "A", "depends_on", Outgoing)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("A outgoing: got %d, want 2", len(out))
	}

	in, err := s.Neighbors(ctx, "B", "depends_on", Incoming)
	if err != nil {
		t.Fatal(err)
	}
	if len(in) != 1 {
		t.Fatalf("B incoming: got %d, want 1", len(in))
	}
	if in[0].From != "A" {
		t.Fatalf("B incoming from: got %q, want A", in[0].From)
	}
}

func TestEdgeRemove(t *testing.T) {
	ctx := context.Background()
	s := tempStore(t)

	s.AddEdge(ctx, "A", "B", "rel", nil)
	s.RemoveEdge(ctx, "A", "B", "rel")

	out, _ := s.Neighbors(ctx, "A", "rel", Outgoing)
	if len(out) != 0 {
		t.Fatalf("expected 0 neighbors after remove, got %d", len(out))
	}
}

func TestWalkDAG(t *testing.T) {
	ctx := context.Background()
	s := tempStore(t)

	// CON-295 -> CON-294 -> CON-293 -> CON-289
	s.AddEdge(ctx, "295", "294", "depends_on", nil)
	s.AddEdge(ctx, "294", "293", "depends_on", nil)
	s.AddEdge(ctx, "293", "289", "depends_on", nil)

	var visited []string
	err := s.Walk(ctx, "295", "depends_on", Outgoing, 0, func(depth int, e Edge) bool {
		visited = append(visited, string(e.To))
		return true
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(visited) != 3 {
		t.Fatalf("walk visited %d, want 3: %v", len(visited), visited)
	}
}

func TestWalkMaxDepth(t *testing.T) {
	ctx := context.Background()
	s := tempStore(t)

	s.AddEdge(ctx, "A", "B", "r", nil)
	s.AddEdge(ctx, "B", "C", "r", nil)
	s.AddEdge(ctx, "C", "D", "r", nil)

	var visited []string
	s.Walk(ctx, "A", "r", Outgoing, 2, func(depth int, e Edge) bool {
		visited = append(visited, string(e.To))
		return true
	})
	if len(visited) != 2 {
		t.Fatalf("walk depth=2 visited %d, want 2: %v", len(visited), visited)
	}
}

func TestDefaultPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	p := DefaultPath("/workspace/a", "/workspace/b")
	if !filepath.HasPrefix(p, filepath.Join(home, ".mosbus")) {
		t.Fatalf("unexpected path: %s", p)
	}
	if filepath.Base(p) != "store.db" {
		t.Fatalf("expected store.db, got %s", filepath.Base(p))
	}
}
