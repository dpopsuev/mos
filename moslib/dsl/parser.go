package dsl

import (
	"strconv"
	"strings"
)

// Parse parses Mos DSL source text into an AST.
// If kw is nil, English defaults are used.
func Parse(src string, kw *KeywordMap) (*File, error) {
	if kw == nil {
		kw = DefaultKeywords()
	}
	p := &parser{lex: NewLexer(src, kw), kw: kw}
	if err := p.next(); err != nil {
		return nil, err
	}
	return p.parseFile()
}

type parser struct {
	lex *Lexer
	kw  *KeywordMap
	cur Token
}

func (p *parser) next() error {
	t, err := p.lex.Next()
	if err != nil {
		return err
	}
	p.cur = t
	return nil
}

func (p *parser) skipTrivia() error {
	for p.cur.Type == TokenNewline || p.cur.Type == TokenComment {
		if err := p.next(); err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) expect(typ TokenType) (Token, error) {
	if p.cur.Type != typ {
		return p.cur, &ParseError{Line: p.cur.Line, Col: p.cur.Col,
			Msg: "expected " + typ.String() + ", got " + p.cur.Type.String()}
	}
	t := p.cur
	return t, p.next()
}

func (p *parser) parseFile() (*File, error) {
	if err := p.skipTrivia(); err != nil {
		return nil, err
	}

	art, err := p.parseArtifact()
	if err != nil {
		return nil, err
	}

	if err := p.skipTrivia(); err != nil {
		return nil, err
	}
	if p.cur.Type != TokenEOF {
		return nil, &ParseError{Line: p.cur.Line, Col: p.cur.Col, Msg: "unexpected content after artifact"}
	}

	return &File{Artifact: art}, nil
}

// parseArtifact accepts any TokenIdent as artifact type (open set).
// Heuristic: if next meaningful token after keyword is TokenString -> named,
// if TokenLBrace -> unnamed.
func (p *parser) parseArtifact() (Node, error) {
	if p.cur.Type != TokenIdent {
		return nil, &ParseError{Line: p.cur.Line, Col: p.cur.Col,
			Msg: "expected artifact keyword, got " + p.cur.Type.String()}
	}

	kind := p.kw.machineKeyword(p.cur.Value)
	line := p.cur.Line

	if err := p.next(); err != nil {
		return nil, err
	}
	if err := p.skipTrivia(); err != nil {
		return nil, err
	}

	switch p.cur.Type {
	case TokenString:
		nameTok, err := p.expect(TokenString)
		if err != nil {
			return nil, err
		}
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		items, err := p.parseBlockBody()
		if err != nil {
			return nil, err
		}
		return &ArtifactBlock{Kind: kind, Name: nameTok.Value, Items: items, Line: line}, nil

	case TokenLBrace:
		items, err := p.parseBlockBody()
		if err != nil {
			return nil, err
		}
		return &ArtifactBlock{Kind: kind, Items: items, Line: line}, nil

	default:
		return nil, &ParseError{Line: p.cur.Line, Col: p.cur.Col,
			Msg:      "expected string or '{' after artifact keyword, got " + p.cur.Type.String(),
			Expected: []string{"string", "{"},
			Got:      p.cur.Type.String()}
	}
}

func (p *parser) parseBlockBody() ([]Node, error) {
	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	var items []Node
	for {
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenRBrace {
			if err := p.next(); err != nil {
				return nil, err
			}
			return items, nil
		}
		if p.cur.Type == TokenEOF {
			return nil, &ParseError{Msg: "unterminated block"}
		}

		item, err := p.parseBlockItem()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
}

// isKeywordToken returns true for tokens that represent grammar keywords but
// might also appear as field names (e.g. "feature = ..." in a lexicon
// keywords block).
func isKeywordToken(t TokenType) bool {
	switch t {
	case TokenFeature, TokenBackground, TokenScenario,
		TokenGiven, TokenWhen, TokenThen,
		TokenGroup, TokenSpec, TokenInclude:
		return true
	}
	return false
}

func (p *parser) parseBlockItem() (Node, error) {
	// Keyword tokens can be field names when followed by '=', so we
	// normalize them to an ident-like token and use lookahead. Spec and
	// feature get special treatment only if NOT followed by '='.
	if isKeywordToken(p.cur.Type) {
		ident := Token{Type: TokenIdent, Value: p.cur.Value, Line: p.cur.Line, Col: p.cur.Col, Offset: p.cur.Offset}
		savedType := p.cur.Type
		if err := p.next(); err != nil {
			return nil, err
		}
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenEquals {
			return p.parseFieldAfterIdent(ident)
		}
		// Not a field — dispatch to keyword-specific parser.
		switch savedType {
		case TokenSpec:
			// parseSpecBlock expects cur to be right after "spec", already consumed.
			return p.parseSpecBlockAfterKeyword(ident.Line)
		case TokenFeature:
			return p.parseFeatureBlockAfterKeyword(ident.Line)
		default:
			// Other keywords (given, when, etc.) used as block names.
			switch p.cur.Type {
			case TokenLBrace:
				return p.parseNestedBlockAfterIdent(ident, "")
			case TokenString:
				title := p.cur
				if err := p.next(); err != nil {
					return nil, err
				}
				if err := p.skipTrivia(); err != nil {
					return nil, err
				}
				if p.cur.Type == TokenLBrace {
					return p.parseNestedBlockAfterIdent(ident, title.Value)
				}
				return nil, &ParseError{Line: title.Line, Col: title.Col,
					Msg: "expected '{' after block title, got " + p.cur.Type.String()}
			default:
				return nil, &ParseError{Line: ident.Line, Col: ident.Col,
					Msg: "expected '=', '{', or string after " + ident.Value + ", got " + p.cur.Type.String()}
			}
		}
	}

	if p.cur.Type != TokenIdent {
		return nil, &ParseError{Line: p.cur.Line, Col: p.cur.Col,
			Msg: "expected identifier, got " + p.cur.Type.String()}
	}

	ident := p.cur
	if err := p.next(); err != nil {
		return nil, err
	}
	if err := p.skipTrivia(); err != nil {
		return nil, err
	}

	switch p.cur.Type {
	case TokenEquals:
		return p.parseFieldAfterIdent(ident)
	case TokenLBrace:
		return p.parseNestedBlockAfterIdent(ident, "")
	case TokenString:
		title := p.cur
		if err := p.next(); err != nil {
			return nil, err
		}
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenLBrace {
			return p.parseNestedBlockAfterIdent(ident, title.Value)
		}
		return nil, &ParseError{Line: title.Line, Col: title.Col,
			Msg: "expected '{' after block title, got " + p.cur.Type.String()}
	default:
		return nil, &ParseError{Line: ident.Line, Col: ident.Col,
			Msg: "expected '=', '{', or string after " + ident.Value + ", got " + p.cur.Type.String()}
	}
}

func (p *parser) parseFieldAfterIdent(ident Token) (*Field, error) {
	if err := p.next(); err != nil { // skip =
		return nil, err
	}
	if err := p.skipTrivia(); err != nil {
		return nil, err
	}

	val, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	return &Field{Key: ident.Value, Value: val, Line: ident.Line}, nil
}

func (p *parser) parseNestedBlockAfterIdent(ident Token, title string) (*Block, error) {
	items, err := p.parseBlockBody()
	if err != nil {
		return nil, err
	}
	return &Block{Name: ident.Value, Title: title, Items: items, Line: ident.Line}, nil
}

func (p *parser) parseValue() (Value, error) {
	switch p.cur.Type {
	case TokenString:
		text := p.cur.Value
		triple := strings.Contains(text, "\n")
		if triple {
			text = strings.TrimLeft(text, "\n")
		}
		v := &StringVal{Text: text, Triple: triple}
		if err := p.next(); err != nil {
			return nil, err
		}
		return v, nil
	case TokenInteger:
		n, _ := strconv.ParseInt(p.cur.Value, 10, 64)
		v := &IntegerVal{Raw: p.cur.Value, Val: n}
		if err := p.next(); err != nil {
			return nil, err
		}
		return v, nil
	case TokenFloat:
		f, _ := strconv.ParseFloat(p.cur.Value, 64)
		v := &FloatVal{Raw: p.cur.Value, Val: f}
		if err := p.next(); err != nil {
			return nil, err
		}
		return v, nil
	case TokenBool:
		v := &BoolVal{Val: p.cur.Value == "true"}
		if err := p.next(); err != nil {
			return nil, err
		}
		return v, nil
	case TokenDateTime:
		v := &DateTimeVal{Raw: p.cur.Value}
		if err := p.next(); err != nil {
			return nil, err
		}
		return v, nil
	case TokenLBracket:
		return p.parseList()
	case TokenLBrace:
		return p.parseInlineTable()
	default:
		return nil, &ParseError{Line: p.cur.Line, Col: p.cur.Col,
			Msg:      "expected value, got " + p.cur.Type.String(),
			Expected: []string{"string", "number", "boolean", "list", "{"},
			Got:      p.cur.Type.String()}
	}
}

func (p *parser) parseList() (Value, error) {
	if err := p.next(); err != nil { // skip [
		return nil, err
	}

	var items []Value
	for {
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenRBracket {
			if err := p.next(); err != nil {
				return nil, err
			}
			return &ListVal{Items: items}, nil
		}
		if p.cur.Type == TokenEOF {
			return nil, &ParseError{Msg: "unterminated list"}
		}

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		items = append(items, val)

		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenComma {
			if err := p.next(); err != nil {
				return nil, err
			}
		}
	}
}

func (p *parser) parseInlineTable() (Value, error) {
	if err := p.next(); err != nil { // skip {
		return nil, err
	}

	var fields []*Field
	for {
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenRBrace {
			if err := p.next(); err != nil {
				return nil, err
			}
			return &InlineTableVal{Fields: fields}, nil
		}
		if p.cur.Type == TokenEOF {
			return nil, &ParseError{Msg: "unterminated inline table"}
		}

		if p.cur.Type != TokenIdent {
			return nil, &ParseError{Line: p.cur.Line, Col: p.cur.Col,
				Msg: "expected field name in inline table"}
		}

		ident := p.cur
		if err := p.next(); err != nil {
			return nil, err
		}
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}

		f, err := p.parseFieldAfterIdent(ident)
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)

		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenComma {
			if err := p.next(); err != nil {
				return nil, err
			}
		}
	}
}

