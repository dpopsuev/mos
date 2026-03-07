package artifact_test

import (
	"os"
	"testing"

	"github.com/dpopsuev/mos/testkit/wire"
)

func TestMain(m *testing.M) {
	wire.Init()
	os.Exit(m.Run())
}
