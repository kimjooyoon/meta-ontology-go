package syntax

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

// AssemblyDecl belongs to its activity. It authorizes only declared typed paths
// and finite expectations; model choices and execution observations are separate.
type AssemblyDecl struct {
	Span Span
	Spec assemblyspec.Spec
}

func (d *AssemblyDecl) SourceSpan() Span { return d.Span }

func (p *Parser) parseAssembly() *AssemblyDecl {
	keyword := p.advance()
	d := &AssemblyDecl{Span: keyword.Span}
	p.expect(TokenLBrace, "{", DiagUnexpectedDeclaration)
	seenAttempts, seenSeed := false, false
	for !p.at(TokenRBrace) && !p.at(TokenEOF) {
		if !p.at(TokenIdentifier) {
			p.error(DiagUnexpectedDeclaration, p.peek().Span, "expected assembly choice, case, attempts or seed")
			p.advance()
			continue
		}
		field := p.advance()
		switch field.Value {
		case "choice":
			if len(d.Spec.Choices) == 16 {
				p.error(DiagUnexpectedDeclaration, field.Span, "assembly exceeds 16 choices")
				p.skipAssemblyRemainder()
				continue
			}
			d.Spec.Choices = append(d.Spec.Choices, p.parseAssemblyChoice())
		case "case":
			input := p.assemblyInteger("case input")
			p.expect(TokenArrow, "->", DiagExpectedArrow)
			expected := p.assemblyInteger("case expected value")
			if len(d.Spec.Cases) == 128 {
				p.error(DiagUnexpectedDeclaration, field.Span, "assembly exceeds 128 cases")
				p.skipAssemblyRemainder()
				continue
			}
			d.Spec.Cases = append(d.Spec.Cases, assemblyspec.Case{Input: input, Expected: expected})
		case "attempts":
			value := p.assemblyInteger("attempt budget")
			if seenAttempts || value < 1 || value > 64 {
				p.error(DiagUnexpectedDeclaration, field.Span, "assembly needs one attempts field in 1..64")
			}
			seenAttempts, d.Spec.MaxAttempts = true, int(value)
		case "seed":
			if seenSeed {
				p.error(DiagUnexpectedDeclaration, field.Span, "duplicate assembly seed")
			}
			seenSeed, d.Spec.Seed = true, p.expectString().Name
		default:
			p.error(DiagUnexpectedDeclaration, field.Span, "unknown assembly field "+field.Value)
		}
	}
	closing := p.expect(TokenRBrace, "}", DiagUnexpectedDeclaration)
	if !closing.Span.IsEmpty() {
		d.Span.End = closing.Span.End
	}
	if err := d.Spec.Validate(); err != nil {
		p.error(DiagUnexpectedDeclaration, d.Span, err.Error())
	}
	return d
}

func (p *Parser) parseAssemblyChoice() assemblyspec.Choice {
	c := assemblyspec.Choice{ID: p.expectString().Name}
	c.Kind = p.expectIdentifier("assembly choice kind", DiagExpectedIdentifier).Name
	p.assemblyKeyword("at")
	occurrence := p.assemblyInteger("choice occurrence")
	if occurrence < 0 || occurrence > 127 {
		p.error(DiagUnexpectedDeclaration, p.peek().Span, "assembly occurrence must be 0..127")
	}
	c.Occurrence = int(occurrence)
	if p.at(TokenIdentifier) && p.peek().Value == "alternative" {
		p.advance()
		c.Alternative = p.expectString().Name
	}
	p.assemblyKeyword("intent")
	c.Intent = p.expectString().Name
	return c
}

func (p *Parser) assemblyKeyword(word string) {
	token := p.expect(TokenIdentifier, word, DiagExpectedIdentifier)
	if token.Kind == TokenIdentifier && token.Value != word {
		p.error(DiagUnexpectedDeclaration, token.Span, "expected "+word)
	}
}

func (p *Parser) assemblyInteger(label string) int64 {
	token := p.expectString()
	value, err := strconv.ParseInt(token.Name, 10, 64)
	if err != nil {
		p.error(DiagUnexpectedDeclaration, token.Span, label+" must be a quoted int64 decimal")
	}
	return value
}

func (p *Parser) skipAssemblyRemainder() {
	for !p.at(TokenRBrace) && !p.at(TokenEOF) {
		p.advance()
	}
}

func formatAssembly(output *strings.Builder, d *AssemblyDecl) error {
	if d == nil {
		return nil
	}
	if err := d.Spec.Validate(); err != nil {
		return err
	}
	output.WriteString(" assembling {\n")
	for _, c := range d.Spec.Choices {
		fmt.Fprintf(output, "    choice %s %s at %s", quoteString(c.ID), c.Kind, quoteString(strconv.Itoa(c.Occurrence)))
		if c.Alternative != "" {
			fmt.Fprintf(output, " alternative %s", quoteString(c.Alternative))
		}
		fmt.Fprintf(output, " intent %s\n", quoteString(c.Intent))
	}
	for _, c := range d.Spec.Cases {
		fmt.Fprintf(output, "    case %s -> %s\n", quoteString(strconv.FormatInt(c.Input, 10)),
			quoteString(strconv.FormatInt(c.Expected, 10)))
	}
	fmt.Fprintf(output, "    attempts %s\n", quoteString(strconv.Itoa(d.Spec.MaxAttempts)))
	if d.Spec.Seed != "" {
		fmt.Fprintf(output, "    seed %s\n", quoteString(d.Spec.Seed))
	}
	output.WriteByte('}')
	return nil
}
