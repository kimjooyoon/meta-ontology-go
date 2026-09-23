package lsp

import "strings"

// completionExpectedSymbolKind derives the narrow declaration context from
// source text while keeping semantic candidates owned by the parsed IR. It is
// intentionally conservative: only activity entity positions are narrowed;
// all other positions retain the existing keyword and symbol vocabulary.
func completionExpectedSymbolKind(source string, position Position, usePosition bool) (SymbolKind, bool) {
	if !usePosition {
		return 0, false
	}
	offset, err := PositionToOffset(source, position)
	if err != nil || offset < 0 || offset > len(source) {
		return 0, false
	}
	lineStart := strings.LastIndexByte(source[:offset], '\n') + 1
	prefix := strings.TrimSpace(source[lineStart:offset])
	if !strings.HasPrefix(prefix, "activity ") {
		return 0, false
	}
	open := strings.LastIndexByte(prefix, '(')
	close := strings.LastIndexByte(prefix, ')')
	if open >= 0 && close < open {
		return SymbolClass, true
	}
	if arrow := strings.LastIndex(prefix, "->"); arrow >= 0 {
		return SymbolClass, true
	}
	return 0, false
}
