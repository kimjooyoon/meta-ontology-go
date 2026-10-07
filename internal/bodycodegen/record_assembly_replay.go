package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"slices"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

func bindRecordAssemblyScope(common *CompletenessReceipt, result Report) {
	r := result.RecordAssembly
	raw, _ := json.Marshal(r)
	common.Scope["record_assembly"] = map[string]any{"schema": r.Schema, "observation_sha256": digest(raw),
		"contract_sha256": r.ContractSHA256, "test_suite_sha256": r.TestSuiteSHA256,
		"finite_passed": r.Passed, "finite_total": r.Total, "fields_passed": r.FieldsPassed, "fields_total": r.FieldsTotal,
		"local_model_predictions": r.ModelCalls, "external_provider_calls": 0}
	common.Scope["domain_scope"] = "one source-owned pure record assembly with 1..16 typed inputs and required string fields"
	common.Scope["decision_mode"], common.Scope["decision_provider"] = "deterministic_finite_tdd", "deterministic"
	if r.ModelCalls > 0 {
		common.Scope["decision_mode"], common.Scope["decision_provider"] = "local_ordinal_prediction_then_finite_tdd", "tiny_go"
		if r.Context != nil && r.Context.FeatureVersion == jointdecision.RecordFieldFeatureVersion {
			common.Scope["decision_mode"] = "local_field_expression_prediction_then_finite_tdd"
		}
		if r.Context != nil && r.Context.FeatureVersion == jointdecision.RecordSharedFeatureVersion {
			common.Scope["decision_mode"] = "local_shared_field_prediction_then_finite_tdd"
		}
		if r.Context != nil && r.Context.FeatureVersion == jointdecision.RecordOriginSharedFeatureVersion {
			common.Scope["decision_mode"] = "local_source_origin_prediction_then_finite_tdd"
		}
	}
}

// Replay reconstructs recorded selections and finite observations. The model
// identity and historical prediction are carried as observations, not re-run.
func realizeRecordAssembly(ctx context.Context, filename string, source []byte, prior Result) (Realization, error) {
	if ctx == nil || len(prior.Source) > 256<<10 || len(prior.GoooSource) > 128<<10 {
		return Realization{}, fmt.Errorf("record realization requires bounded source and context")
	}
	r := prior.Report.RecordAssembly
	if r.Schema != recordAssemblySchema || prior.Report.Schema != schema || prior.Report.Decision != "PASS" ||
		prior.Report.BodyPaths != nil || prior.Report.BodyFill != nil || prior.Report.BodySearch != nil ||
		!prior.Report.TypecheckPassed || !prior.Report.DeterministicReplay || r.OriginalSourceSHA256 != digest(source) {
		return Realization{}, fmt.Errorf("record assembly observation is missing or unbound")
	}
	plan, err := prepareRecordAssembly(ctx, filename, source, prior.Report.Activity)
	if err != nil {
		return Realization{}, err
	}
	expected := newRecordAssemblyReceipt(source, plan)
	plans, origin, err := recordHistoryPlans(ctx, filename, source, plan, r.ControlHistory)
	if err != nil {
		return Realization{}, err
	}
	if r.RankingSourceSHA256 != origin {
		return Realization{}, fmt.Errorf("record ranking source differs")
	}
	expected.RankingSourceSHA256 = origin
	if err = verifyRecordAssemblyRanking(r, expected, plans[0]); err != nil {
		return Realization{}, err
	}
	expected.Ranking = append([]uint16(nil), r.Ranking...)
	if err = replayRecordSearchStages(ctx, plans, expected, r.ControlHistory, r.Control); err != nil {
		return Realization{}, err
	}
	if err = verifyRecordAssemblyObservations(expected, r); err != nil {
		return Realization{}, err
	}
	return realizeRecordAssemblyProjection(ctx, filename, source, prior, plan, expected)
}

