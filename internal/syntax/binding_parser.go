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
		Span:     startSpan(p.filename, keyword.Span.Start, end),
		Producer: producer,
		Consumer: consumer,
	}
}

func (p *Parser) parseBindingEndpoint() BindingEndpoint {
	activity := p.expectIdentifier("binding activity", DiagExpectedIdentifier)
	p.expect(TokenDot, ".", DiagExpectedDot)
	port := p.expectIdentifier("binding port", DiagExpectedIdentifier)
	end := activity.Span.End
	if !port.Span.IsEmpty() {
		end = port.Span.End
	}
	return BindingEndpoint{
		Span:     startSpan(p.filename, activity.Span.Start, end),
		Activity: activity,
		Port:     port,
	}
}
