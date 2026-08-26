package feedbackstate

import "testing"

func TestSemanticUseCases(t *testing.T) {
	tests := []struct {
		name, decision, source, from, to, wantDecision, wantReason, wantSnapshot string
		previous, descents                                                       int
	}{
		{"fixed point", decisionFixed, decisionFixed, "exact_operation", "exact_operation", "READY", "PREDECESSOR_SEMANTIC_SNAPSHOT_READY", decisionFixed, 0, 0},
		{"improvement", decisionImprove, decisionImprove, "exact_operation", "exact_operation", "READY", "PREDECESSOR_SEMANTIC_SNAPSHOT_READY", decisionImprove, 0, 0},
		{"lower resolution", decisionClosed, decisionClosed, "exact_operation", "operation_class", "READY", "PREDECESSOR_SEMANTIC_SNAPSHOT_READY", decisionLower, 0, 1},
		{"coarsest fail closed", decisionClosed, decisionClosed, "invariant_only", "invariant_only", "READY", "PREDECESSOR_SEMANTIC_SNAPSHOT_READY", decisionClosed, 2, 2},
		{"unknown decision", "UNKNOWN", "UNKNOWN", "exact_operation", "exact_operation", decisionClosed, "FEEDBACK_SEMANTIC_DECISION_UNKNOWN", "", 0, 0},
		{"false fixed point", decisionFixed, decisionClosed, "exact_operation", "exact_operation", decisionClosed, "FALSE_FIXED_POINT_REJECTED", "", 0, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := fixture(test.decision, test.source, test.from, test.to, test.previous, test.descents)
			report := Evaluate(input)
			if report.Decision != test.wantDecision || report.Reason != test.wantReason {
				t.Fatalf("got %s/%s", report.Decision, report.Reason)
			}
			if test.wantSnapshot != "" && (report.Snapshot == nil || report.Snapshot.Decision != test.wantSnapshot) {
				t.Fatalf("snapshot = %#v", report.Snapshot)
			}
			if len(report.Indicators) != 8 || len(report.Proofs) != 3 || report.ReportDigest == "" {
				t.Fatalf("incomplete meta evidence: %#v", report)
			}
		})
	}
}

func TestBindingAndWriteEffectsFailClosed(t *testing.T) {
	input := fixture(decisionFixed, decisionFixed, "exact_operation", "exact_operation", 0, 0)
	input.PayloadDigest = "sha256:wrong"
	if report := Evaluate(input); report.Reason != "PREDECESSOR_PAYLOAD_DIGEST_MISMATCH" {
		t.Fatal(report.Reason)
	}
	input = fixture(decisionFixed, decisionFixed, "exact_operation", "exact_operation", 0, 0)
	input.RepositoryWrites = 1
	if report := Evaluate(input); report.Reason != "PREDECESSOR_WRITE_EFFECT" {
		t.Fatal(report.Reason)
	}
}
