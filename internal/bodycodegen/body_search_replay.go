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
	ctx, localCalls := captureSearchCalls(ctx, filename, source, prior.Report.Activity)
	plan, err := replaySourceIRSearchPlan(ctx, filename, source, prior)
	if err != nil {
		return err
	}
	p, err := prepareSearchBody(filename, source, prior.Report.Activity, plan.HoleID)
	if err != nil {
		return fmt.Errorf("IR body-search source cannot be replayed: %w", err)
	}
	if p.activityID != prior.Report.ActivityID {
		return fmt.Errorf("IR body-search activity identity differs")
	}
	hole := bodyFillHoleToken(plan.HoleID)
	if err := replayIRSearchAttempts(ctx, p, prior, plan); err != nil {
		return err
	}
	completedBody, err := replaceIdentifier(p.body, hole, search.SelectedExpression)
	if err != nil {
		return fmt.Errorf("IR body-search selected expression cannot be restored")
	}
	completedSource, err := replaceActivityProgram(source, p.activity.ValueProgramSpan, completedBody)
	if err != nil {
		return fmt.Errorf("IR body-search completed source cannot be reconstructed")
	}
	if p.file == nil || digest(completedSource) != prior.Report.SourceDigest {
		return fmt.Errorf("IR body-search completed source digest differs")
	}
	replayed, err := GenerateWithPlanner(ctx, filename, completedSource, prior.Report.Activity, "", "")
	if err != nil || replayed.Source != prior.Source || replayed.Report.GeneratedDigest != prior.Report.GeneratedDigest ||
		replayed.Report.ReplayDigest != prior.Report.ReplayDigest || replayed.Report.ActivityID != prior.Report.ActivityID ||
		replayed.Report.ProgramDigest != prior.Report.ProgramDigest ||
		!EqualRetainedConditionOrigins(prior.Report.BodyConditions, replayed.Report.BodyConditions) ||
		!EqualRetainedReadOrigins(prior.Report.BodyReads, replayed.Report.BodyReads) ||
		!EqualRetainedConstructionOrigins(prior.Report.BodyConstructions, replayed.Report.BodyConstructions) ||
		!EqualRetainedLocalOrigins(prior.Report.BodyLocals, replayed.Report.BodyLocals) ||
		!EqualRetainedFieldOrigins(prior.Report.BodyFields, replayed.Report.BodyFields) ||
		!EqualRetainedReturnOrigins(prior.Report.BodyReturns, replayed.Report.BodyReturns) {
		return fmt.Errorf("IR body-search Go projection does not replay")
	}
	if err := replayIRSearchFiniteEvidence(ctx, prior, plan); err != nil {
		return err
	}
	if search.LocalCalls != nil && !equalLocalCallEvidence(search.LocalCalls, localCalls) {
		return fmt.Errorf("IR search local call observations differ from the source replay")
	}
	return nil
}
