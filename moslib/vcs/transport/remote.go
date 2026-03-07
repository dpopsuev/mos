package transport

import (
	"fmt"

	gitconfig "github.com/go-git/go-git/v5/config"

	"github.com/dpopsuev/mos/moslib/vcs"
)

type MosRemoteConfig struct {
	Name string
	URL  string
}

func AddRemote(repo *vcs.Repository, name, url string) error {
	gr, err := repo.RequireGit()
	if err != nil {
		return err
	}
	_, err = gr.CreateRemote(&gitconfig.RemoteConfig{
		Name: name,
		URLs: []string{url},
		Fetch: []gitconfig.RefSpec{
			gitconfig.RefSpec(fmt.Sprintf("+refs/mos/*:refs/mos/remotes/%s/*", name)),
		},
	})
	if err != nil {
		return fmt.Errorf("add remote %q: %w", name, err)
	}
	return nil
}

func RemoveRemote(repo *vcs.Repository, name string) error {
	gr, err := repo.RequireGit()
	if err != nil {
		return err
	}
	if err := gr.DeleteRemote(name); err != nil {
		return fmt.Errorf("remove remote %q: %w", name, err)
	}
	return nil
}

func ListRemotes(repo *vcs.Repository) ([]MosRemoteConfig, error) {
	gr, err := repo.RequireGit()
	if err != nil {
		return nil, err
	}
	remotes, err := gr.Remotes()
	if err != nil {
		return nil, fmt.Errorf("list remotes: %w", err)
	}
	var result []MosRemoteConfig
	for _, r := range remotes {
		cfg := r.Config()
		url := ""
		if len(cfg.URLs) > 0 {
			url = cfg.URLs[0]
		}
		result = append(result, MosRemoteConfig{Name: cfg.Name, URL: url})
	}
	return result, nil
}
