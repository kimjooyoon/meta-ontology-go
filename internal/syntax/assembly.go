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
	seenAttempts, seenSeed, seenBaseline, seenSearch, seenFillPlan := false, false, false, false, false
	for !p.at(TokenRBrace) && !p.at(TokenEOF) {
		if !p.at(TokenIdentifier) {
			p.error(DiagUnexpectedDeclaration, p.peek().Span, "expected an assembly field")
			p.advance()
			continue
		}
		field := p.advance()
		if p.parseAssemblyCheckpointField(d, field, &seenBaseline) {
			continue
		}
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
		case "holdout_case":
			input := p.assemblyInteger("holdout case input")
			p.expect(TokenArrow, "->", DiagExpectedArrow)
			expected := p.assemblyInteger("holdout case expected value")
			if len(d.Spec.HoldoutCases) == 128 {
				p.error(DiagUnexpectedDeclaration, field.Span, "assembly exceeds 128 holdout cases")
				p.skipAssemblyRemainder()
				continue
			}
			d.Spec.HoldoutCases = append(d.Spec.HoldoutCases, assemblyspec.Case{Input: input, Expected: expected})
		case "search":
			if seenSearch {
				p.error(DiagUnexpectedDeclaration, field.Span, "duplicate assembly IR search")
			}
			seenSearch = true
			d.Spec.Search = p.parseAssemblySearch()
		case "source_fill":
			if seenFillPlan {
				p.error(DiagUnexpectedDeclaration, field.Span, "duplicate source body-fill plan")
			}
			seenFillPlan = true
			d.Spec.FillPlan = p.parseAssemblyFillPlan()
		case "value_case":
			input := p.assemblyValue()
			p.expect(TokenArrow, "->", DiagExpectedArrow)
			expected := p.assemblyValue()
			if len(d.Spec.ValueCases) == 128 {
				p.error(DiagUnexpectedDeclaration, field.Span, "assembly exceeds 128 value cases")
				p.skipAssemblyRemainder()
				continue
			}
			d.Spec.ValueCases = append(d.Spec.ValueCases, assemblyspec.ValueCase{Inputs: input, Expected: expected})
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

func (p *Parser) parseAssemblySearch() *assemblyspec.Search {
	p.assemblyKeyword("hole")
	search := &assemblyspec.Search{HoleID: p.expectString().Name}
	p.assemblyKeyword("grammar")
	search.Grammar = p.expectString().Name
	p.assemblyKeyword("intent")
	search.Intent = p.expectString().Name
	p.assemblyKeyword("max_candidates")
	value := p.assemblyInteger("maximum candidate count")
	if value < 2 || value > 16 {
		p.error(DiagUnexpectedDeclaration, p.peek().Span, "IR search max_candidates must be 2..16")
		return search
	}
	search.MaxCandidates = int(value)
	return search
}

func (p *Parser) parseAssemblyFillPlan() *assemblyspec.FillPlan {
	p.assemblyKeyword("intent")
	plan := &assemblyspec.FillPlan{Intent: p.expectString().Name}
	p.expect(TokenLBrace, "{", DiagUnexpectedDeclaration)
	seenDerive := false
	for !p.at(TokenRBrace) && !p.at(TokenEOF) {
		field := p.expectIdentifier("source fill declaration", DiagExpectedIdentifier)
		switch field.Name {
		case "hole":
			if len(plan.Holes) == 8 {
				p.error(DiagUnexpectedDeclaration, field.Span, "source fill plan exceeds 8 holes")
				p.skipAssemblyRemainder()
				continue
			}
			plan.Holes = append(plan.Holes, assemblyspec.FillHole{ID: p.expectString().Name})
		case "derive":
			if seenDerive {
				p.error(DiagUnexpectedDeclaration, field.Span, "duplicate source fill derive clause")
			}
			seenDerive = true
			generation := &assemblyspec.FillGeneration{}
			if p.at(TokenIdentifier) && p.peek().Value == "assignments" {
				p.advance()
				p.assemblyKeyword("max_candidates")
				maxCandidates := p.assemblyInteger("maximum complete candidates")
				if maxCandidates < 2 || maxCandidates > 16 {
					p.error(DiagUnexpectedDeclaration, field.Span, "source fill max_candidates must be 2..16")
				} else {
					generation.MaxCandidates = int(maxCandidates)
				}
				p.expect(TokenLBrace, "{", DiagUnexpectedDeclaration)
				for !p.at(TokenRBrace) && !p.at(TokenEOF) {
					grammarField := p.expectIdentifier("per-hole grammar", DiagExpectedIdentifier)
					if grammarField.Name != "hole" {
						p.error(DiagUnexpectedDeclaration, grammarField.Span, "expected hole grammar")
						p.advance()
						continue
					}
					holeGrammar := assemblyspec.FillHoleGrammar{HoleID: p.expectString().Name}
					p.assemblyKeyword("grammar")
					holeGrammar.Grammar = p.expectString().Name
					p.assemblyKeyword("max_expressions")
					maxExpressions := p.assemblyInteger("maximum expressions for hole")
					if maxExpressions < 2 || maxExpressions > 16 {
						p.error(DiagUnexpectedDeclaration, grammarField.Span, "source fill max_expressions must be 2..16")
					} else {
						holeGrammar.MaxExpressions = int(maxExpressions)
					}
					generation.HoleGrammars = append(generation.HoleGrammars, holeGrammar)
				}
				p.expect(TokenRBrace, "}", DiagUnexpectedDeclaration)
			} else {
				p.assemblyKeyword("grammar")
				generation.Grammar = p.expectString().Name
				p.assemblyKeyword("max_expressions")
				maxExpressions := p.assemblyInteger("maximum expressions per hole")
				if maxExpressions < 2 || maxExpressions > 16 {
					p.error(DiagUnexpectedDeclaration, field.Span, "source fill max_expressions must be 2..16")
				} else {
					generation.MaxExpressions = int(maxExpressions)
				}
				p.assemblyKeyword("max_candidates")
				maxCandidates := p.assemblyInteger("maximum complete candidates")
				if maxCandidates < 2 || maxCandidates > 16 {
					p.error(DiagUnexpectedDeclaration, field.Span, "source fill max_candidates must be 2..16")
				} else {
					generation.MaxCandidates = int(maxCandidates)
				}
			}
			plan.Generation = generation
		case "candidate":
			if len(plan.Candidates) == 16 {
				p.error(DiagUnexpectedDeclaration, field.Span, "source fill plan exceeds 16 candidates")
				p.skipAssemblyRemainder()
				continue
			}
			candidate := assemblyspec.FillCandidate{ID: p.expectString().Name}
			p.expect(TokenLBrace, "{", DiagUnexpectedDeclaration)
			for !p.at(TokenRBrace) && !p.at(TokenEOF) {
				fillField := p.expectIdentifier("candidate fill", DiagExpectedIdentifier)
				if fillField.Name != "fill" {
					p.error(DiagUnexpectedDeclaration, fillField.Span, "expected candidate fill")
					p.advance()
					continue
				}
				candidate.Fills = append(candidate.Fills, assemblyspec.HoleFilling{
					HoleID: p.expectString().Name, Expression: p.expectString().Name,
				})
			}
			p.expect(TokenRBrace, "}", DiagUnexpectedDeclaration)
			plan.Candidates = append(plan.Candidates, candidate)
		default:
			p.error(DiagUnexpectedDeclaration, field.Span, "unknown source fill declaration "+field.Name)
		}
	}
	p.expect(TokenRBrace, "}", DiagUnexpectedDeclaration)
	return plan
}

func (p *Parser) parseAssemblyCheckpointField(d *AssemblyDecl, field Token, seenBaseline *bool) bool {
	switch field.Value {
	case "baseline":
		if *seenBaseline {
			p.error(DiagUnexpectedDeclaration, field.Span, "duplicate assembly baseline")
		}
		*seenBaseline, d.Spec.Baseline = true, p.expectString().Name
		if strings.TrimSpace(d.Spec.Baseline) == "" {
			p.error(DiagUnexpectedDeclaration, field.Span, "assembly baseline must be nonempty")
		}
	case "picked":
		id := p.expectString().Name
		p.expect(TokenArrow, "->", DiagExpectedArrow)
		label := p.expectString().Name
		if len(d.Spec.Picked) == 16 {
			p.error(DiagUnexpectedDeclaration, field.Span, "assembly exceeds 16 picked labels")
			p.skipAssemblyRemainder()
			return true
		}
		d.Spec.Picked = append(d.Spec.Picked, assemblyspec.Pick{ID: id, Label: label})
	default:
		return false
	}
	return true
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

func (p *Parser) assemblyValue() string {
	token := p.expectString()
	value, err := assemblyspec.CanonicalValue(token.Name)
	if err != nil {
		p.error(DiagUnexpectedDeclaration, token.Span, err.Error())
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
	if d.Spec.Baseline != "" {
		fmt.Fprintf(output, "    baseline %s\n", quoteString(d.Spec.Baseline))
		for _, pick := range d.Spec.Picked {
			fmt.Fprintf(output, "    picked %s -> %s\n", quoteString(pick.ID), quoteString(pick.Label))
		}
	}
	for _, c := range d.Spec.Choices {
		fmt.Fprintf(output, "    choice %s %s at %s", quoteString(c.ID), c.Kind, quoteString(strconv.Itoa(c.Occurrence)))
		if c.Alternative != "" {
			fmt.Fprintf(output, " alternative %s", quoteString(c.Alternative))
		}
		fmt.Fprintf(output, " intent %s\n", quoteString(c.Intent))
	}
	if d.Spec.Search != nil {
		search := d.Spec.Search
		fmt.Fprintf(output, "    search hole %s grammar %s intent %s max_candidates %s\n",
			quoteString(search.HoleID), quoteString(search.Grammar), quoteString(search.Intent),
			quoteString(strconv.Itoa(search.MaxCandidates)))
	}
	if d.Spec.FillPlan != nil {
		plan := d.Spec.FillPlan
		fmt.Fprintf(output, "    source_fill intent %s {\n", quoteString(plan.Intent))
		for _, hole := range plan.Holes {
			fmt.Fprintf(output, "        hole %s\n", quoteString(hole.ID))
		}
		if plan.Generation != nil {
			if len(plan.Generation.HoleGrammars) != 0 {
				fmt.Fprintf(output, "        derive assignments max_candidates %s {\n", quoteString(strconv.Itoa(plan.Generation.MaxCandidates)))
				for _, grammar := range plan.Generation.HoleGrammars {
					fmt.Fprintf(output, "            hole %s grammar %s max_expressions %s\n",
						quoteString(grammar.HoleID), quoteString(grammar.Grammar), quoteString(strconv.Itoa(grammar.MaxExpressions)))
				}
				output.WriteString("        }\n")
			} else {
				fmt.Fprintf(output, "        derive grammar %s max_expressions %s max_candidates %s\n",
					quoteString(plan.Generation.Grammar), quoteString(strconv.Itoa(plan.Generation.MaxExpressions)),
					quoteString(strconv.Itoa(plan.Generation.MaxCandidates)))
			}
		}
		for _, candidate := range plan.Candidates {
			fmt.Fprintf(output, "        candidate %s {\n", quoteString(candidate.ID))
			for _, fill := range candidate.Fills {
				fmt.Fprintf(output, "            fill %s %s\n", quoteString(fill.HoleID), quoteString(fill.Expression))
			}
			output.WriteString("        }\n")
		}
		output.WriteString("    }\n")
	}
	for _, c := range d.Spec.Cases {
		fmt.Fprintf(output, "    case %s -> %s\n", quoteString(strconv.FormatInt(c.Input, 10)),
			quoteString(strconv.FormatInt(c.Expected, 10)))
	}
	for _, c := range d.Spec.HoldoutCases {
		fmt.Fprintf(output, "    holdout_case %s -> %s\n", quoteString(strconv.FormatInt(c.Input, 10)),
			quoteString(strconv.FormatInt(c.Expected, 10)))
	}
	for _, c := range d.Spec.ValueCases {
		fmt.Fprintf(output, "    value_case %s -> %s\n", quoteString(c.Inputs), quoteString(c.Expected))
	}
	if d.Spec.FillPlan == nil {
		fmt.Fprintf(output, "    attempts %s\n", quoteString(strconv.Itoa(d.Spec.MaxAttempts)))
	}
	if d.Spec.Seed != "" {
		fmt.Fprintf(output, "    seed %s\n", quoteString(d.Spec.Seed))
	}
	output.WriteByte('}')
	return nil
}

// FormatAssembly renders only the contract, for a locality-preserving source edit.
func FormatAssembly(d *AssemblyDecl) (string, error) {
	var output strings.Builder
	if err := formatAssembly(&output, d); err != nil {
		return "", err
	}
	return strings.TrimPrefix(output.String(), " "), nil
}
