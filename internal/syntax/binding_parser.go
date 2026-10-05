package syntax

func (p *Parser) parseBinding() BindingDecl {
	keyword := p.advance()
	producer := p.parseBindingEndpoint()
	p.expect(TokenArrow, "->", DiagExpectedArrow)
	consumer := p.parseBindingEndpoint()
	end := keyword.Span.End
	if !producer.Span.IsEmpty() {
		end = producer.Span.End
	}
	if !consumer.Span.IsEmpty() {
		end = consumer.Span.End
	}
	return BindingDecl{
		Feedback: keyword.Value == "feedback",
		Span:     startSpan(p.filename, keyword.Span.Start, end),
		Producer: producer,
		Consumer: consumer,
	}
}

func (p *Parser) parseBindingEndpoint() BindingEndpoint {
	first := p.expectIdentifier("binding activity or package alias", DiagExpectedIdentifier)
	p.expect(TokenDot, ".", DiagExpectedDot)
	second := p.expectIdentifier("binding activity or port", DiagExpectedIdentifier)
	packageAlias := ""
	activity := first
	port := second
	end := second.Span.End
	if p.at(TokenDot) {
		p.advance()
		packageAlias = first.Name
		activity = second
		port = p.expectIdentifier("binding port", DiagExpectedIdentifier)
		end = port.Span.End
	}
	return BindingEndpoint{
		Span:         startSpan(p.filename, first.Span.Start, end),
		PackageAlias: packageAlias,
		Activity:     activity,
		Port:         port,
	}
}
