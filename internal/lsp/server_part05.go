package lsp

import (
	"context"
)

func (server *Server) didClose(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params DidCloseTextDocumentParams
	if err := decodeParams(request.Params, &params); err != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid didClose parameters"), nil, nil
	}
	server.cancelRequestsForURI(params.TextDocument.URI)
	server.mu.Lock()
	if _, exists := server.documents[params.TextDocument.URI]; !exists {
		server.mu.Unlock()
		return responseOrNil(request.ID, invalidParams, "Document is not open"), nil, nil
	}
	delete(server.documents, params.TextDocument.URI)
	server.mu.Unlock()
	notification, err := diagnosticsNotification(params.TextDocument.URI, nil)
	return nil, oneNotification(notification, err), err
}
func (server *Server) hoverRequestContext(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params TextDocumentPositionParams
	if err := decodeParams(request.Params, &params); err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid hover parameters"), nil, nil
	}
	if err := server.refresh(ctx, params.TextDocument.URI); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	hover, _ := server.hover(params)
	return resultResponse(request.ID, hover), nil, nil
}
func (server *Server) completionRequest(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params TextDocumentPositionParams
	if err := decodeParams(request.Params, &params); err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid completion parameters"), nil, nil
	}
	if err := server.refresh(ctx, params.TextDocument.URI); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	return resultResponse(request.ID, server.completion(params.TextDocument.URI)), nil, nil
}
func (server *Server) definitionRequest(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params TextDocumentPositionParams
	if err := decodeParams(request.Params, &params); err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid definition parameters"), nil, nil
	}
	return resultResponse(request.ID, server.definition(params)), nil, nil
}
