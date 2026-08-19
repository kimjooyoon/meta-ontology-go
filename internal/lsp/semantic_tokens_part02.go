package lsp

import (
	"sort"
)

func canonicalSemanticTokenSpans(spans []semanticTokenSpan) []semanticTokenSpan {
	result := append([]semanticTokenSpan(nil), spans...)
	sort.Slice(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.rangeValue.Start != second.rangeValue.Start {
			return positionLess(first.rangeValue.Start, second.rangeValue.Start)
		}
		if first.rangeValue.End != second.rangeValue.End {
			return positionLess(first.rangeValue.End, second.rangeValue.End)
		}
		return first.tokenType < second.tokenType
	})
	unique := result[:0]
	for _, span := range result {
		if len(unique) > 0 && unique[len(unique)-1] == span {
			continue
		}
		unique = append(unique, span)
	}
	return unique
}
func semanticTokenData(spans []semanticTokenSpan) []uint32 {
	data := make([]uint32, 0, len(spans)*5)
	previousLine, previousCharacter := 0, 0
	for _, span := range spans {
		lineDelta := span.rangeValue.Start.Line - previousLine
		characterDelta := span.rangeValue.Start.Character
		if lineDelta == 0 {
			characterDelta -= previousCharacter
		}
		if lineDelta < 0 || characterDelta < 0 {
			continue
		}
		length := span.rangeValue.End.Character - span.rangeValue.Start.Character
		if length <= 0 {
			continue
		}
		data = append(data, uint32(lineDelta), uint32(characterDelta), uint32(length), uint32(span.tokenType), 0)
		previousLine, previousCharacter = span.rangeValue.Start.Line, span.rangeValue.Start.Character
	}
	return data
}
func validSemanticTokenRange(value Range) bool {
	return value.Start.Line >= 0 && value.Start.Character >= 0 &&
		value.End.Line == value.Start.Line && value.End.Character > value.Start.Character
}
func semanticTokenType(kind SymbolKind) int {
	switch kind {
	case SymbolClass:
		return 0
	case SymbolFunction:
		return 1
	default:
		return 3
	}
}
