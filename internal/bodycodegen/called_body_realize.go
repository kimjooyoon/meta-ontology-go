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
	file, diagnostics := ParseBodyFile(filename, []byte(realized.Source))
	if diagnostics.HasErrors() || file == nil {
		return Realization{}, fmt.Errorf("called body realization has syntax errors")
	}
	activity, err := sourceBodyActivity(file, prior.Report.Activity)
	if err != nil {
		return Realization{}, err
	}
	if activity.Assembly != nil {
		span := activity.Assembly.Span
		if span.Start.Offset < activity.ValueProgramSpan.End.Offset || span.End.Offset > len(realized.Source) ||
			span.Start.Offset > span.End.Offset {
			return Realization{}, fmt.Errorf("called assembly checkpoint spans differ")
		}
		realized.Source = realized.Source[:span.Start.Offset] + realized.Source[span.End.Offset:]
		realized.RealizedSourceSHA256 = digest([]byte(realized.Source))
	}
	return realized, nil
}
