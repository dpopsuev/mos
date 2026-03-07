package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/linter"
)

// --- test helpers ---

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const validFeatureBlock = `
  feature "Test" {
    scenario "ok" {
      given {
        a thing
      }
      when {
        it runs
      }
      then {
        it works
      }
    }
  }
`

func validCstDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mos := filepath.Join(root, ".mos")

	writeFile(t, filepath.Join(mos, "config.mos"), `
config {
  mos {
    version = 1
  }
  backend {
    type = "git"
  }
}
`)
	writeFile(t, filepath.Join(mos, "declaration.mos"), `
declaration {
  name = "test"
  created = 2026-01-01T00:00:00Z
  authors = ["alice"]
}
`)
	writeFile(t, filepath.Join(mos, "lexicon", "default.mos"), `
lexicon {
  terms {
    governance = "the process of governing"
  }
}
`)

	if err := os.MkdirAll(filepath.Join(mos, "resolution"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(mos, "resolution", "layers.mos"), `
layers {
  layer "repository" {
    level = 1
  }
}
`)

	if err := os.MkdirAll(filepath.Join(mos, "rules", "mechanical"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(mos, "rules", "mechanical", "no-binaries.mos"), `
rule "R-001" {
  name = "No binaries"
  type = "mechanical"
  scope = "repository"
  enforcement = "error"
`+validFeatureBlock+`
}
`)

	return root
}

type mockClient struct {
	in  *bytes.Buffer
	out *bytes.Buffer
}

func newMockClient() *mockClient {
	return &mockClient{
		in:  &bytes.Buffer{},
		out: &bytes.Buffer{},
	}
}

func (c *mockClient) send(method string, id *int, params any) {
	msg := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if id != nil {
		msg["id"] = *id
	}
	if params != nil {
		msg["params"] = params
	}
	body, _ := json.Marshal(msg)
	fmt.Fprintf(c.in, "Content-Length: %d\r\n\r\n%s", len(body), body)
}

func (c *mockClient) readResponse() map[string]any {
	raw := c.out.String()
	idx := strings.Index(raw, "\r\n\r\n")
	if idx < 0 {
		return nil
	}

	bodies := strings.SplitAfter(raw, "\r\n\r\n")
	for i := 1; i < len(bodies); i++ {
		body := bodies[i]
		nextHeader := strings.Index(body, "Content-Length:")
		if nextHeader > 0 {
			body = body[:nextHeader]
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(body), &result); err == nil {
			return result
		}
	}
	return nil
}

func (c *mockClient) readAllResponses() []map[string]any {
	raw := c.out.String()
	var results []map[string]any

	for raw != "" {
		idx := strings.Index(raw, "\r\n\r\n")
		if idx < 0 {
			break
		}
		raw = raw[idx+4:]

		nextHeader := strings.Index(raw, "Content-Length:")
		var body string
		if nextHeader >= 0 {
			body = raw[:nextHeader]
			raw = raw[nextHeader:]
		} else {
			body = raw
			raw = ""
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(body), &result); err == nil {
			results = append(results, result)
		}
	}
	return results
}

func intPtr(v int) *int { return &v }

// --- Transport tests ---

func TestTransportRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	transport := NewTransport(strings.NewReader(""), &buf)

	err := transport.WriteResponse(&ResponseMessage{
		JSONRPC: "2.0",
		ID:      rawID(1),
		Result:  map[string]string{"test": "value"},
	})
	if err != nil {
		t.Fatalf("WriteResponse: %v", err)
	}

	output := buf.String()
	if !strings.HasPrefix(output, "Content-Length:") {
		t.Errorf("missing Content-Length header, got: %s", output)
	}
	if !strings.Contains(output, `"test":"value"`) && !strings.Contains(output, `"test": "value"`) {
		t.Errorf("missing result in output: %s", output)
	}
}

func TestTransportRead(t *testing.T) {
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	input := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)

	transport := NewTransport(strings.NewReader(input), &bytes.Buffer{})
	msg, err := transport.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if msg.Method != "initialize" {
		t.Errorf("method = %q, want initialize", msg.Method)
	}
}

// --- Document store tests ---

