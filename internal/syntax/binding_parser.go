package syntax

func (p *Parser) parseBinding() BindingDecl {
	keyword := p.advance()
	result := BindingDecl{Span: keyword.Span}
	left := p.expectIdentifier("binding source activity", DiagExpectedIdentifier)
	p.expect(TokenDot, ".", DiagUnexpectedDeclaration)
	source := p.expectIdentifier("binding source port", DiagExpectedIdentifier)
	p.expect(TokenArrow, "->", DiagExpectedArrow)
	right := p.expectIdentifier("binding target activity", DiagExpectedIdentifier)
	p.expect(TokenDot, ".", DiagUnexpectedDeclaration)
	target := p.expectIdentifier("binding target port", DiagExpectedIdentifier)
	result.SourceActivity, result.SourcePort = left.Name, source.Name
	result.TargetActivity, result.TargetPort = right.Name, target.Name
	result.SourceActivitySpan, result.SourcePortSpan = left.Span, source.Span
	result.TargetActivitySpan, result.TargetPortSpan = right.Span, target.Span
	if !target.Span.IsEmpty() { result.Span.End = target.Span.End }
	return result
}
