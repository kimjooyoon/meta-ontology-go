package opentofuobservation

func FixedCounterexamples() []Counterexample {
	return []Counterexample{
		{ID: "first-run-execute", Expected: DecisionPass, Decision: DecisionPass, Resolution: ResolutionExact, Reason: "NO_PRIOR_RECEIPT"},
		{ID: "exact-prior-reuse", Expected: DecisionPass, Decision: DecisionPass, Resolution: ResolutionExact, Reason: "EXACT_PRIOR_RECEIPT"},
		{ID: "asset-checksum-mismatch", Expected: DecisionRefuted, Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: "ASSET_CHECKSUM_MISMATCH"},
		{ID: "plan-replay-digest-mismatch", Expected: DecisionRefuted, Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: "PLAN_REPLAY_DIGEST_MISMATCH"},
		{ID: "test-event-replay-mismatch", Expected: DecisionRefuted, Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: "TEST_EVENT_REPLAY_DIGEST_MISMATCH"},
		{ID: "cache-marker-only", Expected: DecisionRefuted, Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: "CACHE_MARKER_WITHOUT_EXACT_DIGESTS"},
		{ID: "reuse-digest-mismatch", Expected: DecisionRefuted, Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: "REUSE_ELIGIBILITY_DIGEST_MISMATCH"},
		{ID: "prior-receipt-not-reusable", Expected: DecisionRefuted, Decision: DecisionRefuted, Resolution: ResolutionExact, Reason: "PRIOR_RECEIPT_NOT_REUSABLE"},
		unknownCase("missing-receipt", "DEBUGGING", "READ_RECEIPT", "RECEIPT_MISSING", "DIRECT_MISSING", "REEXECUTE_OPENTOFU_COMMAND", nil),
		unknownCase("missing-input", "FIXTURE", "READ_INPUT", "FIXTURE_INPUT_MISSING", "DIRECT_MISSING", "RESTORE_OPENTOFU_FIXTURE", nil),
		unknownCase("prior-receipt-missing", "REUSE", "READ_PRIOR_RECEIPT", "PRIOR_RECEIPT_MISSING", "DIRECT_MISSING", "RESTORE_PRIOR_RECEIPT", nil),
		{ID: "unknown-top-decision", Expected: DecisionFailClosed, Decision: DecisionFailClosed, Resolution: ResolutionLower, Reason: "UNKNOWN_TOP_DECISION"},
		{ID: "malformed-receipt", Expected: DecisionFailClosed, Decision: DecisionFailClosed, Resolution: ResolutionLower, Reason: "MALFORMED_RECEIPT"},
		{ID: "malformed-prior-receipt", Expected: DecisionFailClosed, Decision: DecisionFailClosed, Resolution: ResolutionLower, Reason: "PRIOR_RECEIPT_TOP_DECISION_INVALID"},
	}
}

func unknownCase(id, stage, step, reason, class, next string, blocked []string) Counterexample {
	if blocked == nil {
		blocked = []string{}
	}
	return Counterexample{ID: id, Expected: DecisionUnknown, Decision: DecisionUnknown, Resolution: ResolutionLower, Reason: reason,
		Unknown: &Unknown{Stage: stage, Step: step, Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: blocked}}
}
