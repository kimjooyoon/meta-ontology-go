package nonmonotonicrefutationoracle

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/nonmonotonicrefutation"
)

func sourceFixture(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../../examples/nonmonotonic-refutation/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestIndependentOracleReplaysNonMonotonicHistory(t *testing.T) {
	source := sourceFixture(t)
	value, err := producer.Produce("examples/nonmonotonic-refutation/main.gooo", source, 0)
	if err != nil {
		t.Fatal(err)
	}
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
	if !strings.Contains(report.Transitions[2].EvidenceBasis, "observed=0") || report.Transitions[2].EvidenceDigest == "" {
		t.Fatalf("refutation basis = %#v", report.Transitions[2])
	}
	if report.Transitions[1].PreviousDigest != report.Transitions[0].TransitionDigest ||
		report.Transitions[2].PreviousDigest != report.Transitions[1].TransitionDigest {
		t.Fatalf("transition chain = %#v", report.Transitions[:3])
	}
	if report.Cases[2].StatusHistory[2] != "REFUTED" || report.Cases[2].CurrentStatus != "DISCHARGED" {
		t.Fatalf("re-evaluation history = %#v", report.Cases[2])
	}
	if report.Conformance.Decision != "PASS" || report.SubjectResolution.Resolution != "PARTIAL" {
		t.Fatalf("resolution split = %#v / %#v", report.Conformance, report.SubjectResolution)
	}
}

func TestIndependentOracleRejectsSourceDigestMismatch(t *testing.T) {
	source := sourceFixture(t)
	value, err := producer.Produce("examples/nonmonotonic-refutation/main.gooo", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	changed := append([]byte(nil), source...)
	changed = append(changed, []byte("\n# changed source\n")...)
	report, err := Judge(data, changed)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" || report.Reason != "SOURCE_BINDING_MISMATCH" {
		t.Fatalf("report = %#v", report)
	}
}
