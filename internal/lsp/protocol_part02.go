package lsp

type SemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}
type SemanticTokensOptions struct {
	Schema string               `json:"schema"`
	Legend SemanticTokensLegend `json:"legend"`
	Full   bool                 `json:"full"`
}

// ServerCapabilities advertises only implemented document/workspace-read
// features. Workspace edits and source maps remain deferred.
type ServerCapabilities struct {
	TextDocumentSync        TextDocumentSyncOptions `json:"textDocumentSync"`
	HoverProvider           bool                    `json:"hoverProvider,omitempty"`
	CompletionProvider      *CompletionOptions      `json:"completionProvider,omitempty"`
	DefinitionProvider      bool                    `json:"definitionProvider,omitempty"`
	DocumentSymbolProvider  bool                    `json:"documentSymbolProvider,omitempty"`
	ReferencesProvider      bool                    `json:"referencesProvider,omitempty"`
	WorkspaceSymbolProvider *WorkspaceSymbolOptions `json:"workspaceSymbolProvider,omitempty"`
	SemanticTokensProvider  *SemanticTokensOptions  `json:"semanticTokensProvider,omitempty"`
}
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo,omitempty"`
}
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}
type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}
type SemanticTokensParams struct {
	TextDocument *TextDocumentIdentifier `json:"textDocument"`
}
type SemanticTokens struct {
	ResultID string   `json:"resultId,omitempty"`
	Data     []uint32 `json:"data"`
}
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}
type ReferenceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     *Position              `json:"position"`
	Context      ReferenceContext       `json:"context"`
}
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}
