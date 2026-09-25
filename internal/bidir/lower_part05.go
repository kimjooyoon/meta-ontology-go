package bidir

import (
	"context"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func adaptSyntaxActivity(ctx context.Context, activity *syntax.ActivityDecl) (Declaration, error) {
	if err := checkLowerContext(ctx); err != nil {
		return Declaration{}, err
	}
	declaration := Declaration{Kind: ActivityKind, Name: activity.Name, Span: toSourceSpan(activity.Span)}
	if activity.ValueProgramPresent || activity.ValueProgram != "" {
		declaration.Attributes = map[string]string{ActivityValueProgramAttribute: activity.ValueProgram}
	}
	if len(activity.Inputs) == 0 && len(activity.Parameters) != 0 {
		return Declaration{}, fmt.Errorf("activity %q uses unsupported legacy-only Parameters; canonical Inputs is required", activity.Name)
	}
	for _, input := range activity.Inputs {
		if err := checkLowerContext(ctx); err != nil {
			return Declaration{}, err
		}
		declaration.Inputs = append(declaration.Inputs, Reference{Name: input.Name, Span: toSourceSpan(input.Span)})
	}
	if activity.Output == "" && activity.Result.Name != "" {
		return Declaration{}, fmt.Errorf("activity %q uses unsupported legacy-only Result; canonical Output is required", activity.Name)
	}
	if activity.Output != "" {
		declaration.Outputs = append(declaration.Outputs, Reference{Name: activity.Output, Span: toSourceSpan(activity.Span)})
	}
	return declaration, nil
}
func toSourceSpan(span syntax.Span) SourceSpan {
	return SourceSpan{
		File:        span.Filename,
		Start:       span.Start.Offset,
		End:         span.End.Offset,
		StartLine:   span.Start.Line,
		StartColumn: span.Start.Column,
		EndLine:     span.End.Line,
		EndColumn:   span.End.Column,
	}
}
func toSemanticSpan(span SourceSpan) semantic.Span {
	return semantic.Span{
		File:  span.File,
		Start: semantic.Position{Offset: span.Start, Line: span.StartLine, Column: span.StartColumn},
		End:   semantic.Position{Offset: span.End, Line: span.EndLine, Column: span.EndColumn},
	}
}
