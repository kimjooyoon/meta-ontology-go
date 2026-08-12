package lsp

import (
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func adaptSyntaxResult(uri, source string, file *syntax.File, diagnostics syntax.Diagnostics) (ParseResult, error) {
	result := ParseResult{File: file}
	for _, diagnostic := range diagnostics.SortBySpan() {
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
	for _, input := range canonicalActivityParameters(activity) {
		if err := appendReference(result, source, input.Name, input.Span); err != nil {
			return err
		}
	}
	output := canonicalActivityOutput(source, activity)
	return appendReference(result, source, output.Name, output.Span)
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

func canonicalActivityParameters(activity *syntax.ActivityDecl) []syntax.NameRef {
	return activity.Inputs
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
