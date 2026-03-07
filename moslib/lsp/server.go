package lsp

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpopsuev/mos/moslib/linter"
	"github.com/dpopsuev/mos/moslib/names"
)

// Server is the Mos LSP server for .mos DSL files.
type Server struct {
	transport *Transport
	docs      *DocumentStore
	ctx       *linter.ProjectContext
	rootPath  string
	shutdown  bool
}

// NewServer creates a new LSP server over the given reader/writer.
func NewServer(r io.Reader, w io.Writer) *Server {
	return &Server{
		transport: NewTransport(r, w),
		docs:      NewDocumentStore(),
	}
}

// Run reads messages until exit or error.
func (s *Server) Run() error {
	for {
		msg, err := s.transport.Read()
		if err != nil {
			if s.shutdown {
				return nil
			}
			return err
		}

		if err := s.dispatch(msg); err != nil {
			if err == errExit {
				return nil
			}
			return err
		}
	}
}

var errExit = io.EOF

func (s *Server) dispatch(msg *RequestMessage) error {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "initialized":
		return nil
	case "shutdown":
		s.shutdown = true
		return s.respond(msg, nil)
	case "exit":
		if s.shutdown {
			os.Exit(0)
		}
		os.Exit(1)
		return errExit
	case "textDocument/didOpen":
		return s.handleDidOpen(msg)
	case "textDocument/didChange":
		return s.handleDidChange(msg)
	case "textDocument/didClose":
		return s.handleDidClose(msg)
	case "textDocument/completion":
		return s.handleCompletion(msg)
	case "textDocument/hover":
		return s.handleHover(msg)
	case "textDocument/definition":
		return s.handleDefinition(msg)
	default:
		if msg.ID != nil {
			return s.respondError(msg, -32601, "method not found: "+msg.Method)
		}
		return nil
	}
}

func (s *Server) handleInitialize(msg *RequestMessage) error {
	var params struct {
		RootURI string `json:"rootUri"`
	}
	if msg.Params != nil {
		_ = json.Unmarshal(msg.Params, &params)
	}

	if params.RootURI != "" {
		s.rootPath = URIToPath(params.RootURI)
	}

	s.loadProjectContext()

	result := map[string]any{
		"capabilities": map[string]any{
			"textDocumentSync": map[string]any{
				"openClose": true,
				"change":    1, // Full sync
			},
			"completionProvider": map[string]any{
				"triggerCharacters": []string{"[", ".", "="},
			},
			"hoverProvider":      true,
			"definitionProvider": true,
		},
	}

	return s.respond(msg, result)
}

func (s *Server) loadProjectContext() {
	if s.rootPath == "" {
		return
	}
	mosDir := filepath.Join(s.rootPath, names.MosDir)
	ctx, err := linter.LoadContext(mosDir)
	if err == nil {
		s.ctx = ctx
	}
}

type textDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type didOpenParams struct {
	TextDocument textDocumentItem `json:"textDocument"`
}

func (s *Server) handleDidOpen(msg *RequestMessage) error {
	var params didOpenParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil
	}

	s.docs.Open(params.TextDocument.URI, params.TextDocument.Text)
	s.publishDiagnostics(params.TextDocument.URI)
	return nil
}

