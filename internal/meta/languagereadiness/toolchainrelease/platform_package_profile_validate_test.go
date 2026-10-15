package toolchainrelease

import "testing"

func TestComputationPackageRejectsIncompleteSourceAndResults(t *testing.T) {
	mutations := []computationMutation{
		{"schema", "other", false}, {"decision", "OBSERVED", false}, {"error", "failed", false},
		{"manifest_digest", "other", false}, {"cases_digest", "other", false},
		{"result.program.entry.package_path", "other", false}, {"result.program.entry.activity", "other", false},
		{"result.program.lowered_gooo_source", "other", false},
		{"result.program.workspace_image.packages.0.sources.0.source_digest", "other", false},
		{"result.composition.generated_sha256", "other", false},
		{"result.execution_source_sha256", "other", false},
		{"result.runtime.generated_sha256", "other", false},
		{"result.runtime.stage", "FAILED", false}, {"result.runtime.failure", "failed", false},
		{"result.runtime.model_calls", 1, false}, {"result.runtime.model_calls", nil, true},
		{"result.runtime.finite_passed", 1, false}, {"result.runtime.finite_total", nil, true},
		{"result.runtime.projection_replayed", false, false}, {"result.runtime.runtime_replayed", nil, true},
		{"result.runtime.traces", []any{}, false},
		{"result.runtime.traces.0.case_index", 1, false}, {"result.runtime.traces.0.case_index", nil, true},
		{"result.runtime.traces.0.deliveries", []any{}, false},
		{"result.runtime.traces.0.deliveries.0.activity_id", "other", false},
		{"result.runtime.traces.0.deliveries.0.input", nil, true},
		{"result.runtime.traces.0.deliveries.0.actual", nil, true},
		{"result.runtime.traces.0.deliveries.0.expected", nil, true},
		{"result.runtime.traces.0.deliveries.0.passed", false, false},
	}
	for _, p := range computationProfiles {
		t.Run(p.Name, func(t *testing.T) { rejectComputationMutations(t, p, mutations) })
	}
}
