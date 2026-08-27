package judge

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/producer"
)

const testHead = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestJudgeSeparatesPreservationViolationAndMissingEvidence(t *testing.T) {
	cases := []struct {
		id, decision, resolution, reason, status string
	}{
		{"preserved-translation", model.DecisionAllowed, model.ResolutionExact, "ALL_INVARIANTS_DISCHARGED", model.StatusDischarged},
		{"semantic-violation", model.DecisionRefuted, model.ResolutionInvariant, "SEMANTIC_POSTCONDITION_REFUTED", model.StatusRefuted},
		{"missing-regression-witness", model.DecisionBlocked, model.ResolutionLower, "REGRESSION_WITNESS_MISSING", model.StatusOpen},
	}
	for _, test := range cases {
		t.Run(test.id, func(t *testing.T) {
			receipt, err := producer.Build([]byte("package fixture\n"), testHead, test.id)
			if err != nil {
				t.Fatal(err)
			}
			judgment := Judge(receipt)
			if !judgment.Independent || judgment.Decision != test.decision || judgment.Resolution != test.resolution || judgment.Reason != test.reason || judgment.Status != test.status {
				t.Fatalf("judgment=%+v", judgment)
			}
		})
	}
}

func TestJudgeDoesNotTrustResealedDecision(t *testing.T) {
	receipt, err := producer.Build([]byte("package fixture\n"), testHead, "semantic-violation")
	if err != nil {
		t.Fatal(err)
	}
	receipt.Decision = model.DecisionAllowed
	receipt.Resolution = model.ResolutionExact
	receipt.Reason = "ALL_INVARIANTS_DISCHARGED"
	receipt = model.SealReceipt(receipt)
	judgment := Judge(receipt)
	if judgment.Independent || judgment.Reason != "DECLARED_DECISION_MISMATCH" {
		t.Fatalf("resealed forged decision was accepted: %+v", judgment)
	}
}

func TestJudgeRejectsEscalatedWriteEffect(t *testing.T) {
	receipt, err := producer.Build([]byte("package fixture\n"), testHead, "approved-artifact")
	if err != nil {
		t.Fatal(err)
	}
	receipt.RepositoryWrites = 1
	receipt = model.SealReceipt(receipt)
	judgment := Judge(receipt)
	if judgment.Independent || judgment.Reason != "WRITE_BOUNDARY_ESCALATED" {
		t.Fatalf("write escalation was accepted: %+v", judgment)
	}
}
