package bodycodegen

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// VerifyTypedPathProjection reconstructs the declared selection from the original
// source and document without loading a model. When an observation loop was used,
// its bounded candidate observations and oracle labels are independently replayed.
// This validates the emitted pure program; it does not attest earlier model execution.
func VerifyTypedPathProjection(ctx context.Context, filename string, source []byte,
	document pathplan.Document, prior Result) error {
	return replayTypedPathProjection(ctx, filename, source, document, prior, nil)
}

func replayTypedPathProjection(ctx context.Context, filename string, source []byte,
	document pathplan.Document, prior Result, completedSource *[]byte) error {
	if err := validateTypedPathReplayInput(ctx, source, prior); err != nil {
		return err
	}
	r, p := prior.Report, prior.Report.BodyPaths
	prepared, err := document.Prepare()
	if err != nil {
		return err
	}
	observed := &BodyPathReceipt{OriginalSourceSHA256: digest(source)}
	if err := bindTypedPathDocument(document, observed); err != nil {
		return err
	}
	if observed.DocumentSHA256 != p.DocumentSHA256 || observed.TestSuiteSHA256 != p.TestSuiteSHA256 ||
		observed.DeclaredTestCases != p.DeclaredTestCases || len(bodyPathSelection(p).Choices) != len(document.Plan.Decisions) {
		return fmt.Errorf("typed-path document, finite suite or complete selection differs")
	}
	bound, err := bindTypedPathSource(ctx, filename, source, r.Activity, prepared, observed)
	if err != nil {
		return err
	}
	effectiveCases, err := replayPathObservation(ctx, filename, source, r.Activity, prepared, document.TestCases, p)
	if err != nil {
		return err
	}
	if err := replayPathResolution(ctx, document, prepared, p); err != nil {
		return err
	}
	selected, err := prepared.Compile(bodyPathSelection(p).Choices)
	if err != nil {
		return err
	}
	if p.SourceFormat != "" && bound.activity.Assembly == nil {
		return fmt.Errorf("assembly checkpoint requires its source-owned contract")
	}
	completed, err := selectedTypedPathSource(source, bound.activity, selected.GoooBody(), bodyPathSelection(p).Choices,
		p.SourceFormat == sourceAssemblyCheckpointFormat)
	if err != nil {
		return err
	}
	if (p.SourceFormat == "" && prior.GoooSource != "") ||
		(p.SourceFormat != "" && prior.GoooSource != string(completed)) {
		return fmt.Errorf("generation Gooo source does not match the declared selection")
	}
	replayed, err := GenerateWithPlanner(ctx, filename, completed, r.Activity, "", "")
	if err != nil {
		return err
	}
	if err := compareTypedPathReplayProjection(prior, replayed, bound.base.Report.ActivityID); err != nil {
		return err
	}
	conditions, err := selectedPathConditions(ctx, prepared, bodyPathSelection(p).Choices,
		selected, r.Activity, []byte(replayed.Source))
	if err != nil || !reflect.DeepEqual(conditions, p.Conditions) {
		return fmt.Errorf("source condition observations do not replay: %v", err)
	}
	if conditions != nil && !sameConditionResults(conditions.Results, bodyPathSelection(p).Conditions) {
		return fmt.Errorf("selection condition observations differ from declared source")
	}
	if err := verifyTypedPathFiniteObservations(ctx, source, completed, replayed, bound, effectiveCases, p); err != nil {
		return err
	}
	if completedSource != nil {
		*completedSource = completed
	}
	return nil
}
