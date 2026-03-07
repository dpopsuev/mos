package world_test

import (
	"testing"

	"github.com/dpopsuev/mos/moslib/primitive"
	"github.com/dpopsuev/mos/testkit/network"
	"github.com/dpopsuev/mos/testkit/world"
)

func TestTwoUserArtifactAmendment(t *testing.T) {
	bus := network.NewBus()
	w := world.New(t).
		WithUsers("alice", "bob").
		WithNetwork(bus).
		Build()
	defer w.Close()

	// Create a shared repo
	w.CreateRepo("governance")

	// Alice creates an artifact
	alice := w.User("alice")
	a, err := primitive.NewArtifact(
		"no-unused-imports",
		"rule",
		"No Unused Imports",
		`Feature: No Unused Imports
  Scenario: Clean file passes
    Given a Go source file with no unused imports
    When the rule is evaluated
    Then the result is pass`,
		alice.User,
	)
	if err != nil {
		t.Fatalf("new artifact: %v", err)
	}
	if err := alice.Store.Create(a); err != nil {
		t.Fatalf("alice create: %v", err)
	}

	// Alice pushes
	w.Push("alice", "create rule: no-unused-imports")

	// Bob pulls to get Alice's artifact
	w.Pull("bob")

	// Bob reads the artifact from his store (now synced via git)
	bob := w.User("bob")
	bobArtifact, err := bob.Store.Read("no-unused-imports")
	if err != nil {
		t.Fatalf("bob read: %v", err)
	}
	if bobArtifact.Identity.Version != 1 {
		t.Fatalf("bob sees version %d, want 1", bobArtifact.Identity.Version)
	}

	// Bob amends the artifact
	if err := bobArtifact.Amend(bob.User, func(a *primitive.Artifact) {
		a.Title = "No Unused Imports (Strict)"
		a.Status = "active"
	}); err != nil {
		t.Fatalf("bob amend: %v", err)
	}
	if err := bob.Store.Write(bobArtifact); err != nil {
		t.Fatalf("bob write: %v", err)
	}

	// Bob pushes
	w.Push("bob", "amend rule: no-unused-imports")

	// Alice pulls
	w.Pull("alice")

	// Both users see version 2
	w.AssertArtifactVersion("alice", "no-unused-imports", 2)
	w.AssertArtifactVersion("bob", "no-unused-imports", 2)

	// Verify the amendment details
	aliceArtifact, _ := alice.Store.Read("no-unused-imports")
	if aliceArtifact.Title != "No Unused Imports (Strict)" {
		t.Errorf("alice title = %q, want amended title", aliceArtifact.Title)
	}
	if aliceArtifact.Status != "active" {
		t.Errorf("alice status = %q, want active", aliceArtifact.Status)
	}
	if aliceArtifact.Identity.LastAmendedBy != bob.User.Fingerprint() {
		t.Errorf("last_amended_by = %q, want bob's fingerprint", aliceArtifact.Identity.LastAmendedBy)
	}

	// Verify event bus recorded the pushes
	bus.Recorder().AssertReceived(t, "alice", "push")
	bus.Recorder().AssertReceived(t, "bob", "push")
}

func TestThreeUserCreateAmendRatify(t *testing.T) {
	bus := network.NewBus()
	w := world.New(t).
		WithUsers("alice", "bob", "charlie").
		WithNetwork(bus).
		Build()
	defer w.Close()

	w.CreateRepo("project")

	// Alice creates a rule
	alice := w.User("alice")
	rule, _ := primitive.NewArtifact(
		"require-tests",
		"rule",
		"Require Tests",
		"Feature: Require Tests\n  Scenario: Tests exist\n    Given a package\n    Then it has test files",
		alice.User,
	)
	alice.Store.Create(rule)
	w.Push("alice", "create rule: require-tests")

	// Bob pulls, proposes an amendment (bill)
	w.Pull("bob")
	bob := w.User("bob")
	bobRule, _ := bob.Store.Read("require-tests")
	bobRule.Amend(bob.User, func(a *primitive.Artifact) {
		a.Spec.Feature = "Feature: Require Tests\n  Scenario: Coverage above 80%\n    Given a package\n    Then test coverage is above 80%"
	})
	bob.Store.Write(bobRule)
	w.Push("bob", "amend rule: require-tests")

	// Alice pulls, ratifies (signs the amended version)
	w.Pull("alice")
	aliceRule, _ := alice.Store.Read("require-tests")
	aliceRule.Amend(alice.User, func(a *primitive.Artifact) {
		a.Status = "ratified"
	})
	alice.Store.Write(aliceRule)
	w.Push("alice", "ratify rule: require-tests")

	// Charlie pulls -- should see the ratified version
	w.Pull("charlie")
	w.AssertArtifactVersion("charlie", "require-tests", 3)

	charlie := w.User("charlie")
	charlieRule, _ := charlie.Store.Read("require-tests")
	if charlieRule.Status != "ratified" {
		t.Errorf("charlie sees status %q, want ratified", charlieRule.Status)
	}

	// Event bus saw all pushes
	bus.Recorder().AssertCount(t, "alice", 3)
	bus.Recorder().AssertCount(t, "bob", 3)
	bus.Recorder().AssertCount(t, "charlie", 3)
}

func TestPartitionedUserMissesEvents(t *testing.T) {
	bus := network.NewBus()
	w := world.New(t).
		WithUsers("alice", "bob").
		WithNetwork(bus).
		Build()
	defer w.Close()

	w.CreateRepo("partition-test")

	// Partition bob
	bus.Partition("bob")

	// Alice creates and pushes
	alice := w.User("alice")
	rule, _ := primitive.NewArtifact("partitioned-rule", "rule", "P", "Feature: P", alice.User)
	alice.Store.Create(rule)
	w.Push("alice", "create rule while bob partitioned")

	// Bob's recorder should not have the push event
	bus.Recorder().AssertNotReceived(t, "bob", "push")
	bus.Recorder().AssertReceived(t, "alice", "push")

	// Heal bob
	bus.Heal("bob")

	// Alice pushes another event
	rule.Amend(alice.User, func(a *primitive.Artifact) { a.Status = "active" })
	alice.Store.Write(rule)
	w.Push("alice", "activate rule")

	// Now bob should receive the new push
	bus.Recorder().AssertReceived(t, "bob", "push")
}
