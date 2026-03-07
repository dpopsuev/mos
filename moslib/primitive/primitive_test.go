package primitive_test

import (
	"path/filepath"
	"testing"

	"github.com/dpopsuev/mos/moslib/primitive"
	"github.com/dpopsuev/mos/testkit/user"
)

func TestNewArtifact(t *testing.T) {
	u := user.NewUser(t, "alice")

	a, err := primitive.NewArtifact("test-rule", "rule", "Test Rule", `
Feature: Test Rule
  Scenario: Always passes
    Given nothing
    Then everything is fine
`, u)
	if err != nil {
		t.Fatalf("new artifact: %v", err)
	}

	if a.ID != "test-rule" {
		t.Errorf("id = %q, want test-rule", a.ID)
	}
	if a.Kind != "rule" {
		t.Errorf("kind = %q, want rule", a.Kind)
	}
	if a.Status != "draft" {
		t.Errorf("status = %q, want draft", a.Status)
	}
	if a.Identity.Version != 1 {
		t.Errorf("version = %d, want 1", a.Identity.Version)
	}
	if a.Identity.CreatedBy != u.Fingerprint() {
		t.Errorf("created_by = %q, want %q", a.Identity.CreatedBy, u.Fingerprint())
	}
	if a.Identity.CreationSignature == "" {
		t.Error("creation signature is empty")
	}
}

func TestAmendArtifact(t *testing.T) {
	alice := user.NewUser(t, "alice")
	bob := user.NewUser(t, "bob")

	a, err := primitive.NewArtifact("my-rule", "rule", "My Rule", "Feature: X", alice)
	if err != nil {
		t.Fatalf("new artifact: %v", err)
	}

	err = a.Amend(bob, func(a *primitive.Artifact) {
		a.Title = "My Amended Rule"
		a.Status = "active"
	})
	if err != nil {
		t.Fatalf("amend: %v", err)
	}

	if a.Identity.Version != 2 {
		t.Errorf("version = %d, want 2", a.Identity.Version)
	}
	if a.Identity.LastAmendedBy != bob.Fingerprint() {
		t.Errorf("last_amended_by = %q, want %q", a.Identity.LastAmendedBy, bob.Fingerprint())
	}
	if a.Title != "My Amended Rule" {
		t.Errorf("title = %q, want My Amended Rule", a.Title)
	}
	if a.Status != "active" {
		t.Errorf("status = %q, want active", a.Status)
	}
}

func TestFSStoreRoundTrip(t *testing.T) {
	u := user.NewUser(t, "alice")
	dir := filepath.Join(t.TempDir(), "artifacts")

	store, err := primitive.NewFSStore(dir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	a, err := primitive.NewArtifact("roundtrip-test", "rule", "Roundtrip", "Feature: RT", u)
	if err != nil {
		t.Fatalf("new artifact: %v", err)
	}

	if err := store.Create(a); err != nil {
		t.Fatalf("create: %v", err)
	}

	loaded, err := store.Read("roundtrip-test")
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if loaded.ID != a.ID {
		t.Errorf("id = %q, want %q", loaded.ID, a.ID)
	}
	if loaded.Title != a.Title {
		t.Errorf("title = %q, want %q", loaded.Title, a.Title)
	}
	if loaded.Identity.Version != a.Identity.Version {
		t.Errorf("version = %d, want %d", loaded.Identity.Version, a.Identity.Version)
	}
	if loaded.Identity.CreatedBy != a.Identity.CreatedBy {
		t.Errorf("created_by = %q, want %q", loaded.Identity.CreatedBy, a.Identity.CreatedBy)
	}
	if loaded.Spec.Feature != a.Spec.Feature {
		t.Errorf("spec mismatch")
	}
}

func TestFSStoreCreateDuplicate(t *testing.T) {
	u := user.NewUser(t, "alice")
	dir := filepath.Join(t.TempDir(), "artifacts")

	store, err := primitive.NewFSStore(dir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	a, _ := primitive.NewArtifact("dup-test", "rule", "Dup", "Feature: D", u)
	if err := store.Create(a); err != nil {
		t.Fatalf("first create: %v", err)
	}

	if err := store.Create(a); err == nil {
		t.Fatal("expected error on duplicate create")
	}
}

func TestFSStoreList(t *testing.T) {
	u := user.NewUser(t, "alice")
	dir := filepath.Join(t.TempDir(), "artifacts")

	store, err := primitive.NewFSStore(dir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	for _, id := range []string{"alpha", "beta", "gamma"} {
		a, _ := primitive.NewArtifact(id, "rule", id, "Feature: "+id, u)
		if err := store.Create(a); err != nil {
			t.Fatalf("create %s: %v", id, err)
		}
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("list length = %d, want 3", len(list))
	}
}

func TestFSStoreAmendAndReread(t *testing.T) {
	u := user.NewUser(t, "alice")
	bob := user.NewUser(t, "bob")
	dir := filepath.Join(t.TempDir(), "artifacts")

	store, err := primitive.NewFSStore(dir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	a, _ := primitive.NewArtifact("amend-test", "rule", "Original", "Feature: V1", u)
	if err := store.Create(a); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := a.Amend(bob, func(a *primitive.Artifact) {
		a.Title = "Amended"
		a.Spec.Feature = "Feature: V2"
	}); err != nil {
		t.Fatalf("amend: %v", err)
	}

	if err := store.Write(a); err != nil {
		t.Fatalf("write amended: %v", err)
	}

	loaded, err := store.Read("amend-test")
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if loaded.Identity.Version != 2 {
		t.Errorf("version = %d, want 2", loaded.Identity.Version)
	}
	if loaded.Title != "Amended" {
		t.Errorf("title = %q, want Amended", loaded.Title)
	}
	if loaded.Spec.Feature != "Feature: V2" {
		t.Errorf("spec = %q, want Feature: V2", loaded.Spec.Feature)
	}
}
