package toolchainrelease

func addComputationFillFixture(f computationFixture) {
	sha := computationDigest([]byte(f.Program.Source))
	generated := computationDigest([]byte("synthetic Go output"))
	f.Receipt["result"].(map[string]any)["body_fills"] = []any{map[string]any{
		"input_source_sha256": sha,
		"activity": map[string]any{"package_path": f.Profile.Package, "activity": f.Profile.Activity,
			"lowered_name": f.Program.Entry.LoweredName},
		"generation": map[string]any{"source": "synthetic Go output", "gooo_source": f.Program.Source,
			"report": map[string]any{"activity": f.Program.Entry.LoweredName, "activity_id": f.Profile.ID,
				"decision": "PASS", "typecheck_passed": true, "deterministic_replay": true,
				"generated_digest": generated, "replay_digest": generated,
				"body_fill": map[string]any{"schema": "gooo/body-codegen-ir-fill-plan/v3-record",
					"decision": map[string]any{"schema": "gooo/typed-decision-receipt/v1", "mode": "deterministic_fallback",
						"selected": "lower", "provider": "deterministic", "fallback_reason": "NOT_CONFIGURED", "request_sha256": sha},
					"external_provider_calls_known": true,
					"original_source_digest":        sha, "ir_plan_sha256": sha, "selected_candidate_id": "equal",
					"hole_fills":        []any{map[string]any{"hole_id": "match", "expression": "int64(input.text[offset]) == input.needle"}},
					"test_cases_passed": 6, "test_cases_total": 6, "holdout_cases_passed": 4, "holdout_cases_total": 4,
					"local_model_predictions": 0, "external_provider_calls": 0}}}}}
}
