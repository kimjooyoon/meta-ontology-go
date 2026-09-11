package languagereadiness

import "testing"

func TestConceptGovernedRefactoringRequiresProducedOperationReceipt(t *testing.T) {
	obligation := obligations[5]
	without := evaluateObligation(obligation, []conceptEvidence{{
		ID: obligation.ConceptID, Stage: "OPERATING", CodeBindings: []string{"internal/meta/metricstrategy"},
		MetricBindings: []string{"gooo.metric.meta.concept-operation-binding-bps.v1"},
		UseCases: []useCaseEvidence{{ID: "operation-binding", Trigger: "receipt", ExpectedOutcome: "PASS"}},
	}}, evidenceDigests{})
	if without.Status != "NOT_SATISFIED" || without.Reason != "CONCEPT_OPERATION_BINDING_RECEIPT_REQUIRED" {
		t.Fatalf("without receipt = %+v", without)
	}
	with := evaluateObligation(obligation, []conceptEvidence{{
		ID: obligation.ConceptID, Stage: "OPERATING", CodeBindings: []string{"internal/meta/metricstrategy"},
		MetricBindings: []string{"gooo.metric.meta.concept-operation-binding-bps.v1"},
		UseCases: []useCaseEvidence{{ID: "operation-binding", Trigger: "receipt", ExpectedOutcome: "PASS"}},
	}}, evidenceDigests{conceptOperation: "sha256:receipt"})
	if with.Status != "SATISFIED" {
		t.Fatalf("with receipt = %+v", with)
	}
}
