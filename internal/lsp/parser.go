package lsp

import (
	"context"
	"fmt"

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
	if err := ctx.Err(); err != nil {
		return ParseResult{}, err
	}
	file, diagnostics := syntax.ParseFile(uri, source)
	result := ParseResult{File: file}
	for _, diagnostic := range diagnostics {
		mapped, err := syntaxDiagnostic(source, diagnostic)
		if err != nil {
			return ParseResult{}, err
		}
		result.Diagnostics = append(result.Diagnostics, mapped)
	}
	for _, declaration := range syntaxDeclarations(file) {
		if err := appendDeclaration(&result, source, declaration); err != nil {
			return ParseResult{}, err
		}
	}
	return result, nil
}

func syntaxDiagnostic(source string, diagnostic syntax.Diagnostic) (Diagnostic, error) {
	rangeValue, err := syntaxRange(source, diagnostic.Span)
	if err != nil {
		return Diagnostic{}, err
	}
	severity := DiagnosticError
	if diagnostic.Severity == syntax.SeverityWarning {
		severity = DiagnosticWarning
	}
	return Diagnostic{
		Range: rangeValue, Severity: severity, Code: string(diagnostic.Code),
		Source: "gooo", Message: diagnostic.Message,
	}, nil
}

func syntaxDeclarations(file *syntax.File) []syntax.Declaration {
	if file == nil {
		return nil
	}
	if file.Declarations != nil {
		return file.Declarations
	}
	return file.Decls
}

func appendDeclaration(result *ParseResult, source string, declaration syntax.Declaration) error {
	switch value := declaration.(type) {
	case *syntax.EntityDecl:
		return appendEntity(result, source, value)
	case *syntax.ActivityDecl:
		return appendActivity(result, source, value)
	default:
		return nil
	}
}

func appendEntity(result *ParseResult, source string, entity *syntax.EntityDecl) error {
	rangeValue, err := syntaxRange(source, entity.Span)
	if err != nil {
		return err
	}
	selection, err := syntaxRange(source, entity.NameSpan)
	if err != nil {
		return err
	}
	result.Symbols = append(result.Symbols, Symbol{
		Name: entity.Name, ID: entity.ID, Kind: SymbolClass,
		Detail: "entity " + entity.Name, Range: rangeValue, SelectionRange: selection,
	})
	return nil
}

func appendActivity(result *ParseResult, source string, activity *syntax.ActivityDecl) error {
	rangeValue, err := syntaxRange(source, activity.Span)
	if err != nil {
		return err
	}
	selection, err := syntaxRange(source, activity.NameSpan)
	if err != nil {
		return err
	}
	result.Symbols = append(result.Symbols, Symbol{
		Name: activity.Name, Kind: SymbolFunction, Detail: "activity " + activity.Name,
		Range: rangeValue, SelectionRange: selection,
	})
	for _, input := range activity.Inputs {
		if err := appendReference(result, source, input.Name, input.Span); err != nil {
			return err
		}
	}
	if err := appendReference(result, source, activity.Result.Name, activity.Result.Span); err != nil {
		return err
	}
	return nil
}

func appendReference(result *ParseResult, source, name string, span syntax.Span) error {
	if name == "" || span.IsEmpty() {
		return nil
	}
	rangeValue, err := syntaxRange(source, span)
	if err != nil {
		return err
	}
	result.References = append(result.References, Reference{Name: name, Range: rangeValue})
	return nil
}

func syntaxRange(source string, span syntax.Span) (Range, error) {
	start, err := OffsetToPosition(source, span.Start.Offset)
	if err != nil {
		return Range{}, fmt.Errorf("lsp: invalid syntax start span: %w", err)
	}
	end, err := OffsetToPosition(source, span.End.Offset)
	if err != nil {
		return Range{}, fmt.Errorf("lsp: invalid syntax end span: %w", err)
	}
	return Range{Start: start, End: end}, nil
}
