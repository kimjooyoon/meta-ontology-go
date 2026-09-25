package lsp

const jsonRPCVersion = "2.0"
const WorkspaceSymbolProtocolSchema = "gooo/lsp-workspace-symbol/v1"
const SemanticTokensProtocolSchema = "gooo/lsp-semantic-tokens/v1"
const (
	parseError      = -32700
	invalidRequest  = -32600
	methodNotFound  = -32601
	invalidParams   = -32602
	internalError   = -32603
	contentModified = -32801
)

// Position is a zero-based LSP position measured in UTF-16 code units.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range is a half-open LSP source range.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}
type DiagnosticSeverity int

const (
	DiagnosticError   DiagnosticSeverity = 1
	DiagnosticWarning DiagnosticSeverity = 2
)

type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity,omitempty"`
	Code     string             `json:"code,omitempty"`
	Source   string             `json:"source,omitempty"`
	Message  string             `json:"message"`
	filename string
	start    int
	end      int
	spanned  bool
}
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId,omitempty"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}
type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}
type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitempty"`
	RangeLength *int   `json:"rangeLength,omitempty"`
	Text        string `json:"text"`
}
type InitializeParams struct {
	RootURI string `json:"rootUri,omitempty"`
}
type TextDocumentSyncOptions struct {
	OpenClose bool `json:"openClose"`
	Change    int  `json:"change"`
}
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}
type WorkspaceSymbolOptions struct {
	Schema string `json:"schema"`
}
