package lsp

import (
	"encoding/json"
	"fmt"
)

func applyChanges(source string, changes []TextDocumentContentChangeEvent) (string, error) {
	if len(changes) == 0 {
		return "", ErrInvalidRange
	}
	for _, change := range changes {
		if change.Range == nil {
			if change.RangeLength != nil && (*change.RangeLength < 0 || *change.RangeLength != utf16Length(source)) {
				return "", ErrInvalidRange
			}
			source = change.Text
			continue
		}
		start, end, err := ValidateRange(source, *change.Range)
		if err != nil {
			return "", err
		}
		if change.RangeLength != nil && (*change.RangeLength < 0 || *change.RangeLength != utf16Length(source[start:end])) {
			return "", ErrInvalidRange
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
