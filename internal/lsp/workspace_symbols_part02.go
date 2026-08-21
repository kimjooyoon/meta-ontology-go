package lsp

import (
	"sort"
)

func canonicalWorkspaceSymbols(symbols []WorkspaceSymbol) []WorkspaceSymbol {
	result := make([]WorkspaceSymbol, len(symbols))
	copy(result, symbols)
	sort.SliceStable(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.Location.URI != second.Location.URI {
			return first.Location.URI < second.Location.URI
		}
		if first.Name != second.Name {
			return first.Name < second.Name
		}
		if first.ID != second.ID {
			return first.ID < second.ID
		}
		if first.Location.Range.Start != second.Location.Range.Start {
			return positionLess(first.Location.Range.Start, second.Location.Range.Start)
		}
		if first.Location.Range.End != second.Location.Range.End {
			return positionLess(first.Location.Range.End, second.Location.Range.End)
		}
		if first.Detail != second.Detail {
			return first.Detail < second.Detail
		}
		return first.Kind < second.Kind
	})
	return result
}
