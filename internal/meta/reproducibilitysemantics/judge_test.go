package reproducibilitysemantics

import (
	"strings"
	"testing"
)

const fixtureSource = `package reproducibilitysemantics
namespace reproducibilitysemantics
entity ByteArtifact id "gooo://reproducibility-semantics/entity/byte-artifact"
entity MeaningClaim id "gooo://reproducibility-semantics/entity/meaning-claim"
entity WitnessCase id "gooo://reproducibility-semantics/entity/witness-case"
entity BothClaimsDischarged id "gooo://reproducibility-semantics/entity/both-discharged"
entity ReproducibleButWrong id "gooo://reproducibility-semantics/entity/reproducible-but-wrong"
entity MeaningfulButUnreproduced id "gooo://reproducibility-semantics/entity/meaningful-but-unreproduced"
entity ClaimsOpen id "gooo://reproducibility-semantics/entity/claims-open"
activity SeparateClaims(ByteArtifact, MeaningClaim) -> WitnessCase
`

func TestIndependentJudgeDischargesSeparatedClaims(t *testing.T) {
	head := strings.Repeat("a", 40)
	source := []byte(fixtureSource)
	receipt := Produce("fixture.gooo", head, source)
	judgment := Judge("fixture.gooo", head, source, receipt)
	if err := ValidateJudgment("fixture.gooo", head, source, receipt, judgment); err != nil {
		t.Fatal(err)
	}
	if judgment.Summary.CaseMatrix.Numerator != 4 || judgment.Summary.CaseMatrix.Denominator != 4 ||
		judgment.Summary.ByteClaim.Numerator != 2 || judgment.Summary.MeaningClaim.Numerator != 2 ||
		judgment.Summary.JointClaim.Numerator != 1 || judgment.Summary.Counterexamples.Numerator != 2 ||
		judgment.Summary.OpenCases.Numerator != 1 {
		t.Fatalf("summary = %#v", judgment.Summary)
	}
}

func TestMeaningDriftIsRefutedDespiteByteEquality(t *testing.T) {
	head := strings.Repeat("b", 40)
	source := []byte(fixtureSource)
	receipt := Produce("fixture.gooo", head, source)
	receipt.Cases[1].Meaning.Observed = receipt.Cases[1].Meaning.Expected
	judgment := Judge("fixture.gooo", head, source, receipt)
	if judgment.Decision != StatusRefuted || judgment.Reason != "RECEIPT_REASON_DRIFT" {
		t.Fatalf("judgment = %#v", judgment)
	}
}

func TestMissingSourceContractRefutesReceipt(t *testing.T) {
	head := strings.Repeat("c", 40)
	source := []byte("package unrelated\n")
	receipt := Produce("fixture.gooo", head, source)
	judgment := Judge("fixture.gooo", head, source, receipt)
	if judgment.Decision != StatusRefuted || judgment.Reason != "SOURCE_OR_HEAD_INVALID" {
		t.Fatalf("judgment = %#v", judgment)
	}
}
