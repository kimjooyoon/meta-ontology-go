package lsp

import (
	"sort"
)

func referenceTargetAmbiguous(symbols []Symbol, target string) bool {
	matches := 0
	for _, symbol := range symbols {
		if symbol.Name == target || symbol.ID == target {
			matches++
		}
	}
	return matches > 1
}
func canonicalReferenceLocations(uri, target string, references []Reference) []Location {
	result := make([]Location, 0)
	for _, reference := range references {
		if reference.Name == target {
			result = append(result, Location{URI: uri, Range: reference.Range})
		}
	}
	sort.SliceStable(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.Range.Start != second.Range.Start {
			return positionLess(first.Range.Start, second.Range.Start)
		}
		return positionLess(first.Range.End, second.Range.End)
	})
	return result
}
func canonicalReferenceLocationsForTarget(uri, targetID, targetName string, symbols []Symbol, references []Reference, includeDeclaration bool) []Location {
	result := make([]Location, 0, len(references)+1)
	for _, reference := range references {
		if targetID != "" {
			if reference.ID != targetID {
				continue
			}
		} else if reference.Name != targetName {
			continue
		}
		result = append(result, Location{URI: uri, Range: reference.Range})
	}
	if includeDeclaration {
		for _, symbol := range symbols {
			if symbol.SelectionRange == (Range{}) {
				continue
			}
			if targetID != "" {
				if symbol.ID != targetID {
					continue
				}
			} else if symbol.Name != targetName {
				continue
			}
			result = append(result, Location{URI: uri, Range: symbol.SelectionRange})
			break
		}
	}
	sort.SliceStable(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.URI != second.URI {
			return first.URI < second.URI
		}
		if first.Range.Start != second.Range.Start {
			return positionLess(first.Range.Start, second.Range.Start)
		}
		return positionLess(first.Range.End, second.Range.End)
	})
	return result
}
