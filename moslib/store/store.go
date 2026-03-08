package store

import "context"

type NodeID string

type EdgeRel string

type Direction int

const (
	Outgoing Direction = iota
	Incoming
	Both
)

type KV struct {
	Key   string
	Value []byte
}

type Edge struct {
	From NodeID
	To   NodeID
	Rel  EdgeRel
	Meta []byte
}

type WalkFn func(depth int, edge Edge) (cont bool)

// Store is the persistence interface all Locus components program against.
// Designed for two backends: embedded local (bbolt) and remote shared (gRPC/Sophia).
type Store interface {
	Get(ctx context.Context, bucket, key string) ([]byte, error)
	Put(ctx context.Context, bucket, key string, value []byte) error
	Delete(ctx context.Context, bucket, key string) error
	List(ctx context.Context, bucket, prefix string) ([]KV, error)

	AddEdge(ctx context.Context, from, to NodeID, rel EdgeRel, meta []byte) error
	RemoveEdge(ctx context.Context, from, to NodeID, rel EdgeRel) error
	Neighbors(ctx context.Context, id NodeID, rel EdgeRel, dir Direction) ([]Edge, error)
	Walk(ctx context.Context, root NodeID, rel EdgeRel, dir Direction, maxDepth int, fn WalkFn) error

	Close() error
}