func TestDocumentStore(t *testing.T) {
	ds := NewDocumentStore()
	uri := "file:///test.mos"

	ds.Open(uri, "initial")
	content, ok := ds.Get(uri)
	if !ok || content != "initial" {
		t.Errorf("Get after Open = %q, %v", content, ok)
	}

	ds.Update(uri, "updated")
	content, _ = ds.Get(uri)
	if content != "updated" {
		t.Errorf("Get after Update = %q", content)
	}

	ds.Close(uri)
	_, ok = ds.Get(uri)
	if ok {
		t.Error("Get after Close should return false")
	}
}

// --- LSP Initialize test ---

func TestInitializeHandshake(t *testing.T) {
	root := validCstDir(t)
	client := newMockClient()

	client.send("initialize", intPtr(1), map[string]any{
		"rootUri": PathToURI(root),
	})
	client.send("initialized", nil, nil)
	client.send("shutdown", intPtr(2), nil)

	srv := NewServer(client.in, client.out)
	_ = srv.Run()

	responses := client.readAllResponses()
	if len(responses) < 1 {
		t.Fatal("expected at least 1 response")
	}

	initResp := responses[0]
	result, ok := initResp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result map, got %T", initResp["result"])
	}

	caps, ok := result["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("expected capabilities map, got %T", result["capabilities"])
	}

	if _, ok := caps["completionProvider"]; !ok {
		t.Error("missing completionProvider capability")
	}
	if _, ok := caps["hoverProvider"]; !ok {
		t.Error("missing hoverProvider capability")
	}
	if _, ok := caps["definitionProvider"]; !ok {
		t.Error("missing definitionProvider capability")
	}
}

// --- Diagnostics test ---

func TestDiagnosticsOnDidOpen(t *testing.T) {
	root := validCstDir(t)
	badRulePath := filepath.Join(root, ".mos", "rules", "mechanical", "bad.mos")
	writeFile(t, badRulePath, `
rule "R-BAD" {
  name = "Bad"
  type = "mechanical"
  scope = "repository"
  enforcement = "error"
}
`)

	client := newMockClient()
	client.send("initialize", intPtr(1), map[string]any{"rootUri": PathToURI(root)})
	client.send("initialized", nil, nil)
	client.send("textDocument/didOpen", nil, map[string]any{
		"textDocument": map[string]any{
			"uri":        PathToURI(badRulePath),
			"languageId": "mos",
			"version":    1,
			"text":       "rule \"R-BAD\" {\n  name = \"Bad\"\n  type = \"mechanical\"\n  scope = \"repository\"\n  enforcement = \"error\"\n}\n",
		},
	})
	client.send("shutdown", intPtr(2), nil)

	srv := NewServer(client.in, client.out)
	_ = srv.Run()

	responses := client.readAllResponses()
	foundDiag := false
	for _, resp := range responses {
		if resp["method"] == "textDocument/publishDiagnostics" {
			foundDiag = true
			params, _ := resp["params"].(map[string]any)
			diags, _ := params["diagnostics"].([]any)
			if len(diags) == 0 {
				t.Error("expected diagnostics for missing feature/spec block")
			}
			for _, d := range diags {
				dm, _ := d.(map[string]any)
				msg, _ := dm["message"].(string)
				if strings.Contains(msg, "feature") || strings.Contains(msg, "spec") {
					return
				}
			}
			t.Error("expected diagnostic about missing feature/spec")
		}
	}
	if !foundDiag {
		t.Error("no publishDiagnostics notification received")
	}
}

// --- Completion tests ---

func TestCompletionRuleTopLevelFields(t *testing.T) {
	path := "/project/.mos/rules/mechanical/test.mos"
	content := ""
	items := Complete(path, content, 0, 0)

	found := false
	for _, item := range items {
		if item.Label == "name" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected top-level rule fields in completions, got %v", items)
	}
}

func TestCompletionRuleSubBlocks(t *testing.T) {
	path := "/project/.mos/rules/mechanical/test.mos"
	content := "rule \"test\" {\n"
	items := Complete(path, content, 1, 0)

	found := false
	for _, item := range items {
		if item.Label == "harness" || item.Label == "feature" || item.Label == "spec" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected rule sub-blocks in completions, got %v", items)
	}
}

func TestCompletionConfigBlocks(t *testing.T) {
	path := "/project/.mos/config.mos"
	items := Complete(path, "", 0, 0)

	found := false
	for _, item := range items {
		if item.Label == "mos" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected mos block in completions, got %v", items)
	}
}

