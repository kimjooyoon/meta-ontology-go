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
	result.Headers = append([]Symbol(nil), result.Headers...)
	result.Symbols = append([]Symbol(nil), result.Symbols...)
	result.References = append([]Reference(nil), result.References...)
	result.Diagnostics = append([]Diagnostic(nil), result.Diagnostics...)
	return result
}

type ParseResult struct {
	File        *syntax.File
	Headers     []Symbol
	Symbols     []Symbol
	References  []Reference
	Diagnostics []Diagnostic

	// semanticChecked distinguishes the authoritative syntax adapter from
	// test/integration parsers supplied through ParserFunc. A checked result
	// never falls back to name-only links when lowering was rejected.
	semanticChecked bool
	semanticValid   bool
}
type SymbolKind int
