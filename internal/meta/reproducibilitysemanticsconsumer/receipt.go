package reproducibilitysemanticsconsumer

import (
	"regexp"
	"strings"
)

var shaPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func receiptShape(sourcePath, headSHA string, source []byte, semanticDigest string, receipt Receipt) string {
	if receipt.Schema != ReceiptSchema || receipt.Version != 1 || receipt.ContractID != ContractID {
		return "RECEIPT_SCHEMA_INVALID"
	}
	if receipt.SourcePath != sourcePath || receipt.SourceDigest != digestBytes(source) || receipt.HeadSHA != headSHA {
		return "RECEIPT_SOURCE_BINDING_INVALID"
	}
	if !commitPattern.MatchString(headSHA) || !strings.Contains(string(source), "activity SeparateClaims(ByteArtifact, MeaningClaim) -> WitnessCase") {
		return "SOURCE_OR_HEAD_INVALID"
	}
	for _, declaration := range []string{"entity ByteArtifact id", "entity MeaningClaim id", "entity WitnessCase id", "entity BothClaimsDischarged id", "entity ReproducibleButWrong id", "entity MeaningfulButUnreproduced id", "entity ClaimsOpen id"} {
		if !strings.Contains(string(source), declaration) {
			return "GOOO_SOURCE_CONTRACT_INCOMPLETE"
		}
	}
	if receipt.Producer != ProducerID || receipt.Consumer != ConsumerID || receipt.MetaOperation != MetaOperation || receipt.ProofChoice != ProofComposition ||
		receipt.Stage != "receipt" || receipt.Step != "produce" || receipt.Reason != "CLAIM_CHANNELS_SEPARATED" {
		return "RECEIPT_PROVENANCE_INVALID"
	}
	if receipt.SemanticDigest == "" {
		return "DIGEST_ONLY_REFUTED"
	}
	if receipt.SemanticDigest != semanticDigest {
		return "SEMANTIC_DIGEST_BINDING_INVALID"
	}
	if len(receipt.Cases) != CaseCount || len(receipt.Proofs) != 4 || receipt.Authority != (Authority{}) {
		return "RECEIPT_DENOMINATOR_OR_AUTHORITY_INVALID"
	}
	want := receipt.ReceiptDigest
	if want == "" || sealReceipt(receipt).ReceiptDigest != want || !shaPattern.MatchString(receipt.SourceDigest) {
		return "RECEIPT_DIGEST_INVALID"
	}
	return ""
}

func validateReceiptProofs(receipt Receipt) string {
	if receipt.Proofs[0].Choice != ProofByte || receipt.Proofs[0].Claim != "byte equality is evidence only for byte reproducibility" || receipt.Proofs[0].MetaOperation != "compare-byte-digests" || receipt.Proofs[0].Stage != "proof" || receipt.Proofs[0].Step != "byte" || receipt.Proofs[0].Reason != "BYTE_CHANNEL_ONLY" || receipt.Proofs[0].Status != StatusDischarged || receipt.Proofs[0].EvidenceDigest != digestValue(byteEvidence(receipt.Cases)) {
		return "BYTE_PROOF_INVALID"
	}
	if receipt.Proofs[1].Choice != ProofMeaning || receipt.Proofs[1].Claim != "meaning equality requires an independent meaning oracle" || receipt.Proofs[1].MetaOperation != "compare-meaning-oracle-digests" || receipt.Proofs[1].Stage != "proof" || receipt.Proofs[1].Step != "meaning" || receipt.Proofs[1].Reason != "MEANING_CHANNEL_ONLY" || receipt.Proofs[1].Status != StatusDischarged || receipt.Proofs[1].EvidenceDigest != digestValue(meaningEvidence(receipt.Cases)) {
		return "MEANING_PROOF_INVALID"
	}
	if receipt.Proofs[2].Choice != ProofComposition || receipt.Proofs[2].Claim != "the two claims have distinct evidence and failure paths" || receipt.Proofs[2].MetaOperation != MetaOperation || receipt.Proofs[2].Stage != "proof" || receipt.Proofs[2].Step != "compose" || receipt.Proofs[2].Reason != "NON_IDENTITY_EXHIBITED" || receipt.Proofs[2].Status != StatusDischarged || receipt.Proofs[2].EvidenceDigest != digestValue(receipt.Cases) {
		return "COMPOSITION_PROOF_INVALID"
	}
	if receipt.Proofs[3].Choice != ProofSemantic || receipt.Proofs[3].Claim != "case values are derived from parsed and lowered Gooo source" || receipt.Proofs[3].MetaOperation != "parse-and-lower-gooo-source" || receipt.Proofs[3].Stage != "proof" || receipt.Proofs[3].Step != "source" || receipt.Proofs[3].Reason != "SOURCE_SEMANTIC_CAUSALITY" || receipt.Proofs[3].Status != StatusDischarged || receipt.Proofs[3].EvidenceDigest != receipt.SemanticDigest {
		return "SEMANTIC_PROOF_INVALID"
	}
	return ""
}

func byteEvidence(cases []Case) []Evidence {
	result := make([]Evidence, len(cases))
	for index, item := range cases {
		result[index] = item.Byte
	}
	return result
}

func meaningEvidence(cases []Case) []MeaningEvidence {
	result := make([]MeaningEvidence, len(cases))
	for index, item := range cases {
		result[index] = item.Meaning
	}
	return result
}

func judgeProofs(receipt Receipt, judgment Judgment) []Proof {
	return []Proof{
		{Choice: ProofByte, Claim: "consumer recomputed byte equality independently", MetaOperation: "compare-byte-digests", Stage: "judge", Step: "byte", Reason: "BYTE_REPLAY_INDEPENDENT", EvidenceDigest: digestValue(judgment.Cases), Status: StatusDischarged},
		{Choice: ProofMeaning, Claim: "consumer recomputed meaning equality independently", MetaOperation: "compare-meaning-oracle-digests", Stage: "judge", Step: "meaning", Reason: "MEANING_REPLAY_INDEPENDENT", EvidenceDigest: digestValue(judgment.Cases), Status: StatusDischarged},
		{Choice: ProofComposition, Claim: "consumer preserved the two failure paths and four-case matrix", MetaOperation: MetaOperation, Stage: "judge", Step: "compose", Reason: "MATRIX_REPLAY_INDEPENDENT", EvidenceDigest: receipt.ReceiptDigest, Status: StatusDischarged},
		{Choice: ProofSemantic, Claim: "consumer replayed parsed and lowered Gooo source", MetaOperation: "parse-and-lower-gooo-source", Stage: "judge", Step: "source", Reason: "SOURCE_REPLAY_INDEPENDENT", EvidenceDigest: judgment.SemanticDigest, Status: StatusDischarged},
	}
}
