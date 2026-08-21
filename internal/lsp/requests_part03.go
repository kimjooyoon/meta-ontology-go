package lsp

import (
	"encoding/json"
)

func requestDocumentURI(request requestEnvelope) string {
	var params struct {
		TextDocument *TextDocumentIdentifier `json:"textDocument"`
	}
	if decodeParams(request.Params, &params) != nil || params.TextDocument == nil {
		return ""
	}
	return params.TextDocument.URI
}
func (server *Server) canRunAsync(request requestEnvelope) bool {
	if request.ID == nil {
		return false
	}
	if _, ok := server.parser.(ContextParser); !ok {
		return false
	}
	if _, isSyntaxParser := server.parser.(SyntaxParser); isSyntaxParser {
		return false
	}
	switch request.Method {
	case "textDocument/hover", "textDocument/completion", "textDocument/definition", "textDocument/semanticTokens/full":
		return true
	default:
		return false
	}
}
func decodeRequest(payload []byte) (requestEnvelope, bool) {
	var request requestEnvelope
	if err := json.Unmarshal(payload, &request); err != nil {
		return requestEnvelope{}, false
	}
	return request, true
}
