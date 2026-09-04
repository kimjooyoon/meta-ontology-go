package syntax

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
		case p.at(TokenIdentifier) && p.peek().Value == "policy":
			file.Decls = append(file.Decls, p.parsePolicy())
			file.Declarations = file.Decls
		default:
			p.error(DiagUnexpectedDeclaration, p.peek().Span, "expected entity or activity declaration")
			p.advance()
		}
	}
}
