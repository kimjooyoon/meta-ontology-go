package toolchainlsp

import (
	"bytes"

	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
)

const sessionURI = "file:///billing.gooo"
const sessionSource = "package billing\nnamespace billing\nentity Order id \"billing://entity/order\"\nactivity PayOrder(Order) -> Order\n"

func executeServerSession() ([]byte, []rpcMessage, error) {
	var input, output bytes.Buffer
	request := func(id int, method string, params any) error { return appendRPC(&input, &id, method, params) }
	notify := func(method string, params any) error { return appendRPC(&input, nil, method, params) }
	document := map[string]any{"uri": sessionURI}
	steps := []func() error{
		func() error { return request(1, "initialize", map[string]any{}) },
		func() error {
			return notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": sessionURI, "languageId": "gooo", "version": 1, "text": sessionSource}})
		},
		func() error {
			return request(2, "textDocument/hover", map[string]any{"textDocument": document, "position": map[string]any{"line": 2, "character": 8}})
		},
		func() error {
			return request(3, "textDocument/completion", map[string]any{"textDocument": document, "position": map[string]any{"line": 3, "character": 14}})
		},
		func() error {
			return request(4, "textDocument/definition", map[string]any{"textDocument": document, "position": map[string]any{"line": 3, "character": 14}})
		},
		func() error {
			return request(5, "textDocument/documentSymbol", map[string]any{"textDocument": document})
		},
		func() error { return request(6, "workspace/symbol", map[string]any{"query": "Order"}) },
		func() error {
			return request(7, "textDocument/references", map[string]any{"textDocument": document, "position": map[string]any{"line": 2, "character": 8}, "context": map[string]any{"includeDeclaration": true}})
		},
		func() error {
			return request(8, "textDocument/semanticTokens/full", map[string]any{"textDocument": document})
		},
		func() error {
			return notify("textDocument/didChange", map[string]any{"textDocument": map[string]any{"uri": sessionURI, "version": 2}, "contentChanges": []map[string]any{{"text": "package billing\nnamespace billing\nentity Order id \"unterminated"}}})
		},
		func() error {
			return request(9, "textDocument/rename", map[string]any{"textDocument": document, "position": map[string]any{"line": 2, "character": 8}, "newName": "Invoice"})
		},
		func() error { return notify("textDocument/didClose", map[string]any{"textDocument": document}) },
		func() error { return request(10, "shutdown", nil) },
		func() error { return notify("exit", nil) },
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return nil, nil, err
		}
	}
	if err := lsp.NewServer().Serve(&input, &output); err != nil {
		return nil, nil, err
	}
	raw := append([]byte(nil), output.Bytes()...)
	messages, err := readRPC(raw)
	return raw, messages, err
}
