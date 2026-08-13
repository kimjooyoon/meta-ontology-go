package lsp

import "encoding/json"

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
	ID             string     `json:"id,omitempty"`
	Name           string     `json:"name"`
	Detail         string     `json:"detail,omitempty"`
	Kind           SymbolKind `json:"kind"`
	Range          Range      `json:"range"`
	SelectionRange Range      `json:"selectionRange"`
}

type WorkspaceSymbol struct {
	ID       string     `json:"id,omitempty"`
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
