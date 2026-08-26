package lsp

import (
	"context"
	"encoding/json"
)

func (server *Server) dispatch(ctx context.Context, payload []byte) (*responseEnvelope, [][]byte, error) {
	var request requestEnvelope
	if err := json.Unmarshal(payload, &request); err != nil {
		return errorResponse(nil, parseError, "Parse error"), nil, nil
	}
	if request.JSONRPC != jsonRPCVersion || request.Method == "" {
		return responseOrNil(request.ID, invalidRequest, "Invalid Request"), nil, nil
	}
	shutdown, _ := server.lifecycleState()
	if shutdown && request.Method != "exit" {
		return responseOrNil(request.ID, invalidRequest, "server is shut down"), nil, nil
	}
	switch request.Method {
	case "initialize":
		if server.isInitialized() {
			return responseOrNil(request.ID, invalidRequest, "server is already initialized"), nil, nil
		}
		return server.initialize(request)
	case "initialized", "$/cancelRequest":
		return nil, nil, nil
	case "shutdown":
		return server.shutdownRequest(request), nil, nil
	case "exit":
		server.markExited()
		return nil, nil, nil
	case "textDocument/didOpen":
		return server.didOpen(ctx, request)
	case "textDocument/didChange":
		return server.didChange(ctx, request)
	case "textDocument/didClose":
		return server.didClose(request)
	case "textDocument/hover":
		return server.hoverRequestContext(ctx, request)
	case "textDocument/completion":
		return server.completionRequest(ctx, request)
	case "textDocument/definition":
		return server.definitionRequest(request)
	case "textDocument/documentSymbol":
		return server.documentSymbolRequest(request)
	case "textDocument/references":
		return server.referencesRequest(request)
	case "workspace/symbol":
		return server.workspaceSymbolRequest(request)
	case "textDocument/semanticTokens/full":
		return server.semanticTokensRequest(ctx, request)
	case "textDocument/rename", "textDocument/formatting":
		return responseOrNil(request.ID, methodNotFound, "method is deferred by this LSP baseline"), nil, nil
	default:
		return responseOrNil(request.ID, methodNotFound, "Method not found"), nil, nil
	}
}
func (server *Server) shutdownRequest(request requestEnvelope) *responseEnvelope {
	server.mu.Lock()
	server.shutdown = true
	server.mu.Unlock()
	if request.ID == nil {
		return nil
	}
	return resultResponse(request.ID, nil)
}
