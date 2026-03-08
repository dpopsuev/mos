package harness

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dpopsuev/mos/moslib/store"
)

func TestSameWorkspaceSharedState(t *testing.T) {
	h, err := Start(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer h.Cleanup()

	ctx := context.Background()
	a, err := h.ConnectClient([]string{"/ws/shared"})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	b, err := h.ConnectClient([]string{"/ws/shared"})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	if err := a.Put(ctx, "data", "key", []byte("from-a")); err != nil {
		t.Fatal(err)
	}

	val, err := b.Get(ctx, "data", "key")
	if err != nil {
		t.Fatal(err)
	}
	if string(val) != "from-a" {
		t.Errorf("shared workspace: expected from-a, got %s", string(val))
	}
}

func TestDifferentWorkspaceIsolation(t *testing.T) {
	h, err := Start(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer h.Cleanup()

	ctx := context.Background()
	a, err := h.ConnectClient([]string{"/ws/project-a"})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	b, err := h.ConnectClient([]string{"/ws/project-b"})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	a.Put(ctx, "secrets", "api-key", []byte("secret-a"))
	b.Put(ctx, "secrets", "api-key", []byte("secret-b"))

	valA, _ := a.Get(ctx, "secrets", "api-key")
	valB, _ := b.Get(ctx, "secrets", "api-key")

	if string(valA) != "secret-a" {
		t.Errorf("cross-contamination: workspace A got %s", string(valA))
	}
	if string(valB) != "secret-b" {
		t.Errorf("cross-contamination: workspace B got %s", string(valB))
	}

	itemsA, _ := a.List(ctx, "secrets", "")
	for _, item := range itemsA {
		if string(item.Value) == "secret-b" {
			t.Error("workspace A contains workspace B data")
		}
	}

	itemsB, _ := b.List(ctx, "secrets", "")
	for _, item := range itemsB {
		if string(item.Value) == "secret-a" {
			t.Error("workspace B contains workspace A data")
		}
	}
}

func TestConcurrentWritersSameWorkspace(t *testing.T) {
	h, err := Start(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer h.Cleanup()

	const N = 30
	var wg sync.WaitGroup

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c, err := h.ConnectClient([]string{"/ws/concurrent"})
			if err != nil {
				t.Errorf("connect %d: %v", idx, err)
				return
			}
			defer c.Close()

			key := string(rune('A' + idx))
			c.Put(context.Background(), "writes", key, []byte(key))
		}(i)
	}
	wg.Wait()

	reader, err := h.ConnectClient([]string{"/ws/concurrent"})
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	items, _ := reader.List(context.Background(), "writes", "")
	if len(items) != N {
		t.Errorf("expected %d items after concurrent writes, got %d", N, len(items))
	}
}

func TestDisconnectReconnectPersistence(t *testing.T) {
	h, err := Start(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer h.Cleanup()

	ctx := context.Background()
	c1, err := h.ConnectClient([]string{"/ws/persist"})
	if err != nil {
		t.Fatal(err)
	}
	c1.Put(ctx, "durable", "key", []byte("persisted"))
	c1.Close()

	c2, err := h.ConnectClient([]string{"/ws/persist"})
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Close()

	val, _ := c2.Get(ctx, "durable", "key")
	if string(val) != "persisted" {
		t.Errorf("expected persisted, got %s", string(val))
	}
}

func TestEdgeIsolationBetweenWorkspaces(t *testing.T) {
	h, err := Start(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer h.Cleanup()

	ctx := context.Background()
	a, _ := h.ConnectClient([]string{"/ws/edges-a"})
	defer a.Close()
	b, _ := h.ConnectClient([]string{"/ws/edges-b"})
	defer b.Close()

	a.AddEdge(ctx, "X", "Y", "dep", nil)
	b.AddEdge(ctx, "M", "N", "dep", nil)

	edgesA, _ := a.Neighbors(ctx, "X", "dep", store.Outgoing)
	if len(edgesA) != 1 || edgesA[0].To != "Y" {
		t.Errorf("workspace A edges wrong: %+v", edgesA)
	}

	edgesACross, _ := a.Neighbors(ctx, "M", "dep", store.Outgoing)
	if len(edgesACross) != 0 {
		t.Error("workspace A sees workspace B edges")
	}

	edgesB, _ := b.Neighbors(ctx, "M", "dep", store.Outgoing)
	if len(edgesB) != 1 || edgesB[0].To != "N" {
		t.Errorf("workspace B edges wrong: %+v", edgesB)
	}
}

func TestSessionManagerIdleReaping(t *testing.T) {
	mgr := store.NewSessionManager(100 * time.Millisecond)
	defer mgr.Close()

	sess, err := mgr.Connect([]string{"/ws/reap"})
	if err != nil {
		t.Fatal(err)
	}

	mgr.Disconnect(sess.ID)

	time.Sleep(300 * time.Millisecond)

	if mgr.ActiveSessions() != 0 {
		t.Errorf("expected 0 active sessions, got %d", mgr.ActiveSessions())
	}
}
