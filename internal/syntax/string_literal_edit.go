package syntax

import (
	"fmt"
	"math"
	"strconv"
)

// ReplaceStringLiteral replaces exactly one parsed literal, preserving every
// byte outside its source span. It never interprets replacement as source code.
func ReplaceStringLiteral(source []byte, span Span, replacement string) ([]byte, error) {
	start, end := span.Start.Offset, span.End.Offset
	if start < 0 || end <= start || end > len(source) {
		return nil, fmt.Errorf("string literal has an invalid source span")
	}
	if source[start] != '"' && source[start] != '`' {
		return nil, fmt.Errorf("source span is not a string literal")
	}
	if _, err := strconv.Unquote(string(source[start:end])); err != nil {
		return nil, fmt.Errorf("source span is not a string literal: %w", err)
	}
	quoted := strconv.Quote(replacement)
	size, err := replacementSourceSize(len(source), end-start, len(quoted))
	if err != nil {
		return nil, err
	}
	updated := make([]byte, 0, size)
	updated = append(updated, source[:start]...)
	updated = append(updated, quoted...)
	return append(updated, source[end:]...), nil
}

func replacementSourceSize(source, removed, inserted int) (int, error) {
	if source < 0 || removed < 0 || removed > source || inserted < 0 {
		return 0, fmt.Errorf("invalid source replacement sizes")
	}
	remaining := source - removed
	if inserted > math.MaxInt-remaining {
		return 0, fmt.Errorf("source replacement exceeds addressable size")
	}
	return remaining + inserted, nil
}
