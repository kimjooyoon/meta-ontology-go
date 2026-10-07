package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

// ValidateSourceIRSearch checks the finite grammar and at least one typed
// candidate. The residual grammar observes pure training-input probes; preflight
// does not select a final body or call a provider.
func ValidateSourceIRSearch(ctx context.Context, filename string, source []byte, activity string,
	spec *assemblyspec.Spec) error {
	if ctx == nil {
		return fmt.Errorf("source IR search requires a context")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	plan, err := sourceIRSearchPlan(spec)
	if err != nil {
		return err
	}
	if err := validateIRBodySearchPlan(plan); err != nil {
		return err
	}
	if _, err = generateIRBodySearchCandidatesForSource(ctx, filename, source, activity, &plan); err != nil {
		return err
	}
	file, _, id, body, err := prepareBodySearch(filename, source, activity, plan.HoleID)
	if err != nil {
		return err
	}
	var last error
	for _, candidate := range plan.Candidates {
		if err := ctx.Err(); err != nil {
			return err
		}
		filled, err := replaceIdentifier(body, bodyFillHoleToken(plan.HoleID), candidate.Expression)
		if err != nil {
			return err
		}
		_, last = generateRoute(file.Package.Name, activity, id, "int64", "int64", filled, preserveRoute)
		if last == nil {
			return nil
		}
	}
	return fmt.Errorf("source IR search has no typed candidate: %w", last)
}

func realizeSourceIRSearch(ctx context.Context, filename string, source []byte, prior Result) (Realization, error) {
	if err := VerifyIRBodySearchProjection(ctx, filename, source, prior); err != nil {
		return Realization{}, err
	}
	spec, err := SourceAssembly(ctx, filename, source, prior.Report.Activity)
	if err != nil {
		return Realization{}, err
	}
	_, activity, _, body, err := prepareBodySearch(filename, source, prior.Report.Activity, spec.Search.HoleID)
	if err != nil {
		return Realization{}, err
	}
	search := prior.Report.BodySearch
	filled, err := replaceIdentifier(body, bodyFillHoleToken(spec.Search.HoleID), search.SelectedExpression)
	if err != nil {
		return Realization{}, err
	}
	start, end := activity.Assembly.Span.Start.Offset, activity.Assembly.Span.End.Offset
	if start < activity.ValueProgramSpan.End.Offset || end < start || end > len(source) {
		return Realization{}, fmt.Errorf("source IR search spans are inconsistent")
	}
	if start > 0 && source[start-1] == ' ' {
		start--
	}
	stripped := append(append([]byte(nil), source[:start]...), source[end:]...)
	completed, err := replaceActivityProgram(stripped, activity.ValueProgramSpan, filled)
	if err != nil {
		return Realization{}, err
	}
	return Realization{Schema: "gooo/body-realization/v1", Source: string(completed),
		OriginalSourceSHA256: digest(source), GenerationSourceSHA256: prior.Report.SourceDigest,
		RealizedSourceSHA256: digest(completed), DocumentSHA256: search.IRPlanSHA256,
		TestSuiteSHA256: search.TrainingSuiteSHA256, ActivityID: prior.Report.ActivityID,
		FinitePassed: search.TrainingPassed, FiniteTotal: search.TrainingTotal}, nil
}
