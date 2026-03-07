//go:build integration

package forge_test

import (
	"testing"

	"github.com/dpopsuev/mos/testkit/forge"
	"github.com/dpopsuev/mos/testkit/gitcompat"
)

// TestGiteaGovernanceRoundTrip verifies NEED-2026-004 criterion C7:
// governance objects stored via GitStore survive a push-through-Gitea-clone-back
// cycle, proving that refs/mos/* are accepted and served by a real HTTP forge.
func TestGiteaGovernanceRoundTrip(t *testing.T) {
	f := forge.Gitea(t)
	gitcompat.AssertGovernanceRoundTrip(t, f)
}
