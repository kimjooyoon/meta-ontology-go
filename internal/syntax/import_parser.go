package syntax

import "strings"

func (p *Parser) parseImport() ImportDecl {
	keyword := p.advance()
	path := p.expectString()
	if strings.TrimSpace(path.Name) == "" {
		p.error(DiagUnexpectedDeclaration, path.Span, "package import path must be nonempty")
	}
	return ImportDecl{Span: startSpan(p.filename, keyword.Span.Start, path.Span.End), Path: path.Name}
}
