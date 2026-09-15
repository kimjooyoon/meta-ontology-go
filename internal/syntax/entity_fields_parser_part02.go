package syntax

func (p *Parser) parseEntityField() (FieldDecl, bool) {
	keyword := p.advance()
	name := p.expectFieldIdentifier()
	p.expect(TokenID, "id", DiagExpectedID)
	id := p.expectString()
	p.expectEntityFieldsWord("type", DiagExpectedFieldType)
	typeRef := p.expectTypeReference()
	presence := p.expectFieldPresence()
	cardinality := p.expectFieldCardinality()
	if name.Span.IsEmpty() || id.Span.IsEmpty() || typeRef.Span.IsEmpty() || presence.Span.IsEmpty() || cardinality.Span.IsEmpty() {
		return FieldDecl{}, false
	}
	return FieldDecl{
		Span:            startSpan(p.filename, keyword.Span.Start, cardinality.Span.End),
		ID:              id.Name,
		Name:            name.Name,
		TypeRef:         typeRef,
		Presence:        FieldPresence(presence.Name),
		Cardinality:     FieldCardinality(cardinality.Name),
		IDSpan:          id.Span,
		NameSpan:        name.Span,
		PresenceSpan:    presence.Span,
		CardinalitySpan: cardinality.Span,
	}, true
}
func (p *Parser) expectFieldIdentifier() NameRef {
	p.skipIllegal()
	if p.at(TokenIdentifier) && !isEntityFieldsReservedWord(p.peek().Value) {
		token := p.advance()
		return NameRef{Span: token.Span, Name: token.Value}
	}
	p.error(DiagExpectedIdentifier, p.peek().Span, "expected field name")
	return NameRef{}
}
func (p *Parser) expectTypeReference() TypeRefDecl {
	p.skipIllegal()
	if p.at(TokenString) || (p.at(TokenIdentifier) && !isEntityFieldsReservedWord(p.peek().Value)) {
		token := p.advance()
		return TypeRefDecl{Span: token.Span, Spelling: token.Value}
	}
	p.error(DiagExpectedFieldType, p.peek().Span, "expected type reference")
	return TypeRefDecl{}
}
func (p *Parser) expectFieldPresence() NameRef {
	if p.atEntityFieldsWord("required") || p.atEntityFieldsWord("optional") {
		token := p.advance()
		return NameRef{Span: token.Span, Name: token.Value}
	}
	p.error(DiagExpectedFieldPresence, p.peek().Span, "expected required or optional")
	return NameRef{}
}
func (p *Parser) expectFieldCardinality() NameRef {
	if p.atEntityFieldsWord("one") || p.atEntityFieldsWord("many") {
		token := p.advance()
		return NameRef{Span: token.Span, Name: token.Value}
	}
	p.error(DiagExpectedFieldCardinality, p.peek().Span, "expected one or many")
	return NameRef{}
}
func (p *Parser) expectEntityFieldsWord(word string, code DiagnosticCode) Token {
	if p.atEntityFieldsWord(word) {
		return p.advance()
	}
	p.error(code, p.peek().Span, "expected "+word)
	return Token{}
}
func (p *Parser) atEntityFieldsWord(word string) bool {
	return p.at(TokenIdentifier) && p.peek().Value == word
}
