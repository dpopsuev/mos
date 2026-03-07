package vcs

import (
	"fmt"
	"io"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const mosRefPrefix = "refs/mos/"

// GitStore is an ObjectStore that writes governance objects as native Git
// objects in .git/objects/ and manages refs under refs/mos/*.
type GitStore struct {
	repo *git.Repository
}

var _ ObjectStore = (*GitStore)(nil)

// NewGitStore opens (or initializes) a Git repository at path and returns
// a GitStore backed by it.
func NewGitStore(path string) (*GitStore, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		repo, err = git.PlainInit(path, false)
		if err != nil {
			return nil, fmt.Errorf("gitstore: cannot open or init repo at %s: %w", path, err)
		}
	}
	return &GitStore{repo: repo}, nil
}

// NewGitStoreFromRepo wraps an existing go-git Repository.
func NewGitStoreFromRepo(repo *git.Repository) *GitStore {
	return &GitStore{repo: repo}
}

func (g *GitStore) StoreBlob(data []byte) (Hash, error) {
	obj := g.repo.Storer.NewEncodedObject()
	obj.SetType(plumbing.BlobObject)
	obj.SetSize(int64(len(data)))
	w, err := obj.Writer()
	if err != nil {
		return ZeroHash, err
	}
	if _, err := w.Write(data); err != nil {
		w.Close()
		return ZeroHash, err
	}
	w.Close()
	gh, err := g.repo.Storer.SetEncodedObject(obj)
	if err != nil {
		return ZeroHash, err
	}
	return gitHashToHash(gh), nil
}

