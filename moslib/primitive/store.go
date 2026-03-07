package primitive

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Store persists and retrieves artifacts.
type Store interface {
	Create(artifact *Artifact) error
	Read(id string) (*Artifact, error)
	Write(artifact *Artifact) error
	List() ([]*Artifact, error)
}

// FSStore is a filesystem-backed Store that writes TOML files to a directory.
type FSStore struct {
	Dir string
}

// NewFSStore creates a store rooted at the given directory.
// The directory is created if it does not exist.
func NewFSStore(dir string) (*FSStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}
	return &FSStore{Dir: dir}, nil
}

func (s *FSStore) path(id string) string {
	return filepath.Join(s.Dir, id+".toml")
}

// Create writes a new artifact to disk. Fails if the artifact already exists.
func (s *FSStore) Create(a *Artifact) error {
	p := s.path(a.ID)
	if _, err := os.Stat(p); err == nil {
		return fmt.Errorf("artifact %q already exists", a.ID)
	}
	return s.Write(a)
}

// Read loads an artifact from disk by ID.
func (s *FSStore) Read(id string) (*Artifact, error) {
	p := s.path(id)

	var wrapper artifactFile
	if _, err := toml.DecodeFile(p, &wrapper); err != nil {
		return nil, fmt.Errorf("read artifact %q: %w", id, err)
	}
	return wrapper.toArtifact(), nil
}

// Write persists an artifact to disk, overwriting any existing file.
func (s *FSStore) Write(a *Artifact) error {
	p := s.path(a.ID)
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("write artifact %q: %w", a.ID, err)
	}
	defer f.Close()

	wrapper := newArtifactFile(a)
	if err := toml.NewEncoder(f).Encode(wrapper); err != nil {
		return fmt.Errorf("encode artifact %q: %w", a.ID, err)
	}
	return nil
}

// List returns all artifacts in the store directory.
func (s *FSStore) List() ([]*Artifact, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		return nil, fmt.Errorf("list artifacts: %w", err)
	}

	var artifacts []*Artifact
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".toml" {
			continue
		}
		id := e.Name()[:len(e.Name())-len(".toml")]
		a, err := s.Read(id)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, a)
	}
	return artifacts, nil
}

// artifactFile is the TOML-level representation matching the schema structure.
type artifactFile struct {
	Artifact artifactSection `toml:"artifact"`
	Spec     Spec            `toml:"spec"`
	Identity Identity        `toml:"identity"`
}

type artifactSection struct {
	ID     string   `toml:"id"`
	Kind   string   `toml:"kind"`
	Title  string   `toml:"title"`
	Status string   `toml:"status"`
	Scope  []string `toml:"scope,omitempty"`
}

func newArtifactFile(a *Artifact) artifactFile {
	return artifactFile{
		Artifact: artifactSection{
			ID:     a.ID,
			Kind:   a.Kind,
			Title:  a.Title,
			Status: a.Status,
			Scope:  a.Scope,
		},
		Spec:     a.Spec,
		Identity: a.Identity,
	}
}

func (f *artifactFile) toArtifact() *Artifact {
	return &Artifact{
		ID:       f.Artifact.ID,
		Kind:     f.Artifact.Kind,
		Title:    f.Artifact.Title,
		Status:   f.Artifact.Status,
		Scope:    f.Artifact.Scope,
		Spec:     f.Spec,
		Identity: f.Identity,
	}
}
