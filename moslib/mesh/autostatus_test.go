package mesh

import (
	"testing"
)

func TestDetectMainBranch_Fallback(t *testing.T) {
	branch := detectMainBranch(t.TempDir())
	if branch != "main" {
		t.Errorf("expected fallback to 'main', got %q", branch)
	}
}

func TestCollectMainCommits_NoRepo(t *testing.T) {
	commits := collectMainCommits(t.TempDir(), "main")
	if commits != nil {
		t.Errorf("expected nil for non-repo dir, got %v", commits)
	}
}
