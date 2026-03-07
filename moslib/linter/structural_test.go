package linter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunStructuralChecks_Integration(t *testing.T) {
	dir := t.TempDir()
	mosDir := filepath.Join(dir, ".mos", "rules", "mechanical")
	if err := os.MkdirAll(mosDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a minimal rule with nesting_depth config
	ruleContent := `rule "flatten-nesting" {
  config {
    metric = "nesting_depth"
    ceiling = 2
    language "go" { ceiling = 2 }
  }
}
`
	if err := os.WriteFile(filepath.Join(mosDir, "flatten-nesting.mos"), []byte(ruleContent), 0644); err != nil {
		t.Fatal(err)
	}
	// Create a Go file with depth 3 (exceeds ceiling 2)
	goFile := filepath.Join(dir, "pkg.go")
	goCode := `package pkg
func Deep() {
	if true {
		if true {
			if true {
				return
			}
		}
	}
}
`
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		t.Fatal(err)
	}

	diags := RunStructuralChecks(dir, filepath.Join(dir, ".mos"))
	var nestingDiags []Diagnostic
	for _, d := range diags {
		if d.Rule == "flatten-nesting" {
			nestingDiags = append(nestingDiags, d)
		}
	}
	if len(nestingDiags) == 0 {
		t.Fatal("expected at least one flatten-nesting diagnostic, got none")
	}
	found := false
	for _, d := range nestingDiags {
		if d.Line == 2 && d.File == "pkg.go" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected diagnostic for pkg.go Deep, got %v", nestingDiags)
	}
}

func TestAnalyzeNestingDepth_WithinCeiling(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "pkg.go")
	// Depth 2: if inside if
	code := `package pkg

func Foo() {
	if true {
		if false {
			return
		}
	}
}
`
	if err := os.WriteFile(f, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeNestingDepth(dir, 3, "**/*.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Errorf("expected no violations for depth 2 with ceiling 3, got %d: %v", len(violations), violations)
	}
}

func TestAnalyzeNestingDepth_ExceedsCeiling(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "pkg.go")
	code := `package pkg

func Bar() {
	if true {
		if true {
			if true {
				if true {
					if true {
						return
					}
				}
			}
		}
	}
}
`
	if err := os.WriteFile(f, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeNestingDepth(dir, 3, "**/*.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation for depth 5 with ceiling 3, got %d: %v", len(violations), violations)
	}
	v := violations[0]
	if v.Depth != 5 {
		t.Errorf("violation Depth = %d, want 5", v.Depth)
	}
	if v.Ceiling != 3 {
		t.Errorf("violation Ceiling = %d, want 3", v.Ceiling)
	}
	if v.Func != "Bar" {
		t.Errorf("violation Func = %q, want Bar", v.Func)
	}
}

func TestAnalyzeParamsPerFunction_Violation(t *testing.T) {
	dir := t.TempDir()
	code := `package pkg

func TooManyParams(a, b, c int, d string, e bool, f float64) {}
func OkParams(a int, b string) {}
`
	if err := os.WriteFile(filepath.Join(dir, "pkg.go"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeParamsPerFunction(dir, 5, "**/*.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Count != 6 {
		t.Errorf("expected count=6, got %d", violations[0].Count)
	}
	if violations[0].Func != "TooManyParams" {
		t.Errorf("expected func=TooManyParams, got %s", violations[0].Func)
	}
}

func TestAnalyzeParamsPerFunction_NoViolation(t *testing.T) {
	dir := t.TempDir()
	code := `package pkg

func Ok(a int, b string) {}
`
	if err := os.WriteFile(filepath.Join(dir, "pkg.go"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeParamsPerFunction(dir, 5, "**/*.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d", len(violations))
	}
}

func TestAnalyzeFanOut_Violation(t *testing.T) {
	dir := t.TempDir()
	code := `package pkg

import (
	"fmt"
	"strings"
	"os"
)

func Use() {
	fmt.Println(strings.Join(os.Args, " "))
}
`
	if err := os.WriteFile(filepath.Join(dir, "pkg.go"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeFanOut(dir, 2, "**/*.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Count != 3 {
		t.Errorf("expected count=3, got %d", violations[0].Count)
	}
}

func TestAnalyzeFanOut_NoViolation(t *testing.T) {
	dir := t.TempDir()
	code := `package pkg

import "fmt"

func Use() { fmt.Println("hi") }
`
	if err := os.WriteFile(filepath.Join(dir, "pkg.go"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeFanOut(dir, 5, "**/*.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d", len(violations))
	}
}

func TestAnalyzeLOCPerFile_Violation(t *testing.T) {
	dir := t.TempDir()
	code := "package pkg\n\nfunc A() {}\nfunc B() {}\nfunc C() {}\nfunc D() {}\nfunc E() {}\n"
	if err := os.WriteFile(filepath.Join(dir, "pkg.go"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeLOCPerFile(dir, 3, "**/*.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
}

func TestAnalyzeLOCPerFile_NoViolation(t *testing.T) {
	dir := t.TempDir()
	code := "package pkg\n\nfunc A() {}\n"
	if err := os.WriteFile(filepath.Join(dir, "pkg.go"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	violations, err := AnalyzeLOCPerFile(dir, 500, "**/*.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d", len(violations))
	}
}

func TestCountLOC(t *testing.T) {
	source := `package pkg

// This is a comment
/* Block comment
   spans multiple lines */

func A() {}
func B() {}
`
	got := countLOC(source)
	if got != 3 {
		t.Errorf("expected 3 LOC, got %d", got)
	}
}
