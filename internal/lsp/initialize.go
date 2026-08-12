package lsp

func (server *Server) initialize(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params InitializeParams
	if err := decodeParams(request.Params, &params); err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid initialize parameters"), nil, nil
	}
	server.mu.Lock()
	server.initialized = true
	server.mu.Unlock()
	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync:        TextDocumentSyncOptions{OpenClose: true, Change: 2},
			HoverProvider:           true,
			CompletionProvider:      &CompletionOptions{},
			DefinitionProvider:      true,
			DocumentSymbolProvider:  true,
			ReferencesProvider:      true,
			WorkspaceSymbolProvider: &WorkspaceSymbolOptions{Schema: WorkspaceSymbolProtocolSchema},
		},
		ServerInfo: ServerInfo{Name: "gooo-lsp", Version: "current-ddaf"},
	}
	return resultResponse(request.ID, result), nil, nil
}
