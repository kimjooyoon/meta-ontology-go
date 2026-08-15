package generator

import "strings"

type entityFieldSourceSpan struct {
	name string
	span SourceSpan
}

func validateFieldSource(entity Entity, field Field, sourceURI string, previousStart int, hasPrevious bool) error {
	span := field.Source
	if !validSourceSpan(span) {
		return entityFieldsError(entityFieldsIncompleteDiagnostic, field, "source span must be a non-zero, half-open source range")
	}
	if sourceURI != "" && span.URI != sourceURI {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, "field origins cross source snapshots")
	}
	if hasPrevious && span.Start.Offset <= previousStart {
		return entityFieldsError(entityFieldsIllegalReorderDiagnostic, field, "field declaration order is not source ordered")
	}
	if entity.Source.URI != "" && entity.Source.URI != span.URI {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, "field origin does not match its entity source snapshot")
	}
	spans := requiredEntityFieldSourceSpans(field)
	for _, subspan := range spans {
		if sourceSpanIsZero(subspan.span) {
			return entityFieldsError(entityFieldsIncompleteDiagnostic, field, subspan.name+" source span is required")
		}
		if err := validateFieldSubspan(field, subspan.span, subspan.name); err != nil {
			return err
		}
	}
	if !sourceSpansAreOrdered(spans) {
		return entityFieldsError(entityFieldsIllegalReorderDiagnostic, field, "field subspans are not in lexical declaration order")
	}
	if !sourceSpanIsZero(field.NameSource) && field.NameSource != field.NameSpan {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, "name source spans disagree")
	}
	return nil
}

func requiredEntityFieldSourceSpans(field Field) []entityFieldSourceSpan {
	return []entityFieldSourceSpan{
		{name: "ID", span: field.IDSpan},
		{name: "name", span: field.NameSpan},
		{name: "type", span: field.TypeRefSpan},
		{name: "presence", span: field.PresenceSpan},
		{name: "cardinality", span: field.CardinalitySpan},
	}
}

func sourceSpansAreOrdered(spans []entityFieldSourceSpan) bool {
	var previous SourceSpan
	for index, current := range spans {
		if index > 0 && (!sourceSpanIsZero(current.span) && !sourceSpanIsZero(previous) && previous.End.Offset > current.span.Start.Offset) {
			return false
		}
		previous = current.span
	}
	return true
}

func validSourceSpan(span SourceSpan) bool {
	return strings.TrimSpace(span.URI) != "" && validSourcePosition(span.Start) && validSourcePosition(span.End) && span.End.Offset > span.Start.Offset && sourceLocationBefore(span.Start, span.End)
}

func sourceSpanIsZero(span SourceSpan) bool {
	return span == (SourceSpan{})
}

func validateFieldSubspan(field Field, subspan SourceSpan, label string) error {
	if !validSourceSpan(subspan) || subspan.URI != field.Source.URI || !sourcePositionWithin(subspan.Start, field.Source.Start, field.Source.End) || !sourcePositionWithin(subspan.End, field.Source.Start, field.Source.End) || subspan.Start.Offset < field.Source.Start.Offset || subspan.End.Offset > field.Source.End.Offset {
		return entityFieldsError(entityFieldsUnrepresentableDiagnostic, field, label+" source span is not an exact subspan of the field origin")
	}
	return nil
}

func validSourcePosition(position Position) bool {
	return position.Line > 0 && position.Column > 0 && position.Offset >= 0
}

func sourceLocationBefore(left, right Position) bool {
	if left.Line != right.Line {
		return left.Line < right.Line
	}
	return left.Column < right.Column
}

func sourcePositionWithin(position, start, end Position) bool {
	return position.Offset >= start.Offset && position.Offset <= end.Offset && !sourceLocationBefore(position, start) && !sourceLocationBefore(end, position)
}

func fieldNameSource(field Field) SourceSpan {
	return field.NameSpan
}
