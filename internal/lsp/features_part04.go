package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"unicode/utf8"
)

func featureErrorResponse(id json.RawMessage, err error, ctx context.Context) (*responseEnvelope, [][]byte, error) {
	if errors.Is(err, ErrStaleResult) {
		return responseOrNil(id, contentModified, "content modified during request"), nil, nil
	}
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}
	return responseOrNil(id, internalError, err.Error()), nil, nil
}
func symbolNamed(symbols []Symbol, name string) (Symbol, bool) {
	for _, symbol := range symbols {
		if symbol.Name == name {
			return symbol, true
		}
	}
	return Symbol{}, false
}
func uniqueCompletionItems(items []CompletionItem) []CompletionItem {
	result := make([]CompletionItem, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if _, exists := seen[item.Label]; exists {
			continue
		}
		seen[item.Label] = struct{}{}
		result = append(result, item)
	}
	return result
}
func wordAt(source string, position Position) (string, int, int, bool) {
	offset, err := PositionToOffset(source, position)
	if err != nil {
		return "", 0, 0, false
	}
	start, end := identifierBounds(source, offset)
	if start == end {
		return "", 0, 0, false
	}
	return source[start:end], start, end, true
}
func identifierBounds(source string, offset int) (int, int) {
	if offset > 0 && (offset == len(source) || !identifierAt(source, offset)) {
		_, size := utf8.DecodeLastRuneInString(source[:offset])
		if size > 0 && identifierAt(source, offset-size) {
			offset -= size
		}
	}
	start := offset
	for start > 0 {
		value, size := utf8.DecodeLastRuneInString(source[:start])
		if !isIdentifier(value) {
			break
		}
		start -= size
	}
	end := offset
	for end < len(source) {
		value, size := utf8.DecodeRuneInString(source[end:])
		if !isIdentifier(value) {
			break
		}
		end += size
	}
	return start, end
}
