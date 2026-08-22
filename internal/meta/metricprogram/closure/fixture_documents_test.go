package closure_test

func fixtureProgram(source []byte) map[string]any {
	return map[string]any{
		"schema": "gooo/metric-meta-program/v1", "repository": fixtureRepository,
		"subject_sha": fixtureSHA, "strategy_digest": fixtureDigest("b"),
		"strategy_verification_digest": fixtureDigest("f"),
		"execution_policy":             "READ_ONLY_META_PROGRAM",
		"root_policy": map[string]any{
			"counts_applicability":   "OBSERVED",
			"topology_applicability": "NOT_APPLICABLE",
			"topology_reason":        "ROOT_TOPOLOGY_EXEMPT",
			"readme_requirement":     "NOT_APPLICABLE",
		},
		"registry_digest": fixtureDigest("c"), "source_digest": bytesDigest(source),
		"semantic_digest": fixtureDigest("d"), "operations": fixtureOperations(),
		"bindings": fixtureBindings(), "steps": fixtureSteps(),
		"selection": map[string]any{
			"proof_choice": "REGRESSION", "decision": "HOLD_FIXED_POINT",
			"meta_operation": "terminate-at-fixed-point",
		},
		"coverage": map[string]any{
			"binding_count": fixtureBindingCount, "resolved_binding_count": fixtureBindingCount,
			"registry_operation_count": 8, "referenced_operation_count": 8,
			"selection_operation_resolved": true, "status": "COMPLETE",
		},
		"repository_workspace_writes": false, "promotion_authorized": false,
		"digest": fixtureDigest("a"),
	}
}

func fixtureVerification(program map[string]any) map[string]any {
	return map[string]any{
		"schema":      "gooo/metric-meta-program-verification/v1",
		"subject_sha": fixtureSHA, "strategy_digest": program["strategy_digest"],
		"program_digest": program["digest"], "registry_digest": program["registry_digest"],
		"source_digest": program["source_digest"], "semantic_digest": program["semantic_digest"],
			"binding_count": fixtureBindingCount, "operation_count": 8, "step_count": 4,
		"status": "VERIFIED", "repository_workspace_writes": false,
		"promotion_authorized": false, "digest": fixtureDigest("9"),
	}
}
