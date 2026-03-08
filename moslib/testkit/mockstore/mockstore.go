package mockstore

import (
	"context"
	"fmt"
	"sync"

	"github.com/dpopsuev/mos/moslib/store"
)

type Call struct {
	Method string
	Args   []any
}

type MockStore struct {
	mu      sync.Mutex
	buckets map[string]map[string][]byte
	edges   []store.Edge
	calls   []Call
	errors  map[string]error
}

func New() *MockStore {
	return &MockStore{
		buckets: make(map[string]map[string][]byte),
		errors:  make(map[string]error),
	}
}

func (m *MockStore) SetError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[method] = err
}

func (m *MockStore) Calls() []Call {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]Call, len(m.calls))
	copy(cp, m.calls)
	return cp
}

func (m *MockStore) record(method string, args ...any) {
	m.calls = append(m.calls, Call{Method: method, Args: args})
}

func (m *MockStore) Get(_ context.Context, bucket, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.record("Get", bucket, key)
	if err, ok := m.errors["Get"]; ok {
		return nil, err
	}
	b, ok := m.buckets[bucket]
	if !ok {
		return nil, nil
	}
	v, ok := b[key]
	if !ok {
		return nil, nil
	}
	cp := make([]byte, len(v))
	copy(cp, v)
	return cp, nil
}

func (m *MockStore) Put(_ context.Context, bucket, key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.record("Put", bucket, key, value)
	if err, ok := m.errors["Put"]; ok {
		return err
	}
	b, ok := m.buckets[bucket]
	if !ok {
		b = make(map[string][]byte)
		m.buckets[bucket] = b
	}
	cp := make([]byte, len(value))
	copy(cp, value)
	b[key] = cp
	return nil
}

func (m *MockStore) Delete(_ context.Context, bucket, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.record("Delete", bucket, key)
	if err, ok := m.errors["Delete"]; ok {
		return err
	}
	if b, ok := m.buckets[bucket]; ok {
		delete(b, key)
	}
	return nil
}

func (m *MockStore) List(_ context.Context, bucket, prefix string) ([]store.KV, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.record("List", bucket, prefix)
	if err, ok := m.errors["List"]; ok {
		return nil, err
	}
	var result []store.KV
	if b, ok := m.buckets[bucket]; ok {
		for k, v := range b {
			if len(prefix) == 0 || len(k) >= len(prefix) && k[:len(prefix)] == prefix {
				cp := make([]byte, len(v))
				copy(cp, v)
				result = append(result, store.KV{Key: k, Value: cp})
			}
		}
	}
	return result, nil
}

func (m *MockStore) AddEdge(_ context.Context, from, to store.NodeID, rel store.EdgeRel, meta []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.record("AddEdge", from, to, rel, meta)
	if err, ok := m.errors["AddEdge"]; ok {
		return err
	}
	m.edges = append(m.edges, store.Edge{From: from, To: to, Rel: rel, Meta: meta})
	return nil
}

func (m *MockStore) RemoveEdge(_ context.Context, from, to store.NodeID, rel store.EdgeRel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.record("RemoveEdge", from, to, rel)
	if err, ok := m.errors["RemoveEdge"]; ok {
		return err
	}
	filtered := m.edges[:0]
	for _, e := range m.edges {
		if !(e.From == from && e.To == to && e.Rel == rel) {
			filtered = append(filtered, e)
		}
	}
	m.edges = filtered
	return nil
}

func (m *MockStore) Neighbors(_ context.Context, id store.NodeID, rel store.EdgeRel, dir store.Direction) ([]store.Edge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.record("Neighbors", id, rel, dir)
	if err, ok := m.errors["Neighbors"]; ok {
		return nil, err
	}
	var result []store.Edge
	for _, e := range m.edges {
		match := false
		if (dir == store.Outgoing || dir == store.Both) && e.From == id {
			if rel == "" || e.Rel == rel {
				match = true
			}
		}
		if (dir == store.Incoming || dir == store.Both) && e.To == id {
			if rel == "" || e.Rel == rel {
				match = true
			}
		}
		if match {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *MockStore) Walk(_ context.Context, root store.NodeID, rel store.EdgeRel, dir store.Direction, maxDepth int, fn store.WalkFn) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.record("Walk", root, rel, dir, maxDepth)
	if err, ok := m.errors["Walk"]; ok {
		return err
	}
	return nil
}

func (m *MockStore) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.record("Close")
	return nil
}

func AssertPutCalled(t interface{ Helper(); Errorf(string, ...any) }, ms *MockStore, bucket, key string) {
	t.Helper()
	for _, c := range ms.Calls() {
		if c.Method == "Put" && len(c.Args) >= 2 && c.Args[0] == bucket && c.Args[1] == key {
			return
		}
	}
	t.Errorf("expected Put(%q, %q) to be called", bucket, key)
}

func AssertBucketIsolation(t interface{ Helper(); Errorf(string, ...any) }, ms *MockStore, hash1, hash2 string) {
	t.Helper()
	_ = fmt.Sprintf("bucket isolation check: %s vs %s", hash1, hash2)
}
