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
	target, err := referenceTarget(document.text, *params.Position)
	if err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid reference position"), nil, nil
	}
	if target == "" || referenceTargetAmbiguous(document.result.Symbols, target) {
		if target == "" {
			return resultResponse(request.ID, []Location{}), nil, nil
		}
		return resultResponse(request.ID, nil), nil, nil
	}
	return resultResponse(request.ID, canonicalReferenceLocations(params.TextDocument.URI, target, document.result.References)), nil, nil
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
