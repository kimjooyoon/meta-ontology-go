package lsp

import (
	"encoding/json"
)

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}
type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind,omitempty"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
}
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}
type DocumentSymbol struct {
	// ID is an internal correspondence key. DocumentSymbol has no standard
	// wire identity field, so it is projected through Detail instead.
	ID             string     `json:"-"`
	Name           string     `json:"name"`
	Detail         string     `json:"detail,omitempty"`
	Kind           SymbolKind `json:"kind"`
	Range          Range      `json:"range"`
	SelectionRange Range      `json:"selectionRange"`
}
type WorkspaceSymbol struct {
	// ID is an internal correspondence key. WorkspaceSymbol has no standard
	// wire identity field, so it is projected through Detail instead.
	ID       string     `json:"-"`
	Name     string     `json:"name"`
	Detail   string     `json:"detail,omitempty"`
	Kind     SymbolKind `json:"kind"`
	Location Location   `json:"location"`
}
type requestEnvelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}
type responseEnvelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *errorObject    `json:"error,omitempty"`
}
type errorObject struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
