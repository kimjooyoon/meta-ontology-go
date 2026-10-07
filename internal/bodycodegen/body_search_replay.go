package bodycodegen

import (
	"context"
	"fmt"
)

// VerifyIRBodySearchProjection binds a prior selected expression to its
// original source and deterministically reconstructs its emitted Go before
// native execution. The source-owned candidate set and selected finite scores
// are replayed without repeating model selection or attesting prior model calls.
func VerifyIRBodySearchProjection(ctx context.Context, filename string, source []byte, prior Result) error {
	search := prior.Report.BodySearch
	if search == nil || search.OriginalSourceDigest != digest(source) || search.SelectedExpression == "" {
		return fmt.Errorf("IR body-search source binding is missing or changed")
	}
	plan, err := replaySourceIRSearchPlan(ctx, filename, source, prior)
	if err != nil {
		return err
	}
	file, activity, activityID, body, err := prepareBodySearch(filename, source, prior.Report.Activity, plan.HoleID)
	if err != nil {
		return fmt.Errorf("IR body-search source cannot be replayed: %w", err)
	}
	if activityID != prior.Report.ActivityID {
		return fmt.Errorf("IR body-search activity identity differs")
	}
	hole := bodyFillHoleToken(plan.HoleID)
	completedBody, err := replaceIdentifier(body, hole, search.SelectedExpression)
	if err != nil {
		return fmt.Errorf("IR body-search selected expression cannot be restored")
	}
	completedSource, err := replaceActivityProgram(source, activity.ValueProgramSpan, completedBody)
	if err != nil {
		return fmt.Errorf("IR body-search completed source cannot be reconstructed")
	}
	if file == nil || digest(completedSource) != prior.Report.SourceDigest {
		return fmt.Errorf("IR body-search completed source digest differs")
	}
	replayed, err := GenerateWithPlanner(ctx, filename, completedSource, prior.Report.Activity, "", "")
	if err != nil || replayed.Source != prior.Source || replayed.Report.GeneratedDigest != prior.Report.GeneratedDigest ||
		replayed.Report.ReplayDigest != prior.Report.ReplayDigest || replayed.Report.ActivityID != prior.Report.ActivityID ||
		replayed.Report.ProgramDigest != prior.Report.ProgramDigest {
		return fmt.Errorf("IR body-search Go projection does not replay")
	}
	return replayIRSearchFiniteEvidence(ctx, prior, plan)
}