type didChangeParams struct {
	TextDocument struct {
		URI     string `json:"uri"`
		Version int    `json:"version"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
}

func (s *Server) handleDidChange(msg *RequestMessage) error {
	var params didChangeParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil
	}

	if len(params.ContentChanges) > 0 {
		s.docs.Update(params.TextDocument.URI, params.ContentChanges[len(params.ContentChanges)-1].Text)
	}

	s.publishDiagnostics(params.TextDocument.URI)
	return nil
}

type didCloseParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
}

func (s *Server) handleDidClose(msg *RequestMessage) error {
	var params didCloseParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil
	}

	s.docs.Close(params.TextDocument.URI)

	return s.transport.WriteNotification(&NotificationMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/publishDiagnostics",
		Params: map[string]any{
			"uri":         params.TextDocument.URI,
			"diagnostics": []any{},
		},
	})
}

func (s *Server) publishDiagnostics(uri string) {
	filePath := URIToPath(uri)

	mosDir := ""
	if s.rootPath != "" {
		mosDir = filepath.Join(s.rootPath, names.MosDir)
	} else {
		mosDir = findMosDir(filePath)
	}

	if mosDir == "" {
		return
	}

	l := &linter.Linter{}
	diags, err := l.Lint(filepath.Dir(mosDir))
	if err != nil {
		return
	}

	var lspDiags []map[string]any
	for _, d := range diags {
		if !strings.HasSuffix(d.File, filepath.Base(filePath)) && d.File != filePath {
			normDiag := filepath.Clean(d.File)
			normFile := filepath.Clean(filePath)
			if normDiag != normFile {
				continue
			}
		}

		var severity int
		switch d.Severity {
		case linter.SeverityWarning:
			severity = 2
		case linter.SeverityInfo:
			severity = 3
		default:
			severity = 1 // error
		}

		lspDiags = append(lspDiags, map[string]any{
			"range": map[string]any{
				"start": map[string]int{"line": max(0, d.Line-1), "character": 0},
				"end":   map[string]int{"line": max(0, d.Line-1), "character": 0},
			},
			"severity": severity,
			"source":   "mos-lint",
			"code":     d.Rule,
			"message":  d.Message,
		})
	}

	if lspDiags == nil {
		lspDiags = []map[string]any{}
	}

	_ = s.transport.WriteNotification(&NotificationMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/publishDiagnostics",
		Params: map[string]any{
			"uri":         uri,
			"diagnostics": lspDiags,
		},
	})
}

type textDocumentPositionParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position struct {
		Line      int `json:"line"`
		Character int `json:"character"`
	} `json:"position"`
}

func (s *Server) handleCompletion(msg *RequestMessage) error {
	var params textDocumentPositionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.respond(msg, nil)
	}

	content, ok := s.docs.Get(params.TextDocument.URI)
	if !ok {
		return s.respond(msg, nil)
	}

	path := URIToPath(params.TextDocument.URI)
	items := Complete(path, content, params.Position.Line, params.Position.Character)

	return s.respond(msg, map[string]any{
		"isIncomplete": false,
		"items":        items,
	})
}

func (s *Server) handleHover(msg *RequestMessage) error {
	var params textDocumentPositionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.respond(msg, nil)
	}

	content, ok := s.docs.Get(params.TextDocument.URI)
	if !ok {
		return s.respond(msg, nil)
	}

	path := URIToPath(params.TextDocument.URI)
	result := Hover(path, content, params.Position.Line, params.Position.Character, s.ctx)
	if result == nil {
		return s.respond(msg, nil)
	}

	return s.respond(msg, map[string]any{
		"contents": result.Contents,
	})
}

func (s *Server) handleDefinition(msg *RequestMessage) error {
	var params textDocumentPositionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.respond(msg, nil)
	}

	content, ok := s.docs.Get(params.TextDocument.URI)
	if !ok {
		return s.respond(msg, nil)
	}

	path := URIToPath(params.TextDocument.URI)
	loc := Definition(path, content, params.Position.Line, params.Position.Character, s.ctx)
	if loc == nil {
		return s.respond(msg, nil)
	}

	return s.respond(msg, loc)
}

func (s *Server) respond(msg *RequestMessage, result any) error {
	return s.transport.WriteResponse(&ResponseMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	})
}

func (s *Server) respondError(msg *RequestMessage, code int, message string) error {
	return s.transport.WriteResponse(&ResponseMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Error:   &ResponseError{Code: code, Message: message},
	})
}

func findMosDir(filePath string) string {
	dir := filepath.Dir(filePath)
	for {
		if filepath.Base(dir) == names.MosDir {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
