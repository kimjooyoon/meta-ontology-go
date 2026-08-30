package lsp

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

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
		IdentityRange: identityRange, HasIdentity: hasIdentity,
		identityRange: identityRange, hasIdentity: hasIdentity,
	})
	for _, field := range entity.Fields {
		fieldRange, err := syntaxRange(source, field.Span)
		if err != nil {
			return err
		}
		fieldSelection, err := syntaxRange(source, field.NameSpan)
		if err != nil {
			return err
		}
		fieldIdentity, err := syntaxRange(source, field.IDSpan)
		if err != nil {
			return err
		}
			result.Symbols = append(result.Symbols, Symbol{
				Name: field.Name, ID: field.ID, Kind: SymbolField,
				Detail: "field " + field.Name, Range: fieldRange, SelectionRange: fieldSelection,
				IdentityRange: fieldIdentity, HasIdentity: true,
				identityRange: fieldIdentity, hasIdentity: true,
		})
	}
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
	if err := appendValueProgramFieldReference(result, source, activity); err != nil {
		return err
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
