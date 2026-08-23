package toolchainlsp

import "github.com/kimjooyoon/meta-ontology-go/internal/lsp"

func observeCapabilities(message rpcMessage) (bool, int) {
	result, ok := resultAs[lsp.InitializeResult](message)
	if !ok {
		return false, 0
	}
	capability := result.Capabilities
	count := boolCount(
		capability.TextDocumentSync.OpenClose && capability.TextDocumentSync.Change == 2,
		capability.HoverProvider,
		capability.CompletionProvider != nil,
		capability.DefinitionProvider,
		capability.DocumentSymbolProvider,
		capability.ReferencesProvider,
		capability.WorkspaceSymbolProvider != nil && capability.WorkspaceSymbolProvider.Schema == lsp.WorkspaceSymbolProtocolSchema,
		capability.SemanticTokensProvider != nil && capability.SemanticTokensProvider.Full && capability.SemanticTokensProvider.Schema == lsp.SemanticTokensProtocolSchema,
	)
	return count == 8 && result.ServerInfo.Name == "gooo-lsp", count
}
