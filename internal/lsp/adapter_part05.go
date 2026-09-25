package lsp

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func appendReference(result *ParseResult, source, name string, span syntax.Span, id string) error {
	if name == "" || span.IsEmpty() {
		return nil
	}
	rangeValue, err := syntaxRange(source, span)
	if err != nil {
		return err
	}
	result.References = append(result.References, Reference{Name: name, ID: id, Range: rangeValue})
	return nil
}
func syntaxRange(source string, span syntax.Span) (Range, error) {
	start, err := OffsetToPosition(source, span.Start.Offset)
	if err != nil {
		return Range{}, fmt.Errorf("lsp: invalid syntax start span: %w", err)
	}
	end, err := OffsetToPosition(source, span.End.Offset)
	if err != nil {
		return Range{}, fmt.Errorf("lsp: invalid syntax end span: %w", err)
	}
	return Range{Start: start, End: end}, nil
}
