package artifact

import (
	"strings"
	"testing"
)

func TestLexiconAddAndList(t *testing.T) {
	root := setupScaffold(t)

	if err := AddTerm(root, "pillar", "A test category"); err != nil {
		t.Fatalf("AddTerm failed: %v", err)
	}
	if err := AddTerm(root, "component", "A product component"); err != nil {
		t.Fatalf("AddTerm failed: %v", err)
	}

	terms, err := ListTerms(root)
	if err != nil {
		t.Fatalf("ListTerms failed: %v", err)
	}
	if len(terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(terms))
	}
	if terms[0].Key != "component" || terms[0].Description != "A product component" {
		t.Errorf("unexpected first term: %+v", terms[0])
	}
	if terms[1].Key != "pillar" || terms[1].Description != "A test category" {
		t.Errorf("unexpected second term: %+v", terms[1])
	}
}

func TestLexiconAddUpdate(t *testing.T) {
	root := setupScaffold(t)

	AddTerm(root, "pillar", "original")
	if err := AddTerm(root, "pillar", "updated description"); err != nil {
		t.Fatalf("AddTerm update failed: %v", err)
	}

	terms, err := ListTerms(root)
	if err != nil {
		t.Fatalf("ListTerms failed: %v", err)
	}
	if len(terms) != 1 {
		t.Fatalf("expected 1 term after update, got %d", len(terms))
	}
	if terms[0].Description != "updated description" {
		t.Errorf("expected updated description, got %q", terms[0].Description)
	}
}

func TestLexiconRemove(t *testing.T) {
	root := setupScaffold(t)

	AddTerm(root, "pillar", "A test category")
	AddTerm(root, "component", "A product component")

	if err := RemoveTerm(root, "pillar"); err != nil {
		t.Fatalf("RemoveTerm failed: %v", err)
	}

	terms, err := ListTerms(root)
	if err != nil {
		t.Fatalf("ListTerms failed: %v", err)
	}
	if len(terms) != 1 {
		t.Fatalf("expected 1 term after remove, got %d", len(terms))
	}
	if terms[0].Key != "component" {
		t.Errorf("expected component to remain, got %q", terms[0].Key)
	}
}

func TestLexiconRemoveNotFound(t *testing.T) {
	root := setupScaffold(t)

	err := RemoveTerm(root, "nonexistent")
	if err == nil {
		t.Fatal("expected error removing nonexistent term, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestLexiconListEmpty(t *testing.T) {
	root := setupScaffold(t)

	terms, err := ListTerms(root)
	if err != nil {
		t.Fatalf("ListTerms failed: %v", err)
	}
	if len(terms) != 0 {
		t.Errorf("expected 0 terms from empty lexicon, got %d", len(terms))
	}
}
