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
		if symbol.Name != name || symbol.Kind != SymbolFunction {
			continue
		}
		if id == "" {
			return nil, true, fmt.Errorf("activity rename requires a resolved reference; start at the declaration")
		}
		if symbol.ID == id {
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

func activityCallRenameTarget(document document, uri string, position Position) (bodycodegen.SourceActivityReference, bool, error) {
	offset, err := PositionToOffset(document.text, position)
	if err != nil {
		return bodycodegen.SourceActivityReference{}, false, err
	}
	return bodycodegen.ResolveActivityReference(context.Background(), uri, []byte(document.text), offset)
}
