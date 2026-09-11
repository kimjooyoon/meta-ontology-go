package syntax

import (
	"fmt"
	"strconv"
)

func (p *Parser) parsePolicy() *PolicyDecl {
	keyword := p.advance()
	name := p.expectIdentifier("policy name", DiagExpectedIdentifier)
	p.expect(TokenID, "id", DiagExpectedID)
	id := p.expectString()
	policy := &PolicyDecl{
		Span:     startSpan(p.filename, keyword.Span.Start, keyword.Span.End),
		Name:     name.Name,
		ID:       id.Name,
		NameSpan: name.Span,
		IDSpan:   id.Span,
	}
	if !name.Span.IsEmpty() {
		policy.Span.End = name.Span.End
	}
	if !id.Span.IsEmpty() {
		policy.Span.End = id.Span.End
	}
	p.expect(TokenLBrace, "{", DiagUnexpectedDeclaration)
	for !p.at(TokenRBrace) && !p.at(TokenEOF) {
		p.skipIllegal()
		if !p.at(TokenIdentifier) {
			p.error(DiagUnexpectedDeclaration, p.peek().Span, "expected policy node")
			p.advance()
			continue
		}
		switch p.peek().Value {
		case "state":
			policy.States = append(policy.States, p.parseState())
		case "transition":
			policy.Transitions = append(policy.Transitions, p.parseTransition())
		case "case":
			policy.Cases = append(policy.Cases, p.parseCase())
		default:
			p.error(DiagUnexpectedDeclaration, p.peek().Span, fmt.Sprintf("unknown policy node %q", p.peek().Value))
			p.advance()
		}
	}
	closing := p.expect(TokenRBrace, "}", DiagUnexpectedDeclaration)
	if !closing.Span.IsEmpty() {
		policy.Span.End = closing.Span.End
	}
	return policy
}

func (p *Parser) parseState() StateDecl {
	keyword := p.advance()
	name := p.expectString()
	return StateDecl{Span: startSpan(p.filename, keyword.Span.Start, name.Span.End), Name: name.Name, NameSpan: name.Span}
}

func (p *Parser) parseTransition() TransitionDecl {
	keyword := p.advance()
	from := p.expectString()
	p.expect(TokenArrow, "->", DiagExpectedArrow)
	to := p.expectString()
	return TransitionDecl{
		Span: startSpan(p.filename, keyword.Span.Start, to.Span.End), From: from.Name, To: to.Name,
		FromSpan: from.Span, ToSpan: to.Span,
	}
}

func (p *Parser) parseCase() CaseDecl {
	keyword := p.advance()
	name := p.expectString()
	current := CaseDecl{
		Span: startSpan(p.filename, keyword.Span.Start, name.Span.End), Name: name.Name, NameSpan: name.Span,
	}
	p.expect(TokenLBrace, "{", DiagUnexpectedDeclaration)
	for !p.at(TokenRBrace) && !p.at(TokenEOF) {
		p.skipIllegal()
		if !p.at(TokenIdentifier) {
			p.error(DiagUnexpectedDeclaration, p.peek().Span, "expected case node")
			p.advance()
			continue
		}
		switch p.peek().Value {
		case "evidence":
			current.Evidence = append(current.Evidence, p.parseEvidence())
		case "resolution":
			if current.Resolution != nil {
				p.error(DiagUnexpectedDeclaration, p.peek().Span, "case has duplicate resolution")
				p.skipPolicyBlock()
			} else {
				resolution := p.parseResolution()
				current.Resolution = &resolution
			}
		default:
			p.error(DiagUnexpectedDeclaration, p.peek().Span, fmt.Sprintf("unknown case node %q", p.peek().Value))
			p.advance()
		}
	}
	closing := p.expect(TokenRBrace, "}", DiagUnexpectedDeclaration)
	if !closing.Span.IsEmpty() {
		current.Span.End = closing.Span.End
	}
	return current
}

func (p *Parser) parseEvidence() EvidenceDecl {
	keyword := p.advance()
	name := p.expectString()
	value := p.expectString()
	return EvidenceDecl{
		Span: startSpan(p.filename, keyword.Span.Start, value.Span.End), Name: name.Name, Value: value.Name,
		NameSpan: name.Span, ValueSpan: value.Span,
	}
}

func (p *Parser) parseResolution() ResolutionDecl {
	keyword := p.advance()
	resolution := ResolutionDecl{Span: startSpan(p.filename, keyword.Span.Start, keyword.Span.End)}
	p.expect(TokenLBrace, "{", DiagUnexpectedDeclaration)
	for !p.at(TokenRBrace) && !p.at(TokenEOF) {
		p.skipIllegal()
		if !p.at(TokenIdentifier) {
			p.error(DiagUnexpectedDeclaration, p.peek().Span, "expected resolution field")
			p.advance()
			continue
		}
		field := p.advance().Value
		switch field {
		case "decision":
			resolution.Decision = p.expectString().Name
		case "stage":
			resolution.Stage = p.expectString().Name
		case "step":
			value := p.expectString()
			step, err := strconv.Atoi(value.Name)
			if err != nil {
				p.error(DiagUnexpectedDeclaration, value.Span, "resolution step must be a quoted integer")
			} else {
				resolution.Step = step
			}
		case "reason":
			resolution.Reason = p.expectString().Name
		case "decision_stage":
			resolution.DecisionStage = p.expectString().Name
		case "decision_step":
			value := p.expectString()
			step, err := strconv.Atoi(value.Name)
			if err != nil {
				p.error(DiagUnexpectedDeclaration, value.Span, "decision step must be a quoted integer")
			} else {
				resolution.DecisionStep = step
			}
		case "decision_reason":
			resolution.DecisionReason = p.expectString().Name
		case "unknown_class":
			resolution.UnknownClass = p.expectString().Name
		case "next_operation":
			resolution.NextOperation = p.expectString().Name
		case "blocked_by":
			for p.at(TokenString) {
				resolution.BlockedBy = append(resolution.BlockedBy, p.advance().Value)
			}
		case "role":
			resolution.Role = p.expectString().Name
		case "meta_operation":
			resolution.MetaOperation = p.expectString().Name
		case "proof_choice":
			resolution.ProofChoice = p.expectString().Name
		case "claim":
			resolution.Claim = p.expectString().Name
		default:
			p.error(DiagUnexpectedDeclaration, p.peek().Span, fmt.Sprintf("unknown resolution field %q", field))
			if p.at(TokenString) {
				p.advance()
			}
		}
	}
	closing := p.expect(TokenRBrace, "}", DiagUnexpectedDeclaration)
	if !closing.Span.IsEmpty() {
		resolution.Span.End = closing.Span.End
	}
	return resolution
}

func (p *Parser) skipPolicyBlock() {
	if p.at(TokenIdentifier) {
		p.advance()
	}
	if p.at(TokenLBrace) {
		depth := 0
		for !p.at(TokenEOF) {
			switch p.advance().Kind {
			case TokenLBrace:
				depth++
			case TokenRBrace:
				depth--
				if depth == 0 {
					return
				}
			}
		}
	}
}
