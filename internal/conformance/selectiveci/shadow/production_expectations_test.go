package shadow

type productionExpectation struct {
	status    string
	stage     string
	component string
	reason    string
	vector    *productionOutput
}

func expectedProduction(name string) productionExpectation {
	if name == "positive-selective" || name == "permutation-stable" {
		vector := baselineProductionVector()
		return productionExpectation{status: "SHADOW_SELECTIVE", stage: "SELECTIVE", component: "all", reason: "VERIFIED", vector: &vector}
	}
	if name == "argv-injection-looking-data" || name == "plan-command-injection-is-data" {
		vector := injectionProductionVector()
		return productionExpectation{status: "SHADOW_SELECTIVE", stage: "SELECTIVE", component: "all", reason: "VERIFIED", vector: &vector}
	}
	values := map[string]productionExpectation{
		"snapshot-binding-manifest-mismatch":            {"FULL_SUITE_FALLBACK", "SNAPSHOT_BINDING", "base_manifest", "MANIFEST_MISMATCH", nil},
		"snapshot-binding-stale-analyzer-digest":        {"FULL_SUITE_FALLBACK", "INPUT", "base_snapshot", "selectiveci.malformed-digest", nil},
		"registry-binding-mismatch":                     {"FULL_SUITE_FALLBACK", "REGISTRY_BINDING", "base_snapshot", "REGISTRY_DIGEST_MISMATCH", nil},
		"registry-binding-lane-mismatch":                {"FULL_SUITE_FALLBACK", "REGISTRY_BINDING", "lane", "REGISTRY_DIGEST_MISMATCH", nil},
		"plan-digest-tamper":                            {"FULL_SUITE_FALLBACK", "PLAN", "planner", "FRONTIER_BLOCKED", nil},
		"plan-unknown":                                  {"FULL_SUITE_FALLBACK", "PLAN", "planner", "FRONTIER_BLOCKED", nil},
		"plan-changed-roots-mismatch":                   {"FULL_SUITE_FALLBACK", "PLAN", "planner", "FRONTIER_BLOCKED", nil},
		"plan-selection-union-invalid":                  {"FULL_SUITE_FALLBACK", "PLAN", "planner", "FRONTIER_BLOCKED", nil},
		"plan-proof-source-binding-mismatch":            {"FULL_SUITE_FALLBACK", "PLAN_PROOF_BINDING", "snapshots", "PROOF_SNAPSHOT_BINDING_MISMATCH", nil},
		"plan-proof-semantic-binding-mismatch":          {"FULL_SUITE_FALLBACK", "PLAN_PROOF_BINDING", "snapshots", "PROOF_SNAPSHOT_BINDING_MISMATCH", nil},
		"plan-proof-plan-digest-mismatch":               {"FULL_SUITE_FALLBACK", "PLAN_PROOF_BINDING", "plan_digest", "PROOF_PLAN_DIGEST_MISMATCH", nil},
		"plan-proof-changed-roots-mismatch":             {"FULL_SUITE_FALLBACK", "PLAN_PROOF_BINDING", "changed_root_ids", "PROOF_CHANGED_ROOT_IDS_MISMATCH", nil},
		"plan-proof-selected-union-mismatch":            {"FULL_SUITE_FALLBACK", "PLAN_PROOF_BINDING", "selected_command_ids", "PROOF_SELECTED_COMMAND_IDS_MISMATCH", nil},
		"proof-fail-closed":                             {"FULL_SUITE_FALLBACK", "PROOF_FAIL_CLOSED", "proof", "SELECTIVE_CI_V1_RECEIPT_MISMATCH", nil},
		"proof-unknown":                                 {"FULL_SUITE_FALLBACK", "PROOF_UNKNOWN", "proof", "SELECTIVE_CI_V1_MISSING_INPUT", nil},
		"lane-unknown":                                  {"FULL_SUITE_FALLBACK", "LANE_UNKNOWN", "lane", "MISSING_INPUT", nil},
		"lane-ineligible":                               {"FULL_SUITE_FALLBACK", "LANE_INELIGIBLE", "lane", "ACTIVE_LEASE", nil},
		"lane-digest-tamper":                            {"FULL_SUITE_FALLBACK", "LANE_UNKNOWN", "lane", "INVALID_COUNT", nil},
		"malformed-unknown-field":                       {"FULL_SUITE_FALLBACK", "INPUT", "evidence_input", "UNKNOWN_FIELD", nil},
		"malformed-duplicate-key":                       {"FULL_SUITE_FALLBACK", "INPUT", "plan_input", "DUPLICATE_FIELD", nil},
		"malformed-trailing-json":                       {"FULL_SUITE_FALLBACK", "INPUT", "lane_input", "TRAILING_DATA", nil},
		"incomplete-required-field":                     {"FULL_SUITE_FALLBACK", "INPUT", "base_snapshot", "selectiveci.invalid-status", nil},
		"precedence-input-over-snapshot":                {"FULL_SUITE_FALLBACK", "INPUT", "evidence_input", "MALFORMED", nil},
		"precedence-snapshot-over-registry":             {"FULL_SUITE_FALLBACK", "SNAPSHOT_BINDING", "base_manifest", "MANIFEST_MISMATCH", nil},
		"precedence-registry-over-plan":                 {"FULL_SUITE_FALLBACK", "REGISTRY_BINDING", "base_snapshot", "REGISTRY_DIGEST_MISMATCH", nil},
		"precedence-plan-over-proof-fail":               {"FULL_SUITE_FALLBACK", "PLAN", "planner", "FRONTIER_BLOCKED", nil},
		"precedence-plan-proof-over-proof-fail":         {"FULL_SUITE_FALLBACK", "PLAN_PROOF_BINDING", "snapshots", "PROOF_SNAPSHOT_BINDING_MISMATCH", nil},
		"precedence-proof-unknown-over-lane-ineligible": {"FULL_SUITE_FALLBACK", "PROOF_UNKNOWN", "proof", "SELECTIVE_CI_V1_MISSING_INPUT", nil},
		"precedence-lane-unknown-over-ineligible":       {"FULL_SUITE_FALLBACK", "LANE_UNKNOWN", "lane", "MISSING_INPUT", nil},
	}
	return values[name]
}

