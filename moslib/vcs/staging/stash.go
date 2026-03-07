package staging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dpopsuev/mos/moslib/vcs"
)

type StashEntry struct {
	Message    string   `json:"message"`
	Time       time.Time `json:"time"`
	IndexTree  vcs.Hash  `json:"index_tree"`
	WorkTree   vcs.Hash  `json:"work_tree"`
	HeadCommit vcs.Hash  `json:"head_commit"`
	Branch     string    `json:"branch"`
}

const stashFile = ".mos/vcs/stash.json"

func stashPath(root string) string {
	return filepath.Join(root, stashFile)
}

func loadStash(root string) ([]StashEntry, error) {
	data, err := os.ReadFile(stashPath(root))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var entries []StashEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("corrupt stash file: %w", err)
	}
	return entries, nil
}

func saveStash(root string, entries []StashEntry) error {
	p := stashPath(root)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

func Stash(repo *vcs.Repository, message string) error {
	idx, err := LoadIndex(repo.Root)
	if err != nil {
		return fmt.Errorf("load index: %w", err)
	}
	if len(idx.Entries) == 0 {
		return fmt.Errorf("nothing to stash (empty index)")
	}

	indexTree, err := idx.BuildTree(repo.Store)
	if err != nil {
		return fmt.Errorf("build index tree: %w", err)
	}

	workEntries, err := SnapshotWorkingTree(repo.Root, repo.Store)
	if err != nil {
		return fmt.Errorf("snapshot working tree: %w", err)
	}
	workIdx := &Index{Entries: workEntries}
	workTree, err := workIdx.BuildTree(repo.Store)
	if err != nil {
		return fmt.Errorf("build work tree: %w", err)
	}

	headCommit, _ := vcs.ResolveHead(repo)
	branch := vcs.CurrentBranch(repo.Root)

	if message == "" {
		message = fmt.Sprintf("WIP on %s", branch)
	}

	entry := StashEntry{
		Message:    message,
		Time:       time.Now().UTC(),
		IndexTree:  indexTree,
		WorkTree:   workTree,
		HeadCommit: headCommit,
		Branch:     branch,
	}

	stack, err := loadStash(repo.Root)
	if err != nil {
		return err
	}
	stack = append([]StashEntry{entry}, stack...)
	if err := saveStash(repo.Root, stack); err != nil {
		return err
	}

	if !headCommit.IsZero() {
		cd, err := repo.Store.ReadCommit(headCommit)
		if err != nil {
			return fmt.Errorf("read HEAD commit: %w", err)
		}
		flatMap, err := FlattenTree(repo.Store, cd.Tree, "")
		if err != nil {
			return fmt.Errorf("flatten HEAD tree: %w", err)
		}
		idx.Entries = nil
		for path, hash := range flatMap {
			idx.Set(path, hash, vcs.ModeRegular)
		}
	} else {
		idx.Entries = nil
	}
	return idx.Save()
}

func StashPop(repo *vcs.Repository) (*StashEntry, error) {
	return stashApplyAt(repo, 0, true)
}

func StashApply(repo *vcs.Repository, index int) (*StashEntry, error) {
	return stashApplyAt(repo, index, false)
}

func stashApplyAt(repo *vcs.Repository, index int, drop bool) (*StashEntry, error) {
	stack, err := loadStash(repo.Root)
	if err != nil {
		return nil, err
	}
	if index < 0 || index >= len(stack) {
		return nil, fmt.Errorf("stash@{%d}: not found (stack has %d entries)", index, len(stack))
	}

	entry := stack[index]

	flatIndex, err := FlattenTree(repo.Store, entry.IndexTree, "")
	if err != nil {
		return nil, fmt.Errorf("flatten stash index tree: %w", err)
	}
	idx, err := LoadIndex(repo.Root)
	if err != nil {
		return nil, err
	}
	idx.Entries = nil
	for path, hash := range flatIndex {
		idx.Set(path, hash, vcs.ModeRegular)
	}
	if err := idx.Save(); err != nil {
		return nil, err
	}

	if drop {
		stack = append(stack[:index], stack[index+1:]...)
		if err := saveStash(repo.Root, stack); err != nil {
			return nil, err
		}
	}

	return &entry, nil
}

func StashList(repo *vcs.Repository) ([]StashEntry, error) {
	return loadStash(repo.Root)
}

func StashDrop(repo *vcs.Repository, index int) error {
	stack, err := loadStash(repo.Root)
	if err != nil {
		return err
	}
	if index < 0 || index >= len(stack) {
		return fmt.Errorf("stash@{%d}: not found (stack has %d entries)", index, len(stack))
	}
	stack = append(stack[:index], stack[index+1:]...)
	return saveStash(repo.Root, stack)
}
