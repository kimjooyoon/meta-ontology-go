package lsp

import (
	"testing"
)

func TestDidCloseCancelsRequestsByDocumentURI(t *testing.T) {
	methods := []string{"textDocument/hover", "textDocument/completion", "textDocument/semanticTokens/full"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) { runClosedFeature(t, method) })
	}
}
