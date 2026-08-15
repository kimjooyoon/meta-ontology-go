package lsp

import "sort"

func (server *Server) referencesRequest(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params ReferenceParams
	if decodeParams(request.Params, &params) != nil || params.TextDocument.URI == "" || params.Position == nil {
		return responseOrNil(request.ID, invalidParams, "Invalid reference parameters"), nil, nil
	}
	document, exists := server.referenceDocument(params.TextDocument.URI)
	if !exists {
		return resultResponse(request.ID, nil), nil, nil
	}
	if document.result.semanticChecked && !document.result.semanticValid {
		return resultResponse(request.ID, nil), nil, nil
	}
	targetID, target, err := referenceTargetForDocument(document, *params.Position)
	if err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid reference position"), nil, nil
	}
	if target == "" || (targetID == "" && referenceTargetAmbiguous(allSymbols(document.result), target)) {
		if target == "" {
			return resultResponse(request.ID, []Location{}), nil, nil
		}
		return resultResponse(request.ID, nil), nil, nil
	}
	return resultResponse(request.ID, canonicalReferenceLocationsForTarget(
		params.TextDocument.URI, targetID, target, allSymbols(document.result), document.result.References,
		params.Context.IncludeDeclaration,
	)), nil, nil
}

func (server *Server) referenceDocument(uri string) (document, bool) {
	server.mu.RLock()
	value, exists := server.documents[uri]
	if exists {
		copyValue := documentCopy(value)
		server.mu.RUnlock()
		return copyValue, true
	}
	server.mu.RUnlock()
	return document{}, false
}

func referenceTarget(source string, position Position) (string, error) {
	offset, err := PositionToOffset(source, position)
	if err != nil {
		return "", err
	}
	start, end := identifierBounds(source, offset)
	if start == end {
		return "", nil
	}
	return source[start:end], nil
}

func referenceTargetForDocument(document document, position Position) (string, string, error) {
	target, err := referenceTarget(document.text, position)
	if err != nil {
		return "", "", err
	}
	for _, symbol := range allSymbols(document.result) {
		if symbol.hasIdentity && positionInRange(position, symbol.identityRange) {
			return symbol.ID, symbol.Name, nil
		}
		if symbol.Name == target && positionInRange(position, symbol.SelectionRange) {
			return symbol.ID, symbol.Name, nil
		}
	}
	for _, reference := range document.result.References {
		if reference.Name == target && positionInRange(position, reference.Range) {
			return reference.ID, reference.Name, nil
		}
	}
	return "", target, nil
}

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
