package syntax

// Parser turns tokens into a deterministic, source-spanned AST. It performs
// syntax recovery at declaration boundaries so editor clients can continue to
// use a partial tree while a file is being edited.
type Parser struct {
	filename string
	source   string
	tokens   Tokens

	index       int
	diagnostics Diagnostics
	parsed      bool
	file        *File
}

// NewParser creates a parser for an unnamed source.
func NewParser(source string) *Parser {
	return NewParserFile("", source)
}

// NewParserFile creates a parser for a named source.
func NewParserFile(filename, source string) *Parser {
	tokens, diagnostics := LexFile(filename, source)
	return &Parser{
		filename:    filename,
		source:      source,
		tokens:      tokens,
		diagnostics: diagnostics,
	}
}

// Parse parses the source once. Repeated calls return copies of the diagnostic
// slice and the same immutable-by-convention AST pointer.
func (p *Parser) Parse() (*File, Diagnostics) {
	if p.parsed {
		return p.file, append(Diagnostics(nil), p.diagnostics...)
	}
	p.parsed = true
	p.file = p.parseFile()
	return p.file, append(Diagnostics(nil), p.diagnostics...)
}

// Diagnostics returns the parser's current diagnostics. Calling Parse first
// is recommended because parser diagnostics are added during parsing.
func (p *Parser) Diagnostics() Diagnostics {
	return append(Diagnostics(nil), p.diagnostics...)
}

// Parse parses an unnamed .gooo source file.
func Parse(source string) (*File, Diagnostics) {
	return NewParser(source).Parse()
}

// ParseFile parses a named .gooo source file.
func ParseFile(filename, source string) (*File, Diagnostics) {
	return NewParserFile(filename, source).Parse()
}

func (p *Parser) parseFile() *File {
	start := Position{Offset: 0, Line: 1, Column: 1}
	end := p.tokens[len(p.tokens)-1].Span.End
	file := &File{Span: startSpan(p.filename, start, end)}

	p.skipIllegal()
	if p.at(TokenPackage) {
		file.Package = p.parsePackage()
	} else {
		p.error(DiagExpectedPackage, p.peek().Span, "expected package declaration")
	}

	p.skipIllegal()
	if p.at(TokenNamespace) {
		file.Namespace = p.parseNamespace()
	} else {
		p.error(DiagExpectedNamespace, p.peek().Span, "expected namespace declaration")
	}

	for {
		p.skipIllegal()
		switch {
		case p.at(TokenEOF):
			file.Declarations = file.Decls
			return file
		case p.at(TokenEntity):
			file.Decls = append(file.Decls, p.parseEntity())
			file.Declarations = file.Decls
		case p.at(TokenActivity):
			file.Decls = append(file.Decls, p.parseActivity())
			file.Declarations = file.Decls
		default:
			p.error(DiagUnexpectedDeclaration, p.peek().Span, "expected entity or activity declaration")
			p.advance()
		}
	}
}

func (p *Parser) parsePackage() *PackageDecl {
	keyword := p.advance()
	name := p.expectIdentifier("package name", DiagExpectedIdentifier)
	end := keyword.Span.End
	if !name.Span.IsEmpty() {
		end = name.Span.End
	}
	return &PackageDecl{
		Span:     startSpan(p.filename, keyword.Span.Start, end),
		Name:     name.Name,
		NameSpan: name.Span,
	}
}

func (p *Parser) parseNamespace() *NamespaceDecl {
	keyword := p.advance()
	name := p.expectIdentifier("namespace name", DiagExpectedIdentifier)
	end := keyword.Span.End
	if !name.Span.IsEmpty() {
		end = name.Span.End
	}
	return &NamespaceDecl{
		Span:     startSpan(p.filename, keyword.Span.Start, end),
		Name:     name.Name,
		NameSpan: name.Span,
	}
}

func (p *Parser) parseEntity() *EntityDecl {
	keyword := p.advance()
	name := p.expectIdentifier("entity name", DiagExpectedIdentifier)
	p.expect(TokenID, "id", DiagExpectedID)
	id := p.expectString()

	end := keyword.Span.End
	if !name.Span.IsEmpty() {
		end = name.Span.End
	}
	if !id.Span.IsEmpty() {
		end = id.Span.End
	}
	return &EntityDecl{
		Span:     startSpan(p.filename, keyword.Span.Start, end),
		Name:     name.Name,
		ID:       id.Name,
		NameSpan: name.Span,
		IDSpan:   id.Span,
	}
}

