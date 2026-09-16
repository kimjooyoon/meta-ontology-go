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
	result.SourceActivity, result.SourcePort = left.Value, source.Value
	result.TargetActivity, result.TargetPort = right.Value, target.Value
	if !target.Span.IsEmpty() { result.Span.End = target.Span.End }
	return result
}
