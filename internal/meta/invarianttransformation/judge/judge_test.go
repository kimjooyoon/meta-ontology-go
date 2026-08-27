package judge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/producer"
)

const testHead = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

const testSource = `package invarianttransformation
namespace meta
entity Transformation id "gooo://invariant-transformation/value/transformation"
activity PreservedTranslation() -> Transformation computes "case=preserved-translation;input=2;candidate=add:1;expected=3;invariant=candidate-output-equals-expected;replay=add:1;effect=none"
activity SemanticViolation() -> Transformation computes "case=semantic-violation;input=2;candidate=add:2;expected=3;invariant=candidate-output-equals-expected;replay=add:2;effect=none"
activity MissingRegressionWitness() -> Transformation computes "case=missing-regression-witness;input=2;candidate=add:1;expected=3;invariant=candidate-output-equals-expected;replay=unavailable;effect=none"
activity ApprovedArtifact() -> Transformation computes "case=approved-artifact;input=5;candidate=add:1;expected=6;invariant=candidate-output-equals-expected;replay=add:1;effect=approved-artifact"
`

func TestJudgeSeparatesPreservationViolationAndMissingEvidence(t *testing.T) {
	cases := []struct {
		id, decision, resolution, reason, status string
	}{
		{"preserved-translation", model.DecisionAllowed, model.ResolutionExact, "ALL_INVARIANTS_DISCHARGED", model.StatusDischarged},
		{"semantic-violation", model.DecisionRefuted, model.ResolutionInvariant, "SEMANTIC_POSTCONDITION_REFUTED", model.StatusRefuted},
		{"missing-regression-witness", model.DecisionBlocked, model.ResolutionLower, "REGRESSION_REPLAY_RECIPE_UNAVAILABLE", model.StatusOpen},
	}
	for _, test := range cases {
		t.Run(test.id, func(t *testing.T) {
			receipt, err := producer.Build([]byte(testSource), testHead, test.id)
			if err != nil {
				t.Fatal(err)
			}
			judgment := Judge(receipt, []byte(testSource))
			if !judgment.Independent || judgment.Decision != test.decision || judgment.Resolution != test.resolution || judgment.Reason != test.reason || judgment.Status != test.status {
				t.Fatalf("judgment=%+v", judgment)
			}
		})
	}
}

func TestJudgeDoesNotTrustResealedDecision(t *testing.T) {
	receipt, err := producer.Build([]byte(testSource), testHead, "semantic-violation")
	if err != nil {
		t.Fatal(err)
	}
	receipt.Decision = model.DecisionAllowed
	receipt.Resolution = model.ResolutionExact
	receipt.Reason = "ALL_INVARIANTS_DISCHARGED"
	receipt = model.SealReceipt(receipt)
	judgment := Judge(receipt, []byte(testSource))
	if judgment.Independent || judgment.Reason != "DECLARED_DECISION_MISMATCH" {
		t.Fatalf("resealed forged decision was accepted: %+v", judgment)
	}
}

func TestJudgeRejectsEscalatedWriteEffect(t *testing.T) {
	receipt, err := producer.Build([]byte(testSource), testHead, "approved-artifact")
	if err != nil {
		t.Fatal(err)
	}
	receipt.RepositoryWrites = 1
	receipt = model.SealReceipt(receipt)
	judgment := Judge(receipt, []byte(testSource))
	if judgment.Independent || judgment.Reason != "WRITE_BOUNDARY_ESCALATED" {
		t.Fatalf("write escalation was accepted: %+v", judgment)
	}
}

func TestJudgeReplaysSemanticEvidence(t *testing.T) {
	receipt, err := producer.Build([]byte(testSource), testHead, "preserved-translation")
	if err != nil {
		t.Fatal(err)
	}
	receipt.Evidence.SemanticAfterDigest = model.DigestBytes([]byte("tampered-after"))
	receipt = model.SealReceipt(receipt)
	judgment := Judge(receipt, []byte(testSource))
	if judgment.Independent || judgment.Reason != "TRANSFORMATION_EVIDENCE_INVALID" {
		t.Fatalf("tampered semantic evidence was accepted: %+v", judgment)
	}
}

func TestSourceValueChangeChangesJudgment(t *testing.T) {
	actualSource, err := os.ReadFile(filepath.Join("..", "..", "..", "..", model.SourcePath))
	if err != nil {
		t.Fatal(err)
	}
	mutatedSource := strings.Replace(string(actualSource), "expected=3", "expected=4", 1)
	receipt, err := producer.Build([]byte(mutatedSource), testHead, "preserved-translation")
	if err != nil {
		t.Fatal(err)
	}
	judgment := Judge(receipt, []byte(mutatedSource))
	if !judgment.Independent || judgment.Decision != model.DecisionRefuted || judgment.Reason != "SEMANTIC_POSTCONDITION_REFUTED" {
		t.Fatalf("source meaning change did not refute receipt: %+v", judgment)
	}
}
