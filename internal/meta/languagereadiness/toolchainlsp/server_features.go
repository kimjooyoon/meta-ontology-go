package toolchainlsp

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
)

func observeFeatures(responses map[string]rpcMessage) (map[string]observation, int) {
	result := map[string]observation{}
	hover, hoverOK := resultAs[*lsp.Hover](responses["2"])
	hoverOK = hoverOK && hover != nil && strings.Contains(hover.Contents.Value, "entity Order")
	result["hover"] = observation{"ENTITY_ORDER", hoverOK}
	completion, completionOK := resultAs[lsp.CompletionList](responses["3"])
	foundCompletion := false
	for _, item := range completion.Items { if item.Label == "Order" { foundCompletion = true } }
	completionOK = completionOK && foundCompletion
	result["completion"] = observation{"ORDER_ITEM", completionOK}
	definition, definitionOK := resultAs[[]lsp.Location](responses["4"])
	definitionOK = definitionOK && len(definition) == 1 && definition[0].URI == sessionURI
	result["definition"] = observation{"ONE_LOCAL_LOCATION", definitionOK}
	documentSymbols, documentOK := resultAs[[]lsp.DocumentSymbol](responses["5"])
	documentOK = documentOK && documentSymbolNamed(documentSymbols, "Order")
	result["document-symbol"] = observation{"ORDER_SYMBOL", documentOK}
	workspaceSymbols, workspaceOK := resultAs[[]lsp.WorkspaceSymbol](responses["6"])
	workspaceOK = workspaceOK && workspaceSymbolNamed(workspaceSymbols, "Order")
	result["workspace-symbol"] = observation{"ORDER_WORKSPACE_SYMBOL", workspaceOK}
	references, referencesOK := resultAs[[]lsp.Location](responses["7"])
	referencesOK = referencesOK && len(references) > 0
	result["references"] = observation{"NONEMPTY_LOCAL_REFERENCES", referencesOK}
	tokens, tokensOK := resultAs[lsp.SemanticTokens](responses["8"])
	tokensOK = tokensOK && len(tokens.Data) > 0
	result["semantic-tokens"] = observation{"NONEMPTY_DELTA_TOKENS", tokensOK}
	return result, boolCount(hoverOK, completionOK, definitionOK, documentOK, workspaceOK, referencesOK, tokensOK)
}

func documentSymbolNamed(symbols []lsp.DocumentSymbol, name string) bool {
	for _, symbol := range symbols { if symbol.Name == name { return true } }
	return false
}

func workspaceSymbolNamed(symbols []lsp.WorkspaceSymbol, name string) bool {
	for _, symbol := range symbols { if symbol.Name == name && symbol.Location.URI == sessionURI { return true } }
	return false
}
