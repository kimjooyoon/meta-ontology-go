package producer

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

const testHead = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

const testSource = `package invarianttransformation
namespace meta
entity Transformation id "gooo://invariant-transformation/value/transformation"
activity PreservedTranslation() -> Transformation computes "case=preserved-translation;kind=PRESERVED;input=2;candidate=add:1;expected=3;invariant=candidate-output-equals-expected;invariant-id=candidate-output-equals-expected-v1;domain=bounded-fixture-input-domain-v1;replay=add:1;effect=none"
activity SemanticViolation() -> Transformation computes "case=semantic-violation;kind=VIOLATION;input=2;candidate=add:2;expected=3;invariant=candidate-output-equals-expected;invariant-id=candidate-output-equals-expected-v1;domain=bounded-fixture-input-domain-v1;replay=add:2;effect=none"
activity MissingRegressionWitness() -> Transformation computes "case=missing-regression-witness;kind=EVIDENCE_MISSING;input=2;candidate=add:1;expected=3;invariant=candidate-output-equals-expected;invariant-id=candidate-output-equals-expected-v1;domain=bounded-fixture-input-domain-v1;replay=unavailable;effect=none"
activity ApprovedArtifact() -> Transformation computes "case=approved-artifact;kind=APPROVED_ARTIFACT;input=5;candidate=add:1;expected=6;invariant=candidate-output-equals-expected;invariant-id=candidate-output-equals-expected-v1;domain=bounded-fixture-input-domain-v1;replay=add:1;effect=approved-artifact"
`

func TestBuildUsesSourceRecipeAndProvisionalEffect(t *testing.T) {
	receipt, err := Build([]byte(testSource), testHead, "preserved-translation")
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != model.DecisionAllowed || receipt.Resolution != model.ResolutionExact || len(receipt.Claims) != 4 || len(receipt.Values) != 4 || receipt.Phase != model.ReceiptProvisional || len(receipt.Effects) != 0 {
		t.Fatalf("receipt=%+v", receipt)
	}
	if receipt.Evidence.SemanticSourceDigest == "" || receipt.Evidence.InputValue != 2 || receipt.Evidence.CandidateOperation != "add:1" || receipt.Evidence.CandidateResult != 3 || receipt.Evidence.ReplayCount != 2 || receipt.Evidence.BaselineDigest != receipt.Evidence.ReplayDigest || !receipt.Evidence.RegressionWitnessPresent {
		t.Fatalf("source fixture was not executed: %+v", receipt.Evidence)
	}
	for index, claim := range receipt.Claims {
		if claim.ID != "preserved-translation::"+model.CanonicalValueSpecs()[index].ID || claim.Status != model.StatusDischarged || len(claim.Transitions) != 1 || claim.Transitions[0].CurrentTransitionDigest == "" {
			t.Fatalf("claim[%d]=%+v", index, claim)
		}
	}
}

func TestBuildDoesNotCreateApprovedArtifact(t *testing.T) {
	receipt, err := Build([]byte(testSource), testHead, "approved-artifact")
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Effects) != 0 || receipt.Phase != model.ReceiptProvisional || receipt.TempArtifactWriteAuthorized {
		t.Fatalf("producer emitted effect: %+v", receipt)
	}
}

func TestBuildDerivesMissingReplayFromUnavailableRecipe(t *testing.T) {
	receipt, err := Build([]byte(testSource), testHead, "missing-regression-witness")
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Evidence.ReplayCount != 1 || receipt.Evidence.ReplayFailureStage != "REGRESSION" || receipt.Evidence.ReplayFailureStep != "execute-replay" || receipt.Evidence.ReplayFailureReason != "REGRESSION_REPLAY_RECIPE_UNAVAILABLE" || receipt.Decision != model.DecisionBlocked || receipt.Resolution != model.ResolutionLower || receipt.Claims[3].Status != model.StatusOpen {
		t.Fatalf("missing replay evidence=%+v receipt=%+v", receipt.Evidence, receipt)
	}
}
