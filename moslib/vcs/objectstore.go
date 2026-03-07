package vcs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Hash [32]byte

var ZeroHash Hash

func NewHash(data []byte) Hash {
	return sha256.Sum256(data)
}

func ParseHash(s string) (Hash, error) {
	if len(s) != 64 {
		return ZeroHash, fmt.Errorf("invalid hash length %d, want 64", len(s))
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return ZeroHash, fmt.Errorf("invalid hex: %w", err)
	}
	var h Hash
	copy(h[:], b)
	return h, nil
}

func (h Hash) String() string { return hex.EncodeToString(h[:]) }
func (h Hash) IsZero() bool   { return h == ZeroHash }
func (h Hash) Short() string  { return h.String()[:12] }

type ObjectType int

const (
	ObjectBlob ObjectType = iota
	ObjectTree
	ObjectCommit
)

func (t ObjectType) String() string {
	switch t {
	case ObjectBlob:
		return "blob"
	case ObjectTree:
		return "tree"
	case ObjectCommit:
		return "commit"
	default:
		return "unknown"
	}
}

type TreeEntry struct {
	Name string
	Hash Hash
	Mode uint32 // 0100644 for regular file, 040000 for directory
}

type CommitData struct {
	Tree    Hash
	Parents []Hash
	Author  string
	Email   string
	Time    time.Time
	Message string
}

type Ref struct {
	Name string
	Hash Hash
}

// ObjectStore is the core content-addressable storage abstraction.
// All governance operations are backend-agnostic through this interface.
type ObjectStore interface {
	StoreBlob(data []byte) (Hash, error)
	StoreTree(entries []TreeEntry) (Hash, error)
	StoreCommit(c CommitData) (Hash, error)

	ReadBlob(h Hash) ([]byte, error)
	ReadTree(h Hash) ([]TreeEntry, error)
	ReadCommit(h Hash) (*CommitData, error)

	HasObject(h Hash) bool
	TypeOf(h Hash) (ObjectType, error)

	UpdateRef(name string, h Hash) error
	ResolveRef(name string) (Hash, error)
	ListRefs(prefix string) ([]Ref, error)
	DeleteRef(name string) error

	AllObjects() ([]Hash, error)
}

// ErrFSStoreNoRemote is returned when remote operations are attempted on the fs backend.
var ErrFSStoreNoRemote = fmt.Errorf("remote operations require git backend; run 'mos vcs migrate --to git'")

// ModeRegular is the mode for regular files in tree entries.
const ModeRegular uint32 = 0100644

// ModeDir is the mode for directories in tree entries.
const ModeDir uint32 = 040000
