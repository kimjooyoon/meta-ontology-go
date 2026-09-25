package lsp

import (
	"sort"
)

func canonicalDocumentSymbols(symbols []Symbol) []DocumentSymbol {
	result := make([]DocumentSymbol, 0, len(symbols))
	for _, symbol := range symbols {
		result = append(result, DocumentSymbol{
			ID: symbol.ID, Name: symbol.Name, Detail: symbolDetail(symbol), Kind: symbol.Kind,
			Range: symbol.Range, SelectionRange: symbol.SelectionRange,
		})
	}
	sort.SliceStable(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.Range.Start != second.Range.Start {
			return positionLess(first.Range.Start, second.Range.Start)
		}
		if first.Range.End != second.Range.End {
			return positionLess(first.Range.End, second.Range.End)
		}
		if first.SelectionRange.Start != second.SelectionRange.Start {
			return positionLess(first.SelectionRange.Start, second.SelectionRange.Start)
		}
		if first.SelectionRange.End != second.SelectionRange.End {
			return positionLess(first.SelectionRange.End, second.SelectionRange.End)
		}
		if first.ID != second.ID {
			return first.ID < second.ID
		}
		if first.Name != second.Name {
			return first.Name < second.Name
		}
		if first.Detail != second.Detail {
			return first.Detail < second.Detail
		}
		return first.Kind < second.Kind
	})
	return result
}
