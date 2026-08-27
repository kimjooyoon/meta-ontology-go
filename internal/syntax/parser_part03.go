package syntax

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

func (p *Parser) parseFreshness() *FreshnessDecl {
	keyword := p.advance()
	kind := p.expectIdentifier("freshness policy kind", DiagExpectedIdentifier)
	valueCount := 1
	if kind.Name == "axes" {
		valueCount = 6
	}
	values := make([]NameRef, 0, valueCount)
	end := keyword.Span.End
	if !kind.Span.IsEmpty() {
		end = kind.Span.End
	}
	for index := 0; index < valueCount; index++ {
		value := p.expectIdentifier("freshness policy value", DiagExpectedIdentifier)
		if !value.Span.IsEmpty() {
			end = value.Span.End
		}
		values = append(values, value)
	}
	return &FreshnessDecl{Span: startSpan(p.filename, keyword.Span.Start, end), Kind: kind.Name, Values: values}
}
func (p *Parser) expect(kind TokenKind, expected string, code DiagnosticCode) Token {
	p.skipIllegal()
	if p.at(kind) {
		return p.advance()
	}
	p.error(code, p.peek().Span, "expected "+expected)
	return Token{}
}
