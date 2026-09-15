package formatter

import (
	"reflect"
)

// FormatAST safely adapts and formats a concrete AST.
func (f Formatter) FormatAST(ast any) Result {
	if isNilAST(ast) {
		return Result{Diagnostics: Diagnostics{missingASTDiagnostic()}}
	}
	if f.Adapter == nil || isNilAST(f.Adapter) {
		return Result{Diagnostics: Diagnostics{{Severity: SeverityError, Code: CodeMissingAdapter, Message: "cannot format AST without an AST adapter"}}}
	}
	document, diagnostics := f.Adapter.Adapt(ast)
	if document == nil {
		if len(diagnostics) == 0 {
			diagnostics = append(diagnostics, Diagnostic{Severity: SeverityError, Code: CodeAdapterReturnedNil, Message: "AST adapter returned no document"})
		}
		return Result{Diagnostics: diagnostics}
	}
	result := f.FormatDocument(document)
	result.Diagnostics = append(append(Diagnostics(nil), diagnostics...), result.Diagnostics...)
	if hasErrors(diagnostics) {
		result.Source = ""
	}
	return result
}
func isNilAST(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
func missingASTDiagnostic() Diagnostic {
	return Diagnostic{Severity: SeverityError, Code: CodeMissingAST, Message: "cannot format because the AST is nil"}
}
func hasErrors(diagnostics Diagnostics) bool { return diagnostics.HasErrors() }