func verifyRecordAssemblyObservations(expected, r *RecordAssemblyReceipt) error {
	observedCases, err := canonicalRecordCaseValues(r.Cases)
	if err != nil {
		return err
	}
	checks := []struct {
		name    string
		matches bool
	}{
		{"attempted combinations", reflect.DeepEqual(expected.Attempts, r.Attempts)},
		{"case values", reflect.DeepEqual(expected.Cases, observedCases)},
		{"field choices", reflect.DeepEqual(expected.Choices, r.Choices)},
		{"selected combination", expected.SelectedMask == r.SelectedMask},
		{"case counts", expected.Passed == r.Passed && expected.Total == r.Total},
		{"field counts", expected.FieldsPassed == r.FieldsPassed && expected.FieldsTotal == r.FieldsTotal},
		{"finite status", expected.Status == r.Status},
		{"policy decisions", reflect.DeepEqual(expected.Control, r.Control)},
		{"policy history", reflect.DeepEqual(expected.ControlHistory, r.ControlHistory)},
		{"continuation counts", reflect.DeepEqual(expected.Continuation, r.Continuation)},
	}
	for _, check := range checks {
		if !check.matches {
			return fmt.Errorf("record %s do not replay", check.name)
		}
	}
	return nil
}

func realizeRecordAssemblyProjection(ctx context.Context, filename string, source []byte, prior Result,
	plan recordAssemblyPlan, expected *RecordAssemblyReceipt) (Realization, error) {
	r := prior.Report.RecordAssembly
	generated, err := emitRecordAssembly(ctx, filename, source, plan, expected)
	if err != nil {
		return Realization{}, err
	}
	a, b := prior.Report, generated.Report
	if generated.Source != prior.Source || generated.GoooSource != prior.GoooSource || expected.SelectedSourceSHA256 != r.SelectedSourceSHA256 ||
		a.ActivityID != b.ActivityID || a.SourceDigest != b.SourceDigest || a.ProgramDigest != b.ProgramDigest ||
		a.GeneratedDigest != b.GeneratedDigest || a.GeneratedDigest != a.ReplayDigest || a.InputType != b.InputType || a.OutputType != b.OutputType ||
		!reflect.DeepEqual(a.InputParameters, b.InputParameters) || !reflect.DeepEqual(a.RecordTypes, b.RecordTypes) ||
		!reflect.DeepEqual(a.CallClosure, b.CallClosure) {
		return Realization{}, fmt.Errorf("record selected source, typed signature or projection differs")
	}
	if err = verifyRecordCommonReceipt(prior.Report); err != nil {
		return Realization{}, err
	}
	return Realization{Schema: "gooo/body-realization/v1", Source: generated.GoooSource, OriginalSourceSHA256: digest(source),
		GenerationSourceSHA256: a.SourceDigest, RealizedSourceSHA256: r.SelectedSourceSHA256,
		DocumentSHA256: r.ContractSHA256, TestSuiteSHA256: r.TestSuiteSHA256, ActivityID: a.ActivityID, FinitePassed: r.Passed, FiniteTotal: r.Total}, nil
}