func TestCompletionDSLKeywords(t *testing.T) {
	path := "/project/.mos/rules/mechanical/test.mos"
	content := "rule \"test\" {\n  feature \"x\" {\n    scenario \"s\" {\n"
	items := Complete(path, content, 3, 4)

	found := false
	for _, item := range items {
		if item.Label == "given" || item.Label == "when" || item.Label == "then" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected DSL keywords in completions, got %v", items)
	}
}

// --- Hover tests ---

func TestHoverFieldDoc(t *testing.T) {
	content := "mos {\n  version = 1\n}\n"
	result := Hover("/project/.mos/config.mos", content, 1, 4, nil)
	if result == nil {
		t.Fatal("expected hover result")
	}
	if !strings.Contains(result.Contents, "Schema version") {
		t.Errorf("expected schema version doc, got %q", result.Contents)
	}
}

func TestHoverVocabularyTerm(t *testing.T) {
	ctx := &linter.ProjectContext{
		Lexicon: &linter.MergedLexicon{
			Terms: map[string]string{
				"governance": "the process of governing",
			},
		},
	}
	content := "declaration {\n  governance\n}\n"
	result := Hover("/project/.mos/declaration.mos", content, 1, 4, ctx)
	if result == nil {
		t.Fatal("expected hover result for lexicon term")
	}
	if !strings.Contains(result.Contents, "governance") {
		t.Errorf("expected lexicon definition, got %q", result.Contents)
	}
}

func TestHoverRuleID(t *testing.T) {
	ctx := &linter.ProjectContext{
		RuleIDs: map[string]string{
			"R-001": "/project/.mos/rules/mechanical/no-binaries.mos",
		},
	}
	content := "execution {\n  rules_override = [\"R-001\"]\n}\n"
	result := Hover("/project/.mos/contracts/active/CON-001/contract.mos", content, 1, 23, ctx)
	if result == nil {
		t.Fatal("expected hover result for rule ID")
	}
	if !strings.Contains(result.Contents, "R-001") {
		t.Errorf("expected rule ID info, got %q", result.Contents)
	}
}

// --- Definition tests ---

func TestDefinitionRuleID(t *testing.T) {
	ctx := &linter.ProjectContext{
		RuleIDs: map[string]string{
			"R-001": "/project/.mos/rules/mechanical/no-binaries.mos",
		},
	}
	content := "execution {\n  rules_override = [\"R-001\"]\n}\n"
	loc := Definition("/project/.mos/contracts/active/CON-001/contract.mos", content, 1, 23, ctx)
	if loc == nil {
		t.Fatal("expected definition location")
	}
	if !strings.Contains(loc.URI, "no-binaries.mos") {
		t.Errorf("expected rule file URI, got %q", loc.URI)
	}
}

func TestDefinitionInclude(t *testing.T) {
	ctx := &linter.ProjectContext{}
	content := "spec {\n  include \"spec.feature\"\n}\n"
	loc := Definition("/project/.mos/rules/mechanical/test.mos", content, 1, 12, ctx)
	if loc == nil {
		t.Fatal("expected definition location for include")
	}
	if !strings.Contains(loc.URI, "spec.feature") {
		t.Errorf("expected spec.feature URI, got %q", loc.URI)
	}
}

// --- Artifact detection tests ---

func TestDetectArtifactKind(t *testing.T) {
	tests := []struct {
		path string
		want ArtifactKind
	}{
		{"/proj/.mos/config.mos", ArtifactConfig},
		{"/proj/.mos/declaration.mos", ArtifactDeclaration},
		{"/proj/.mos/rules/mechanical/r.mos", ArtifactRule},
		{"/proj/.mos/contracts/active/CON-001/contract.mos", ArtifactContract},
		{"/proj/random.txt", ArtifactUnknown},
	}
	for _, tt := range tests {
		got := DetectArtifactKind(tt.path)
		if got != tt.want {
			t.Errorf("DetectArtifactKind(%q) = %d, want %d", tt.path, got, tt.want)
		}
	}
}

// --- URI helpers ---

func TestURIConversion(t *testing.T) {
	path := "/home/user/project/.mos/config.mos"
	uri := PathToURI(path)
	if !strings.HasPrefix(uri, "file://") {
		t.Errorf("PathToURI should produce file:// URI, got %q", uri)
	}

	back := URIToPath(uri)
	if back != path {
		t.Errorf("URIToPath(PathToURI(%q)) = %q", path, back)
	}
}

func rawID(id int) *json.RawMessage {
	raw := json.RawMessage(fmt.Sprintf("%d", id))
	return &raw
}
