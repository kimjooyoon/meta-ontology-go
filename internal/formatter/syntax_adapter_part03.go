package formatter

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// FormatSyntax formats a current syntax AST through the recovered formatter.
func FormatSyntax(file *syntax.File) Result {
	return FormatAST(file, SyntaxAdapter{})
}
func syntaxActivityInputs(activity *syntax.ActivityDecl) ([]syntax.NameRef, *Diagnostic) {
	if activity.Inputs != nil && activity.Parameters != nil && !sameSyntaxNames(activity.Inputs, activity.Parameters) {
		return nil, syntaxDiagnostic(CodeInvalidDocument, "activity "+activity.Name+" has conflicting input aliases", activity.Span)
	}
	inputs := activity.Parameters
	if inputs == nil {
		inputs = activity.Inputs
	}
	return inputs, nil
}
func syntaxActivityOutput(activity *syntax.ActivityDecl) (string, *Diagnostic) {
	output := activity.Result.Name
	if output != "" && activity.Output != "" && output != activity.Output {
		return "", syntaxDiagnostic(CodeInvalidDocument, "activity "+activity.Name+" has conflicting result aliases", activity.Span)
	}
	if output == "" {
		output = activity.Output
	}
	return output, nil
}
func sameSyntaxDeclarations(left, right []syntax.Declaration) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
func sameSyntaxNames(left, right []syntax.NameRef) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Name != right[index].Name {
			return false
		}
	}
	return true
}
func unsupportedDeclarationDiagnostic(message string, span syntax.Span) Diagnostics {
	return Diagnostics{*syntaxDiagnostic(CodeUnsupportedSyntax, message, span)}
}
func syntaxDiagnostic(code DiagnosticCode, message string, span syntax.Span) *Diagnostic {
	return &Diagnostic{
		Severity: SeverityError,
		Code:     code,
		Message:  message,
		Span:     syntaxSpan(span),
	}
}
