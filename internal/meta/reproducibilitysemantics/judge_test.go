package reproducibilitysemantics_test

import (
	"encoding/json"
	"strings"
	"testing"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/reproducibilitysemantics"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/reproducibilitysemanticsconsumer"
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
	receipt := producer.Produce("fixture.gooo", head, source)
	raw := receiptJSON(t, receipt)
	judgment := consumer.Judge("fixture.gooo", head, source, raw)
	if err := consumer.ValidateJudgment("fixture.gooo", head, source, raw, judgment); err != nil {
		t.Fatal(err)
	}
	if judgment.Summary.CaseMatrix.Numerator != 4 || judgment.Summary.CaseMatrix.Denominator != 4 ||
		judgment.Summary.ByteClaim.Numerator != 2 || judgment.Summary.MeaningClaim.Numerator != 2 ||
		judgment.Summary.JointClaim.Numerator != 1 || judgment.Summary.Counterexamples.Numerator != 2 ||
		judgment.Summary.OpenCases.Numerator != 1 || judgment.Summary.SourceDigestBinding.Numerator != 4 ||
		judgment.Summary.SemanticCausality.Numerator != 4 {
		t.Fatalf("summary = %#v", judgment.Summary)
	}
	if judgment.ConformanceDecision != consumer.StatusDischarged || judgment.ConformanceResolution != "EXACT" ||
		judgment.SubjectDecision != consumer.StatusOpen || judgment.SubjectResolution != "LOWER_RESOLUTION" ||
		judgment.SubjectReason != "OPEN_EVIDENCE_REMAINS" {
		t.Fatalf("resolution = %#v", judgment)
	}
	if judgment.Cases[0].ByteTransition.From != consumer.StatusOpen || judgment.Cases[0].ByteTransition.To != consumer.StatusDischarged ||
		judgment.Cases[0].ByteTransition.Coordinate.Numerator != 1 || judgment.Cases[0].ByteTransition.Coordinate.Denominator != 1 ||
		judgment.Cases[1].MeaningTransition.To != consumer.StatusRefuted || judgment.Cases[3].JointTransition.To != consumer.StatusOpen ||
		judgment.Cases[3].JointTransition.EvidenceDigest == "" {
		t.Fatalf("transitions = %#v", judgment.Cases)
	}
}

func TestMeaningDriftIsRefutedDespiteByteEquality(t *testing.T) {
	head := strings.Repeat("b", 40)
	source := []byte(fixtureSource)
	receipt := producer.Produce("fixture.gooo", head, source)
	receipt.Cases[1].Meaning.Observed = receipt.Cases[1].Meaning.Expected
	receipt = producer.SealReceipt(receipt)
	judgment := consumer.Judge("fixture.gooo", head, source, receiptJSON(t, receipt))
	if judgment.Decision != consumer.StatusRefuted || judgment.Reason != "SEMANTIC_CAUSALITY_INVALID" {
		t.Fatalf("judgment = %#v", judgment)
	}
}

func TestMissingSourceContractRefutesReceipt(t *testing.T) {
	head := strings.Repeat("c", 40)
	source := []byte("package unrelated\n")
	receipt := producer.Produce("fixture.gooo", head, source)
	judgment := consumer.Judge("fixture.gooo", head, source, receiptJSON(t, receipt))
	if judgment.Decision != consumer.StatusRefuted || judgment.Reason != "GOOO_SOURCE_SEMANTICS_INVALID" {
		t.Fatalf("judgment = %#v", judgment)
	}
}

func receiptJSON(t *testing.T, receipt producer.Receipt) []byte {
	t.Helper()
	raw, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
