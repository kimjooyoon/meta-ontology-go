package bodycodegen

import (
	"context"
	"fmt"
)

// RealizeCalledAssembly verifies the retained construction, then exposes its
// chosen body as a fixed callable declaration. The original contract, choices
// and partial scores remain in the caller's saved construction step.
func RealizeCalledAssembly(ctx context.Context, filename string, source []byte, prior Result) (Realization, error) {
	realized, err := RealizeSourceAssembly(ctx, filename, source, prior)
	if err != nil {
		return Realization{}, err
	}
	realized.Source, err = fixedCalledSource(filename, realized.Source, prior.Report.Activity)
	if err != nil {
		return Realization{}, err
	}
	realized.RealizedSourceSHA256 = digest([]byte(realized.Source))
	return realized, nil
}

func fixedCalledSource(filename, source, name string) (string, error) {
	file, diagnostics := ParseBodyFile(filename, []byte(source))
	if diagnostics.HasErrors() || file == nil {
		return "", fmt.Errorf("called body realization has syntax errors")
	}
	activity, err := sourceBodyActivity(file, name)
	if err != nil {
		return "", err
	}
	if activity.Assembly != nil {
		span := activity.Assembly.Span
		if span.Start.Offset < activity.ValueProgramSpan.End.Offset || span.End.Offset > len(source) ||
			span.Start.Offset > span.End.Offset {
			return "", fmt.Errorf("called assembly checkpoint spans differ")
		}
		source = source[:span.Start.Offset] + source[span.End.Offset:]
	}
	return source, nil
}
