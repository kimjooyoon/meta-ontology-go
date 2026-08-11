package lsp

import (
	"encoding/json"
	"fmt"
)

func applyChanges(source string, changes []TextDocumentContentChangeEvent) (string, error) {
	for _, change := range changes {
		if change.Range == nil {
			source = change.Text
			continue
		}
		start := positionOffset(source, change.Range.Start)
		end := positionOffset(source, change.Range.End)
		if start > end {
			return "", fmt.Errorf("change range starts after it ends")
		}
		source = source[:start] + change.Text + source[end:]
	}
	return source, nil
}

func diagnosticsNotification(uri string, diagnostics []Diagnostic) ([]byte, error) {
	if diagnostics == nil {
		diagnostics = []Diagnostic{}
	}
	notification := struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params"`
	}{
		JSONRPC: jsonRPCVersion,
		Method:  "textDocument/publishDiagnostics",
		Params:  PublishDiagnosticsParams{URI: uri, Diagnostics: diagnostics},
	}
	payload, err := json.Marshal(notification)
	if err != nil {
		return nil, fmt.Errorf("lsp: encode diagnostics: %w", err)
	}
	return payload, nil
}
