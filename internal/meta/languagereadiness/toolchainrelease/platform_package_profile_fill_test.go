package toolchainrelease

import "testing"

func TestComputationPackageRejectsUnboundFill(t *testing.T) {
	prefix := "result.body_fills.0."
	mutations := []computationMutation{
		{"result.body_fills", []any{}, false},
		{prefix + "input_source_sha256", "other", false},
		{prefix + "activity.activity", "other", false}, {prefix + "activity.lowered_name", "other", false},
		{prefix + "generation.source", "changed", false}, {prefix + "generation.gooo_source", "changed", false},
		{prefix + "generation.report.activity_id", "other", false},
		{prefix + "generation.report.activity", "other", false},
		{prefix + "generation.report.decision", "OBSERVED", false},
		{prefix + "generation.report.typecheck_passed", false, false},
		{prefix + "generation.report.deterministic_replay", false, false},
		{prefix + "generation.report.replay_digest", "other", false},
		{prefix + "generation.report.body_fill", nil, true},
	}
	for _, m := range []computationMutation{
		{"schema", "other", false}, {"original_source_digest", "other", false}, {"ir_plan_sha256", "other", false},
		{"decision", nil, true}, {"decision.mode", "sampled", false}, {"decision.selected", "other", false},
		{"decision.provider", "other", false}, {"decision.request_sha256", "other", false},
		{"external_provider_calls_known", nil, true},
		{"selected_candidate_id", "lower", false}, {"hole_fills", []any{}, false},
		{"hole_fills.0.hole_id", "other", false}, {"hole_fills.0.expression", "false", false},
		{"test_cases_passed", 5, false}, {"test_cases_total", nil, true},
		{"holdout_cases_passed", 3, false}, {"holdout_cases_total", nil, true},
		{"local_model_predictions", 1, false}, {"local_model_predictions", nil, true},
		{"external_provider_calls", 1, false}, {"external_provider_calls", nil, true},
	} {
		m.path = prefix + "generation.report.body_fill." + m.path
		mutations = append(mutations, m)
	}
	rejectComputationMutations(t, computationProfiles[1], mutations)
	f := newComputationFixture(t, computationProfiles[0])
	addComputationFillFixture(f)
	if _, err := f.validate(false, boundedTextSmokeBinding{}); err == nil {
		t.Fatal("unexpected body fill accepted")
	}
}
