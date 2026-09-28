package syntax

// Severity classifies a diagnostic. The syntax layer currently emits errors;
// warnings remain part of the API so future syntax extensions do not need an
// incompatible result type.
type Severity uint8

const (
	SeverityError Severity = iota
	SeverityWarning
)

func (s Severity) String() string {
	if s == SeverityWarning {
		return "warning"
	}
	return "error"
}

// DiagnosticCode is a stable machine-readable diagnostic identifier.
type DiagnosticCode string

const (
	DiagUnexpectedCharacter   DiagnosticCode = "lex.unexpected-character"
	DiagUnterminatedComment   DiagnosticCode = "lex.unterminated-comment"
	DiagUnterminatedString    DiagnosticCode = "lex.unterminated-string"
	DiagInvalidEscape         DiagnosticCode = "lex.invalid-escape"
	DiagInvalidUTF8           DiagnosticCode = "lex.invalid-utf8"
	DiagExpectedPackage       DiagnosticCode = "parse.expected-package"
	DiagExpectedNamespace     DiagnosticCode = "parse.expected-namespace"
	DiagExpectedIdentifier    DiagnosticCode = "parse.expected-identifier"
	DiagExpectedID            DiagnosticCode = "parse.expected-id"
	DiagExpectedString        DiagnosticCode = "parse.expected-string"
	DiagExpectedLeftParen     DiagnosticCode = "parse.expected-left-paren"
	DiagExpectedRightParen    DiagnosticCode = "parse.expected-right-paren"
	DiagExpectedComma         DiagnosticCode = "parse.expected-comma"
	DiagExpectedArrow         DiagnosticCode = "parse.expected-arrow"
	DiagExpectedDot           DiagnosticCode = "parse.expected-dot"
	DiagExpectedResult        DiagnosticCode = "parse.expected-result"
	DiagUnexpectedDeclaration DiagnosticCode = "parse.unexpected-token"
)
const (
	DiagEntityFieldsDeferred      DiagnosticCode = "parse.entity-fields-deferred"
	DiagEntityFieldsConfiguration DiagnosticCode = "parse.entity-fields-configuration"
	DiagExpectedFieldsLeftBrace   DiagnosticCode = "parse.expected-fields-left-brace"
	DiagExpectedEntityField       DiagnosticCode = "parse.expected-entity-field"
	DiagExpectedFieldType         DiagnosticCode = "parse.expected-field-type"
	DiagExpectedFieldPresence     DiagnosticCode = "parse.expected-field-presence"
	DiagExpectedFieldCardinality  DiagnosticCode = "parse.expected-field-cardinality"
	DiagEntityFieldsUnterminated  DiagnosticCode = "parse.entity-fields-unterminated"
	DiagEntityFieldsTrailing      DiagnosticCode = "parse.entity-fields-trailing"
)

// Diagnostic describes a recoverable lexical or syntactic problem.
type Diagnostic struct {
	Severity Severity
	Code     DiagnosticCode
	Message  string
	Span     Span
}

func (d Diagnostic) Error() string {
	return d.String()
}

// String formats a diagnostic deterministically.
func (d Diagnostic) String() string {
	return d.Span.String() + ": " + d.Severity.String() + " " + string(d.Code) + ": " + d.Message
}

// Diagnostics is an ordered list of diagnostics. Lexer diagnostics precede
// parser diagnostics, and diagnostics produced within each phase follow
// source order.
type Diagnostics []Diagnostic
