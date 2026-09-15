package lsp

import (
	"context"
)

func (server *Server) didOpen(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params DidOpenTextDocumentParams
	if err := decodeParams(request.Params, &params); err != nil || params.TextDocument.URI == "" || params.TextDocument.Version < 0 {
		return responseOrNil(request.ID, invalidParams, "Invalid didOpen parameters"), nil, nil
	}
	server.mu.RLock()
	_, exists := server.documents[params.TextDocument.URI]
	server.mu.RUnlock()
	if exists {
		return responseOrNil(request.ID, invalidParams, "Document is already open"), nil, nil
	}
	result, err := server.parse(ctx, params.TextDocument.URI, params.TextDocument.Text)
	if err != nil {
		return parseFailure(request.ID, ctx, err), nil, errIfCanceled(ctx, err)
	}
	server.mu.Lock()
	server.documents[params.TextDocument.URI] = &document{
		version: params.TextDocument.Version, text: params.TextDocument.Text, result: result,
	}
	server.mu.Unlock()
	notification, err := diagnosticsNotification(params.TextDocument.URI, result.Diagnostics)
	return nil, oneNotification(notification, err), err
}
func (server *Server) didChange(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params DidChangeTextDocumentParams
	if err := decodeParams(request.Params, &params); err != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid didChange parameters"), nil, nil
	}
	server.mu.RLock()
	document, exists := server.documents[params.TextDocument.URI]
	if exists {
		documentVersion, documentText := document.version, document.text
		server.mu.RUnlock()
		return server.changeDocument(ctx, request, params, documentVersion, documentText)
	}
	server.mu.RUnlock()
	return responseOrNil(request.ID, invalidParams, "Document is not open"), nil, nil
}
func (server *Server) changeDocument(ctx context.Context, request requestEnvelope, params DidChangeTextDocumentParams, version int, source string) (*responseEnvelope, [][]byte, error) {
	if params.TextDocument.Version <= version {
		return responseOrNil(request.ID, invalidParams, ErrInvalidVersion.Error()), nil, nil
	}
	text, err := applyChanges(source, params.ContentChanges)
	if err != nil {
		return responseOrNil(request.ID, invalidParams, err.Error()), nil, nil
	}
	result, err := server.parse(ctx, params.TextDocument.URI, text)
	if err != nil {
		return parseFailure(request.ID, ctx, err), nil, errIfCanceled(ctx, err)
	}
	server.mu.Lock()
	document, exists := server.documents[params.TextDocument.URI]
	if !exists || document.version != version || document.text != source {
		server.mu.Unlock()
		return nil, nil, nil
	}
	document.version, document.text, document.result, document.cacheKey = params.TextDocument.Version, text, result, documentCacheKey{}
	server.mu.Unlock()
	notification, err := diagnosticsNotification(params.TextDocument.URI, result.Diagnostics)
	return nil, oneNotification(notification, err), err
}
