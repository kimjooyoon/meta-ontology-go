package syntax

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func (p *Parser) parseAssemblyCriteria() *assemblyspec.Criteria {
	criteria := &assemblyspec.Criteria{}
	p.expect(TokenLBrace, "{", DiagUnexpectedDeclaration)
	for !p.at(TokenRBrace) && !p.at(TokenEOF) {
		field := p.expectIdentifier("criteria field", DiagExpectedIdentifier)
		switch field.Name {
		case "predicate":
			id := p.expectString().Name
			p.expect(TokenActivity, "activity", DiagUnexpectedDeclaration)
			activity := p.expectString().Name
			if len(criteria.Predicates) == 8 {
				p.error(DiagUnexpectedDeclaration, field.Span, "criteria exceeds eight predicates")
			} else {
				criteria.Predicates = append(criteria.Predicates, assemblyspec.Criterion{ID: id, Activity: activity})
			}
		case "input", "holdout_input":
			input := p.assemblyValue()
			if len(criteria.Inputs)+len(criteria.HoldoutInputs) == 128 {
				p.error(DiagUnexpectedDeclaration, field.Span, "criteria exceeds 128 input tuples")
			} else if field.Name == "input" {
				criteria.Inputs = append(criteria.Inputs, input)
			} else {
				criteria.HoldoutInputs = append(criteria.HoldoutInputs, input)
			}
		default:
			p.error(DiagUnexpectedDeclaration, field.Span, "unknown criteria field "+field.Name)
			p.advance()
		}
	}
	p.expect(TokenRBrace, "}", DiagUnexpectedDeclaration)
	return criteria
}

func formatAssemblyCriteria(out *strings.Builder, criteria *assemblyspec.Criteria) {
	if criteria == nil {
		return
	}
	out.WriteString("    criteria {\n")
	for _, predicate := range criteria.Predicates {
		fmt.Fprintf(out, "        predicate %s activity %s\n", quoteString(predicate.ID), quoteString(predicate.Activity))
	}
	for _, input := range criteria.Inputs {
		fmt.Fprintf(out, "        input %s\n", quoteString(input))
	}
	for _, input := range criteria.HoldoutInputs {
		fmt.Fprintf(out, "        holdout_input %s\n", quoteString(input))
	}
	out.WriteString("    }\n")
}