func baselineProductionVector() productionOutput {
	return productionOutput{
		SchemaVersion: "gooo/selective-ci-shadow/v1", Command: "selective-ci shadow", Status: "SHADOW_SELECTIVE", Stage: "SELECTIVE", Component: "all", Reason: "VERIFIED",
		ExecutionAuthorized: false, ShadowOnly: true,
		BaseSourceDigest:   "sha256:032c1647b5477d4c9d3dc0c3dc8a9e47244fb67c4014d57323d99912d00b1971",
		HeadSourceDigest:   "sha256:397911681c8e05e8c65f4220baf97395efe7b2e4d42ff04543110a217752dd34",
		BaseSemanticDigest: "4e441c035f24ba9ee8b96c1267e76995d647e08f1ed2f672253ee51de95e1aaa",
		HeadSemanticDigest: "e6728e17a22d5a6cc9616580c1398cabf5e5bc48efa8effc80f7098b50434273",
		RegistryDigest:     "28209a518aeebf80b595f63b674f5a3e31ab703b948c1877aaeb041936e085e9",
		PlanDigest:         "2c9db682d9e0c7adb80281ec110ee773184926811a9defc6ef65f58ef13d259e",
		ProofStatus:        "VERIFIED", ProofCode: "SELECTIVE_CI_V1_VERIFIED",
		ChangedSemanticIDs: []string{"urn:gooo:shadow/entity/order"},
		SelectedCommands:   []productionCommand{{ID: "urn:gooo:shadow/command/test", Argv: []string{"gooo-shadow-sentinel", "never-run"}}},
		SelectedGuards:     []productionCommand{}, SelectedWorkIDs: []string{"4b2e873bb8692ecedf29f9415745b7bd1a06fee1ec17b71e003e74de1e01a81c"},
		ResourceReceipts: []productionResourceReceipt{{CommandID: "urn:gooo:shadow/command/test", CPUWorkUnits: 100, MemoryBytes: 1000}},
		Lane:             productionLane{Decision: "ELIGIBLE", Reason: "ELIGIBLE", RegistryDigest: "28209a518aeebf80b595f63b674f5a3e31ab703b948c1877aaeb041936e085e9", BaseSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", LaneHeadSHA: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", LaneID: "urn:gooo:shadow/lane/main"},
		CanonicalDigest:  "da84bd348cc7cd506377c2823babfc5dba025226c86cee7b07aa37a6220e5d51",
	}
}

func injectionProductionVector() productionOutput {
	value := baselineProductionVector()
	value.BaseSourceDigest = "sha256:a7833e9907231e0cac60514be4fe280da6965857e7bcda68293bd379e14b8d70"
	value.HeadSourceDigest = "sha256:30f950d9a5f8eaabf50ac914469687b4ce20d78622d224d5f9d4b93da7ddfb39"
	value.RegistryDigest = "989f8295246f04651fa738e763e372c8f21e56dc0672cb902e71961d180a3562"
	value.SelectedCommands = []productionCommand{{ID: "urn:gooo:shadow/command/test", Argv: []string{"sh", "-c", "echo SAFE; touch /tmp/gooo-shadow-must-not-run"}}}
	value.Lane.RegistryDigest = value.RegistryDigest
	value.CanonicalDigest = "b5703c6d2c120cbc9750b1e37cd6f5c0b5732a683780e27514429f08c88012a5"
	return value
}
