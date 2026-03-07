package artifact

import (
	"testing"

	"github.com/dpopsuev/mos/moslib/dsl"
)

func TestPathGetTopLevel(t *testing.T) {
	src := `contract "CON-PATH" {
  title = "Path test"
  status = "draft"
}
`
	f, err := dsl.Parse(src, nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	ab := f.Artifact.(*dsl.ArtifactBlock)
	val, err := PathGet(ab, "title")
	if err != nil {
		t.Fatalf("PathGet title: %v", err)
	}
	if val != "Path test" {
		t.Errorf("expected 'Path test', got %q", val)
	}
}

func TestPathGetNested(t *testing.T) {
	src := `contract "CON-NEST" {
  title = "Nested"
  status = "draft"
  coverage {
    unit {
      applies = true
    }
  }
}
`
	f, err := dsl.Parse(src, nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	ab := f.Artifact.(*dsl.ArtifactBlock)
	val, err := PathGet(ab, "coverage.unit.applies")
	if err != nil {
		t.Fatalf("PathGet nested: %v", err)
	}
	if val != "true" {
		t.Errorf("expected 'true', got %q", val)
	}
}

func TestPathSetCreatesBlocks(t *testing.T) {
	src := `contract "CON-SET" {
  title = "Set test"
  status = "draft"
}
`
	f, err := dsl.Parse(src, nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	ab := f.Artifact.(*dsl.ArtifactBlock)
	if err := PathSet(ab, "coverage.unit.applies", "true"); err != nil {
		t.Fatalf("PathSet: %v", err)
	}
	val, err := PathGet(ab, "coverage.unit.applies")
	if err != nil {
		t.Fatalf("PathGet after set: %v", err)
	}
	if val != "true" {
		t.Errorf("expected 'true', got %q", val)
	}
}

func TestPathAppend(t *testing.T) {
	src := `contract "CON-APP" {
  title = "Append test"
  status = "draft"
  scope {
    depends_on = ["CON-A"]
  }
}
`
	f, err := dsl.Parse(src, nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	ab := f.Artifact.(*dsl.ArtifactBlock)
	if err := PathAppend(ab, "scope.depends_on", "CON-B"); err != nil {
		t.Fatalf("PathAppend: %v", err)
	}
	val, err := PathGet(ab, "scope.depends_on")
	if err != nil {
		t.Fatalf("PathGet after append: %v", err)
	}
	if val != "CON-A,CON-B" {
		t.Errorf("expected 'CON-A,CON-B', got %q", val)
	}
}

func TestParseValueStripsOuterQuotes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"SPR-2026-012"`, "SPR-2026-012"},
		{"SPR-2026-012", "SPR-2026-012"},
		{`""`, ""},
		{"plain", "plain"},
	}
	for _, tt := range tests {
		v := parseValue(tt.input)
		sv, ok := v.(*dsl.StringVal)
		if !ok {
			t.Errorf("parseValue(%q): expected StringVal, got %T", tt.input, v)
			continue
		}
		if sv.Text != tt.want {
			t.Errorf("parseValue(%q).Text = %q, want %q", tt.input, sv.Text, tt.want)
		}
	}
}

func TestPathSetIdempotent(t *testing.T) {
	src := `contract "CON-IDEM" {
  title = "Idempotent test"
  status = "draft"
}
`
	f, err := dsl.Parse(src, nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	ab := f.Artifact.(*dsl.ArtifactBlock)

	if err := PathSet(ab, "sprint", "SPR-2026-012"); err != nil {
		t.Fatalf("first set: %v", err)
	}
	if err := PathSet(ab, "sprint", "SPR-2026-012"); err != nil {
		t.Fatalf("second set: %v", err)
	}
	val, err := PathGet(ab, "sprint")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if val != "SPR-2026-012" {
		t.Errorf("expected SPR-2026-012, got %q", val)
	}
}
