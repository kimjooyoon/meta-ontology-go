package lsp

import (
	"context"
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type loweredSymbolKey struct {
	start int
	end   int
	kind  semantic.Kind
	name  string
}

// adaptSyntaxResult is retained as a small, context-free test seam. The live
// SyntaxParser path uses adaptSyntaxResultContext so lowering participates in
// request cancellation.
func adaptSyntaxResult(uri, source string, file *syntax.File, diagnostics syntax.Diagnostics) (ParseResult, error) {
	return adaptSyntaxResultContext(context.Background(), uri, source, file, diagnostics)
}
func adaptSyntaxResultContext(ctx context.Context, uri, source string, file *syntax.File, diagnostics syntax.Diagnostics) (ParseResult, error) {
	result := ParseResult{File: file}
	for _, diagnostic := range diagnostics.SortBySpan() {
		mapped, err := syntaxDiagnostic(source, diagnostic)
		if err != nil {
			return ParseResult{}, err
		}
		result.Diagnostics = append(result.Diagnostics, mapped)
	}

	result.semanticChecked = file != nil && (file.Package != nil || file.Namespace != nil || len(syntaxDeclarations(file)) > 0)
	ids := make(map[loweredSymbolKey]string)
	names := make(map[string]string)
	if result.semanticChecked && !diagnostics.HasErrors() && file.Package != nil && file.Namespace != nil {
		ir, err := bidir.LowerContext(ctx, canonicalSyntaxFile(file))
		if err != nil {
			if errors.Is(err, bidir.ErrLowerCanceled) {
				return ParseResult{}, err
			}
			result.Diagnostics = append(result.Diagnostics, semanticDiagnostic(uri, source, file, err))
		} else {
			result.semanticValid = true
			ids, names = loweredIdentities(ir)
		}
	}

	if file != nil {
		if err := appendHeaderSymbols(&result, source, file); err != nil {
			return ParseResult{}, err
		}
		for _, declaration := range syntaxDeclarations(file) {
			if err := appendDeclaration(&result, source, declaration, ids, names); err != nil {
				return ParseResult{}, err
			}
		}
	}
	return normalizeParseResult(uri, source, result), nil
}
func normalizeParseResult(uri, source string, result ParseResult) ParseResult {
	result.Diagnostics = canonicalDiagnosticOrder(uri, source, result.Diagnostics)
	return result
}
