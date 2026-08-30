package lsp

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	SymbolFile      SymbolKind = 1
	SymbolNamespace SymbolKind = 3
	SymbolPackage   SymbolKind = 4
	SymbolClass     SymbolKind = 5
	SymbolFunction  SymbolKind = 12
	SymbolField     SymbolKind = 8
	SymbolKeyword   SymbolKind = 14
	SymbolText      SymbolKind = 1
)

type Symbol struct {
	Name           string
	ID             string
	Kind           SymbolKind
	Detail         string
	Range          Range
	SelectionRange Range

	// identityRange preserves the canonical syntax span of an explicit
	// semantic identity literal for local navigation. Activity identities are
	// intentionally absent here because their IDs belong to semantic lowering.
	identityRange Range
	hasIdentity   bool
}
type Reference struct {
	Name  string
	ID    string
	Range Range
}

// SyntaxParser consumes internal/syntax directly; it does not duplicate its
// lexer, parser, AST, source spans, or diagnostic codes.
type SyntaxParser struct{}

func (SyntaxParser) Parse(uri, source string) ParseResult {
	result, _ := (SyntaxParser{}).ParseContext(context.Background(), uri, source)
	return result
}
func (SyntaxParser) ParseContext(ctx context.Context, uri, source string) (ParseResult, error) {

	if err := ctx.Err(); err != nil {
		return ParseResult{}, err
	}
	file, diagnostics := syntax.ParseFile(uri, source)
	return adaptSyntaxResultContext(ctx, uri, source, file, diagnostics)
}
