package main

import (
	"os"
	"testing"

	conformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/operationconformance"
)

func TestSplitGoBehavioralCorpus(t *testing.T) {
	contract, baseline := conformanceEvidence(t)
	raw, err := os.ReadFile("../../internal/meta/operationconformance/testdata/split-go-behavior-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	corpus, err := conformance.ParseBehavioralCorpus(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range corpus.Cases {
		t.Run(test.ID, func(t *testing.T) {
			evidence := cloneConformanceEvidence(baseline)
			mutateConformanceEvidence(&evidence, test.Mutation)
			report := conformance.Evaluate(contract, evidence)
			if report.Decision != test.ExpectedDecision {
				t.Fatalf("decision=%s want=%s", report.Decision, test.ExpectedDecision)
			}
			if test.ExpectedIndicator != "" &&
				indicatorDecision(report, test.ExpectedIndicator) != test.ExpectedIndicatorDecision {
				t.Fatalf("indicator %s did not resolve %s", test.ExpectedIndicator, test.ExpectedIndicatorDecision)
			}
			if err := conformance.Validate(report, contract); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSplitGoBaselineEmitsSixRuntimeIndicators(t *testing.T) {
	contract, evidence := conformanceEvidence(t)
	report := conformance.Evaluate(contract, evidence)
	if report.Decision != conformance.DecisionPass || report.Summary.PassCount != 6 ||
		report.Summary.RuntimeObservedIndicatorCount != 6 || len(report.Indicators) != 6 {
		t.Fatalf("summary=%+v decision=%s", report.Summary, report.Decision)
	}
}
