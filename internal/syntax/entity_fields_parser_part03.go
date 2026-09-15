package syntax

func isEntityFieldsReservedWord(value string) bool {
	switch value {
	case "fields", "field", "type", "required", "optional", "one", "many":
		return true
	default:
		return false
	}
}
func (p *Parser) atEntityFieldsBoundary() bool {
	return p.at(TokenEOF) || p.at(TokenEntity) || p.at(TokenActivity)
}
func (p *Parser) recoverEntityFields(depth int) {
	for !p.at(TokenEOF) {
		token := p.peek()
		if depth == 0 && (p.at(TokenEntity) || p.at(TokenActivity)) {
			return
		}
		if depth == 1 && (p.at(TokenEntity) || p.at(TokenActivity)) {
			return
		}
		switch token.Kind {
		case TokenLBrace:
			depth++
		case TokenRBrace:
			if depth > 0 {
				depth--
			}
			if depth == 0 {
				p.advance()
				return
			}
		}
		p.advance()
	}
}
func (p *Parser) hasDiagnosticFrom(offset int) bool {
	for _, diagnostic := range p.diagnostics {
		if diagnostic.Span.Start.Offset >= offset {
			return true
		}
	}
	return false
}
func (p *Parser) discardDiagnosticsFrom(offset int) {
	kept := p.diagnostics[:0]
	for _, diagnostic := range p.diagnostics {
		if diagnostic.Span.Start.Offset < offset {
			kept = append(kept, diagnostic)
		}
	}
	p.diagnostics = kept
}
