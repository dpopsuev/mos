package transport

import (
	"errors"
	"fmt"

	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"

	"github.com/dpopsuev/mos/moslib/vcs"
)

type PushOpts struct {
	Force  bool
	All    bool
	Branch string
}

func Push(repo *vcs.Repository, remoteName string, opts PushOpts) error {
	gr, err := repo.RequireGit()
	if err != nil {
		return err
	}
	if remoteName == "" {
		remoteName = "origin"
	}

	var refSpecs []gitconfig.RefSpec
	if opts.All {
		spec := gitconfig.RefSpec("refs/mos/heads/*:refs/mos/heads/*")
		if opts.Force {
			spec = gitconfig.RefSpec("+refs/mos/heads/*:refs/mos/heads/*")
		}
		refSpecs = append(refSpecs, spec)
	} else {
		branch := opts.Branch
		if branch == "" {
			branch = vcs.CurrentBranch(repo.Root)
			if branch == "" {
				return fmt.Errorf("push: HEAD is detached; specify a branch")
			}
		}
		src := "refs/mos/heads/" + branch
		dst := "refs/mos/heads/" + branch
		spec := gitconfig.RefSpec(src + ":" + dst)
		if opts.Force {
			spec = gitconfig.RefSpec("+" + src + ":" + dst)
		}
		refSpecs = append(refSpecs, spec)
	}

	refSpecs = append(refSpecs, gitconfig.RefSpec("refs/mos/tags/*:refs/mos/tags/*"))

	pushOpts := &git.PushOptions{
		RemoteName: remoteName,
		RefSpecs:   refSpecs,
	}
	if err := gr.Push(pushOpts); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return nil
		}
		return fmt.Errorf("push to %s: %w", remoteName, err)
	}
	return nil
}
