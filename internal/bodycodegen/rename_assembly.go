package bodycodegen

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func renameAssemblyReferences(a *syntax.ActivityDecl, source []byte, from, to string,
	result *SourceRenameResult, add func(syntax.Span, string, string)) error {
	if a.Assembly == nil {
		return nil
	}
	spec := a.Assembly.Spec
	if spec.FillPlan != nil || spec.Search != nil {
		return fmt.Errorf("activity %s: source_fill/search rename needs a candidate-scope mapping", a.Name)
	}
	start, end := a.Assembly.Span.Start.Offset, a.Assembly.Span.End.Offset
	tokens, diagnostics := syntax.Lex(string(source[start:end]))
	if diagnostics.HasErrors() {
		return diagnostics.Error()
	}
	choiceIndex := -1
	bindings, err := enclosingReferenceBindings([]string{a.ValueProgram, spec.Baseline}, len(a.Inputs))
	if err != nil {
		return err
	}
	for i, t := range tokens {
		if t.Kind == syntax.TokenIdentifier && t.Value == "choice" {
			choiceIndex++
		}
		if i == 0 || t.Kind != syntax.TokenString {
			continue
		}
		previous := tokens[i-1]
		updated, count := t.Value, 0
		switch {
		case previous.Kind == syntax.TokenActivity || previous.Kind == syntax.TokenIdentifier && previous.Value == "oracle":
			if t.Value == from {
				updated = to
			}
		case previous.Kind == syntax.TokenIdentifier && previous.Value == "baseline":
			updated, count, err = renameBodyCalls(t.Value, len(a.Inputs), from, to)
		case previous.Kind == syntax.TokenIdentifier && previous.Value == "alternative":
			if choiceIndex < 0 || choiceIndex >= len(spec.Choices) {
				return fmt.Errorf("unresolved choice span")
			}
			kind := spec.Choices[choiceIndex].Kind
			if kind != "field_value" && kind != "field_update" {
				continue
			}
			updated, count, err = renameBodyCalls("return "+t.Value, len(a.Inputs), from, to)
			updated = strings.TrimPrefix(updated, "return ")
			if count > 0 && (bindings[from] || bindings[to]) {
				return fmt.Errorf("activity %s: alternative requires precise enclosing scope for rename", a.Name)
			}
		}
		if err != nil {
			return fmt.Errorf("activity %s assembly: %w", a.Name, err)
		}
		if updated == t.Value {
			continue
		}
		span := t.Span
		span.Start.Offset += start
		span.End.Offset += start
		add(span, "assembly_reference", renameString(updated, t.Text))
		result.CallSites += count
	}
	return nil
}
