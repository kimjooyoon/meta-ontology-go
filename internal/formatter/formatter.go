package formatter

import "reflect"

// ASTAdapter bridges a concrete syntax AST to the formatter's parser-neutral
// Document. Implementations may live in internal/syntax or in a future syntax
// version without making this package depend on either one.
type ASTAdapter interface {
	Adapt(ast any) (*Document, Diagnostics)
}

// ASTAdapterFunc adapts a function to ASTAdapter.
type ASTAdapterFunc func(ast any) (*Document, Diagnostics)

func (f ASTAdapterFunc) Adapt(ast any) (*Document, Diagnostics) {
	if f == nil {
		return nil, nil
	}
	return f(ast)
}

// Options controls the small initial surface printer. Options is intentionally
// open for future syntax extensions while keeping default output deterministic.
type Options struct {
	FinalNewline bool
}

// DefaultOptions are used by the package-level Format and FormatAST helpers.
var DefaultOptions = Options{FinalNewline: true}

// Result contains formatted source and any warnings or errors collected while
// adapting and printing the AST.
type Result struct {
	Source      string
	Diagnostics Diagnostics
}

// HasErrors reports whether the result contains an error diagnostic.
func (r Result) HasErrors() bool { return r.Diagnostics.HasErrors() }

// Formatter formats parser-neutral documents and adapted ASTs.
type Formatter struct {
	Adapter ASTAdapter
	Options Options
}

// New constructs a formatter with default options and the supplied adapter.
func New(adapter ASTAdapter) Formatter {
	return Formatter{Adapter: adapter, Options: DefaultOptions}
}

// Format formats a parser-neutral document with default options.
func Format(document *Document) Result {
	return Formatter{Options: DefaultOptions}.FormatDocument(document)
}

// FormatAST adapts and formats an AST with default options.
func FormatAST(ast any, adapter ASTAdapter) Result {
	return New(adapter).FormatAST(ast)
}

// FormatDocument formats an AST-free document value. A nil document is treated
// as missing input and is reported instead of being dereferenced.
func (f Formatter) FormatDocument(document *Document) Result {
	if document == nil {
		return Result{Diagnostics: Diagnostics{missingASTDiagnostic()}}
	}
	diagnostics := document.validate()
	if diagnostics.HasErrors() {
		return Result{Diagnostics: diagnostics}
	}
	return Result{Source: render(*document, f.Options), Diagnostics: diagnostics}
}

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
