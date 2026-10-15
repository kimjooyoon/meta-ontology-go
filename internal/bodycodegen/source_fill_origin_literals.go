package bodycodegen

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type fillOriginLiteral struct {
	value string
	span  FillOriginRange
}

type fillOriginIndex struct {
	intent      FillOriginRange
	candidate   *FillOriginRange
	holes       map[string]FillOriginRange
	expressions map[string]fillOriginLiteral
	grammars    map[string]fillOriginLiteral
}

func fillLiteralOrigin(token syntax.Token, base int) (fillOriginLiteral, error) {
	if token.Kind != syntax.TokenString {
		return fillOriginLiteral{}, fmt.Errorf("source fill origin expects a source literal")
	}
	start, end, err := SourceLiteralByteRange(token.Text, token.Value, 0, len(token.Value))
	return fillOriginLiteral{token.Value, FillOriginRange{base + token.Span.Start.Offset + start, base + token.Span.Start.Offset + end}}, err
}

func sourceFillLiteralOrigins(source []byte, activity *syntax.ActivityDecl, selected string) (fillOriginIndex, error) {
	index := fillOriginIndex{holes: map[string]FillOriginRange{}, expressions: map[string]fillOriginLiteral{}, grammars: map[string]fillOriginLiteral{}}
	span := activity.Assembly.Span
	base := span.Start.Offset
	tokens, diagnostics := syntax.Lex(string(source[base:span.End.Offset]))
	if diagnostics.HasErrors() {
		return index, diagnostics.Error()
	}
	for i, token := range tokens {
		if token.Kind == syntax.TokenIdentifier && token.Value == "source_fill" && i+3 < len(tokens) {
			intent, err := fillLiteralOrigin(tokens[i+2], base)
			if err != nil || intent.value != activity.Assembly.Spec.FillPlan.Intent {
				return index, fmt.Errorf("source fill origin intent differs")
			}
			index.intent = intent.span
			err = scanFillOriginBlock(tokens[i+3:], base, selected, &index)
			return index, err
		}
	}
	return index, fmt.Errorf("source fill origin block is absent")
}

func scanFillOriginBlock(tokens syntax.Tokens, base int, selected string, index *fillOriginIndex) error {
	depth, candidate, grammarHole := 0, "", ""
	for i, token := range tokens {
		switch token.Kind {
		case syntax.TokenLBrace:
			depth++
		case syntax.TokenRBrace:
			depth--
			if depth == 0 {
				return nil
			}
			if depth == 1 {
				candidate, grammarHole = "", ""
			}
		case syntax.TokenIdentifier:
			if i+1 >= len(tokens) {
				return fmt.Errorf("source fill origin has an incomplete declaration")
			}
			next := tokens[i+1]
			switch {
			case depth == 1 && token.Value == "candidate":
				candidate = next.Value
				if candidate == selected {
					literal, err := fillLiteralOrigin(next, base)
					if err != nil {
						return err
					}
					index.candidate = &literal.span
				}
			case token.Value == "hole" && depth == 1:
				index.holes[next.Value] = FillOriginRange{base + token.Span.Start.Offset, base + next.Span.End.Offset}
			case token.Value == "hole" && depth == 2 && candidate == "":
				grammarHole = next.Value
			case token.Value == "grammar" && candidate == "" && (depth == 1 || depth == 2):
				literal, err := fillLiteralOrigin(next, base)
				if err != nil {
					return err
				}
				index.grammars[grammarHole] = literal
			case token.Value == "fill" && depth == 2 && candidate == selected && i+2 < len(tokens):
				literal, err := fillLiteralOrigin(tokens[i+2], base)
				if err != nil {
					return err
				}
				index.expressions[next.Value] = literal
			}
		}
	}
	return fmt.Errorf("source fill origin block is unterminated")
}
