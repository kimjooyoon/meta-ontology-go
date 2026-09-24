package lsp

import "strings"

// completionExpectedSymbolKind derives the narrow declaration context from
// source text while keeping semantic candidates owned by the parsed IR. It is
// intentionally conservative: activity and entity-field positions are
// narrowed; all other positions retain the existing keyword and symbol
// vocabulary.
func completionExpectedSymbolKind(source string, position Position, usePosition bool) (SymbolKind, bool) {
	if !usePosition {
		return 0, false
	}
	offset, err := PositionToOffset(source, position)
	if err != nil || offset < 0 || offset > len(source) {
		return 0, false
	}
	if _, _, end, found := wordAt(source, position); found && offset < end {
		return 0, false
	}
	lineStart := strings.LastIndexByte(source[:offset], '\n') + 1
	prefix := strings.TrimSpace(source[lineStart:offset])
	if kind, ok := completionExpectedEntityFieldKind(source, offset, prefix); ok {
		return kind, true
	}
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

func completionExpectedEntityFieldKind(source string, offset int, prefix string) (SymbolKind, bool) {
	before := source[:offset]
	marker := strings.LastIndex(before, "fields")
	if marker < 0 {
		return 0, false
	}
	open := strings.IndexByte(before[marker+len("fields"):], '{')
	if open < 0 {
		return 0, false
	}
	open += marker + len("fields")
	body := before[open+1:]
	if strings.Count(body, "{") != strings.Count(body, "}") {
		return 0, false
	}
	if strings.Contains(prefix, " required") || strings.Contains(prefix, " optional") || strings.Contains(prefix, " one") || strings.Contains(prefix, " many") {
		return SymbolKeyword, true
	}
	if strings.Contains(prefix, " type ") || strings.HasSuffix(prefix, " type") || strings.HasPrefix(prefix, "type ") {
		return SymbolClass, true
	}
	return SymbolField, true
}