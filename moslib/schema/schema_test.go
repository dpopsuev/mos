package schema

import (
	"slices"
	"testing"
)

func TestDefaultCoreSchemas_ContainsExpectedKinds(t *testing.T) {
	schemas := DefaultCoreSchemas()
	for _, kind := range []string{"contract", "specification"} {
		if _, ok := schemas[kind]; !ok {
			t.Errorf("DefaultCoreSchemas missing kind %q", kind)
		}
	}
}

func TestDefaultCoreSchemas_ContractFields(t *testing.T) {
	s := DefaultCoreSchemas()["contract"]
	if s.Kind != "contract" {
		t.Errorf("Kind = %q, want contract", s.Kind)
	}
	if s.Directory != "contracts" {
		t.Errorf("Directory = %q, want contracts", s.Directory)
	}

	wantLinks := []string{"justifies", "implements", "documents", "sprint", "batch", "parent", "depends_on"}
	for _, name := range wantLinks {
		idx := slices.IndexFunc(s.Fields, func(f FieldSchema) bool { return f.Name == name })
		if idx < 0 {
			t.Errorf("contract schema missing link field %q", name)
			continue
		}
		if !s.Fields[idx].Link {
			t.Errorf("field %q: Link = false, want true", name)
		}
	}
}

func TestDefaultCoreSchemas_ContractRefKinds(t *testing.T) {
	s := DefaultCoreSchemas()["contract"]
	wantRefKinds := map[string]string{
		"justifies":  "specification",
		"sprint":     "sprint",
		"batch":      "batch",
		"parent":     "contract",
		"depends_on": "contract",
	}
	for name, wantKind := range wantRefKinds {
		idx := slices.IndexFunc(s.Fields, func(f FieldSchema) bool { return f.Name == name })
		if idx < 0 {
			t.Fatalf("missing field %q", name)
		}
		if s.Fields[idx].RefKind != wantKind {
			t.Errorf("field %q: RefKind = %q, want %q", name, s.Fields[idx].RefKind, wantKind)
		}
	}
}

func TestDefaultCoreSchemas_ContractFieldsWithoutRefKind(t *testing.T) {
	s := DefaultCoreSchemas()["contract"]
	for _, name := range []string{"implements", "documents"} {
		idx := slices.IndexFunc(s.Fields, func(f FieldSchema) bool { return f.Name == name })
		if idx < 0 {
			t.Fatalf("missing field %q", name)
		}
		if s.Fields[idx].RefKind != "" {
			t.Errorf("field %q: RefKind = %q, want empty", name, s.Fields[idx].RefKind)
		}
	}
}

func TestDefaultCoreSchemas_SpecificationFields(t *testing.T) {
	s := DefaultCoreSchemas()["specification"]
	if s.Kind != "specification" {
		t.Errorf("Kind = %q, want specification", s.Kind)
	}
	if s.Directory != "specifications" {
		t.Errorf("Directory = %q, want specifications", s.Directory)
	}

	wantLinks := []string{"satisfies", "addresses"}
	for _, name := range wantLinks {
		idx := slices.IndexFunc(s.Fields, func(f FieldSchema) bool { return f.Name == name })
		if idx < 0 {
			t.Errorf("specification schema missing link field %q", name)
			continue
		}
		if !s.Fields[idx].Link {
			t.Errorf("field %q: Link = false, want true", name)
		}
	}
}

func TestDefaultCoreSchemas_SpecificationRefKinds(t *testing.T) {
	s := DefaultCoreSchemas()["specification"]
	idx := slices.IndexFunc(s.Fields, func(f FieldSchema) bool { return f.Name == "satisfies" })
	if idx < 0 {
		t.Fatal("missing field satisfies")
	}
	if s.Fields[idx].RefKind != "need" {
		t.Errorf("satisfies.RefKind = %q, want need", s.Fields[idx].RefKind)
	}
}

