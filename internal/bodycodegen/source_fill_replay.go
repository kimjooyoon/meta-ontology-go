package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

func realizeSourceIRBodyFill(ctx context.Context, filename string, source []byte, prior Result) (Realization, error) {
	if ctx == nil || len(prior.Source) > 256<<10 || len(prior.GoooSource) > 128<<10 {
		return Realization{}, fmt.Errorf("body-fill realization requires bounded source and context")
	}
	r := prior.Report
	if r.BodyFill == nil || r.BodyFill.OriginalSourceDigest != digest(source) || r.BodySearch != nil ||
		r.RecordAssembly != nil || r.BodyPaths != nil || r.Schema != schema || r.Decision != "PASS" ||
		!r.TypecheckPassed || !r.DeterministicReplay {
		return Realization{}, fmt.Errorf("body-fill observation is missing or source binding differs")
	}
	spec, err := SourceAssembly(ctx, filename, source, r.Activity)
	if err != nil || !IsSourceIRBodyFill(spec) {
		return Realization{}, fmt.Errorf("body-fill replay requires its original source_fill contract")
	}
	if err := spec.Validate(); err != nil {
		return Realization{}, err
	}
	plan, grammar, err := sourceIRBodyFillPlan(filename, source, r.Activity, spec)
	if err != nil {
		return Realization{}, err
	}
	options := IRBodyFillOptions{recordedDecision: &r.BodyFill.Decision, TinyModelLoadMS: r.BodyFill.Timing.TinyModelLoadMS}
	expected, err := generateWithIRBodyFillOptions(ctx, filename, source, r.Activity, plan, "", "", options, nil)
	if err != nil {
		return Realization{}, fmt.Errorf("replay source body fill: %w", err)
	}
	completed, err := sourceWithoutIRBodyFill(filename, []byte(expected.GoooSource), r.Activity)
	if err != nil {
		return Realization{}, err
	}
	expected.GoooSource, expected.Report.BodyFill.CandidateGeneration = string(completed), grammar
	populateCompletenessReceipt(&expected.Report, "")
	if err := compareSourceFillReplay(prior, expected); err != nil {
		return Realization{}, err
	}
	return Realization{Schema: "gooo/body-realization/v1", Source: expected.GoooSource,
		OriginalSourceSHA256: digest(source), GenerationSourceSHA256: r.SourceDigest,
		RealizedSourceSHA256: digest(completed), DocumentSHA256: r.BodyFill.IRPlanSHA256,
		TestSuiteSHA256: r.BodyFill.TestSuiteSHA256, ActivityID: r.ActivityID,
		FinitePassed: r.BodyFill.TestCasesPassed, FiniteTotal: r.BodyFill.TestCasesTotal}, nil
}

func compareSourceFillReplay(prior, expected Result) error {
	if prior.Source != expected.Source || prior.GoooSource != expected.GoooSource {
		return fmt.Errorf("body-fill selected source or projection differs")
	}
	a, b := prior.Report, expected.Report
	var err error
	a.BodyFill, err = canonicalFillObservation(a.BodyFill)
	if err != nil {
		return err
	}
	b.BodyFill, err = canonicalFillObservation(b.BodyFill)
	if err != nil {
		return err
	}
	a.CompilerSourceSHA, b.CompilerSourceSHA = "", ""
	a.RouteDecisionLatencyMS, b.RouteDecisionLatencyMS = 0, 0
	a.CompletenessReceipt, b.CompletenessReceipt = nil, nil
	if !reflect.DeepEqual(a, b) {
		return fmt.Errorf("body-fill candidate, selected case, holdout or projection evidence does not replay")
	}
	if err := completeness.Validate(prior.Report.CompletenessReceipt); err != nil {
		return err
	}
	actualCommon, _ := json.Marshal(prior.Report.CompletenessReceipt)
	expectedReceipt := buildCompletenessReceipt(prior.Report, "")
	// These are historical environment observations, not replay host requirements.
	for _, key := range []string{"toolchain", "execution_environment"} {
		expectedReceipt.Scope[key] = prior.Report.CompletenessReceipt.Scope[key]
	}
	expectedCommon, _ := json.Marshal(expectedReceipt)
	var actualValue, expectedValue any
	if err := json.Unmarshal(actualCommon, &actualValue); err != nil {
		return err
	}
	if err := json.Unmarshal(expectedCommon, &expectedValue); err != nil {
		return err
	}
	if !reflect.DeepEqual(actualValue, expectedValue) {
		return fmt.Errorf("body-fill completeness evidence differs from its bound observations")
	}
	return nil
}

func canonicalFillObservation(receipt *IRBodyFillReceipt) (*IRBodyFillReceipt, error) {
	value := *receipt
	value.Timing = IRBodyFillTiming{}
	var err error
	value.SelectedValueCaseResults, err = canonicalRecordCaseValues(receipt.SelectedValueCaseResults)
	if err == nil {
		value.SelectedValueHoldoutCaseResults, err = canonicalRecordCaseValues(receipt.SelectedValueHoldoutCaseResults)
	}
	return &value, err
}
