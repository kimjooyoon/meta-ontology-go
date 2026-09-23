package lsp

// PrepareRenameResult identifies the exact client-side range that a later
// rename request may edit. It carries no mutation authority.
type PrepareRenameResult struct {
	Range       Range  `json:"range"`
	Placeholder string `json:"placeholder,omitempty"`
}

func (server *Server) prepareRenameRequest(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params TextDocumentPositionParams
	if decodeParams(request.Params, &params) != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid prepare rename parameters"), nil, nil
	}
	document, exists := server.referenceDocument(params.TextDocument.URI)
	if !exists {
		return resultResponse(request.ID, nil), nil, nil
	}
	if document.result.semanticChecked && !document.result.semanticValid {
		return resultResponse(request.ID, nil), nil, nil
	}
	targetID, targetName, err := referenceTargetForDocument(document, params.Position)
	if err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid prepare rename position"), nil, nil
	}
	if targetID == "" || targetName == "" {
		return resultResponse(request.ID, nil), nil, nil
	}
	prepared, ok := prepareRenameForTarget(params.TextDocument.URI, targetID, targetName, allSymbols(document.result), document.result.References)
	if !ok {
		return resultResponse(request.ID, nil), nil, nil
	}
	return resultResponse(request.ID, prepared), nil, nil
}

func prepareRenameForTarget(uri, targetID, targetName string, symbols []Symbol, references []Reference) (PrepareRenameResult, bool) {
	locations := canonicalReferenceLocationsForTarget(uri, targetID, targetName, symbols, references, true)
	if len(locations) == 0 {
		return PrepareRenameResult{}, false
	}
	return PrepareRenameResult{Range: locations[0].Range, Placeholder: targetName}, true
}
