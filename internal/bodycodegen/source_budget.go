package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// ReviseAssemblyBudget changes only the selected source-owned contract. The
// body, choices, cases, identity and other activities remain authoritative.
func ReviseAssemblyBudget(ctx context.Context, filename string, source []byte, name string, attempts int) ([]byte, error) {
	spec, err := SourceAssembly(ctx, filename, source, name)
	if err != nil {
		return nil, err
	}
	if spec == nil || spec.FillPlan != nil || attempts < 1 || attempts > 64 {
		return nil, fmt.Errorf("budget revision requires a choice/search assembly and 1..64 attempts")
	}
	if attempts == spec.MaxAttempts {
		return append([]byte(nil), source...), nil
	}
	spec.MaxAttempts = attempts
	return rewriteAssemblyContract(filename, source, name, spec)
}

func rewriteAssemblyContract(filename string, source []byte, name string, spec *assemblyspec.Spec) ([]byte, error) {
	contract, err := syntax.FormatAssembly(&syntax.AssemblyDecl{Spec: *spec})
	if err != nil {
		return nil, err
	}
	file, diagnostics := ParseBodyFile(filename, source)
	if diagnostics.HasErrors() || file == nil {
		return nil, fmt.Errorf("budget source: %v", diagnostics)
	}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != name || activity.Assembly == nil {
			continue
		}
		span := activity.Assembly.Span
		if span.Start.Offset < 0 || span.End.Offset < span.Start.Offset || span.End.Offset > len(source) {
			return nil, fmt.Errorf("budget source span is inconsistent")
		}
		next := append([]byte(nil), source[:span.Start.Offset]...)
		next = append(next, contract...)
		next = append(next, source[span.End.Offset:]...)
		if len(next) > 128<<10 {
			return nil, fmt.Errorf("budget revision exceeds the source limit")
		}
		return next, nil
	}
	return nil, fmt.Errorf("budget activity is missing")
}
