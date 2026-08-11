package lsp

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

func (s *Server) hover(params TextDocumentPositionParams) (*Hover, bool) {
	document, ok := s.documents[params.TextDocument.URI]
	if !ok {
		return nil, false
	}
	name := wordAt(document.text, params.Position)
	if name == "" {
		return nil, false
	}
	symbol, ok := symbolNamed(document.result.Symbols, name)
	if !ok {
		return nil, false
	}
	rangeValue := wordRange(document.text, params.Position)
	return &Hover{Contents: MarkupContent{Kind: "plaintext", Value: symbol.Detail}, Range: &rangeValue}, true
}

func (s *Server) completion(uri string) *CompletionList {
	items := []CompletionItem{
		{Label: "package", Kind: int(symbolKeyword), Detail: "gooo keyword"},
		{Label: "namespace", Kind: int(symbolKeyword), Detail: "gooo keyword"},
		{Label: "entity", Kind: int(symbolKeyword), Detail: "gooo keyword"},
		{Label: "activity", Kind: int(symbolKeyword), Detail: "gooo keyword"},
	}
	if document, ok := s.documents[uri]; ok {
		for _, symbol := range document.result.Symbols {
			items = append(items, symbolCompletions(symbol)...)
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Label < items[j].Label })
	return &CompletionList{Items: uniqueCompletions(items)}
}

func (s *Server) definition(params TextDocumentPositionParams) []Location {
	document, ok := s.documents[params.TextDocument.URI]
	if !ok {
		return nil
	}
	name := wordAt(document.text, params.Position)
	if name == "" {
		return nil
	}
	symbol, ok := symbolNamed(document.result.Symbols, name)
	if !ok {
		return nil
	}
	return []Location{{URI: params.TextDocument.URI, Range: symbol.SelectionRange}}
}

func symbolNamed(symbols []Symbol, name string) (Symbol, bool) {
	for _, symbol := range symbols {
		if symbol.Name == name || symbolAliasNamed(symbol.Aliases, name) {
			return symbol, true
		}
	}
	return Symbol{}, false
}

func symbolAliasNamed(aliases []string, name string) bool {
	for _, alias := range aliases {
		if alias == name {
			return true
		}
	}
	return false
}

func symbolCompletions(symbol Symbol) []CompletionItem {
	kind := int(symbolCompletionKind(symbol.Kind))
	items := []CompletionItem{{Label: symbol.Name, Kind: kind, Detail: symbol.Detail}}
	for _, alias := range symbol.Aliases {
		items = append(items, CompletionItem{Label: alias, Kind: kind, Detail: "alias of " + symbol.Name})
	}
	return items
}

func uniqueCompletions(items []CompletionItem) []CompletionItem {
	result := make([]CompletionItem, 0, len(items))
	seen := make(map[string]bool)
	for _, item := range items {
		if !seen[item.Label] {
			seen[item.Label] = true
			result = append(result, item)
		}
	}
	return result
}

func symbolCompletionKind(kind SymbolKind) SymbolKind {
	if kind == symbolClass {
		return symbolClass
	}
	if kind == symbolFunction {
		return symbolFunction
	}
	return symbolText
}

func wordAt(source string, position Position) string {
	offset := positionOffset(source, position)
	start, end := identifierBounds(source, offset)
	return source[start:end]
}

func wordRange(source string, position Position) Range {
	offset := positionOffset(source, position)
	start, end := identifierBounds(source, offset)
	return Range{Start: offsetPosition(source, start), End: offsetPosition(source, end)}
}

func identifierBounds(source string, offset int) (int, int) {
	if offset > len(source) {
		offset = len(source)
	}
	if offset > 0 && (offset == len(source) || !isIdentifierRuneAt(source, offset)) {
		_, size := utf8.DecodeLastRuneInString(source[:offset])
		if size > 0 && isIdentifierRuneAt(source, offset-size) {
			offset -= size
		}
	}
	start := offset
	for start > 0 {
		runeValue, size := utf8.DecodeLastRuneInString(source[:start])
		if !isIdentifier(runeValue) {
			break
		}
		start -= size
	}
	end := offset
	for end < len(source) {
		runeValue, size := utf8.DecodeRuneInString(source[end:])
		if !isIdentifier(runeValue) {
			break
		}
		end += size
	}
	return start, end
}

func isIdentifierRuneAt(source string, offset int) bool {
	if offset < 0 || offset >= len(source) {
		return false
	}
	runeValue, _ := utf8.DecodeRuneInString(source[offset:])
	return isIdentifier(runeValue)
}

func isIdentifier(runeValue rune) bool {
	return runeValue == '_' || unicode.IsLetter(runeValue) || unicode.IsDigit(runeValue)
}

func positionOffset(source string, position Position) int {
	lineStart := 0
	for line := 0; line < position.Line && lineStart < len(source); line++ {
		newline := strings.IndexByte(source[lineStart:], '\n')
		if newline < 0 {
			return len(source)
		}
		lineStart += newline + 1
	}
	if position.Line < 0 || position.Character <= 0 {
		return lineStart
	}
	offset, units := lineStart, 0
	for offset < len(source) && source[offset] != '\n' && source[offset] != '\r' {
		runeValue, size := utf8.DecodeRuneInString(source[offset:])
		length := utf16.RuneLen(runeValue)
		if length < 0 {
			length = 1
		}
		if units+length > position.Character {
			break
		}
		units += length
		offset += size
	}
	return offset
}

func offsetPosition(source string, offset int) Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(source) {
		offset = len(source)
	}
	line, character := 0, 0
	for index := 0; index < offset; {
		runeValue, size := utf8.DecodeRuneInString(source[index:])
		if runeValue == '\n' {
			line++
			character = 0
		} else {
			units := utf16.RuneLen(runeValue)
			if units < 0 {
				units = 1
			}
			character += units
		}
		index += size
	}
	return Position{Line: line, Character: character}
}
