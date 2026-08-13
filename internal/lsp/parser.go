package lsp

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// Parser is the small integration seam used by the LSP document store.
type Parser interface {
	Parse(uri, source string) ParseResult
}

type DocumentParser = Parser

type ParserFunc func(uri, source string) ParseResult

func (function ParserFunc) Parse(uri, source string) ParseResult {
	return function(uri, source)
}

// ContextParser is an optional parser seam for cancellation-aware adapters.
type ContextParser interface {
	ParseContext(ctx context.Context, uri, source string) (ParseResult, error)
}

type ContextParserFunc func(ctx context.Context, uri, source string) (ParseResult, error)

func (function ContextParserFunc) ParseContext(ctx context.Context, uri, source string) (ParseResult, error) {
	return function(ctx, uri, source)
}

func (function ContextParserFunc) Parse(uri, source string) ParseResult {
	result, _ := function(context.Background(), uri, source)
	return result
}

func (server *Server) parse(ctx context.Context, uri, source string) (ParseResult, error) {
	server.parseMu.Lock()
	defer server.parseMu.Unlock()
	if parser, ok := server.parser.(ContextParser); ok {
		result, err := parser.ParseContext(ctx, uri, source)
		return cloneParseResult(normalizeParseResult(uri, source, result)), err
	}
	if err := ctx.Err(); err != nil {
		return ParseResult{}, err
	}
	return cloneParseResult(normalizeParseResult(uri, source, server.parser.Parse(uri, source))), nil
}

// cloneParseResult gives the document store and feature readers ownership of
// the derived LSP projections. The canonical syntax tree remains read-only and
// is intentionally not copied or reinterpreted by this adapter.
func cloneParseResult(result ParseResult) ParseResult {
	result.Symbols = append([]Symbol(nil), result.Symbols...)
	result.References = append([]Reference(nil), result.References...)
	result.Diagnostics = append([]Diagnostic(nil), result.Diagnostics...)
	return result
}

type ParseResult struct {
	File        *syntax.File
	Symbols     []Symbol
	References  []Reference
	Diagnostics []Diagnostic
}

type SymbolKind int

const (
	SymbolFile      SymbolKind = 1
	SymbolNamespace SymbolKind = 3
	SymbolClass     SymbolKind = 5
	SymbolFunction  SymbolKind = 12
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
	// syntax.ParseFile is not interruptible. The default adapter therefore
	// guarantees cancellation only at the parse boundary; ContextParser
	// adapters can provide finer-grained cancellation when needed.
	if err := ctx.Err(); err != nil {
		return ParseResult{}, err
	}
	file, diagnostics := syntax.ParseFile(uri, source)
	return adaptSyntaxResult(uri, source, file, diagnostics)
}
