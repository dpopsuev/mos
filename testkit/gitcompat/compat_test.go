package gitcompat_test

import (
	"testing"

	"github.com/dpopsuev/mos/testkit/gitcompat"
)

func TestGoGitObjectsPassFsck(t *testing.T) {
	gitcompat.AssertObjectValid(t)
}

func TestRoundTripGoGitAndGit(t *testing.T) {
	gitcompat.AssertRoundTrip(t)
}
