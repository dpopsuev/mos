package transport

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"

	"github.com/dpopsuev/mos/moslib/vcs"
)

type CloneResult struct {
	Root   string
	Branch string
}

func Clone(url, destDir string) (*CloneResult, error) {
	if destDir == "" {
		destDir = inferDestFromURL(url)
	}

	repo, err := git.PlainClone(destDir, false, &git.CloneOptions{
		URL:        url,
		NoCheckout: true,
	})
	if err != nil {
		if err2 := cloneFallback(url, destDir); err2 != nil {
			return nil, fmt.Errorf("clone %s: %w (fallback: %v)", url, err, err2)
		}
	} else {
		if err := fetchMosRefs(repo, "origin"); err != nil {
			return nil, fmt.Errorf("clone: fetch mos refs: %w", err)
		}
	}

	cfg := vcs.VCSConfig{Backend: vcs.BackendGit}
	if err := vcs.WriteVCSConfig(destDir, cfg); err != nil {
		return nil, fmt.Errorf("clone: write vcs config: %w", err)
	}

	headPath := filepath.Join(destDir, vcs.SymHeadFile)
	if !fileExists(headPath) {
		if err := vcs.InitHead(destDir); err != nil {
			return nil, fmt.Errorf("clone: init HEAD: %w", err)
		}
	}

	branch := vcs.CurrentBranch(destDir)
	if branch == "" {
		branch = vcs.DefaultBranch
	}

	return &CloneResult{Root: destDir, Branch: branch}, nil
}

func cloneFallback(url, destDir string) error {
	repo, err := git.PlainInit(destDir, false)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}
	_, err = repo.CreateRemote(&gitconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
		Fetch: []gitconfig.RefSpec{
			"+refs/*:refs/*",
		},
	})
	if err != nil {
		return fmt.Errorf("create remote: %w", err)
	}
	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []gitconfig.RefSpec{"+refs/*:refs/*"},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("fetch: %w", err)
	}
	return nil
}

func fetchMosRefs(repo *git.Repository, remoteName string) error {
	err := repo.Fetch(&git.FetchOptions{
		RemoteName: remoteName,
		RefSpecs:   []gitconfig.RefSpec{"+refs/mos/*:refs/mos/*"},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
}

func inferDestFromURL(url string) string {
	base := filepath.Base(url)
	base = strings.TrimSuffix(base, ".git")
	if base == "" || base == "." {
		return "repo"
	}
	return base
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