// --- Spec block (v3: normal-mode grouping) ---

func (p *parser) parseSpecBlock() (*SpecBlock, error) {
	line := p.cur.Line
	if err := p.next(); err != nil { // skip "spec"
		return nil, err
	}
	if err := p.skipTrivia(); err != nil {
		return nil, err
	}
	return p.parseSpecBlockAfterKeyword(line)
}

func (p *parser) parseSpecBlockAfterKeyword(line int) (*SpecBlock, error) {
	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	sb := &SpecBlock{Line: line}

	for {
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenRBrace {
			if err := p.next(); err != nil {
				return nil, err
			}
			return sb, nil
		}
		if p.cur.Type == TokenEOF {
			return nil, &ParseError{Line: line, Msg: "unterminated spec block"}
		}

		switch p.cur.Type {
		case TokenInclude:
			inc, err := p.parseIncludeDirective()
			if err != nil {
				return nil, err
			}
			sb.Includes = append(sb.Includes, inc)
		case TokenFeature:
			feat, err := p.parseFeatureBlock()
			if err != nil {
				return nil, err
			}
			sb.Features = append(sb.Features, feat.(*FeatureBlock))
		default:
			return nil, &ParseError{Line: p.cur.Line, Col: p.cur.Col,
				Msg:      "expected 'include' or 'feature' inside spec block, got " + p.cur.Type.String(),
				Expected: []string{"include", "feature"},
				Got:      p.cur.Type.String()}
		}
	}
}

