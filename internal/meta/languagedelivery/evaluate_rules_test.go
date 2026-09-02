package languagedelivery

import "testing"

func TestObserveRuleDispatchesExecutionCounters(t *testing.T) {
	decoded := decodedEvidence{Execution: ExecutionReceipt{Summary: ExecutionSummary{
		SourceExecutions: 1, DeterministicReplays: 1, DiagnosticRejections: 2,
	}}}
	tests := []struct {
		counter string
		want    int
		reason  string
	}{
		{"source_executions", 1, "SOURCE_EXECUTIONS_OBSERVED"},
		{"deterministic_replays", 1, "SOURCE_EXECUTION_REPLAYS_OBSERVED"},
		{"diagnostic_rejections", 2, "SOURCE_EXECUTION_DIAGNOSTICS_OBSERVED"},
	}
	for _, test := range tests {
		t.Run(test.counter, func(t *testing.T) {
			got, reason := observeRule(EvidenceRule{Kind: EvidenceExecution, Counter: test.counter}, decoded)
			if got != test.want || reason != test.reason {
				t.Fatalf("observeRule() = (%d, %q), want (%d, %q)", got, reason, test.want, test.reason)
			}
		})
	}
}

func TestObserveRuleDispatchesProfileCounter(t *testing.T) {
	decoded := decodedEvidence{Profile: ProfileReceipt{Summary: ProfileSummary{Profiles: 2}}}
	got, reason := observeRule(EvidenceRule{Kind: EvidenceProfile, Counter: "profiles"}, decoded)
	if got != 2 || reason != "PROFILE_RECEIPTS_OBSERVED" {
		t.Fatalf("observeRule() = (%d, %q)", got, reason)
	}
}
