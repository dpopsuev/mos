package vcs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
)

// VCSConfig is persisted at .mos/vcs.json and controls which backend is active.
type VCSConfig struct {
	Backend string `json:"backend"` // "fs" or "git"
}

// Repository is the high-level handle tying a working directory to an ObjectStore.
type Repository struct {
	Root   string
	Store  ObjectStore
	Config VCSConfig
}

const (
	BackendFS  = "fs"
	BackendGit = "git"
	vcsDir     = ".mos/vcs"
	configFile = ".mos/vcs.json"
)

// InitRepo initializes a governance VCS repository at root.
func InitRepo(root, backend string) (*Repository, error) {
	if backend == "" {
		backend = BackendFS
	}
	store, err := openStore(root, backend)
	if err != nil {
		return nil, err
	}
	cfg := VCSConfig{Backend: backend}
	if err := WriteVCSConfig(root, cfg); err != nil {
		return nil, err
	}
	if err := InitHead(root); err != nil {
		return nil, fmt.Errorf("init HEAD: %w", err)
	}
	return &Repository{Root: root, Store: store, Config: cfg}, nil
}

// OpenRepo opens an existing governance VCS repository.
func OpenRepo(root string) (*Repository, error) {
	cfg, err := readVCSConfig(root)
	if err != nil {
		return nil, fmt.Errorf("no vcs repo at %s: %w", root, err)
	}
	store, err := openStore(root, cfg.Backend)
	if err != nil {
		return nil, err
	}
	return &Repository{Root: root, Store: store, Config: cfg}, nil
}

// GitRepo returns the underlying go-git Repository handle.
// Returns nil if the backend is not git.
func (r *Repository) GitRepo() *git.Repository {
	if gs, ok := r.Store.(*GitStore); ok {
		return gs.repo
	}
	return nil
}

// RequireGit returns the go-git Repository or ErrFSStoreNoRemote.
func (r *Repository) RequireGit() (*git.Repository, error) {
	gr := r.GitRepo()
	if gr == nil {
		return nil, ErrFSStoreNoRemote
	}
	return gr, nil
}

func openStore(root, backend string) (ObjectStore, error) {
	switch backend {
	case BackendFS:
		dir := filepath.Join(root, vcsDir, "store")
		return NewFSStore(dir)
	case BackendGit:
		return NewGitStore(root)
	default:
		return nil, fmt.Errorf("unknown backend: %q", backend)
	}
}

func WriteVCSConfig(root string, cfg VCSConfig) error {
	p := filepath.Join(root, configFile)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

func readVCSConfig(root string) (VCSConfig, error) {
	data, err := os.ReadFile(filepath.Join(root, configFile))
	if err != nil {
		return VCSConfig{}, err
	}
	var cfg VCSConfig
	return cfg, json.Unmarshal(data, &cfg)
}
