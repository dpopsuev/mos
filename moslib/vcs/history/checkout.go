package history

import (
	"fmt"

	"github.com/dpopsuev/mos/moslib/vcs"
	"github.com/dpopsuev/mos/moslib/vcs/staging"
)

type CheckoutResult struct {
	Branch   string
	Detached bool
	Hash     vcs.Hash
}

func Checkout(repo *vcs.Repository, target string) (*CheckoutResult, error) {
	if h, err := repo.Store.ResolveRef(vcs.HeadsPrefix + target); err == nil {
		if err := vcs.WriteSymbolicHead(repo.Root, target); err != nil {
			return nil, err
		}
		if err := staging.ResetIndexToCommit(repo, h); err != nil {
			return nil, err
		}
		return &CheckoutResult{Branch: target, Hash: h}, nil
	}

	if h, err := repo.Store.ResolveRef(vcs.TagsPrefix + target); err == nil {
		if err := vcs.WriteDetachedHead(repo.Root, h); err != nil {
			return nil, err
		}
		if err := staging.ResetIndexToCommit(repo, h); err != nil {
			return nil, err
		}
		return &CheckoutResult{Detached: true, Hash: h}, nil
	}

	h, err := vcs.ParseHash(target)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve %q as branch, tag, or hash", target)
	}
	if !repo.Store.HasObject(h) {
		return nil, fmt.Errorf("object %s not found", h.Short())
	}
	if err := vcs.WriteDetachedHead(repo.Root, h); err != nil {
		return nil, err
	}
	if err := staging.ResetIndexToCommit(repo, h); err != nil {
		return nil, err
	}
	return &CheckoutResult{Detached: true, Hash: h}, nil
}

func CheckoutNewBranch(repo *vcs.Repository, name string) (*CheckoutResult, error) {
	head, err := vcs.ResolveHead(repo)
	if err != nil {
		return nil, fmt.Errorf("resolve HEAD: %w", err)
	}
	if err := CreateBranch(repo.Store, name, head); err != nil {
		return nil, err
	}
	if err := vcs.WriteSymbolicHead(repo.Root, name); err != nil {
		return nil, err
	}
	return &CheckoutResult{Branch: name, Hash: head}, nil
}
