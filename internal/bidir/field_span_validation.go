package bidir

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

func validateExactFieldSpans(field Field) error {
	spans := []struct {
		name string
		span SourceSpan
	}{
		{name: "field", span: field.Span},
		{name: "field ID", span: field.IDSpan},
		{name: "field name", span: field.NameSpan},
		{name: "field type", span: field.TypeRefSpan},
		{name: "field presence", span: field.PresenceSpan},
		{name: "field cardinality", span: field.CardinalitySpan},
	}
	for _, item := range spans {
		if err := item.span.Validate(); err != nil {
			return fmt.Errorf("%s span is invalid: %v", item.name, err)
		}
		if strings.TrimSpace(item.span.File) == "" || item.span.End <= item.span.Start {
			return fmt.Errorf("%s span is missing exact source provenance", item.name)
		}
	}
	if field.Span.File != field.IDSpan.File || field.Span.File != field.NameSpan.File || field.Span.File != field.TypeRefSpan.File || field.Span.File != field.PresenceSpan.File || field.Span.File != field.CardinalitySpan.File {
		return errors.New("field subspans cross source snapshots")
	}
	ordered := []SourceSpan{field.IDSpan, field.NameSpan, field.TypeRefSpan, field.PresenceSpan, field.CardinalitySpan}
	sort.SliceStable(ordered, func(left, right int) bool {
		return ordered[left].Start < ordered[right].Start
	})
	for index, span := range ordered {
		if span.Start < field.Span.Start || span.End > field.Span.End {
			return errors.New("field subspan is outside the aggregate field span")
		}
		if index > 0 && ordered[index-1].End > span.Start {
			return errors.New("field subspans overlap or are ambiguous")
		}
	}
	return nil
}