func verifyRecordAssemblyRanking(r, expected *RecordAssemblyReceipt, plan recordAssemblyPlan) error {
	if r.ContractSHA256 != expected.ContractSHA256 || r.TestSuiteSHA256 != expected.TestSuiteSHA256 ||
		len(r.Ranking) != len(expected.Ranking) {
		return fmt.Errorf("record contract or ranking differs")
	}
	var seen [64]bool
	for _, mask := range r.Ranking {
		if int(mask) >= len(r.Ranking) || seen[mask] {
			return fmt.Errorf("record ranking requires each permitted mask once")
		}
		seen[mask] = true
	}
	if !r.ModelRequested {
		if r.Model != nil || r.Context != nil || r.Prediction != nil || r.ModelCalls != 0 || !slices.Equal(r.Ranking, expected.Ranking) {
			return fmt.Errorf("disconnected record selection must preserve deterministic ranking")
		}
		return nil
	}
	if r.Model == nil || !r.Model.Loaded || r.Context == nil || r.ModelCalls < 0 || r.ModelCalls > 1 {
		return fmt.Errorf("record model observation is missing")
	}
	context := recordPlanModelContext(plan, r.Model.FeatureVersion)
	if r.Context.Status == "DECLINED_TO_DETERMINISTIC" && r.Context.Reason == "THREE_CHOICE_MODEL_REQUIRED" {
		context.Status, context.Reason = r.Context.Status, r.Context.Reason
	}
	if !reflect.DeepEqual(context, r.Context) {
		return fmt.Errorf("record source-bound model context differs")
	}
	if context.Status == "DECLINED_TO_DETERMINISTIC" {
		if r.ModelCalls != 0 || r.Prediction != nil || !slices.Equal(r.Ranking, expected.Ranking) {
			return fmt.Errorf("declined context must use deterministic selection")
		}
	} else {
		if r.ModelCalls != 1 || r.PredictNS < 1 || r.Prediction == nil {
			return fmt.Errorf("record model call observation differs")
		}
		if err := verifyRecordPredictionRanking(r); err != nil {
			return err
		}
	}
	return nil
}

func verifyRecordPredictionRanking(r *RecordAssemblyReceipt) error {
	if len(r.Ranking) != 8 {
		return fmt.Errorf("three-choice prediction requires eight field masks")
	}
	var total float64
	for i, p := range r.Prediction.Probabilities {
		if p < 0 || p > 1 || math.IsNaN(float64(p)) || math.IsInf(float64(p), 0) ||
			math.IsNaN(float64(r.Prediction.Logits[i])) || math.IsInf(float64(r.Prediction.Logits[i]), 0) {
			return fmt.Errorf("invalid record prediction values")
		}
		total += float64(p)
	}
	if math.Abs(total-1) > 1e-5 || r.Prediction.Mask != r.Ranking[0] {
		return fmt.Errorf("record prediction proposal or normalization differs")
	}
	for i, mask := range r.Ranking {
		if i == 0 {
			continue
		}
		previous := r.Ranking[i-1]
		a, b := r.Prediction.Probabilities[previous], r.Prediction.Probabilities[mask]
		if a < b || (a == b && previous > mask) {
			return fmt.Errorf("record ranking differs from captured probabilities")
		}
	}
	return nil
}

func verifyRecordCommonReceipt(report Report) error {
	if report.CompletenessReceipt == nil || report.PlanSHA256 != completenessPlanSHA(report) {
		return fmt.Errorf("record plan identity differs")
	}
	if err := completeness.Validate(report.CompletenessReceipt); err != nil {
		return err
	}
	for _, expected := range recordAssemblyDimensions(report.RecordAssembly) {
		index := slices.IndexFunc(report.CompletenessReceipt.Dimensions, func(d CompletenessDimension) bool { return d.ID == expected.ID })
		if index < 0 || !reflect.DeepEqual(report.CompletenessReceipt.Dimensions[index], expected) ||
			!slices.Contains(report.CompletenessReceipt.CoreDimensions, expected.ID) {
			return fmt.Errorf("record completeness field or case counts differ")
		}
	}
	observation := *report.RecordAssembly
	var err error
	observation.Cases, err = canonicalRecordCaseValues(observation.Cases)
	if err != nil {
		return err
	}
	raw, _ := json.Marshal(&observation)
	// JSON decoding stores the nested scope as map[string]any on both sides.
	rawScope, _ := json.Marshal(report.CompletenessReceipt.Scope["record_assembly"])
	var scope map[string]any
	if err := json.Unmarshal(rawScope, &scope); err != nil || scope["observation_sha256"] != digest(raw) ||
		report.CompletenessReceipt.Scope["plan_sha256"] != report.PlanSHA256 ||
		report.CompletenessReceipt.Scope["compiler_source_sha"] != report.CompilerSourceSHA {
		return fmt.Errorf("record completeness observation is unbound")
	}
	return nil
}
