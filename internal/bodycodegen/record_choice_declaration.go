package bodycodegen

import (
	"fmt"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type RecordChoiceDeclarationOrigin struct {
	ChoiceID          string
	Kind              string
	Occurrence        int
	Alternative       string
	AlternativeSource FillOriginRange
	Intent            string
	IntentSource      FillOriginRange
	DeclarationSource FillOriginRange
}

func InspectRecordChoiceDeclarationOrigins(filename string, source []byte, activity string) ([]RecordChoiceDeclarationOrigin, error) {
	if len(source) > 128<<10 {
		return nil, fmt.Errorf("record choice declaration source exceeds 128 KiB")
	}
	decl, err := sourceFillActivity(filename, source, activity)
	if err != nil || decl.Assembly == nil {
		return nil, fmt.Errorf("record choice activity declaration unavailable")
	}
	span := decl.Assembly.Span
	tokens, diagnostics := syntax.Lex(string(source[span.Start.Offset:span.End.Offset]))
	if diagnostics.HasErrors() {
		return nil, diagnostics.Error()
	}
	var result []RecordChoiceDeclarationOrigin
	depth := 0
	for i, token := range tokens {
		switch token.Kind {
		case syntax.TokenLBrace:
			depth++
		case syntax.TokenRBrace:
			depth--
		case syntax.TokenIdentifier:
			if depth == 1 && token.Value == "choice" {
				row, err := recordChoiceLiteralOrigin(tokens[i:], span.Start.Offset)
				if err != nil {
					return nil, err
				}
				result = append(result, row)
			}
		}
	}
	if len(result) != len(decl.Assembly.Spec.Choices) {
		return nil, fmt.Errorf("record choice declaration population differs")
	}
	for i, row := range result {
		expected := decl.Assembly.Spec.Choices[i]
		if row.ChoiceID != expected.ID || row.Kind != expected.Kind || row.Occurrence != expected.Occurrence || row.Alternative != expected.Alternative || row.Intent != expected.Intent {
			return nil, fmt.Errorf("record choice literal differs from parsed declaration")
		}
	}
	return result, nil
}

func recordChoiceLiteralOrigin(tokens syntax.Tokens, base int) (RecordChoiceDeclarationOrigin, error) {
	if len(tokens) < 9 || tokens[3].Value != "at" || tokens[5].Value != "alternative" || tokens[7].Value != "intent" {
		return RecordChoiceDeclarationOrigin{}, fmt.Errorf("record field choice literal shape differs")
	}
	alternative, err := fillLiteralOrigin(tokens[6], base)
	if err != nil {
		return RecordChoiceDeclarationOrigin{}, err
	}
	intent, err := fillLiteralOrigin(tokens[8], base)
	if err != nil {
		return RecordChoiceDeclarationOrigin{}, err
	}
	occurrence, err := strconv.Atoi(tokens[4].Value)
	if err != nil || tokens[1].Kind != syntax.TokenString || occurrence < 0 || occurrence > 127 {
		return RecordChoiceDeclarationOrigin{}, fmt.Errorf("record choice identity or occurrence differs")
	}
	return RecordChoiceDeclarationOrigin{ChoiceID: tokens[1].Value, Kind: tokens[2].Value, Occurrence: occurrence,
		Alternative: alternative.value, AlternativeSource: alternative.span, Intent: intent.value, IntentSource: intent.span,
		DeclarationSource: FillOriginRange{base + tokens[0].Span.Start.Offset, base + tokens[8].Span.End.Offset}}, nil
}
