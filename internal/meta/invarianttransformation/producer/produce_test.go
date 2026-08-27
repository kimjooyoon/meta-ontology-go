package producer

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
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

func TestBuildUsesAllFourInvariantValues(t *testing.T) {
	receipt, err := Build([]byte(testSource), testHead, "preserved-translation")
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != model.DecisionAllowed || receipt.Resolution != model.ResolutionExact || len(receipt.Claims) != 4 || len(receipt.Values) != 4 {
		t.Fatalf("receipt=%+v", receipt)
	}
	if receipt.Evidence.InputValue != 2 || receipt.Evidence.CandidateOperation != "add:1" || receipt.Evidence.CandidateResult != 3 || receipt.Evidence.ExpectedValue != 3 ||
		receipt.Evidence.SemanticBeforeDigest != model.SemanticDigest(2) || receipt.Evidence.SemanticAfterDigest != model.SemanticDigest(3) ||
		receipt.Evidence.BaselineInputValue != 2 || receipt.Evidence.BaselineOperation != "add:1" || receipt.Evidence.BaselineOutput != 3 ||
		receipt.Evidence.ReplayInputValue != 2 || receipt.Evidence.ReplayOperation != "add:1" || receipt.Evidence.ReplayOutput != 3 || receipt.Evidence.ReplayCount != 2 ||
		receipt.Evidence.BaselineDigest != receipt.Evidence.ReplayDigest || !receipt.Evidence.RegressionWitnessPresent {
		t.Fatalf("source fixture was not executed: %+v", receipt.Evidence)
	}
	for index, claim := range receipt.Claims {
		if claim.Status != model.StatusDischarged || len(claim.Transitions) != 1 || claim.Transitions[0].From != model.StatusOpen || claim.Transitions[0].To != model.StatusDischarged {
			t.Fatalf("claim[%d]=%+v", index, claim)
		}
	}
}

func TestBuildRecordsApprovedArtifactWithoutWrite(t *testing.T) {
	receipt, err := Build([]byte(testSource), testHead, "approved-artifact")
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Effects) != 1 || receipt.Effects[0].Kind != model.EffectApproved || receipt.RepositoryWrites != 0 || receipt.MutationAuthority {
		t.Fatalf("approved effect=%+v writes=%d authority=%t", receipt.Effects, receipt.RepositoryWrites, receipt.MutationAuthority)
	}
	effect := receipt.Effects[0]
	data, err := os.ReadFile(effect.ArtifactPath)
	if err != nil || len(data) != effect.ArtifactSize || model.DigestBytes(data) != effect.ArtifactDigest || effect.ArtifactSize == 0 {
		t.Fatalf("approved artifact was not observed: effect=%+v err=%v", effect, err)
	}
}

func TestBuildDerivesMissingReplayFromUnavailableRecipe(t *testing.T) {
	receipt, err := Build([]byte(testSource), testHead, "missing-regression-witness")
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Evidence.ReplayCount != 1 || receipt.Evidence.ReplayFailureStage != "REGRESSION" || receipt.Evidence.ReplayFailureStep != "execute-replay" || receipt.Evidence.ReplayFailureReason != "REGRESSION_REPLAY_RECIPE_UNAVAILABLE" || receipt.Decision != model.DecisionBlocked || receipt.Resolution != model.ResolutionLower {
		t.Fatalf("missing replay evidence=%+v receipt=%+v", receipt.Evidence, receipt)
	}
	if receipt.Claims[3].Status != model.StatusOpen || receipt.Claims[3].Coordinate.Stage != "REGRESSION" || receipt.Claims[3].Coordinate.Step != "execute-replay" {
		t.Fatalf("missing replay claim=%+v", receipt.Claims[3])
	}
}
