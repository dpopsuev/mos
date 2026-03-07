package vcs

import (
	"fmt"
)

// Migrate copies all objects and refs from the current store to a new backend,
// remapping hashes since different backends use different hash algorithms.
func Migrate(repo *Repository, targetBackend string) error {
	if repo.Config.Backend == targetBackend {
		return fmt.Errorf("already using %q backend", targetBackend)
	}

	dst, err := openStore(repo.Root, targetBackend)
	if err != nil {
		return fmt.Errorf("open target store: %w", err)
	}

	objects, err := repo.Store.AllObjects()
	if err != nil {
		return fmt.Errorf("list source objects: %w", err)
	}

	hashMap := map[Hash]Hash{} // old hash -> new hash

	// Phase 1: Copy all blobs (leaves of the DAG).
	for _, h := range objects {
		typ, err := repo.Store.TypeOf(h)
		if err != nil {
			return fmt.Errorf("type of %s: %w", h.Short(), err)
		}
		if typ != ObjectBlob {
			continue
		}
		data, err := repo.Store.ReadBlob(h)
		if err != nil {
			return err
		}
		newH, err := dst.StoreBlob(data)
		if err != nil {
			return err
		}
		hashMap[h] = newH
	}

	// Phase 2: Copy trees (may reference blobs and other trees).
	// Iterate until all trees are resolved (handles nested trees).
	pending := true
	for pending {
		pending = false
		for _, h := range objects {
			if _, mapped := hashMap[h]; mapped {
				continue
			}
			typ, err := repo.Store.TypeOf(h)
			if err != nil {
				return err
			}
			if typ != ObjectTree {
				continue
			}
			entries, err := repo.Store.ReadTree(h)
			if err != nil {
				return err
			}

			allResolved := true
			newEntries := make([]TreeEntry, len(entries))
			for i, e := range entries {
				newH, ok := hashMap[e.Hash]
				if !ok {
					allResolved = false
					break
				}
				newEntries[i] = TreeEntry{Name: e.Name, Hash: newH, Mode: e.Mode}
			}
			if !allResolved {
				pending = true
				continue
			}
			newH, err := dst.StoreTree(newEntries)
			if err != nil {
				return err
			}
			hashMap[h] = newH
		}
	}

	// Phase 3: Copy commits in topological order (parents before children).
	pending = true
	for pending {
		pending = false
		for _, h := range objects {
			if _, mapped := hashMap[h]; mapped {
				continue
			}
			typ, err := repo.Store.TypeOf(h)
			if err != nil {
				return err
			}
			if typ != ObjectCommit {
				continue
			}
			cd, err := repo.Store.ReadCommit(h)
			if err != nil {
				return err
			}

			newTree, treeOK := hashMap[cd.Tree]
			if !treeOK {
				pending = true
				continue
			}

			allParentsOK := true
			newParents := make([]Hash, len(cd.Parents))
			for i, p := range cd.Parents {
				newP, ok := hashMap[p]
				if !ok {
					allParentsOK = false
					break
				}
				newParents[i] = newP
			}
			if !allParentsOK {
				pending = true
				continue
			}

			newCD := CommitData{
				Tree:    newTree,
				Parents: newParents,
				Author:  cd.Author,
				Email:   cd.Email,
				Time:    cd.Time,
				Message: cd.Message,
			}
			newH, err := dst.StoreCommit(newCD)
			if err != nil {
				return err
			}
			hashMap[h] = newH
		}
	}

	// Phase 4: Remap refs.
	refs, err := repo.Store.ListRefs("")
	if err != nil {
		return fmt.Errorf("list refs: %w", err)
	}
	for _, ref := range refs {
		newH, ok := hashMap[ref.Hash]
		if !ok {
			return fmt.Errorf("ref %s points to unmapped hash %s", ref.Name, ref.Hash.Short())
		}
		if err := dst.UpdateRef(ref.Name, newH); err != nil {
			return fmt.Errorf("copy ref %s: %w", ref.Name, err)
		}
	}

	repo.Config.Backend = targetBackend
	if err := WriteVCSConfig(repo.Root, repo.Config); err != nil {
		return fmt.Errorf("update config: %w", err)
	}
	repo.Store = dst

	return nil
}
