package history

import (
	"fmt"
	"strings"

	"github.com/dpopsuev/mos/moslib/vcs"
)

func CreateTag(store vcs.ObjectStore, name string, target vcs.Hash) error {
	if _, err := store.ResolveRef(vcs.TagsPrefix + name); err == nil {
		return fmt.Errorf("tag %q already exists", name)
	}
	return store.UpdateRef(vcs.TagsPrefix+name, target)
}

func ListTags(store vcs.ObjectStore) ([]vcs.Ref, error) {
	refs, err := store.ListRefs(vcs.TagsPrefix)
	if err != nil {
		return nil, err
	}
	var tags []vcs.Ref
	for _, r := range refs {
		tags = append(tags, vcs.Ref{
			Name: strings.TrimPrefix(r.Name, vcs.TagsPrefix),
			Hash: r.Hash,
		})
	}
	return tags, nil
}

func DeleteTag(store vcs.ObjectStore, name string) error {
	if _, err := store.ResolveRef(vcs.TagsPrefix + name); err != nil {
		return fmt.Errorf("tag %q not found", name)
	}
	return store.DeleteRef(vcs.TagsPrefix + name)
}
