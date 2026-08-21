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
