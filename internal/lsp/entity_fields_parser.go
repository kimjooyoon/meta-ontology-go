package lsp

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// EntityFieldsSyntaxParser binds the V1 profile to the normal LSP adapter.
// It shares syntax spans, semantic lowering, symbols, and references with the
// ordinary parser instead of inventing a field-specific protocol.
type EntityFieldsSyntaxParser struct{}

func (EntityFieldsSyntaxParser) Parse(uri, source string) ParseResult {
	result, _ := (EntityFieldsSyntaxParser{}).ParseContext(context.Background(), uri, source)
	return result
}

func (EntityFieldsSyntaxParser) ParseContext(ctx context.Context, uri, source string) (ParseResult, error) {
	if err := ctx.Err(); err != nil {
		return ParseResult{}, err
	}
	support := syntax.EntityFieldsV1Support()
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(uri, source, support)
	return adaptSyntaxResultContextWithSupport(ctx, uri, source, file, diagnostics, support)
}
