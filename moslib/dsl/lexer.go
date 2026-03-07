package dsl

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Lexer tokenizes Mos DSL source text.
// In v3 all keywords are recognized in normal mode. Inside given/when/then
// blocks the lexer switches to stepTextMode where non-empty lines that do not
// start with '}' are emitted as TokenStepText.
type Lexer struct {
	src          string
	pos          int
	line         int
	col          int
	kw           *KeywordMap
	stepTextMode bool
	braceDepth   int // tracks brace nesting while in stepTextMode
}

// NewLexer creates a new streaming lexer for the given source.
// If kw is nil, English defaults are used.
func NewLexer(src string, kw *KeywordMap) *Lexer {
	if kw == nil {
		kw = DefaultKeywords()
	}
	return &Lexer{src: src, line: 1, col: 1, kw: kw}
}

// EnterStepTextMode switches the lexer so that free-text lines inside
// given/when/then blocks are emitted as TokenStepText. The parser calls
// this after consuming the '{' that opens a step block.
func (l *Lexer) EnterStepTextMode() { l.stepTextMode = true; l.braceDepth = 1 }

// Next returns the next token from the source.
func (l *Lexer) Next() (Token, error) {
	if l.stepTextMode {
		return l.nextStepText()
	}
	return l.nextNormal()
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.src[l.pos:])
	return r
}

func (l *Lexer) advance() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	r, size := utf8.DecodeRuneInString(l.src[l.pos:])
	l.pos += size
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func (l *Lexer) tok(typ TokenType, val string, line, col, offset int) Token {
	return Token{Type: typ, Value: val, Line: line, Col: col, Offset: offset}
}

func (l *Lexer) errorf(msg string) error {
	return &ParseError{Line: l.line, Col: l.col, Msg: msg}
}