func (g *GitStore) ReadBlob(h Hash) ([]byte, error) {
	obj, err := g.repo.BlobObject(hashToGitHash(h))
	if err != nil {
		return nil, fmt.Errorf("blob %s not found: %w", h.Short(), err)
	}
	r, err := obj.Reader()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func (g *GitStore) StoreTree(entries []TreeEntry) (Hash, error) {
	tree := &object.Tree{}
	for _, e := range entries {
		tree.Entries = append(tree.Entries, object.TreeEntry{
			Name: e.Name,
			Mode: mossToGitMode(e.Mode),
			Hash: hashToGitHash(e.Hash),
		})
	}

	obj := g.repo.Storer.NewEncodedObject()
	if err := tree.Encode(obj); err != nil {
		return ZeroHash, fmt.Errorf("encode tree: %w", err)
	}
	gh, err := g.repo.Storer.SetEncodedObject(obj)
	if err != nil {
		return ZeroHash, err
	}
	return gitHashToHash(gh), nil
}

func (g *GitStore) ReadTree(h Hash) ([]TreeEntry, error) {
	tree, err := object.GetTree(g.repo.Storer, hashToGitHash(h))
	if err != nil {
		return nil, fmt.Errorf("tree %s not found: %w", h.Short(), err)
	}
	var entries []TreeEntry
	for _, e := range tree.Entries {
		entries = append(entries, TreeEntry{
			Name: e.Name,
			Hash: gitHashToHash(e.Hash),
			Mode: gitToMossMode(e.Mode),
		})
	}
	return entries, nil
}

func (g *GitStore) StoreCommit(c CommitData) (Hash, error) {
	commit := &object.Commit{
		Author: object.Signature{
			Name:  c.Author,
			Email: c.Email,
			When:  c.Time,
		},
		Committer: object.Signature{
			Name:  c.Author,
			Email: c.Email,
			When:  c.Time,
		},
		Message:  c.Message,
		TreeHash: hashToGitHash(c.Tree),
	}
	for _, p := range c.Parents {
		commit.ParentHashes = append(commit.ParentHashes, hashToGitHash(p))
	}

	obj := g.repo.Storer.NewEncodedObject()
	if err := commit.Encode(obj); err != nil {
		return ZeroHash, fmt.Errorf("encode commit: %w", err)
	}
	gh, err := g.repo.Storer.SetEncodedObject(obj)
	if err != nil {
		return ZeroHash, err
	}
	return gitHashToHash(gh), nil
}

func (g *GitStore) ReadCommit(h Hash) (*CommitData, error) {
	commit, err := object.GetCommit(g.repo.Storer, hashToGitHash(h))
	if err != nil {
		return nil, fmt.Errorf("commit %s not found: %w", h.Short(), err)
	}
	cd := &CommitData{
		Tree:    gitHashToHash(commit.TreeHash),
		Author:  commit.Author.Name,
		Email:   commit.Author.Email,
		Time:    commit.Author.When,
		Message: commit.Message,
	}
	for _, p := range commit.ParentHashes {
		cd.Parents = append(cd.Parents, gitHashToHash(p))
	}
	return cd, nil
}

func (g *GitStore) HasObject(h Hash) bool {
	_, err := g.repo.Storer.EncodedObject(plumbing.AnyObject, hashToGitHash(h))
	return err == nil
}

func (g *GitStore) TypeOf(h Hash) (ObjectType, error) {
	obj, err := g.repo.Storer.EncodedObject(plumbing.AnyObject, hashToGitHash(h))
	if err != nil {
		return 0, fmt.Errorf("object %s not found: %w", h.Short(), err)
	}
	switch obj.Type() {
	case plumbing.BlobObject:
		return ObjectBlob, nil
	case plumbing.TreeObject:
		return ObjectTree, nil
	case plumbing.CommitObject:
		return ObjectCommit, nil
	default:
		return 0, fmt.Errorf("unsupported git object type: %s", obj.Type())
	}
}

func (g *GitStore) UpdateRef(name string, h Hash) error {
	refName := plumbing.ReferenceName(mosRefPrefix + name)
	ref := plumbing.NewHashReference(refName, hashToGitHash(h))
	return g.repo.Storer.SetReference(ref)
}

func (g *GitStore) ResolveRef(name string) (Hash, error) {
	refName := plumbing.ReferenceName(mosRefPrefix + name)
	ref, err := g.repo.Storer.Reference(refName)
	if err != nil {
		return ZeroHash, fmt.Errorf("ref %q not found: %w", name, err)
	}
	return gitHashToHash(ref.Hash()), nil
}

func (g *GitStore) ListRefs(prefix string) ([]Ref, error) {
	fullPrefix := mosRefPrefix + prefix
	var refs []Ref
	iter, err := g.repo.Storer.IterReferences()
	if err != nil {
		return nil, err
	}
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		name := string(ref.Name())
		if !strings.HasPrefix(name, fullPrefix) {
			return nil
		}
		shortName := strings.TrimPrefix(name, mosRefPrefix)
		refs = append(refs, Ref{
			Name: shortName,
			Hash: gitHashToHash(ref.Hash()),
		})
		return nil
	})
	return refs, err
}

func (g *GitStore) DeleteRef(name string) error {
	refName := plumbing.ReferenceName(mosRefPrefix + name)
	return g.repo.Storer.RemoveReference(refName)
}

func (g *GitStore) AllObjects() ([]Hash, error) {
	var hashes []Hash
	iter, err := g.repo.Storer.IterEncodedObjects(plumbing.AnyObject)
	if err != nil {
		return nil, err
	}
	err = iter.ForEach(func(obj plumbing.EncodedObject) error {
		hashes = append(hashes, gitHashToHash(obj.Hash()))
		return nil
	})
	return hashes, err
}

// gitHashToHash converts a go-git SHA-1 hash to our SHA-256 Hash type.
// Git uses SHA-1 (20 bytes); we zero-pad to 32 bytes.
// GitStore hashes will differ from FSStore hashes for identical content —
// this is expected since the content-addressing schemes differ.
func gitHashToHash(gh plumbing.Hash) Hash {
	var h Hash
	copy(h[:20], gh[:])
	return h
}

func hashToGitHash(h Hash) plumbing.Hash {
	var gh plumbing.Hash
	copy(gh[:], h[:20])
	return gh
}

func mossToGitMode(mode uint32) filemode.FileMode {
	switch mode {
	case ModeDir:
		return filemode.Dir
	default:
		return filemode.Regular
	}
}

func gitToMossMode(fm filemode.FileMode) uint32 {
	if fm == filemode.Dir {
		return ModeDir
	}
	return ModeRegular
}
