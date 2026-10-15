package syntax

func (p *Parser) parseActivity() *ActivityDecl {
	keyword := p.advance()
	name := p.expectIdentifier("activity name", DiagExpectedIdentifier)
	activity := &ActivityDecl{
		Span:     startSpan(p.filename, keyword.Span.Start, keyword.Span.End),
		Name:     name.Name,
		NameSpan: name.Span,
	}
	if !name.Span.IsEmpty() {
		activity.Span.End = name.Span.End
	}

	p.parseActivityInputs(activity)
	p.finishActivityInputNames(activity)
	p.expect(TokenArrow, "->", DiagExpectedArrow)
	result := p.parseActivityTypeReference("activity result", DiagExpectedResult)
	activity.Result = result
	activity.Parameters = append([]NameRef(nil), activity.Inputs...)
	activity.Output = result.Name
	if !result.Span.IsEmpty() {
		activity.Span.End = result.Span.End
	}
	p.parseActivityIdentity(activity)
	if p.at(TokenIdentifier) && p.peek().Value == "computes" {
		p.advance()
		activity.ValueProgramPresent = true
		p.skipIllegal()
		if p.at(TokenString) {
			program := p.advance()
			activity.ValueProgram = program.Value
			activity.ValueProgramSpan = program.Span
			activity.Span.End = program.Span.End
		} else {
			p.error(DiagExpectedString, p.peek().Span, "expected quoted activity value program")
		}
	}
	if p.at(TokenIdentifier) && p.peek().Value == "assembling" {
		activity.Assembly = p.parseAssembly()
		activity.Span.End = activity.Assembly.Span.End
		if !activity.ValueProgramPresent {
			p.error(DiagUnexpectedDeclaration, activity.Assembly.Span, "assembling requires an activity computes body")
		}
	}
	return activity
}

func (p *Parser) parseActivityInputs(activity *ActivityDecl) {
	p.expect(TokenLParen, "(", DiagExpectedLeftParen)
	if !p.at(TokenRParen) {
		for {
			p.skipIllegal()
			if p.at(TokenIdentifier) {
				parameter, inputName := p.parseActivityInput()
				activity.Inputs = append(activity.Inputs, parameter)
				activity.InputNames = append(activity.InputNames, inputName)
				activity.Span.End = parameter.Span.End
			} else {
				p.error(DiagExpectedIdentifier, p.peek().Span, "expected activity parameter name")
				if p.at(TokenComma) {
					p.advance()
				}
				break
			}

			if p.at(TokenComma) {
				p.advance()
				if p.at(TokenRParen) {
					p.error(DiagExpectedIdentifier, p.peek().Span, "expected activity parameter name after comma")
					break
				}
				continue
			}
			if !p.at(TokenRParen) {
				p.error(DiagExpectedComma, p.peek().Span, "expected comma or closing parenthesis after activity parameter")
				if !p.at(TokenEOF) {
					p.advance()
				}
			}
			break
		}
	}

	closing := p.expect(TokenRParen, ")", DiagExpectedRightParen)
	if !closing.Span.IsEmpty() {
		activity.Span.End = closing.Span.End
	}
}