func TestLinkFieldNames_Contract(t *testing.T) {
	s := DefaultCoreSchemas()["contract"]
	got := s.LinkFieldNames()
	want := []string{"justifies", "implements", "documents", "sprint", "batch", "parent", "depends_on"}
	if len(got) != len(want) {
		t.Fatalf("LinkFieldNames() returned %d names, want %d: %v", len(got), len(want), got)
	}
	for i, name := range want {
		if got[i] != name {
			t.Errorf("LinkFieldNames()[%d] = %q, want %q", i, got[i], name)
		}
	}
}

func TestLinkFieldNames_Specification(t *testing.T) {
	s := DefaultCoreSchemas()["specification"]
	got := s.LinkFieldNames()
	want := []string{"satisfies", "addresses"}
	if len(got) != len(want) {
		t.Fatalf("LinkFieldNames() returned %d names, want %d: %v", len(got), len(want), got)
	}
	for i, name := range want {
		if got[i] != name {
			t.Errorf("LinkFieldNames()[%d] = %q, want %q", i, got[i], name)
		}
	}
}

func TestLinkFieldNames_NoLinks(t *testing.T) {
	s := &ArtifactSchema{
		Kind: "custom",
		Fields: []FieldSchema{
			{Name: "title", Required: true},
			{Name: "status", Enum: []string{"draft", "active"}},
		},
	}
	got := s.LinkFieldNames()
	if len(got) != 0 {
		t.Errorf("LinkFieldNames() = %v, want empty", got)
	}
}

func TestFieldSchema_StructFields(t *testing.T) {
	f := FieldSchema{
		Name:     "status",
		Required: true,
		Enum:     []string{"draft", "active", "complete"},
		Default:  "draft",
		Ordered:  true,
		Transitions: []TransitionDef{
			{From: "draft", To: "active", VerifiedBy: "harness"},
		},
		Link:    false,
		RefKind: "",
	}
	if f.Name != "status" {
		t.Errorf("Name = %q", f.Name)
	}
	if !f.Required {
		t.Error("Required should be true")
	}
	if len(f.Enum) != 3 {
		t.Errorf("Enum length = %d, want 3", len(f.Enum))
	}
	if f.Default != "draft" {
		t.Errorf("Default = %q", f.Default)
	}
	if !f.Ordered {
		t.Error("Ordered should be true")
	}
	if len(f.Transitions) != 1 {
		t.Fatalf("Transitions length = %d, want 1", len(f.Transitions))
	}
	tr := f.Transitions[0]
	if tr.From != "draft" || tr.To != "active" || tr.VerifiedBy != "harness" {
		t.Errorf("Transition = %+v", tr)
	}
	if f.Link {
		t.Error("Link should be false")
	}
}

func TestExpectsDownstream_StructFields(t *testing.T) {
	s := &ArtifactSchema{
		Kind: "need",
		ExpectsDownstream: &ExpectsDownstream{
			Via:      "satisfies",
			After:    "active",
			Severity: "warn",
		},
	}
	ed := s.ExpectsDownstream
	if ed.Via != "satisfies" {
		t.Errorf("Via = %q", ed.Via)
	}
	if ed.After != "active" {
		t.Errorf("After = %q", ed.After)
	}
	if ed.Severity != "warn" {
		t.Errorf("Severity = %q", ed.Severity)
	}
}

func TestArtifactSchema_ActiveAndArchiveStates(t *testing.T) {
	s := &ArtifactSchema{
		Kind:          "contract",
		ActiveStates:  []string{"draft", "active"},
		ArchiveStates: []string{"complete", "cancelled"},
	}
	if !slices.Contains(s.ActiveStates, "draft") {
		t.Error("ActiveStates missing draft")
	}
	if !slices.Contains(s.ArchiveStates, "complete") {
		t.Error("ArchiveStates missing complete")
	}
}

func TestArtifactSchema_UrgencyPropagation(t *testing.T) {
	s := &ArtifactSchema{
		Kind:               "contract",
		UrgencyPropagation: map[string]string{"high": "sprint", "low": "batch"},
	}
	if s.UrgencyPropagation["high"] != "sprint" {
		t.Errorf("UrgencyPropagation[high] = %q", s.UrgencyPropagation["high"])
	}
}
