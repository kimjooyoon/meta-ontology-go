package syntax

func (p *Parser) atEntityFieldsMarker() bool {
	return p.at(TokenIdentifier) && p.peek().Value == "fields"
}

func (p *Parser) rejectEntityFields(marker Token) {
	p.discardDiagnosticsFrom(marker.Span.Start.Offset)
	support := CurrentEntityFieldsSupport()
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

func (p *Parser) discardDiagnosticsFrom(offset int) {
	kept := p.diagnostics[:0]
	for _, diagnostic := range p.diagnostics {
		if diagnostic.Span.Start.Offset < offset {
			kept = append(kept, diagnostic)
		}
	}
	p.diagnostics = kept
}
