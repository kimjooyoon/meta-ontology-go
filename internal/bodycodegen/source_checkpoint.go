package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const sourceAssemblyCheckpointFormat = "gooo/source-assembly-checkpoint/v1"

// Selectors stay relative to their original baseline even when a chosen local,
// branch or declaration moves. Only the selected activity's two spans change.
func selectedTypedPathSource(source []byte, activity *syntax.ActivityDecl, body string,
	choices map[string]string, checkpoint bool) ([]byte, error) {
	if !checkpoint || activity.Assembly == nil {
		return replaceActivityProgram(source, activity.ValueProgramSpan, body)
	}
	spec := activity.Assembly.Spec.Clone()
	if spec.Baseline == "" {
		spec.Baseline = activity.ValueProgram
	}
	spec.Picked = nil
	if len(choices) != len(spec.Choices) {
		return nil, fmt.Errorf("checkpoint requires a complete declared selection")
	}
	for _, choice := range spec.Choices {
		spec.Picked = append(spec.Picked, assemblyspec.Pick{ID: choice.ID, Label: choices[choice.ID]})
	}
	contract, err := syntax.FormatAssembly(&syntax.AssemblyDecl{Spec: *spec})
	if err != nil {
		return nil, err
	}
	span := activity.Assembly.Span
	if span.Start.Offset < activity.ValueProgramSpan.End.Offset || span.End.Offset < span.Start.Offset ||
		span.End.Offset > len(source) {
		return nil, fmt.Errorf("assembly source spans are inconsistent")
	}
	updated := append([]byte(nil), source[:span.Start.Offset]...)
	updated = append(updated, contract...)
	updated = append(updated, source[span.End.Offset:]...)
	updated, err = replaceActivityProgram(updated, activity.ValueProgramSpan, body)
	if err != nil {
		return nil, err
	}
	if len(updated) > 128<<10 {
		return nil, fmt.Errorf("assembly checkpoint exceeds its 128 KiB source budget")
	}
	return updated, nil
}

func sourceAssemblyPlanningSource(ctx context.Context, filename string, source []byte, activity string) ([]byte, error) {
	spec, err := SourceAssembly(ctx, filename, source, activity)
	if err != nil || spec == nil || spec.Baseline == "" {
		return source, err
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return nil, fmt.Errorf("assembly checkpoint source: %v", diagnostics)
	}
	for _, declaration := range file.Declarations {
		if d, ok := declaration.(*syntax.ActivityDecl); ok && d.Name == activity {
			return replaceActivityProgram(source, d.ValueProgramSpan, spec.Baseline)
		}
	}
	return nil, fmt.Errorf("assembly checkpoint activity missing")
}

func verifySourceCheckpoint(ctx context.Context, activity *syntax.ActivityDecl, packageName, activityID string,
	selectedBody string) error {
	actual, err := rewriteLetDeclarations(activity.ValueProgram)
	if err != nil {
		return err
	}
	selected, err := rewriteLetDeclarations(selectedBody)
	if err != nil {
		return err
	}
	route, err := generateRoute(packageName, activity.Name, activityID, "int64", "int64", selected, preserveRoute)
	if err != nil {
		return err
	}
	match, err := typedBodyTreeEquivalence(ctx, activity.Name, actual, route.source)
	if err != nil || !match.Equivalent {
		return fmt.Errorf("assembly computes body differs from its baseline and picked labels: %v", err)
	}
	return nil
}
