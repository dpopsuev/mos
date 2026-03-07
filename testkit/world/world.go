package world

import (
	"path/filepath"
	"testing"

	"github.com/dpopsuev/mos/moslib/primitive"
	"github.com/dpopsuev/mos/testkit/forge"
	"github.com/dpopsuev/mos/testkit/network"
	"github.com/dpopsuev/mos/testkit/user"
)

// World is a test scenario environment composing forge, users, and network.
type World struct {
	t     testing.TB
	forge forge.Forge
	bus   *network.Bus
	users map[string]*Participant
}

// Participant is a user within a world: identity + local clone + artifact store.
type Participant struct {
	User  *user.User
	Store *primitive.FSStore
	// CloneDir is the local git clone directory for this user.
	CloneDir string
}

// Builder constructs a World with a fluent API.
type Builder struct {
	t         testing.TB
	forge     forge.Forge
	bus       *network.Bus
	userNames []string
}

// New starts building a test world.
func New(t testing.TB) *Builder {
	return &Builder{t: t}
}

// WithForge sets the forge for the world.
func (b *Builder) WithForge(f forge.Forge) *Builder {
	b.forge = f
	return b
}

// WithNetwork sets the event bus for the world.
func (b *Builder) WithNetwork(bus *network.Bus) *Builder {
	b.bus = bus
	return b
}

// WithUsers adds named users to the world.
func (b *Builder) WithUsers(names ...string) *Builder {
	b.userNames = append(b.userNames, names...)
	return b
}

// Build constructs the world. Each user gets an identity, a working directory,
// and a primitive store.
func (b *Builder) Build() *World {
	b.t.Helper()

	if b.forge == nil {
		b.forge = forge.InProcess(b.t)
	}
	if b.bus == nil {
		b.bus = network.NewBus()
	}

	w := &World{
		t:     b.t,
		forge: b.forge,
		bus:   b.bus,
		users: make(map[string]*Participant),
	}

	for _, name := range b.userNames {
		u := user.NewUser(b.t, name)
		conDir := filepath.Join(u.WorkDir, ".mos", "artifacts")
		store, err := primitive.NewFSStore(conDir)
		if err != nil {
			b.t.Fatalf("create store for %s: %v", name, err)
		}

		w.users[name] = &Participant{
			User:     u,
			Store:    store,
			CloneDir: u.WorkDir,
		}

		b.bus.Subscribe(name)
	}

	return w
}

// Forge returns the world's forge.
func (w *World) Forge() forge.Forge {
	return w.forge
}

// Bus returns the world's event bus.
func (w *World) Bus() *network.Bus {
	return w.bus
}

// User returns a participant by name.
func (w *World) User(name string) *Participant {
	p, ok := w.users[name]
	if !ok {
		w.t.Fatalf("user %q not found in world", name)
	}
	return p
}

// Users returns all participants.
func (w *World) Users() map[string]*Participant {
	return w.users
}

// CreateRepo creates a repo on the forge and clones it for all users.
func (w *World) CreateRepo(name string) string {
	w.t.Helper()

	repoURL, err := w.forge.CreateRepo(name)
	if err != nil {
		w.t.Fatalf("create repo %s: %v", name, err)
	}

	for userName, p := range w.users {
		cloneDir := filepath.Join(p.User.WorkDir, name)
		if _, err := forge.GitExec("", "clone", repoURL, cloneDir); err != nil {
			w.t.Fatalf("clone %s for %s: %v", name, userName, err)
		}
		p.CloneDir = cloneDir

		conDir := filepath.Join(cloneDir, ".mos", "artifacts")
		store, err := primitive.NewFSStore(conDir)
		if err != nil {
			w.t.Fatalf("create store for %s in %s: %v", userName, name, err)
		}
		p.Store = store
	}

	return repoURL
}

// Push runs git add + commit + push in a user's clone directory.
func (w *World) Push(userName, message string) {
	w.t.Helper()
	p := w.User(userName)

	forge.GitExec(p.CloneDir, "add", "-A")
	if _, err := forge.GitExec(p.CloneDir, "commit", "-m", message); err != nil {
		w.t.Fatalf("%s commit: %v", userName, err)
	}
	if _, err := forge.GitExec(p.CloneDir, "push"); err != nil {
		w.t.Fatalf("%s push: %v", userName, err)
	}

	w.bus.Publish(network.Event{
		Type:   "push",
		Source: userName,
	})
}

// Pull runs git pull in a user's clone directory.
func (w *World) Pull(userName string) {
	w.t.Helper()
	p := w.User(userName)

	if _, err := forge.GitExec(p.CloneDir, "pull"); err != nil {
		w.t.Fatalf("%s pull: %v", userName, err)
	}
}

// AssertArtifactVersion reads an artifact from a user's store and checks its version.
func (w *World) AssertArtifactVersion(userName, artifactID string, wantVersion int) {
	w.t.Helper()
	p := w.User(userName)

	a, err := p.Store.Read(artifactID)
	if err != nil {
		w.t.Fatalf("read artifact %q from %s: %v", artifactID, userName, err)
	}
	if a.Identity.Version != wantVersion {
		w.t.Errorf("%s artifact %q version = %d, want %d",
			userName, artifactID, a.Identity.Version, wantVersion)
	}
}

// Close cleans up the world. Called automatically via t.Cleanup if using the builder.
func (w *World) Close() {
	if w.forge != nil {
		w.forge.Close()
	}
}
