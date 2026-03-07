package survey_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dpopsuev/mos/moslib/survey"
)

func projectRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine project root via runtime.Caller")
	}
	// file is cstlib/survey/dogfood_test.go, root is two dirs up
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func TestDogfoodMos(t *testing.T) {
	root := projectRoot(t)

	sc := &survey.GoScanner{}
	mod, err := sc.Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if mod.Path != "github.com/dpopsuev/mos" {
		t.Fatalf("module path = %q", mod.Path)
	}

	wantPackages := []string{
		"github.com/dpopsuev/mos/cmd/mos",
		"github.com/dpopsuev/mos/testkit/curia",
		"github.com/dpopsuev/mos/moslib/linter",
		"github.com/dpopsuev/mos/moslib/lsp",
		"github.com/dpopsuev/mos/moslib/model",
		"github.com/dpopsuev/mos/moslib/primitive",
		"github.com/dpopsuev/mos/moslib/survey",
		"github.com/dpopsuev/mos/testkit/forge",
		"github.com/dpopsuev/mos/testkit/gitcompat",
		"github.com/dpopsuev/mos/testkit/network",
		"github.com/dpopsuev/mos/testkit/user",
		"github.com/dpopsuev/mos/testkit/world",
	}

	pkgSet := make(map[string]bool)
	for _, p := range mod.Namespaces {
		pkgSet[p.ImportPath] = true
	}

	for _, want := range wantPackages {
		if !pkgSet[want] {
			t.Errorf("missing package %s", want)
		}
	}

	if mod.DependencyGraph == nil {
		t.Fatal("dependency graph is nil")
	}

	wantEdges := []struct {
		from, to string
		external bool
	}{
		{
			"github.com/dpopsuev/mos/moslib/survey",
			"github.com/dpopsuev/mos/moslib/model",
			false,
		},
		{
			"github.com/dpopsuev/mos/testkit/world",
			"github.com/dpopsuev/mos/moslib/primitive",
			false,
		},
	}

	for _, want := range wantEdges {
		found := false
		for _, e := range mod.DependencyGraph.Edges {
			if e.From == want.from && e.To == want.to {
				if e.External != want.external {
					t.Errorf("edge %s -> %s: external = %v, want %v",
						want.from, want.to, e.External, want.external)
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing edge %s -> %s", want.from, want.to)
		}
	}
}
