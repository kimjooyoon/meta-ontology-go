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
