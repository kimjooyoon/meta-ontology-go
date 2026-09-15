package syntax

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
