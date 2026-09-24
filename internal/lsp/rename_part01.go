package lsp

import (
	"slices"
	"unicode"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func (server *Server) renameRequest(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params RenameParams
	if decodeParams(request.Params, &params) != nil || params.TextDocument.URI == "" || !validRenameIdentifier(params.NewName) {
		return responseOrNil(request.ID, invalidParams, "Invalid rename parameters"), nil, nil
	}
	document, exists := server.referenceDocument(params.TextDocument.URI)
	if !exists {
		return resultResponse(request.ID, nil), nil, nil
	}
	if document.result.semanticChecked && !document.result.semanticValid {
		return responseOrNil(request.ID, methodNotFound, "method is deferred by this LSP baseline"), nil, nil
}
	targetID, targetName, err := referenceTargetForDocument(document, params.Position)
	if err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid rename position"), nil, nil
	}
	if targetID == "" || targetName == "" {
		return resultResponse(request.ID, nil), nil, nil
	}
	edits := renameEditsForTarget(params.TextDocument.URI, targetID, targetName, allSymbols(document.result), document.result.References, params.NewName)
	if len(edits) == 0 {
		return resultResponse(request.ID, nil), nil, nil
	}
	return resultResponse(request.ID, WorkspaceEdit{Changes: map[string][]TextEdit{params.TextDocument.URI: edits}}), nil, nil
}

func renameEditsForTarget(uri, targetID, targetName string, symbols []Symbol, references []Reference, newName string) []TextEdit {
	locations := canonicalReferenceLocationsForTarget(uri, targetID, targetName, symbols, references, true)
	edits := make([]TextEdit, 0, len(locations))
	seen := make(map[Range]struct{}, len(locations))
	for _, location := range locations {
		if _, exists := seen[location.Range]; exists {
			continue
		}
		seen[location.Range] = struct{}{}
		edits = append(edits, TextEdit{Range: location.Range, NewText: newName})
	}
	return edits
}

func validRenameIdentifier(value string) bool {
	if value == "" {
		return false
	}
	if containsSyntaxKeyword(value) {
		return false
	}
	for index, character := range value {
		if index == 0 {
			if character != '_' && !unicode.IsLetter(character) {
				return false
			}
			continue
		}
		if !isIdentifier(character) {
			return false
		}
	}
	return true
}

func containsSyntaxKeyword(value string) bool {
	return slices.Contains(syntax.CanonicalKeywordNames(), value)
}