func (p *Parser) parseActivity() *ActivityDecl {
	keyword := p.advance()
	name := p.expectIdentifier("activity name", DiagExpectedIdentifier)
	activity := &ActivityDecl{
		Span:     startSpan(p.filename, keyword.Span.Start, keyword.Span.End),
		Name:     name.Name,
		NameSpan: name.Span,
	}
	if !name.Span.IsEmpty() {
		activity.Span.End = name.Span.End
	}

	p.expect(TokenLParen, "(", DiagExpectedLeftParen)
	if !p.at(TokenRParen) {
		for {
			p.skipIllegal()
			if p.at(TokenIdentifier) {
				parameter := p.advance()
				activity.Inputs = append(activity.Inputs, NameRef{Span: parameter.Span, Name: parameter.Value})
				activity.Span.End = parameter.Span.End
			} else {
				p.error(DiagExpectedIdentifier, p.peek().Span, "expected activity parameter name")
				if p.at(TokenComma) {
					p.advance()
				}
				break
			}

			if p.at(TokenComma) {
				p.advance()
				if p.at(TokenRParen) {
					p.error(DiagExpectedIdentifier, p.peek().Span, "expected activity parameter name after comma")
					break
				}
				continue
			}
			if !p.at(TokenRParen) {
				p.error(DiagExpectedComma, p.peek().Span, "expected comma or closing parenthesis after activity parameter")
				if !p.at(TokenEOF) {
					p.advance()
				}
			}
			break
		}
	}

	closing := p.expect(TokenRParen, ")", DiagExpectedRightParen)
	if !closing.Span.IsEmpty() {
		activity.Span.End = closing.Span.End
	}
	p.expect(TokenArrow, "->", DiagExpectedArrow)
	result := p.expectIdentifier("activity result", DiagExpectedResult)
	activity.Result = result
	activity.Parameters = append([]NameRef(nil), activity.Inputs...)
	activity.Output = result.Name
	if !result.Span.IsEmpty() {
		activity.Span.End = result.Span.End
	}
	return activity
}

func (p *Parser) expect(kind TokenKind, expected string, code DiagnosticCode) Token {
	p.skipIllegal()
	if p.at(kind) {
		return p.advance()
	}
	p.error(code, p.peek().Span, "expected "+expected)
	return Token{}
}

func (p *Parser) expectIdentifier(expected string, code DiagnosticCode) NameRef {
	p.skipIllegal()
	if p.at(TokenIdentifier) {
		token := p.advance()
		return NameRef{Span: token.Span, Name: token.Value}
	}
	p.error(code, p.peek().Span, "expected "+expected)
	return NameRef{}
}

func (p *Parser) expectString() NameRef {
	p.skipIllegal()
	if p.at(TokenString) {
		token := p.advance()
		return NameRef{Span: token.Span, Name: token.Value}
	}
	p.error(DiagExpectedString, p.peek().Span, "expected quoted semantic identifier")
	return NameRef{}
}

func (p *Parser) skipIllegal() {
	for p.at(TokenIllegal) {
		p.advance()
	}
}

func (p *Parser) at(kind TokenKind) bool {
	return p.peek().Kind == kind
}

func (p *Parser) peek() Token {
	if p.index >= len(p.tokens) {
		position := Position{Offset: len(p.source), Line: 1, Column: 1}
		return Token{Kind: TokenEOF, Span: startSpan(p.filename, position, position)}
	}
	return p.tokens[p.index]
}

func (p *Parser) advance() Token {
	token := p.peek()
	if p.index < len(p.tokens) {
		p.index++
	}
	return token
}

func (p *Parser) error(code DiagnosticCode, span Span, message string) {
	p.diagnostics = append(p.diagnostics, Diagnostic{
		Severity: SeverityError,
		Code:     code,
		Message:  message,
		Span:     span,
	})
}
