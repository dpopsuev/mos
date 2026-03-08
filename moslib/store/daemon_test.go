package store

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func startTestDaemon(t *testing.T) (socketPath string, cleanup func()) {
	t.Helper()
	dir := t.TempDir()
	socketPath = filepath.Join(dir, "test.sock")

	d := NewDaemon(socketPath, 30*time.Minute)
	ready := make(chan struct{})
	go func() {
		close(ready)
		d.ListenAndServe()
	}()
	<-ready
	time.Sleep(50 * time.Millisecond)

	return socketPath, func() { d.Shutdown() }
}

func TestDaemonRoundTrip(t *testing.T) {
	sock, cleanup := startTestDaemon(t)
	defer cleanup()

	client, err := Dial(sock, []string{"/tmp/test-workspace"})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	if err := client.Put(ctx, "test-bucket", "key1", []byte("value1")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	val, err := client.Get(ctx, "test-bucket", "key1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(val) != "value1" {
		t.Errorf("expected value1, got %s", string(val))
	}

	if err := client.Put(ctx, "test-bucket", "key2", []byte("value2")); err != nil {
		t.Fatalf("Put key2: %v", err)
	}
	items, err := client.List(ctx, "test-bucket", "key")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	if err := client.Delete(ctx, "test-bucket", "key1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	val, err = client.Get(ctx, "test-bucket", "key1")
	if err != nil {
		t.Fatalf("Get after delete: %v", err)
	}
	if val != nil {
		t.Errorf("expected nil after delete, got %s", string(val))
	}
}

func TestDaemonEdgeOperations(t *testing.T) {
	sock, cleanup := startTestDaemon(t)
	defer cleanup()

	client, err := Dial(sock, []string{"/tmp/test-workspace-edges"})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	if err := client.AddEdge(ctx, "A", "B", "depends_on", nil); err != nil {
		t.Fatalf("AddEdge A->B: %v", err)
	}
	if err := client.AddEdge(ctx, "B", "C", "depends_on", nil); err != nil {
		t.Fatalf("AddEdge B->C: %v", err)
	}

	edges, err := client.Neighbors(ctx, "A", "depends_on", Outgoing)
	if err != nil {
		t.Fatalf("Neighbors: %v", err)
	}
	if len(edges) != 1 || edges[0].To != "B" {
		t.Errorf("expected A->B, got %+v", edges)
	}

	err = client.Walk(ctx, "A", "depends_on", Outgoing, 0, func(depth int, edge Edge) bool {
		return true
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}

	if err := client.RemoveEdge(ctx, "A", "B", "depends_on"); err != nil {
		t.Fatalf("RemoveEdge: %v", err)
	}
	edges, err = client.Neighbors(ctx, "A", "depends_on", Outgoing)
	if err != nil {
		t.Fatalf("Neighbors after remove: %v", err)
	}
	if len(edges) != 0 {
		t.Errorf("expected 0 edges after remove, got %d", len(edges))
	}
}

func TestDaemonSessionIsolation(t *testing.T) {
	sock, cleanup := startTestDaemon(t)
	defer cleanup()

	clientA, err := Dial(sock, []string{"/workspace/project-a"})
	if err != nil {
		t.Fatalf("Dial A: %v", err)
	}
	defer clientA.Close()

	clientB, err := Dial(sock, []string{"/workspace/project-b"})
	if err != nil {
		t.Fatalf("Dial B: %v", err)
	}
	defer clientB.Close()

	ctx := context.Background()

	if err := clientA.Put(ctx, "data", "secret", []byte("project-a-data")); err != nil {
		t.Fatalf("Put A: %v", err)
	}

	if err := clientB.Put(ctx, "data", "secret", []byte("project-b-data")); err != nil {
		t.Fatalf("Put B: %v", err)
	}

	valA, err := clientA.Get(ctx, "data", "secret")
	if err != nil {
		t.Fatalf("Get A: %v", err)
	}
	if string(valA) != "project-a-data" {
		t.Errorf("workspace A contaminated: expected project-a-data, got %s", string(valA))
	}

	valB, err := clientB.Get(ctx, "data", "secret")
	if err != nil {
		t.Fatalf("Get B: %v", err)
	}
	if string(valB) != "project-b-data" {
		t.Errorf("workspace B contaminated: expected project-b-data, got %s", string(valB))
	}
}

func TestDaemonSharedWorkspace(t *testing.T) {
	sock, cleanup := startTestDaemon(t)
	defer cleanup()

	client1, err := Dial(sock, []string{"/workspace/shared"})
	if err != nil {
		t.Fatalf("Dial 1: %v", err)
	}
	defer client1.Close()

	client2, err := Dial(sock, []string{"/workspace/shared"})
	if err != nil {
		t.Fatalf("Dial 2: %v", err)
	}
	defer client2.Close()

	ctx := context.Background()

	if err := client1.Put(ctx, "shared", "key", []byte("from-client-1")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	val, err := client2.Get(ctx, "shared", "key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(val) != "from-client-1" {
		t.Errorf("expected from-client-1, got %s", string(val))
	}
}

func TestDaemonConcurrentWriters(t *testing.T) {
	sock, cleanup := startTestDaemon(t)
	defer cleanup()

	const N = 20
	var wg sync.WaitGroup

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c, err := Dial(sock, []string{"/workspace/concurrent"})
			if err != nil {
				t.Errorf("Dial %d: %v", idx, err)
				return
			}
			defer c.Close()

			key := "key-" + string(rune('a'+idx))
			val := []byte("val-" + string(rune('a'+idx)))
			if err := c.Put(context.Background(), "concurrent", key, val); err != nil {
				t.Errorf("Put %d: %v", idx, err)
			}
		}(i)
	}
	wg.Wait()

	reader, err := Dial(sock, []string{"/workspace/concurrent"})
	if err != nil {
		t.Fatalf("Dial reader: %v", err)
	}
	defer reader.Close()

	items, err := reader.List(context.Background(), "concurrent", "key-")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != N {
		t.Errorf("expected %d items, got %d", N, len(items))
	}
}
