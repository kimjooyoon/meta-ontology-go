package feedbackstate

import (
	"encoding/json"
	"strings"
	"testing"
)

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
			report, err := Evaluate(input)
			if err != nil {
				t.Fatal(err)
			}
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
	if report, err := Evaluate(input); err != nil || report.Reason != "PREDECESSOR_PAYLOAD_DIGEST_MISMATCH" {
		t.Fatal(report.Reason)
	}
	input = fixture(decisionFixed, decisionFixed, "exact_operation", "exact_operation", 0, 0)
	input.RepositoryWrites = 1
	if report, err := Evaluate(input); err != nil || report.Reason != "PREDECESSOR_WRITE_EFFECT" {
		t.Fatal(report.Reason)
	}
}

func TestAggregateRepositoryWritesBoundaries(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	tests := []struct {
		name       string
		counts     []int
		want       int
		wantReason string
	}{
		{name: "zero", counts: []int{0, 0, 0}},
		{name: "positive", counts: []int{1, 0, 0}, want: 1},
		{name: "negative input", counts: []int{-1, 0, 0}, wantReason: "FAIL_CLOSED: NEGATIVE_REPOSITORY_WRITES"},
		{name: "negative outer", counts: []int{0, -1, 0}, wantReason: "FAIL_CLOSED: NEGATIVE_REPOSITORY_WRITES"},
		{name: "negative cancellation", counts: []int{1, -1, 0}, wantReason: "FAIL_CLOSED: NEGATIVE_REPOSITORY_WRITES"},
		{name: "negative nested", counts: []int{0, 0, -1}, wantReason: "FAIL_CLOSED: NEGATIVE_REPOSITORY_WRITES"},
		{name: "overflow", counts: []int{maxInt, maxInt, 2}, wantReason: "FAIL_CLOSED: REPOSITORY_WRITES_OVERFLOW"},
		{name: "negative takes priority", counts: []int{maxInt, maxInt, -1}, wantReason: "FAIL_CLOSED: NEGATIVE_REPOSITORY_WRITES"},
		{name: "maximum safe", counts: []int{maxInt - 1, 0, 1}, want: maxInt},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := AggregateRepositoryWrites(test.counts...)
			if test.wantReason != "" {
				if err == nil || err.Error() != test.wantReason {
					t.Fatalf("got %d/%v", got, err)
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("got %d/%v", got, err)
			}
		})
	}
}

func inputWithRepositoryWrites(t *testing.T, inputWrites, outerWrites, nestedWrites int) Input {
	t.Helper()
	input := fixture(decisionFixed, decisionFixed, "exact_operation", "exact_operation", 0, 0)
	input.RepositoryWrites = inputWrites
	var receipt archivedReceipt
	if err := json.Unmarshal(input.Receipt, &receipt); err != nil {
		t.Fatal(err)
	}
	receipt.RepositoryWrites = outerWrites
	receipt.Report.RepositoryWrites = nestedWrites
	raw, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	input.Receipt, input.PayloadDigest = raw, payloadDigest(raw)
	return input
}

func TestEvaluateRejectsInvalidRepositoryWritesWithoutReport(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	tests := []struct {
		name, wantReason     string
		input, outer, nested int
	}{
		{name: "negative input", wantReason: "NEGATIVE_REPOSITORY_WRITES", input: -1},
		{name: "negative outer", wantReason: "NEGATIVE_REPOSITORY_WRITES", outer: -1},
		{name: "negative cancellation", wantReason: "NEGATIVE_REPOSITORY_WRITES", input: 1, outer: -1},
		{name: "negative nested", wantReason: "NEGATIVE_REPOSITORY_WRITES", nested: -1},
		{name: "overflow", wantReason: "REPOSITORY_WRITES_OVERFLOW", input: maxInt, outer: maxInt, nested: 2},
		{name: "negative takes priority", wantReason: "NEGATIVE_REPOSITORY_WRITES", input: maxInt, outer: maxInt, nested: -1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report, err := Evaluate(inputWithRepositoryWrites(t, test.input, test.outer, test.nested))
			if err == nil || !strings.Contains(err.Error(), test.wantReason) {
				t.Fatalf("got report=%#v err=%v", report, err)
			}
			if report.Decision != "" || report.ReportDigest != "" || len(report.Indicators) != 0 || len(report.Proofs) != 0 {
				t.Fatalf("invalid count published report: %#v", report)
			}
		})
	}
}

func TestEvaluatePreservesWriteEvidenceBoundaries(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	tests := []struct {
		name                       string
		input, outer, nested, want int
		wantDecision, wantReason   string
	}{
		{name: "zero", wantDecision: "READY", wantReason: "PREDECESSOR_SEMANTIC_SNAPSHOT_READY"},
		{name: "positive", input: 1, want: 1, wantDecision: decisionClosed, wantReason: "PREDECESSOR_WRITE_EFFECT"},
		{name: "maximum safe", input: maxInt - 1, nested: 1, want: maxInt, wantDecision: decisionClosed, wantReason: "PREDECESSOR_WRITE_EFFECT"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report, err := Evaluate(inputWithRepositoryWrites(t, test.input, test.outer, test.nested))
			if err != nil || report.Decision != test.wantDecision || report.Reason != test.wantReason || report.Summary.RepositoryWrites != test.want {
				t.Fatalf("got report=%#v err=%v", report, err)
			}
		})
	}
}
