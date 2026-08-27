package reproducibilitysemantics

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var shaPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func Judge(sourcePath, headSHA string, source []byte, receipt Receipt) Judgment {
	judgment := Judgment{
		Schema: JudgmentSchema, Version: 1, ContractID: ContractID, SourcePath: sourcePath,
		SourceDigest: digestBytes(source), HeadSHA: headSHA, ReceiptDigest: receipt.ReceiptDigest,
		Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperation,
		ProofChoice: ProofComposition, Stage: "judge", Step: "independent-replay",
		Reason: "JUDGMENT_NOT_STARTED", Decision: StatusRefuted, Authority: Authority{},
	}
	if reason := receiptShape(sourcePath, headSHA, source, receipt); reason != "" {
		judgment.Reason = reason
		return sealJudgment(judgment)
	}
	judgment.Cases = make([]JudgmentCase, len(receipt.Cases))
	caseIDs := []string{"both-discharged", "reproducible-but-wrong", "meaningful-but-unreproduced", "claims-open"}
	for index, item := range receipt.Cases {
		if item.ID != caseIDs[index] {
			judgment.Reason = "CASE_MATRIX_ORDER_INVALID"
			return sealJudgment(judgment)
		}
		if reason := validateCaseProvenance(item); reason != "" {
			judgment.Reason = reason
			return sealJudgment(judgment)
		}
		byteStatus, byteReason := judgeEvidence(item.Byte.Reference, item.Byte.Candidate)
		meaningStatus, meaningReason := judgeEvidence(item.Meaning.Expected, item.Meaning.Observed)
		status, reason := compose(byteStatus, meaningStatus)
		if byteStatus == StatusOpen {
			byteReason = "BYTE_EVIDENCE_MISSING"
		}
		if meaningStatus == StatusOpen {
			meaningReason = "MEANING_EVIDENCE_MISSING"
		}
		judgment.Cases[index] = JudgmentCase{ID: item.ID, ByteStatus: byteStatus,
			MeaningStatus: meaningStatus, Status: status, Reason: reason}
		if item.Byte.Reason != byteReason || item.Meaning.Reason != meaningReason ||
			item.Status != status || item.Reason != reason {
			judgment.Reason = "RECEIPT_REASON_DRIFT"
			return sealJudgment(judgment)
		}
	}
	judgment.Summary = summarizeJudgment(judgment.Cases)
	if judgment.Summary != receipt.Summary {
		judgment.Reason = "RECEIPT_SUMMARY_DRIFT"
		return sealJudgment(judgment)
	}
	if reason := validateReceiptProofs(receipt); reason != "" {
		judgment.Reason = reason
		return sealJudgment(judgment)
	}
	judgment.Proofs = judgeProofs(receipt, judgment)
	judgment.Decision, judgment.Reason = StatusDischarged, "NON_IDENTITY_EXHIBITED"
	return sealJudgment(judgment)
}

func validateCaseProvenance(item Case) string {
	if item.Stage != "composition" || item.Step != "case" ||
		item.Byte.Producer != ProducerID || item.Byte.Consumer != ConsumerID ||
		item.Byte.MetaOperation != "compare-byte-digests" || item.Byte.ProofChoice != ProofByte ||
		item.Byte.Stage != "evidence" || item.Byte.Step != "byte" ||
		item.Meaning.Producer != ProducerID || item.Meaning.Consumer != ConsumerID ||
		item.Meaning.MetaOperation != "compare-meaning-oracle-digests" || item.Meaning.ProofChoice != ProofMeaning ||
		item.Meaning.Stage != "evidence" || item.Meaning.Step != "meaning" ||
		!validEvidenceDigest(item.Byte.Reference) || !validEvidenceDigest(item.Byte.Candidate) ||
		!validEvidenceDigest(item.Meaning.Expected) || !validEvidenceDigest(item.Meaning.Observed) {
		return "CASE_PROVENANCE_INVALID"
	}
	return ""
}

func validEvidenceDigest(value string) bool {
	return value == "" || shaPattern.MatchString(value)
}

func validateReceiptProofs(receipt Receipt) string {
	if receipt.Proofs[0].Choice != ProofByte || receipt.Proofs[0].MetaOperation != "compare-byte-digests" ||
		receipt.Proofs[0].Stage != "proof" || receipt.Proofs[0].Step != "byte" ||
		receipt.Proofs[0].Reason != "BYTE_CHANNEL_ONLY" || receipt.Proofs[0].Status != StatusDischarged ||
		receipt.Proofs[0].EvidenceDigest != digestValue(byteEvidence(receipt.Cases)) {
		return "BYTE_PROOF_INVALID"
	}
	if receipt.Proofs[1].Choice != ProofMeaning || receipt.Proofs[1].MetaOperation != "compare-meaning-oracle-digests" ||
		receipt.Proofs[1].Stage != "proof" || receipt.Proofs[1].Step != "meaning" ||
		receipt.Proofs[1].Reason != "MEANING_CHANNEL_ONLY" || receipt.Proofs[1].Status != StatusDischarged ||
		receipt.Proofs[1].EvidenceDigest != digestValue(meaningEvidence(receipt.Cases)) {
		return "MEANING_PROOF_INVALID"
	}
	if receipt.Proofs[2].Choice != ProofComposition || receipt.Proofs[2].MetaOperation != MetaOperation ||
		receipt.Proofs[2].Stage != "proof" || receipt.Proofs[2].Step != "compose" ||
		receipt.Proofs[2].Reason != "NON_IDENTITY_EXHIBITED" || receipt.Proofs[2].Status != StatusDischarged ||
		receipt.Proofs[2].EvidenceDigest != digestValue(receipt.Cases) {
		return "COMPOSITION_PROOF_INVALID"
	}
	return ""
}

