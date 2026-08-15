package lsp

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func loweredIdentities(ir semantic.IR) (map[loweredSymbolKey]string, map[string]string) {
	ids := make(map[loweredSymbolKey]string)
	names := make(map[string]string)
	for _, node := range ir.Graph.Nodes() {
		if _, ok := lspSymbolKind(node.Kind); !ok {
			continue
		}
		key := loweredSymbolKey{start: node.Span.Start.Offset, end: node.Span.End.Offset, kind: node.Kind, name: node.Name}
		ids[key] = node.ID.String()
		if _, exists := names[node.Name]; !exists {
			names[node.Name] = node.ID.String()
		}
	}
	return ids, names
}

func lspSymbolKind(kind semantic.Kind) (SymbolKind, bool) {
	switch kind {
	case semantic.Entity:
		return SymbolClass, true
	case semantic.Activity:
		return SymbolFunction, true
	default:
		return 0, false
	}
}

func semanticDiagnostic(uri, source string, file *syntax.File, err error) Diagnostic {
	span := syntax.Span{
		Filename: uri,
		Start:    syntax.Position{Offset: 0, Line: 1, Column: 1},
		End:      syntax.Position{Offset: len(source), Line: 1, Column: 1},
	}
	if file != nil && !file.Span.IsEmpty() {
		span = file.Span
	}
	rangeValue, rangeErr := syntaxRange(source, span)
	if rangeErr != nil {
		rangeValue = Range{}
	}
	return Diagnostic{
		Range: rangeValue, Severity: DiagnosticError, Code: "semantic.lowering",
		Source: "gooo", Message: err.Error(), filename: uri,
		start: span.Start.Offset, end: span.End.Offset, spanned: true,
	}
}
