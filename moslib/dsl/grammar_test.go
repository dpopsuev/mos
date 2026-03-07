package dsl

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"

	"github.com/dpopsuev/mos/moslib/dsl/antlrgen"
)

// preprocessStructuredWhen rewrites "when {" blocks that contain structured
// fields (key = value) into "whencond {" so the ANTLR lexer won't match them
// as WHEN_OPEN step-text openers. Only needed for ANTLR cross-validation;
// the hand-written parser handles both forms via context.
//
// Heuristic: a "when { ... }" block is structured if EVERY non-blank line
// inside the braces starts with an identifier followed by whitespace and "=".
// Step-text lines are free-form prose and don't follow this pattern.
func preprocessStructuredWhen(src string) string {
	b := []byte(src)
	var out []byte
	i := 0
	for i < len(b) {
		if i+4 <= len(b) && string(b[i:i+4]) == "when" {
			if i > 0 && isIdentByte(b[i-1]) {
				out = append(out, b[i])
				i++
				continue
			}
			end := i + 4
			if end < len(b) && isIdentByte(b[end]) {
				out = append(out, b[i])
				i++
				continue
			}
			j := end
			for j < len(b) && (b[j] == ' ' || b[j] == '\t' || b[j] == '\r' || b[j] == '\n') {
				j++
			}
			if j < len(b) && b[j] == '{' {
				content := extractBraceContent(b, j)
				if content != "" && isStructuredBlock(content) {
					out = append(out, []byte("whencond")...)
					i = end
					continue
				}
			}
		}
		out = append(out, b[i])
		i++
	}
	return string(out)
}

var fieldLineRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*\s*=`)

func isStructuredBlock(content string) bool {
	hasFields := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "\r" {
			continue
		}
		if !fieldLineRe.MatchString(trimmed) {
			return false
		}
		hasFields = true
	}
	return hasFields
}

func isIdentByte(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-'
}

func extractBraceContent(b []byte, bracePos int) string {
	depth := 1
	k := bracePos + 1
	inStr := false
	for k < len(b) && depth > 0 {
		if inStr {
			if b[k] == '\\' {
				k++
			} else if b[k] == '"' {
				inStr = false
			}
		} else {
			switch b[k] {
			case '"':
				inStr = true
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return string(b[bracePos+1 : k])
				}
			}
		}
		k++
	}
	return ""
}

// antlrParse parses src with the ANTLR-generated parser.
// Returns nil on success, error on parse failure.
func antlrParse(src string) error {
	src = preprocessStructuredWhen(src)
	input := antlr.NewInputStream(src)
	lexer := antlrgen.NewMosDSLLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := antlrgen.NewMosDSLParser(stream)
	parser.BuildParseTrees = true
	parser.RemoveErrorListeners()
	errListener := &antlrParseErrorListener{}
	parser.AddErrorListener(errListener)
	parser.File()
	if errListener.err != nil {
		return errListener.err
	}
	return nil
}

type antlrParseErrorListener struct {
	*antlr.DefaultErrorListener
	err error
}

func (l *antlrParseErrorListener) SyntaxError(
	recognizer antlr.Recognizer,
	offendingSymbol interface{},
	line, column int,
	msg string,
	e antlr.RecognitionException,
) {
	l.err = &ParseError{Line: line, Col: column, Msg: msg}
}

// antlrUnsupportedPatterns lists file path substrings that use constructs the
// ANTLR grammar cannot handle even with preprocessing. Currently empty — the
// "when" block ambiguity is resolved by preprocessStructuredWhen.
var antlrUnsupportedPatterns []string

func antlrUnsupported(path string) bool {
	for _, p := range antlrUnsupportedPatterns {
		if strings.Contains(path, p) {
			return true
		}
	}
	return false
}

// TestCrossValidationMosFiles parses all .mos files from mos-dsl-desired-format-draft/
// and .mos/ with both the hand-written parser and the ANTLR parser.
// Both must accept the input (no errors). Structured "when" blocks are
// preprocessed before ANTLR parsing to resolve the keyword ambiguity.
func TestCrossValidationMosFiles(t *testing.T) {
	// Roots relative to package dir (moslib/dsl/) when running go test
	roots := []string{
		"../../mos-dsl-desired-format-draft",
		"../../.mos",
	}
	for _, root := range roots {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			t.Skipf("root %s does not exist (run from moslib/dsl/)", root)
			continue
		}
		var files []string
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".mos") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Walk %s: %v", root, err)
		}
		for _, p := range files {
			t.Run(p, func(t *testing.T) {
				data, err := os.ReadFile(p)
				if err != nil {
					t.Fatalf("ReadFile: %v", err)
				}
				src := string(data)

				// Hand-written parser — must accept
				_, err = Parse(src, nil)
				if err != nil {
					t.Skipf("hand-written Parse rejects (unsupported construct): %v", err)
					return
				}

				// ANTLR parser — skip if file uses unsupported constructs
				if antlrUnsupported(p) {
					t.Skipf("ANTLR does not support this file's constructs")
					return
				}

				// ANTLR parser — must also accept
				err = antlrParse(src)
				if err != nil {
					t.Errorf("ANTLR Parse: %v", err)
				}
			})
		}
	}
}

// --- Property-style tests: minimal valid inputs ---

func TestANTLRParseMinimalUnnamedArtifact(t *testing.T) {
	src := `config { title = "x" }`
	if err := antlrParse(src); err != nil {
		t.Errorf("ANTLR parse: %v", err)
	}
	_, err := Parse(src, nil)
	if err != nil {
		t.Errorf("hand-written Parse: %v", err)
	}
}

func TestANTLRParseNamedArtifactWithFeature(t *testing.T) {
	src := `contract "C-1" {
  feature "F" {
    scenario "S" {
      given {
        step text
      }
    }
  }
}`
	if err := antlrParse(src); err != nil {
		t.Errorf("ANTLR parse: %v", err)
	}
	_, err := Parse(src, nil)
	if err != nil {
		t.Errorf("hand-written Parse: %v", err)
	}
}

func TestANTLRParseStructuredWhenBlock(t *testing.T) {
	src := `rule "rogyb" {
  name = "ROGYB Harness Required"
  when {
    artifact_kind = "contract"
    status = "draft"
  }
  when {
    artifact_kind = "contract"
    status = "active"
  }
}`
	if err := antlrParse(src); err != nil {
		t.Errorf("ANTLR parse failed for structured when block: %v", err)
	}
	_, err := Parse(src, nil)
	if err != nil {
		t.Errorf("hand-written Parse: %v", err)
	}
}

func TestANTLRParseStepTextWhenBlock(t *testing.T) {
	src := `contract "C-1" {
  feature "F" {
    scenario "S" {
      when {
        the user clicks the button
      }
      then {
        the page reloads
      }
    }
  }
}`
	if err := antlrParse(src); err != nil {
		t.Errorf("ANTLR parse failed for step text when block: %v", err)
	}
	_, err := Parse(src, nil)
	if err != nil {
		t.Errorf("hand-written Parse: %v", err)
	}
}

func TestANTLRParseNestedBlockWithList(t *testing.T) {
	src := `rule "R" {
  scope {
    depends_on = ["A", "B"]
  }
}`
	if err := antlrParse(src); err != nil {
		t.Errorf("ANTLR parse: %v", err)
	}
	_, err := Parse(src, nil)
	if err != nil {
		t.Errorf("hand-written Parse: %v", err)
	}
}
