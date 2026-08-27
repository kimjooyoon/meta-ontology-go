package nonmonotonicrefutationoracle

import (
	"encoding/json"
	"testing"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/nonmonotonicrefutation"
)

func TestIndependentOracleReplaysNonMonotonicHistory(t *testing.T) {
	source := []byte("package nonmonotonicrefutation\nactivity Observe\n")
	value := producer.Produce("examples/nonmonotonic-refutation/main.gooo", source)
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	report, err := Judge(data, source)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "PASS" || report.Metrics.TransitionTotal != 6 ||
		report.Metrics.DischargedToRefutedTotal != 2 || report.Metrics.RefutedToDischargedTotal != 1 {
		t.Fatalf("report = %#v", report)
	}
	if report.Cases[1].StatusHistory[1] != "DISCHARGED" || report.Cases[1].CurrentStatus != "REFUTED" {
		t.Fatalf("refutation history = %#v", report.Cases[1])
	}
	if report.Cases[2].StatusHistory[2] != "REFUTED" || report.Cases[2].CurrentStatus != "DISCHARGED" {
		t.Fatalf("re-evaluation history = %#v", report.Cases[2])
	}
}

func TestIndependentOracleRejectsSourceDigestMismatch(t *testing.T) {
	source := []byte("package nonmonotonicrefutation\nactivity Observe\n")
	value := producer.Produce("examples/nonmonotonic-refutation/main.gooo", source)
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	report, err := Judge(data, []byte("package nonmonotonicrefutation\nactivity Changed\n"))
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" || report.Reason != "SOURCE_BINDING_MISMATCH" {
		t.Fatalf("report = %#v", report)
	}
}
