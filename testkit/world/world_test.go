package world_test

import (
	"testing"

	"github.com/dpopsuev/mos/moslib/primitive"
	"github.com/dpopsuev/mos/testkit/world"
)

func TestWorldBuilderCreatesUsers(t *testing.T) {
	w := world.New(t).
		WithUsers("alice", "bob", "charlie").
		Build()
	defer w.Close()

	for _, name := range []string{"alice", "bob", "charlie"} {
		p := w.User(name)
		if p.User.Name != name {
			t.Errorf("user name = %q, want %q", p.User.Name, name)
		}
		if p.Store == nil {
			t.Errorf("user %s has nil store", name)
		}
	}
}

func TestWorldCreateRepoAndClone(t *testing.T) {
	w := world.New(t).
		WithUsers("alice", "bob").
		Build()
	defer w.Close()

	w.CreateRepo("test-project")

	for _, name := range []string{"alice", "bob"} {
		p := w.User(name)
		if p.CloneDir == "" {
			t.Errorf("%s has empty clone dir", name)
		}
	}
}

func TestWorldArtifactRoundTrip(t *testing.T) {
	w := world.New(t).
		WithUsers("alice").
		Build()
	defer w.Close()

	alice := w.User("alice")
	a, err := primitive.NewArtifact("test-rule", "rule", "Test", "Feature: X", alice.User)
	if err != nil {
		t.Fatalf("new artifact: %v", err)
	}

	if err := alice.Store.Create(a); err != nil {
		t.Fatalf("create: %v", err)
	}

	w.AssertArtifactVersion("alice", "test-rule", 1)
}
