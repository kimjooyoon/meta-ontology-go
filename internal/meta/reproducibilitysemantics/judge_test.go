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
activity CaseBothClaimsDischarged(ByteArtifact, MeaningClaim) -> WitnessCase computes "case=both-discharged;byte.reference=artifact/canonical/approved/v1;byte.candidate=artifact/canonical/approved/v1;meaning.expected=meaning/charge-and-ledger/v1;meaning.observed=meaning/charge-and-ledger/v1"
activity CaseReproducibleButWrong(ByteArtifact, MeaningClaim) -> WitnessCase computes "case=reproducible-but-wrong;byte.reference=artifact/canonical/approved/v1;byte.candidate=artifact/canonical/approved/v1;meaning.expected=meaning/charge-and-ledger/v1;meaning.observed=meaning/render-approved/v1"
activity CaseMeaningfulButUnreproduced(ByteArtifact, MeaningClaim) -> WitnessCase computes "case=meaningful-but-unreproduced;byte.reference=artifact/canonical/approved/v1;byte.candidate=artifact/canonical/approved/v2;meaning.expected=meaning/charge-and-ledger/v1;meaning.observed=meaning/charge-and-ledger/v1"
activity CaseClaimsOpen(ByteArtifact, MeaningClaim) -> WitnessCase computes "case=claims-open;byte.reference=;byte.candidate=;meaning.expected=;meaning.observed="
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
		judgment.Summary.OpenCases.Numerator != 1 || judgment.Summary.SourceDigestBinding.Numerator != 4 ||
		judgment.Summary.SemanticCausality.Numerator != 4 {
		t.Fatalf("summary = %#v", judgment.Summary)
	}
}

func TestMeaningDriftIsRefutedDespiteByteEquality(t *testing.T) {
	head := strings.Repeat("b", 40)
	source := []byte(fixtureSource)
	receipt := Produce("fixture.gooo", head, source)
	receipt.Cases[1].Meaning.Observed = receipt.Cases[1].Meaning.Expected
	judgment := Judge("fixture.gooo", head, source, receipt)
	if judgment.Decision != StatusRefuted || judgment.Reason != "SEMANTIC_CAUSALITY_INVALID" {
		t.Fatalf("judgment = %#v", judgment)
	}
}

func TestMissingSourceContractRefutesReceipt(t *testing.T) {
	head := strings.Repeat("c", 40)
	source := []byte("package unrelated\n")
	receipt := Produce("fixture.gooo", head, source)
	judgment := Judge("fixture.gooo", head, source, receipt)
	if judgment.Decision != StatusRefuted || judgment.Reason != "GOOO_SOURCE_SEMANTICS_INVALID" {
		t.Fatalf("judgment = %#v", judgment)
	}
}
