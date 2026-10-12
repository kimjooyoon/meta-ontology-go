package syntax

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func (p *Parser) parseAssemblyObservation() *assemblyspec.Observation {
	o := &assemblyspec.Observation{}
	p.expect(TokenLBrace, "{", DiagUnexpectedDeclaration)
	seen := make(map[string]bool)
	for !p.at(TokenRBrace) && !p.at(TokenEOF) {
		field := p.expectIdentifier("observe field", DiagExpectedIdentifier)
		if field.Name != "input" && field.Name != "value_input" && field.Name != "criteria_input" && seen[field.Name] {
			p.error(DiagUnexpectedDeclaration, field.Span, "duplicate observe field "+field.Name)
		}
		seen[field.Name] = true
		switch field.Name {
		case "input":
			input := p.assemblyInteger("observe input")
			if len(o.Inputs) == 32 {
				p.error(DiagUnexpectedDeclaration, field.Span, "observe exceeds 32 inputs")
			} else {
				o.Inputs = append(o.Inputs, input)
			}
		case "value_input":
			input := p.assemblyValue()
			if len(o.ValueInputs) == 32 {
				p.error(DiagUnexpectedDeclaration, field.Span, "observe exceeds 32 value inputs")
			} else {
				o.ValueInputs = append(o.ValueInputs, input)
			}
		case "criteria_input":
			input := p.assemblyValue()
			if len(o.CriteriaInputs) == 32 {
				p.error(DiagUnexpectedDeclaration, field.Span, "observe exceeds 32 criteria inputs")
			} else {
				o.CriteriaInputs = append(o.CriteriaInputs, input)
			}
		case "oracle":
			name := p.expectString()
			if name.Name == "" {
				p.error(DiagUnexpectedDeclaration, name.Span, "observe oracle must be nonempty when declared")
			}
			o.OracleActivity = name.Name
		case "max_candidates":
			value := p.assemblyInteger(field.Name)
			if value < 1 || value > 64 {
				p.error(DiagUnexpectedDeclaration, field.Span, "max_candidates must be 1..64")
			} else {
				o.MaxCandidates = int(value)
			}
		case "max_rounds":
			value := p.assemblyInteger(field.Name)
			if value < 1 || value > 8 {
				p.error(DiagUnexpectedDeclaration, field.Span, "max_rounds must be 1..8")
			} else {
				o.MaxRounds = int(value)
			}
		case "reuse_outputs", "resolve_unique":
			value := p.expectString()
			if value.Name != "true" && value.Name != "false" {
				p.error(DiagUnexpectedDeclaration, value.Span, "observe flag expects quoted true or false")
			}
			if field.Name == "reuse_outputs" {
				o.ReuseProbeOutputs = value.Name == "true"
			} else {
				o.ResolveUniqueCandidate = value.Name == "true"
			}
		default:
			p.error(DiagUnexpectedDeclaration, field.Span, "unknown observe field "+field.Name)
			p.advance()
		}
	}
	p.expect(TokenRBrace, "}", DiagUnexpectedDeclaration)
	return o
}

func formatAssemblyObservation(out *strings.Builder, o *assemblyspec.Observation) {
	if o == nil {
		return
	}
	out.WriteString("    observe {\n")
	for _, input := range o.Inputs {
		fmt.Fprintf(out, "        input %s\n", quoteString(strconv.FormatInt(input, 10)))
	}
	for _, input := range o.ValueInputs {
		fmt.Fprintf(out, "        value_input %s\n", quoteString(input))
	}
	for _, input := range o.CriteriaInputs {
		fmt.Fprintf(out, "        criteria_input %s\n", quoteString(input))
	}
	if o.OracleActivity != "" {
		fmt.Fprintf(out, "        oracle %s\n", quoteString(o.OracleActivity))
	}
	fmt.Fprintf(out, "        max_candidates %s\n        max_rounds %s\n",
		quoteString(strconv.Itoa(o.MaxCandidates)), quoteString(strconv.Itoa(o.MaxRounds)))
	if o.ReuseProbeOutputs {
		out.WriteString("        reuse_outputs \"true\"\n")
	}
	if o.ResolveUniqueCandidate {
		out.WriteString("        resolve_unique \"true\"\n")
	}
	out.WriteString("    }\n")
}
