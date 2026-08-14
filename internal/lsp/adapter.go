package lsp

import (
	"context"
	"errors"
	"fmt"
	"sort"

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

// canonicalDiagnosticOrder preserves syntax source order and adds only the
// LSP view's deterministic tie-breaks for diagnostics sharing a start.
func canonicalDiagnosticOrder(uri, source string, diagnostics []Diagnostic) []Diagnostic {
	result := append([]Diagnostic(nil), diagnostics...)
	for index := range result {
		if result[index].filename == "" {
			result[index].filename = uri
		}
		if !result[index].spanned {
			result[index].start, _ = PositionToOffset(source, result[index].Range.Start)
			result[index].end, _ = PositionToOffset(source, result[index].Range.End)
		}
	}
	sort.SliceStable(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.filename != second.filename {
			return first.filename < second.filename
		}
		if first.start != second.start {
			return first.start < second.start
		}
		if first.end != second.end {
			return first.end < second.end
		}
		if first.Range.Start != second.Range.Start {
			return positionLess(first.Range.Start, second.Range.Start)
		}
		if first.Range.End != second.Range.End {
			return positionLess(first.Range.End, second.Range.End)
		}
		if first.Severity != second.Severity {
			return first.Severity < second.Severity
		}
		if first.Code != second.Code {
			return first.Code < second.Code
		}
		return first.Message < second.Message
	})
	return result
}

func positionLess(first, second Position) bool {
	if first.Line != second.Line {
		return first.Line < second.Line
	}
	return first.Character < second.Character
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
		Source: "gooo", Message: diagnostic.Message, filename: diagnostic.Span.Filename,
		start: diagnostic.Span.Start.Offset, end: diagnostic.Span.End.Offset, spanned: true,
	}, nil
}

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

func appendEntity(result *ParseResult, source string, entity *syntax.EntityDecl, ids map[loweredSymbolKey]string) error {
	rangeValue, err := syntaxRange(source, entity.Span)
	if err != nil {
		return err
	}
	selection, err := syntaxRange(source, entity.NameSpan)
	if err != nil {
		return err
	}
	id := ids[loweredSymbolKey{start: entity.Span.Start.Offset, end: entity.Span.End.Offset, kind: semantic.Entity, name: entity.Name}]
	identityRange := Range{}
	hasIdentity := false
	if id != "" && entity.ID != "" && !entity.IDSpan.IsEmpty() {
		identityRange, err = syntaxRange(source, entity.IDSpan)
		if err != nil {
			return err
		}
		hasIdentity = true
	}
	result.Symbols = append(result.Symbols, Symbol{
		Name: entity.Name, ID: id, Kind: SymbolClass,
		Detail: "entity " + entity.Name, Range: rangeValue, SelectionRange: selection,
		identityRange: identityRange, hasIdentity: hasIdentity,
	})
	return nil
}

func appendActivity(result *ParseResult, source string, activity *syntax.ActivityDecl, ids map[loweredSymbolKey]string, names map[string]string) error {
	rangeValue, err := syntaxRange(source, activity.Span)
	if err != nil {
		return err
	}
	selection, err := syntaxRange(source, activity.NameSpan)
	if err != nil {
		return err
	}
	id := ids[loweredSymbolKey{start: activity.Span.Start.Offset, end: activity.Span.End.Offset, kind: semantic.Activity, name: activity.Name}]
	result.Symbols = append(result.Symbols, Symbol{
		Name: activity.Name, ID: id, Kind: SymbolFunction, Detail: "activity " + activity.Name,
		Range: rangeValue, SelectionRange: selection,
	})
	for _, input := range activity.Inputs {
		if err := appendReference(result, source, input.Name, input.Span, names[input.Name]); err != nil {
			return err
		}
	}
	output := canonicalActivityOutput(source, activity)
	return appendReference(result, source, output.Name, output.Span, names[output.Name])
}

func canonicalActivityOutput(source string, activity *syntax.ActivityDecl) syntax.NameRef {
	if activity.Output == "" {
		return syntax.NameRef{}
	}
	end := activity.Span.End.Offset
	start := end - len(activity.Output)
	if start < 0 || end > len(source) || source[start:end] != activity.Output {
		return syntax.NameRef{Name: activity.Output}
	}
	return syntax.NameRef{
		Name: activity.Output,
		Span: syntax.Span{Filename: activity.Span.Filename, Start: syntax.Position{Offset: start}, End: syntax.Position{Offset: end}},
	}
}

func appendReference(result *ParseResult, source, name string, span syntax.Span, id string) error {
	if name == "" || span.IsEmpty() {
		return nil
	}
	rangeValue, err := syntaxRange(source, span)
	if err != nil {
		return err
	}
	result.References = append(result.References, Reference{Name: name, ID: id, Range: rangeValue})
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
