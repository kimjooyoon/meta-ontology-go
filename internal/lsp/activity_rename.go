package lsp

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func activityRenameEdits(document document, uri, id, name, newName string) ([]TextEdit, bool, error) {
	activity := false
	for _, symbol := range allSymbols(document.result) {
		if symbol.ID == id && symbol.Name == name && symbol.Kind == SymbolFunction {
			activity = true
		}
	}
	if !activity {
		return nil, false, nil
	}
	result, err := bodycodegen.RenameActivitySource(context.Background(), uri, []byte(document.text), name, newName)
	if err != nil {
		return nil, true, err
	}
	if result.ActivityID != id {
		return nil, true, fmt.Errorf("activity identity changed since document analysis")
	}
	edits := make([]TextEdit, 0, len(result.Edits))
	for _, edit := range result.Edits {
		span := syntax.Span{Start: syntax.Position{Offset: edit.Start}, End: syntax.Position{Offset: edit.End}}
		rangeValue, err := syntaxRange(document.text, span)
		if err != nil {
			return nil, true, err
		}
		edits = append(edits, TextEdit{Range: rangeValue, NewText: edit.After})
	}
	return edits, true, nil
}
