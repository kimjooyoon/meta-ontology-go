package lsp

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func syntaxDeclarations(file *syntax.File) []syntax.Declaration {
	if file == nil {
		return nil
	}
	if len(file.Decls) > 0 {
		return file.Decls
	}
	return file.Declarations
}
func canonicalSyntaxFile(file *syntax.File) *syntax.File {
	if file == nil || len(file.Decls) == 0 {
		return file
	}
	clone := *file
	clone.Declarations = append([]syntax.Declaration(nil), file.Decls...)
	return &clone
}
func appendHeaderSymbols(result *ParseResult, source string, file *syntax.File) error {
	if file.Package != nil && file.Package.Name != "" && !file.Package.NameSpan.IsEmpty() {
		rangeValue, err := syntaxRange(source, file.Package.Span)
		if err != nil {
			return err
		}
		selection, err := syntaxRange(source, file.Package.NameSpan)
		if err != nil {
			return err
		}
		result.Headers = append(result.Headers, Symbol{
			Name: file.Package.Name, Kind: SymbolPackage, Detail: "package " + file.Package.Name,
			Range: rangeValue, SelectionRange: selection,
		})
	}
	if file.Namespace != nil && file.Namespace.Name != "" && !file.Namespace.NameSpan.IsEmpty() {
		rangeValue, err := syntaxRange(source, file.Namespace.Span)
		if err != nil {
			return err
		}
		selection, err := syntaxRange(source, file.Namespace.NameSpan)
		if err != nil {
			return err
		}
		result.Headers = append(result.Headers, Symbol{
			Name: file.Namespace.Name, Kind: SymbolNamespace, Detail: "namespace " + file.Namespace.Name,
			Range: rangeValue, SelectionRange: selection,
		})
	}
	return nil
}
func appendDeclaration(result *ParseResult, source string, declaration syntax.Declaration, ids map[loweredSymbolKey]string, names map[string]string) error {
	switch value := declaration.(type) {
	case *syntax.EntityDecl:
		return appendEntity(result, source, value, ids)
	case *syntax.ActivityDecl:
		return appendActivity(result, source, value, ids, names)
	default:
		return nil
	}
}