func receiptShape(sourcePath, headSHA string, source []byte, receipt Receipt) string {
	if receipt.Schema != ReceiptSchema || receipt.Version != 1 || receipt.ContractID != ContractID {
		return "RECEIPT_SCHEMA_INVALID"
	}
	if receipt.SourcePath != sourcePath || receipt.SourceDigest != digestBytes(source) || receipt.HeadSHA != headSHA {
		return "RECEIPT_SOURCE_BINDING_INVALID"
	}
	if !commitPattern.MatchString(headSHA) || !strings.Contains(string(source), "activity SeparateClaims(ByteArtifact, MeaningClaim) -> WitnessCase") {
		return "SOURCE_OR_HEAD_INVALID"
	}
	for _, declaration := range []string{
		"entity ByteArtifact id", "entity MeaningClaim id", "entity WitnessCase id",
		"entity BothClaimsDischarged id", "entity ReproducibleButWrong id",
		"entity MeaningfulButUnreproduced id", "entity ClaimsOpen id",
	} {
		if !strings.Contains(string(source), declaration) {
			return "GOOO_SOURCE_CONTRACT_INCOMPLETE"
		}
	}
	if receipt.Producer != ProducerID || receipt.Consumer != ConsumerID ||
		receipt.MetaOperation != MetaOperation || receipt.ProofChoice != ProofComposition ||
		receipt.Stage != "receipt" || receipt.Step != "produce" || receipt.Reason != "CLAIM_CHANNELS_SEPARATED" {
		return "RECEIPT_PROVENANCE_INVALID"
	}
	if len(receipt.Cases) != CaseCount || len(receipt.Proofs) != 3 || receipt.Authority != (Authority{}) {
		return "RECEIPT_DENOMINATOR_OR_AUTHORITY_INVALID"
	}
	want := receipt.ReceiptDigest
	if want == "" || sealReceipt(receipt).ReceiptDigest != want || !shaPattern.MatchString(receipt.SourceDigest) {
		return "RECEIPT_DIGEST_INVALID"
	}
	return ""
}

func judgeEvidence(reference, candidate string) (string, string) {
	if reference == "" || candidate == "" {
		return StatusOpen, "EVIDENCE_MISSING"
	}
	if reference == candidate {
		return StatusDischarged, "EVIDENCE_MATCH"
	}
	return StatusRefuted, "EVIDENCE_MISMATCH"
}

func summarizeJudgment(cases []JudgmentCase) Summary {
	converted := make([]Case, len(cases))
	for index, item := range cases {
		converted[index] = Case{Byte: Evidence{Status: item.ByteStatus}, Meaning: MeaningEvidence{Status: item.MeaningStatus}, Status: item.Status}
	}
	return summarize(converted)
}

func judgeProofs(receipt Receipt, judgment Judgment) []Proof {
	return []Proof{
		{Choice: ProofByte, Claim: "consumer recomputed byte equality independently", MetaOperation: "compare-byte-digests",
			Stage: "judge", Step: "byte", Reason: "BYTE_REPLAY_INDEPENDENT", EvidenceDigest: digestValue(judgment.Cases), Status: StatusDischarged},
		{Choice: ProofMeaning, Claim: "consumer recomputed meaning equality independently", MetaOperation: "compare-meaning-oracle-digests",
			Stage: "judge", Step: "meaning", Reason: "MEANING_REPLAY_INDEPENDENT", EvidenceDigest: digestValue(judgment.Cases), Status: StatusDischarged},
		{Choice: ProofComposition, Claim: "consumer preserved the two failure paths and four-case matrix", MetaOperation: MetaOperation,
			Stage: "judge", Step: "compose", Reason: "MATRIX_REPLAY_INDEPENDENT", EvidenceDigest: receipt.ReceiptDigest, Status: StatusDischarged},
	}
}

func ValidateJudgment(sourcePath, headSHA string, source []byte, receipt Receipt, judgment Judgment) error {
	want := Judge(sourcePath, headSHA, source, receipt)
	if judgment.Decision != StatusDischarged || judgment.Reason != "NON_IDENTITY_EXHIBITED" {
		return fmt.Errorf("reproducibility semantics judgment failed: %s", judgment.Reason)
	}
	if !reflect.DeepEqual(judgment, want) {
		return fmt.Errorf("reproducibility semantics judgment is not deterministic")
	}
	return nil
}
