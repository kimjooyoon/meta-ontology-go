package syntax

// Parser turns tokens into a deterministic, source-spanned AST. It performs
// syntax recovery at declaration boundaries so editor clients can continue to
// use a partial tree while a file is being edited.
type Parser struct {
	filename string
	source   string
	tokens   Tokens
	eof      Span

	index                int
	diagnostics          Diagnostics
	parsed               bool
	file                 *File
	entityFieldsRejected bool
	entityFieldsSupport  EntityFieldsSupport
	entityFieldsError    error
}

// NewParser creates a parser for an unnamed source.
func NewParser(source string) *Parser {
	return NewParserWithEntityFieldsSupport(source, CurrentEntityFieldsSupport())
}

// NewParserFile creates a parser for a named source.
func NewParserFile(filename, source string) *Parser {
	return NewParserFileWithEntityFieldsSupport(filename, source, CurrentEntityFieldsSupport())
}

// NewParserWithEntityFieldsSupport creates a parser with an explicit,
// profile-bound EntityFields mode. The mode is validated before parsing.
func NewParserWithEntityFieldsSupport(source string, support EntityFieldsSupport) *Parser {
	return NewParserFileWithEntityFieldsSupport("", source, support)
}

// NewParserFileWithEntityFieldsSupport creates a named parser with an
// explicit, profile-bound EntityFields mode.
func NewParserFileWithEntityFieldsSupport(filename, source string, support EntityFieldsSupport) *Parser {
	tokens, diagnostics := LexFile(filename, source)
	return &Parser{
		filename:            filename,
		source:              source,
		tokens:              tokens,
		eof:                 tokens[len(tokens)-1].Span,
		diagnostics:         diagnostics,
		entityFieldsSupport: support,
		entityFieldsError:   support.Validate(),
	}
}

// Parse parses the source once. Repeated calls return copies of the diagnostic
// slice and the same immutable-by-convention AST pointer. Deferred EntityFields
// input returns no AST so unsupported source cannot leak a partial result.
func (p *Parser) Parse() (*File, Diagnostics) {
	if p.parsed {
		return p.file, append(Diagnostics(nil), p.diagnostics...)
	}
	p.parsed = true
	if p.entityFieldsError != nil {
		p.error(DiagEntityFieldsConfiguration, p.eof, p.entityFieldsError.Error())
		return nil, append(Diagnostics(nil), p.diagnostics...)
	}
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

// ParseWithEntityFieldsSupport parses an unnamed source with an explicit
// EntityFields mode.
func ParseWithEntityFieldsSupport(source string, support EntityFieldsSupport) (*File, Diagnostics) {
	return NewParserWithEntityFieldsSupport(source, support).Parse()
}

// ParseFileWithEntityFieldsSupport parses a named source with an explicit
// EntityFields mode.
func ParseFileWithEntityFieldsSupport(filename, source string, support EntityFieldsSupport) (*File, Diagnostics) {
	return NewParserFileWithEntityFieldsSupport(filename, source, support).Parse()
}

func (p *Parser) parseFile() *File {
	start := Position{Offset: 0, Line: 1, Column: 1}
	file := &File{Span: startSpan(p.filename, start, start)}

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
			file.Span.End = p.eof.End
			file.Declarations = file.Decls
			return file
		case p.at(TokenEntity):
			entity := p.parseEntity()
			if p.entityFieldsRejected {
				return nil
			}
			file.Decls = append(file.Decls, entity)
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
	fields := []FieldDecl(nil)
	fieldsPresent := false
	end := keyword.Span.End
	if !name.Span.IsEmpty() {
		end = name.Span.End
	}
	if !id.Span.IsEmpty() {
		end = id.Span.End
	}
	if p.atEntityFieldsMarker() {
		marker := p.advance()
		switch p.entityFieldsSupport.State {
		case EntityFieldsDeferred:
			p.rejectEntityFields(marker)
		case EntityFieldsSupported:
			fields, fieldsPresent, end = p.parseEntityFields(marker)
			if !fieldsPresent {
				fields = nil
			}
		default:
			p.error(DiagEntityFieldsConfiguration, marker.Span, ErrEntityFieldsUnknownState.Error())
		}
	}
	return &EntityDecl{
		Span:          startSpan(p.filename, keyword.Span.Start, end),
		Name:          name.Name,
		ID:            id.Name,
		NameSpan:      name.Span,
		IDSpan:        id.Span,
		Fields:        fields,
		FieldsPresent: fieldsPresent,
	}
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
