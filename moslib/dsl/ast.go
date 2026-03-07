package dsl

import "fmt"

// File is the root AST node representing a single .mos file.
type File struct {
	Artifact Node
	Comments []Comment
}

// Node is implemented by all AST nodes.
type Node interface {
	nodeType() string
}

// Comment holds a # comment.
type Comment struct {
	Text string
	Line int
}

// --- Artifact types ---

// ArtifactBlock represents a top-level artifact: rule "id" { ... }, config { ... }, etc.
// Kind stores the machine-protocol name (always English, e.g. "rule" even if
// the source file used "regla").
type ArtifactBlock struct {
	Kind  string
	Name  string // title string for named artifacts, empty for unnamed
	Items []Node
	Line  int
}

func (a *ArtifactBlock) nodeType() string { return "artifact" }

// --- Block content ---

// Field represents key = value.
type Field struct {
	Key   string
	Value Value
	Line  int
}

func (f *Field) nodeType() string { return "field" }

// Block represents ident { ... } or ident "title" { ... } (a nested section).
// Title is non-empty for named nested blocks like layer "project" { ... }.
type Block struct {
	Name  string
	Title string
	Items []Node
	Line  int
}

func (b *Block) nodeType() string { return "block" }

// --- Values ---

// Value is implemented by all value types.
type Value interface {
	valueType() string
}

// StringVal is a quoted string value.
type StringVal struct {
	Text   string
	Triple bool // true for triple-quoted (""") strings
}

func (s *StringVal) valueType() string { return "string" }

// IntegerVal is an integer value.
type IntegerVal struct {
	Raw string
	Val int64
}

func (i *IntegerVal) valueType() string { return "integer" }

// FloatVal is a floating-point value.
type FloatVal struct {
	Raw string
	Val float64
}

func (f *FloatVal) valueType() string { return "float" }

// BoolVal is a boolean value.
type BoolVal struct {
	Val bool
}

func (b *BoolVal) valueType() string { return "bool" }

// DateTimeVal is a datetime value.
type DateTimeVal struct {
	Raw string
}

func (d *DateTimeVal) valueType() string { return "datetime" }

// ListVal is a list of values: [a, b, c].
type ListVal struct {
	Items []Value
}

func (l *ListVal) valueType() string { return "list" }

// InlineTableVal is an inline table: { key = val, ... }.
type InlineTableVal struct {
	Fields []*Field
}

func (t *InlineTableVal) valueType() string { return "inline_table" }

// --- Spec block (v3: optional grouping for include directives and features) ---

// SpecBlock represents spec { ... } containing include directives and/or
// inline feature blocks. No longer triggers a lexer mode switch.
type SpecBlock struct {
	Includes []*IncludeDirective
	Features []*FeatureBlock
	Line     int
}

func (s *SpecBlock) nodeType() string { return "spec" }

// IncludeDirective represents include "path" inside a spec block.
type IncludeDirective struct {
	Path string
	Line int
}

// --- Feature block (v3: brace-delimited, lowercase) ---

// FeatureBlock represents feature "title" { ... }. Can appear as a direct
// BlockItem in an artifact or inside a SpecBlock.
type FeatureBlock struct {
	Name        string
	Description []string
	Background  *Background
	Groups      []ScenarioContainer
	Line        int
}

func (fb *FeatureBlock) nodeType() string { return "feature" }

// ScenarioContainer is either a standalone Scenario or a Group.
type ScenarioContainer interface {
	scenarioContainer()
}

// Group represents group "title" { ... } containing scenarios.
type Group struct {
	Name      string
	Scenarios []*Scenario
	Line      int
}

func (g *Group) scenarioContainer() {}

// Scenario represents scenario "title" { ... } with arbitrary fields
// and given/when/then step blocks. Fields are open -- the parser accepts
// any identifier = value pair; the linter validates against the lexicon.
type Scenario struct {
	Name   string
	Fields []*Field
	Given  *StepBlock
	When   *StepBlock
	Then   *StepBlock
	Line   int
}

func (s *Scenario) scenarioContainer() {}

func (s *Scenario) fieldVal(key string) string {
	for _, f := range s.Fields {
		if f.Key == key {
			if sv, ok := f.Value.(*StringVal); ok {
				return sv.Text
			}
		}
	}
	return ""
}

// SUT returns the "sut" field value, or empty string if absent.
func (s *Scenario) SUT() string { return s.fieldVal("sut") }

// Test returns the "test" field value, or empty string if absent.
func (s *Scenario) Test() string { return s.fieldVal("test") }

// Case returns the "case" field value, or empty string if absent.
func (s *Scenario) Case() string { return s.fieldVal("case") }

// Actor returns the "actor" field value, or empty string if absent.
func (s *Scenario) Actor() string { return s.fieldVal("actor") }

// Labels returns the "labels" field as a string slice, or nil if absent.
func (s *Scenario) Labels() []string {
	for _, f := range s.Fields {
		if f.Key == "labels" {
			if lv, ok := f.Value.(*ListVal); ok {
				var out []string
				for _, item := range lv.Items {
					if sv, ok := item.(*StringVal); ok {
						out = append(out, sv.Text)
					}
				}
				return out
			}
		}
	}
	return nil
}

// Background represents background { given { ... } }.
type Background struct {
	Given *StepBlock
	Line  int
}

// StepBlock holds the free-text lines inside a given/when/then block.
type StepBlock struct {
	Lines []string
	Line  int
}

// StringValue extracts the text from a Value. If v is a *StringVal, it
// returns the Text field; otherwise it falls back to fmt.Sprintf.
func StringValue(v Value) string {
	if sv, ok := v.(*StringVal); ok {
		return sv.Text
	}
	return fmt.Sprintf("%v", v)
}

// ParseError is returned for syntax errors.
type ParseError struct {
	Line     int
	Col      int
	Msg      string
	Expected []string `json:"expected,omitempty"`
	Got      string   `json:"got,omitempty"`
}

func (e *ParseError) Error() string {
	if e.Col > 0 {
		return fmt.Sprintf("line %d col %d: %s", e.Line, e.Col, e.Msg)
	}
	return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
}
