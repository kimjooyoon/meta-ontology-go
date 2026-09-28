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
			RenameProvider:          true,
			WorkspaceSymbolProvider: &WorkspaceSymbolOptions{Schema: WorkspaceSymbolProtocolSchema},
			SemanticTokensProvider: &SemanticTokensOptions{
				Schema: SemanticTokensProtocolSchema,
				Legend: SemanticTokensLegend{
					TokenTypes:     append([]string(nil), canonicalSemanticTokenTypes...),
					TokenModifiers: []string{},
				},
				Full: true,
			},
			Experimental: map[string]any{
				"goooDocumentProvenance": map[string]string{
					"method": "gooo/documentProvenance", "schema": documentProvenanceSchema,
				},
				"goooSelfImprovementProvenance": map[string]string{
					"method": "gooo/selfImprovementProvenance", "schema": SelfImprovementProvenanceChainSchemaPart01,
				},
				"goooExecutionPlanProvenance": map[string]string{
					"method": "gooo/executionPlanProvenance", "schema": ExecutionPlanProvenanceSchemaPart01,
				},
				"goooReferencesProvenance": map[string]string{
					"method": "gooo/referencesProvenance", "schema": referencesProvenanceSchema,
				},
				"goooDiagnosticProvenance": map[string]string{
					"method": "gooo/diagnosticProvenance", "schema": diagnosticProvenanceSchema,
				},
				"goooStoryProvenance": map[string]string{
					"method": "gooo/storyProvenance", "schema": storyProvenanceSchema,
				},
				"goooCompletionProvenance": map[string]string{
					"method": "gooo/completionProvenance", "schema": completionProvenanceSchema,
				},
			},
		},
		ServerInfo: ServerInfo{Name: "gooo-lsp", Version: "current-ddaf"},
	}
	return resultResponse(request.ID, result), nil, nil
}