func (p *parser) parseIncludeDirective() (*IncludeDirective, error) {
	line := p.cur.Line
	if err := p.next(); err != nil { // skip "include"
		return nil, err
	}
	if err := p.skipTrivia(); err != nil {
		return nil, err
	}

	if p.cur.Type != TokenString {
		return nil, &ParseError{Line: p.cur.Line, Col: p.cur.Col,
			Msg: "expected string after include"}
	}
	path := p.cur.Value
	if err := p.next(); err != nil {
		return nil, err
	}

	return &IncludeDirective{Path: path, Line: line}, nil
}

// --- Feature block ---

func (p *parser) parseFeatureBlock() (Node, error) {
	line := p.cur.Line
	if err := p.next(); err != nil { // skip "feature"
		return nil, err
	}
	if err := p.skipTrivia(); err != nil {
		return nil, err
	}
	return p.parseFeatureBlockAfterKeyword(line)
}

func (p *parser) parseFeatureBlockAfterKeyword(line int) (Node, error) {
	fb := &FeatureBlock{Line: line}

	if p.cur.Type == TokenString {
		fb.Name = p.cur.Value
		if err := p.next(); err != nil {
			return nil, err
		}
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	for {
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenRBrace {
			if err := p.next(); err != nil {
				return nil, err
			}
			return fb, nil
		}
		if p.cur.Type == TokenEOF {
			return nil, &ParseError{Line: line, Msg: "unterminated feature block"}
		}

		switch p.cur.Type {
		case TokenBackground:
			bg, err := p.parseBackgroundBlock()
			if err != nil {
				return nil, err
			}
			fb.Background = bg
		case TokenScenario:
			sc, err := p.parseScenarioBlock()
			if err != nil {
				return nil, err
			}
			fb.Groups = append(fb.Groups, sc)
		case TokenGroup:
			g, err := p.parseGroupBlock()
			if err != nil {
				return nil, err
			}
			fb.Groups = append(fb.Groups, g)
		default:
			return nil, &ParseError{Line: p.cur.Line, Col: p.cur.Col,
				Msg:      "expected 'background', 'scenario', or 'group' inside feature, got " + p.cur.Type.String(),
				Expected: []string{"background", "scenario", "group"},
				Got:      p.cur.Type.String()}
		}
	}
}

// --- Group block ---

func (p *parser) parseGroupBlock() (*Group, error) {
	line := p.cur.Line
	if err := p.next(); err != nil { // skip "group"
		return nil, err
	}
	if err := p.skipTrivia(); err != nil {
		return nil, err
	}

	g := &Group{Line: line}

	if p.cur.Type == TokenString {
		g.Name = p.cur.Value
		if err := p.next(); err != nil {
			return nil, err
		}
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	for {
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenRBrace {
			if err := p.next(); err != nil {
				return nil, err
			}
			return g, nil
		}
		if p.cur.Type == TokenEOF {
			return nil, &ParseError{Line: line, Msg: "unterminated group block"}
		}

		if p.cur.Type != TokenScenario {
			return nil, &ParseError{Line: p.cur.Line, Col: p.cur.Col,
				Msg: "expected 'scenario' inside group, got " + p.cur.Type.String()}
		}

		sc, err := p.parseScenarioBlock()
		if err != nil {
			return nil, err
		}
		g.Scenarios = append(g.Scenarios, sc)
	}
}

// --- Scenario block ---
// Fields are open: any TokenIdent followed by = and a value is accepted.

func (p *parser) parseScenarioBlock() (*Scenario, error) {
	line := p.cur.Line
	if err := p.next(); err != nil { // skip "scenario"
		return nil, err
	}
	if err := p.skipTrivia(); err != nil {
		return nil, err
	}

	sc := &Scenario{Line: line}

	if p.cur.Type == TokenString {
		sc.Name = p.cur.Value
		if err := p.next(); err != nil {
			return nil, err
		}
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	for {
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenRBrace {
			if err := p.next(); err != nil {
				return nil, err
			}
			return sc, nil
		}
		if p.cur.Type == TokenEOF {
			return nil, &ParseError{Line: line, Msg: "unterminated scenario block"}
		}

		switch p.cur.Type {
		case TokenIdent:
			ident := p.cur
			if err := p.next(); err != nil {
				return nil, err
			}
			if err := p.skipTrivia(); err != nil {
				return nil, err
			}
			if _, err := p.expect(TokenEquals); err != nil {
				return nil, err
			}
			if err := p.skipTrivia(); err != nil {
				return nil, err
			}
			val, err := p.parseValue()
			if err != nil {
				return nil, err
			}
			sc.Fields = append(sc.Fields, &Field{Key: ident.Value, Value: val, Line: ident.Line})
		case TokenGiven, TokenWhen, TokenThen:
			ident := Token{Type: TokenIdent, Value: p.cur.Value, Line: p.cur.Line, Col: p.cur.Col, Offset: p.cur.Offset}
			savedType := p.cur.Type
			if err := p.next(); err != nil {
				return nil, err
			}
			if err := p.skipTrivia(); err != nil {
				return nil, err
			}
			if p.cur.Type == TokenEquals {
				f, err := p.parseFieldAfterIdent(ident)
				if err != nil {
					return nil, err
				}
				sc.Fields = append(sc.Fields, f)
			} else {
				sb, err := p.parseStepBlockAfterKeyword(ident.Line)
				if err != nil {
					return nil, err
				}
				switch savedType {
				case TokenGiven:
					sc.Given = sb
				case TokenWhen:
					sc.When = sb
				case TokenThen:
					sc.Then = sb
				}
			}
		default:
			return nil, &ParseError{Line: p.cur.Line, Col: p.cur.Col,
				Msg:      "expected 'given', 'when', 'then', or field inside scenario, got " + p.cur.Type.String(),
				Expected: []string{"given", "when", "then", "field"},
				Got:      p.cur.Type.String()}
		}
	}
}

// --- Background block ---

func (p *parser) parseBackgroundBlock() (*Background, error) {
	line := p.cur.Line
	if err := p.next(); err != nil { // skip "background"
		return nil, err
	}
	if err := p.skipTrivia(); err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	bg := &Background{Line: line}

	for {
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenRBrace {
			if err := p.next(); err != nil {
				return nil, err
			}
			return bg, nil
		}
		if p.cur.Type == TokenEOF {
			return nil, &ParseError{Line: line, Msg: "unterminated background block"}
		}

		if p.cur.Type != TokenGiven {
			return nil, &ParseError{Line: p.cur.Line, Col: p.cur.Col,
				Msg: "expected 'given' inside background, got " + p.cur.Type.String()}
		}

		sb, err := p.parseStepBlock()
		if err != nil {
			return nil, err
		}
		bg.Given = sb
	}
}

// --- Step block ---

func (p *parser) parseStepBlock() (*StepBlock, error) {
	line := p.cur.Line
	if err := p.next(); err != nil { // skip given/when/then keyword
		return nil, err
	}
	if err := p.skipTrivia(); err != nil {
		return nil, err
	}
	return p.parseStepBlockAfterKeyword(line)
}

func (p *parser) parseStepBlockAfterKeyword(line int) (*StepBlock, error) {
	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	// Switch lexer to step-text mode
	p.lex.EnterStepTextMode()
	if err := p.next(); err != nil {
		return nil, err
	}

	sb := &StepBlock{Line: line}

	for {
		if err := p.skipTrivia(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenRBrace {
			if err := p.next(); err != nil {
				return nil, err
			}
			return sb, nil
		}
		if p.cur.Type == TokenEOF {
			return nil, &ParseError{Line: line, Msg: "unterminated step block"}
		}
		if p.cur.Type == TokenStepText {
			sb.Lines = append(sb.Lines, p.cur.Value)
			if err := p.next(); err != nil {
				return nil, err
			}
		} else {
			return nil, &ParseError{Line: p.cur.Line, Col: p.cur.Col,
				Msg: "unexpected token in step block: " + p.cur.Type.String()}
		}
	}
}
