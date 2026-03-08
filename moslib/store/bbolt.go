package store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	bolt "go.etcd.io/bbolt"
)

var edgesBucket = []byte("_edges")

type BoltStore struct {
	db *bolt.DB
}

func Open(path string) (Store, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("open bbolt: %w", err)
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(edgesBucket)
		return err
	}); err != nil {
		db.Close()
		return nil, err
	}
	return &BoltStore{db: db}, nil
}

func DefaultPath(workspaceRoots ...string) string {
	home, _ := os.UserHomeDir()
	if len(workspaceRoots) == 0 {
		cwd, _ := os.Getwd()
		workspaceRoots = []string{cwd}
	}
	sort.Strings(workspaceRoots)
	h := sha256.Sum256([]byte(strings.Join(workspaceRoots, "\n")))
	hash := fmt.Sprintf("%x", h[:6])
	return filepath.Join(home, ".mosbus", hash, "store.db")
}

func (s *BoltStore) Get(_ context.Context, bucket, key string) ([]byte, error) {
	var val []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(key))
		if v != nil {
			val = make([]byte, len(v))
			copy(val, v)
		}
		return nil
	})
	return val, err
}

func (s *BoltStore) Put(_ context.Context, bucket, key string, value []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), value)
	})
}

func (s *BoltStore) Delete(_ context.Context, bucket, key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(key))
	})
}

func (s *BoltStore) List(_ context.Context, bucket, prefix string) ([]KV, error) {
	var result []KV
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		pfx := []byte(prefix)
		for k, v := c.Seek(pfx); k != nil && bytes.HasPrefix(k, pfx); k, v = c.Next() {
			cp := make([]byte, len(v))
			copy(cp, v)
			result = append(result, KV{Key: string(k), Value: cp})
		}
		return nil
	})
	return result, err
}

func edgeKey(from, to NodeID, rel EdgeRel) []byte {
	return []byte(fmt.Sprintf("%s|%s|%s", from, rel, to))
}

func reverseEdgeKey(from, to NodeID, rel EdgeRel) []byte {
	return []byte(fmt.Sprintf("~%s|%s|%s", to, rel, from))
}

func (s *BoltStore) AddEdge(_ context.Context, from, to NodeID, rel EdgeRel, meta []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(edgesBucket)
		if err := b.Put(edgeKey(from, to, rel), meta); err != nil {
			return err
		}
		return b.Put(reverseEdgeKey(from, to, rel), meta)
	})
}

func (s *BoltStore) RemoveEdge(_ context.Context, from, to NodeID, rel EdgeRel) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(edgesBucket)
		if err := b.Delete(edgeKey(from, to, rel)); err != nil {
			return err
		}
		return b.Delete(reverseEdgeKey(from, to, rel))
	})
}

func (s *BoltStore) Neighbors(_ context.Context, id NodeID, rel EdgeRel, dir Direction) ([]Edge, error) {
	var result []Edge
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(edgesBucket)
		c := b.Cursor()

		if dir == Outgoing || dir == Both {
			var pfx []byte
			if rel == "" {
				pfx = []byte(fmt.Sprintf("%s|", id))
			} else {
				pfx = []byte(fmt.Sprintf("%s|%s|", id, rel))
			}
			for k, v := c.Seek(pfx); k != nil && bytes.HasPrefix(k, pfx); k, v = c.Next() {
				e, ok := parseEdgeKey(k, v)
				if ok {
					result = append(result, e)
				}
			}
		}

		if dir == Incoming || dir == Both {
			var pfx []byte
			if rel == "" {
				pfx = []byte(fmt.Sprintf("~%s|", id))
			} else {
				pfx = []byte(fmt.Sprintf("~%s|%s|", id, rel))
			}
			for k, v := c.Seek(pfx); k != nil && bytes.HasPrefix(k, pfx); k, v = c.Next() {
				e, ok := parseReverseEdgeKey(k, v)
				if ok {
					result = append(result, e)
				}
			}
		}

		return nil
	})
	return result, err
}

func parseEdgeKey(key, meta []byte) (Edge, bool) {
	parts := strings.SplitN(string(key), "|", 3)
	if len(parts) != 3 {
		return Edge{}, false
	}
	m := make([]byte, len(meta))
	copy(m, meta)
	return Edge{From: NodeID(parts[0]), Rel: EdgeRel(parts[1]), To: NodeID(parts[2]), Meta: m}, true
}

func parseReverseEdgeKey(key, meta []byte) (Edge, bool) {
	s := string(key)
	if !strings.HasPrefix(s, "~") {
		return Edge{}, false
	}
	parts := strings.SplitN(s[1:], "|", 3)
	if len(parts) != 3 {
		return Edge{}, false
	}
	m := make([]byte, len(meta))
	copy(m, meta)
	return Edge{From: NodeID(parts[2]), Rel: EdgeRel(parts[1]), To: NodeID(parts[0]), Meta: m}, true
}

func (s *BoltStore) Walk(ctx context.Context, root NodeID, rel EdgeRel, dir Direction, maxDepth int, fn WalkFn) error {
	type item struct {
		edge  Edge
		depth int
	}
	visited := map[NodeID]bool{root: true}
	queue := []item{}

	neighbors, err := s.Neighbors(ctx, root, rel, dir)
	if err != nil {
		return err
	}
	for _, e := range neighbors {
		queue = append(queue, item{edge: e, depth: 1})
	}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		target := cur.edge.To
		if dir == Incoming {
			target = cur.edge.From
		}

		if visited[target] {
			continue
		}
		visited[target] = true

		if !fn(cur.depth, cur.edge) {
			return nil
		}

		if maxDepth > 0 && cur.depth >= maxDepth {
			continue
		}

		next, err := s.Neighbors(ctx, target, rel, dir)
		if err != nil {
			return err
		}
		for _, e := range next {
			queue = append(queue, item{edge: e, depth: cur.depth + 1})
		}
	}

	return nil
}

func (s *BoltStore) Close() error {
	return s.db.Close()
}
