package transport

import (
	"errors"
	"fmt"

	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"

	"github.com/dpopsuev/mos/moslib/vcs"
	vcmerge "github.com/dpopsuev/mos/moslib/vcs/merge"
	"github.com/dpopsuev/mos/moslib/vcs/staging"
)

func Fetch(repo *vcs.Repository, remoteName string) error {
	gr, err := repo.RequireGit()
	if err != nil {
		return err
	}
	if remoteName == "" {
		remoteName = "origin"
	}

	fetchOpts := &git.FetchOptions{
		RemoteName: remoteName,
		RefSpecs: []gitconfig.RefSpec{
			gitconfig.RefSpec(fmt.Sprintf("+refs/mos/*:refs/mos/remotes/%s/*", remoteName)),
		},
	}
	if err := gr.Fetch(fetchOpts); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return nil
		}
		return fmt.Errorf("fetch from %s: %w", remoteName, err)
	}
	return nil
}

func Pull(repo *vcs.Repository, remoteName, branch, author, email string) (*vcmerge.MergeResult, error) {
	if _, err := repo.RequireGit(); err != nil {
		return nil, err
	}
	if remoteName == "" {
		remoteName = "origin"
	}
	if branch == "" {
		branch = vcs.CurrentBranch(repo.Root)
		if branch == "" {
			return nil, fmt.Errorf("pull: HEAD is detached; specify a branch")
		}
	}

	if err := Fetch(repo, remoteName); err != nil {
		return nil, fmt.Errorf("pull: %w", err)
	}

	remoteRef := fmt.Sprintf("remotes/%s/heads/%s", remoteName, branch)
	theirHash, err := repo.Store.ResolveRef(remoteRef)
	if err != nil {
		return nil, fmt.Errorf("pull: remote branch %q not found after fetch: %w", branch, err)
	}

	ourHead, err := vcs.ResolveHead(repo)
	if err != nil {
		if err := repo.Store.UpdateRef(vcs.HeadsPrefix+branch, theirHash); err != nil {
			return nil, fmt.Errorf("pull: update ref: %w", err)
		}
		if err := staging.ResetIndexToCommit(repo, theirHash); err != nil {
			return nil, fmt.Errorf("pull: reset index: %w", err)
		}
		return &vcmerge.MergeResult{CommitHash: theirHash, FastForward: true}, nil
	}

	if ourHead == theirHash {
		return &vcmerge.MergeResult{CommitHash: ourHead}, nil
	}

	return vcmerge.MergeHash(repo, theirHash, author, email)
}
