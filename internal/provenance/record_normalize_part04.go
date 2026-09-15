package provenance

import (
	"fmt"
	"strings"
)

func normalizeSourceSpan(span SourceSpan) (SourceSpan, error) {
	if span.URI == "" {
		span.URI = span.File
	}
	if span.URI == "" {
		return SourceSpan{}, fmt.Errorf("source_span.uri is required")
	}
	var err error
	span.URI, err = normalizeIdentifier(span.URI, "source_span.uri")
	if err != nil {
		return SourceSpan{}, err
	}
	span.File = ""
	if err := normalizePosition(span.Start, "source_span.start"); err != nil {
		return SourceSpan{}, err
	}
	if err := normalizePosition(span.End, "source_span.end"); err != nil {
		return SourceSpan{}, err
	}
	if positionAfter(span.Start, span.End) {
		return SourceSpan{}, fmt.Errorf("source_span.end precedes source_span.start")
	}
	return span, nil
}
func normalizePosition(position Position, field string) error {
	if position.Offset < 0 || position.Line < 1 || position.Column < 1 {
		return fmt.Errorf("%s must have offset >= 0 and positive line/column", field)
	}
	return nil
}
func positionAfter(left, right Position) bool {
	if left.Offset != right.Offset {
		return left.Offset > right.Offset
	}
	if left.Line != right.Line {
		return left.Line > right.Line
	}
	return left.Column > right.Column
}
func normalizeAttributes(attributes map[string]string) (map[string]string, error) {
	if len(attributes) == 0 {
		return nil, nil
	}
	result := make(map[string]string, len(attributes))
	for key, value := range attributes {
		key = strings.TrimSpace(key)
		if key == "" || strings.ContainsAny(key, "\r\n") {
			return nil, fmt.Errorf("attribute keys must be non-empty and line-free")
		}
		if _, exists := result[key]; exists {
			return nil, fmt.Errorf("attribute key %q is duplicated after normalization", key)
		}
		if strings.ContainsAny(value, "\r\n") {
			return nil, fmt.Errorf("attribute %q must not contain line breaks", key)
		}
		result[key] = value
	}
	return result, nil
}
