package history

import (
	"fmt"
	"strings"

	"github.com/dpopsuev/mos/moslib/vcs"
)

func CreateBranch(store vcs.ObjectStore, name string, target vcs.Hash) error {
	if _, err := store.ResolveRef(vcs.HeadsPrefix + name); err == nil {
		return fmt.Errorf("branch %q already exists", name)
	}
	return store.UpdateRef(vcs.HeadsPrefix+name, target)
}

func ListBranches(store vcs.ObjectStore) ([]vcs.Ref, error) {
	refs, err := store.ListRefs(vcs.HeadsPrefix)
	if err != nil {
		return nil, err
	}
	var branches []vcs.Ref
	for _, r := range refs {
		branches = append(branches, vcs.Ref{
			Name: strings.TrimPrefix(r.Name, vcs.HeadsPrefix),
			Hash: r.Hash,
		})
	}
	return branches, nil
}

func DeleteBranch(store vcs.ObjectStore, root, name string) error {
	current := vcs.CurrentBranch(root)
	if current == name {
		return fmt.Errorf("cannot delete current branch %q", name)
	}
	if _, err := store.ResolveRef(vcs.HeadsPrefix + name); err != nil {
		return fmt.Errorf("branch %q not found", name)
	}
	return store.DeleteRef(vcs.HeadsPrefix + name)
}