func (l *Lexer) skipSpaces() {
	for l.pos < len(l.src) {
		r := l.peek()
		if r == ' ' || r == '\t' || r == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

// --- Normal mode ---

func (l *Lexer) nextNormal() (Token, error) {
	l.skipSpaces()

	if l.pos >= len(l.src) {
		return l.tok(TokenEOF, "", l.line, l.col, l.pos), nil
	}

	sLine, sCol, sOff := l.line, l.col, l.pos
	ch := l.peek()

	switch {
	case ch == '\n':
		l.advance()
		return l.tok(TokenNewline, "\n", sLine, sCol, sOff), nil
	case ch == '#':
		return l.scanComment(sLine, sCol, sOff), nil
	case ch == '{':
		l.advance()
		return l.tok(TokenLBrace, "{", sLine, sCol, sOff), nil
	case ch == '}':
		l.advance()
		return l.tok(TokenRBrace, "}", sLine, sCol, sOff), nil
	case ch == '[':
		l.advance()
		return l.tok(TokenLBracket, "[", sLine, sCol, sOff), nil
	case ch == ']':
		l.advance()
		return l.tok(TokenRBracket, "]", sLine, sCol, sOff), nil
	case ch == '=':
		l.advance()
		return l.tok(TokenEquals, "=", sLine, sCol, sOff), nil
	case ch == ',':
		l.advance()
		return l.tok(TokenComma, ",", sLine, sCol, sOff), nil
	case ch == '"':
		return l.scanString(sLine, sCol, sOff)
	case ch == '-' || (ch >= '0' && ch <= '9'):
		return l.scanNumberOrDateTime(sLine, sCol, sOff)
	case isIdentStart(ch):
		return l.scanIdentOrKeyword(sLine, sCol, sOff), nil
	default:
		return Token{}, l.errorf("unexpected character: " + string(ch))
	}
}

func (l *Lexer) scanComment(line, col, off int) Token {
	l.advance() // #
	start := l.pos
	for l.pos < len(l.src) && l.peek() != '\n' {
		l.advance()
	}
	return l.tok(TokenComment, l.src[start:l.pos], line, col, off)
}

func (l *Lexer) scanString(line, col, off int) (Token, error) {
	l.advance() // opening "

	if l.pos+1 < len(l.src) && l.src[l.pos] == '"' && l.src[l.pos+1] == '"' {
		l.advance()
		l.advance()
		return l.scanTripleString(line, col, off)
	}

	var b strings.Builder
	for {
		if l.pos >= len(l.src) {
			return Token{}, l.errorf("unterminated string")
		}
		ch := l.advance()
		if ch == '"' {
			return l.tok(TokenString, b.String(), line, col, off), nil
		}
		if ch == '\\' {
			esc, err := l.scanEscape()
			if err != nil {
				return Token{}, err
			}
			b.WriteRune(esc)
		} else {
			b.WriteRune(ch)
		}
	}
}

func (l *Lexer) scanTripleString(line, col, off int) (Token, error) {
	var b strings.Builder
	for {
		if l.pos >= len(l.src) {
			return Token{}, l.errorf("unterminated triple-quoted string")
		}
		if l.pos+2 < len(l.src) &&
			l.src[l.pos] == '"' && l.src[l.pos+1] == '"' && l.src[l.pos+2] == '"' {
			l.advance()
			l.advance()
			l.advance()
			return l.tok(TokenString, b.String(), line, col, off), nil
		}
		b.WriteRune(l.advance())
	}
}

func (l *Lexer) scanEscape() (rune, error) {
	if l.pos >= len(l.src) {
		return 0, l.errorf("unterminated escape sequence")
	}
	ch := l.advance()
	switch ch {
	case 'n':
		return '\n', nil
	case 't':
		return '\t', nil
	case 'r':
		return '\r', nil
	case '"':
		return '"', nil
	case '\\':
		return '\\', nil
	default:
		return 0, l.errorf("unknown escape: \\" + string(ch))
	}
}

func (l *Lexer) scanNumberOrDateTime(line, col, off int) (Token, error) {
	start := l.pos
	if l.peek() == '-' {
		l.advance()
	}

	if l.pos >= len(l.src) || l.peek() < '0' || l.peek() > '9' {
		return Token{}, l.errorf("expected digit after -")
	}

	l.consumeDigits()

	if l.pos < len(l.src) && l.peek() == '-' {
		return l.tryScanDateTime(start, line, col, off)
	}

	if l.pos < len(l.src) && l.peek() == '.' {
		l.advance()
		l.consumeDigits()
		return l.tok(TokenFloat, l.src[start:l.pos], line, col, off), nil
	}

	return l.tok(TokenInteger, l.src[start:l.pos], line, col, off), nil
}

func (l *Lexer) tryScanDateTime(start, line, col, off int) (Token, error) {
	saved, savedLine, savedCol := l.pos, l.line, l.col

	l.advance() // first -
	if !l.isDigit() {
		l.pos, l.line, l.col = saved, savedLine, savedCol
		return l.tok(TokenInteger, l.src[start:saved], line, col, off), nil
	}
	l.consumeDigits() // MM

	if l.pos >= len(l.src) || l.peek() != '-' {
		l.pos, l.line, l.col = saved, savedLine, savedCol
		return l.tok(TokenInteger, l.src[start:saved], line, col, off), nil
	}
	l.advance() // second -
	l.consumeDigits() // DD

	if l.pos >= len(l.src) || l.peek() != 'T' {
		l.pos, l.line, l.col = saved, savedLine, savedCol
		return l.tok(TokenInteger, l.src[start:saved], line, col, off), nil
	}
	l.advance() // T
	l.consumeDigits()
	if l.pos < len(l.src) && l.peek() == ':' {
		l.advance()
		l.consumeDigits()
	}
	if l.pos < len(l.src) && l.peek() == ':' {
		l.advance()
		l.consumeDigits()
	}
	if l.pos < len(l.src) && l.peek() == '.' {
		l.advance()
		l.consumeDigits()
	}
	if l.pos < len(l.src) {
		if l.peek() == 'Z' {
			l.advance()
		} else if l.peek() == '+' || l.peek() == '-' {
			l.advance()
			l.consumeDigits()
			if l.pos < len(l.src) && l.peek() == ':' {
				l.advance()
				l.consumeDigits()
			}
		}
	}

	return l.tok(TokenDateTime, l.src[start:l.pos], line, col, off), nil
}

func (l *Lexer) isDigit() bool {
	return l.pos < len(l.src) && l.peek() >= '0' && l.peek() <= '9'
}

func (l *Lexer) consumeDigits() {
	for l.pos < len(l.src) && l.peek() >= '0' && l.peek() <= '9' {
		l.advance()
	}
}

// scanIdentOrKeyword reads an identifier and then translates it through the
// KeywordMap to see if the human-protocol word maps to a machine-protocol
// keyword. The token type switch operates on machine names.
func (l *Lexer) scanIdentOrKeyword(line, col, off int) Token {
	start := l.pos
	for l.pos < len(l.src) && isIdentContinue(l.peek()) {
		l.advance()
	}
	word := l.src[start:l.pos]

	machine := l.kw.machineKeyword(word)

	switch machine {
	case "true", "false":
		return l.tok(TokenBool, word, line, col, off)
	case "feature":
		return l.tok(TokenFeature, word, line, col, off)
	case "background":
		return l.tok(TokenBackground, word, line, col, off)
	case "scenario":
		return l.tok(TokenScenario, word, line, col, off)
	case "given":
		return l.tok(TokenGiven, word, line, col, off)
	case "when":
		return l.tok(TokenWhen, word, line, col, off)
	case "then":
		return l.tok(TokenThen, word, line, col, off)
	case "group":
		return l.tok(TokenGroup, word, line, col, off)
	case "spec":
		return l.tok(TokenSpec, word, line, col, off)
	case "include":
		return l.tok(TokenInclude, word, line, col, off)
	default:
		return l.tok(TokenIdent, word, line, col, off)
	}
}

func isIdentStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isIdentContinue(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-'
}

// --- Step-text mode ---
// Active inside given/when/then blocks. Emits free-text lines as TokenStepText.
// Tracks brace depth to know when the closing '}' is reached.

func (l *Lexer) nextStepText() (Token, error) {
	l.skipSpaces()

	if l.pos >= len(l.src) {
		return l.tok(TokenEOF, "", l.line, l.col, l.pos), nil
	}

	sLine, sCol, sOff := l.line, l.col, l.pos

	if l.peek() == '\n' {
		l.advance()
		return l.tok(TokenNewline, "\n", sLine, sCol, sOff), nil
	}

	if l.peek() == '#' {
		return l.scanComment(sLine, sCol, sOff), nil
	}

	if l.peek() == '}' {
		l.advance()
		l.braceDepth--
		if l.braceDepth <= 0 {
			l.stepTextMode = false
		}
		return l.tok(TokenRBrace, "}", sLine, sCol, sOff), nil
	}

	// Free-text line
	textStart := l.pos
	for l.pos < len(l.src) && l.peek() != '\n' {
		l.advance()
	}
	text := strings.TrimSpace(l.src[textStart:l.pos])
	if text != "" {
		return l.tok(TokenStepText, text, sLine, sCol, sOff), nil
	}
	return l.tok(TokenNewline, "\n", sLine, sCol, sOff), nil
}
