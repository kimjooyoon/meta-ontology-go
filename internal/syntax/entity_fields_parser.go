package syntax

func (p *Parser) atEntityFieldsMarker() bool {
	return p.at(TokenIdentifier) && p.peek().Value == "fields"
}

func (p *Parser) rejectEntityFields(marker Token) {
	p.discardDiagnosticsFrom(marker.Span.Start.Offset)
	support := p.entityFieldsSupport
	switch support.State {
	case EntityFieldsDeferred:
		if err := support.Validate(); err != nil {
			p.error(DiagEntityFieldsConfiguration, marker.Span, err.Error())
		} else {
			p.error(DiagEntityFieldsDeferred, marker.Span, "entity fields are deferred and unsupported by the public syntax")
		}
	case EntityFieldsSupported:
		if err := support.Validate(); err != nil {
			p.error(DiagEntityFieldsConfiguration, marker.Span, err.Error())
		} else {
			p.error(DiagEntityFieldsConfiguration, marker.Span, ErrEntityFieldsSupportUnavailable.Error())
		}
	default:
		p.error(DiagEntityFieldsConfiguration, marker.Span, ErrEntityFieldsUnknownState.Error())
	}
	p.entityFieldsRejected = true
	p.index = len(p.tokens) - 1
}

func (p *Parser) parseEntityFields(marker Token) ([]FieldDecl, bool, Position) {
	if p.hasDiagnosticFrom(marker.Span.Start.Offset) {
		p.recoverEntityFields(0)
		return nil, false, marker.Span.End
	}
	if !p.at(TokenLBrace) {
		p.error(DiagExpectedFieldsLeftBrace, p.peek().Span, "expected { after fields")
		p.recoverEntityFields(0)
		return nil, false, marker.Span.End
	}
	p.advance()
	fields := make([]FieldDecl, 0)
	for {
		p.skipIllegal()
		switch {
		case p.at(TokenRBrace):
			closing := p.advance()
			if p.hasDiagnosticFrom(marker.Span.Start.Offset) {
				return nil, false, closing.Span.End
			}
			if !p.atEntityFieldsBoundary() {
				p.error(DiagEntityFieldsTrailing, p.peek().Span, "unexpected token after fields block")
				p.recoverEntityFields(0)
				return nil, false, closing.Span.End
			}
			return fields, true, closing.Span.End
		case p.at(TokenEOF), p.at(TokenEntity), p.at(TokenActivity):
			p.error(DiagEntityFieldsUnterminated, p.peek().Span, "expected } to close fields block")
			return nil, false, marker.Span.End
		case p.atEntityFieldsWord("field"):
			field, ok := p.parseEntityField()
			if !ok {
				p.recoverEntityFields(1)
				return nil, false, marker.Span.End
			}
			fields = append(fields, field)
		default:
			p.error(DiagExpectedEntityField, p.peek().Span, "expected field declaration or }")
			p.recoverEntityFields(1)
			return nil, false, marker.Span.End
		}
	}
}

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
