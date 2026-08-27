package reproducibilitysemantics

import (
	"regexp"
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
	declared, semanticDigest, err := deriveJudgeCases(sourcePath, source)
	if err != nil {
		judgment.Reason = "GOOO_SOURCE_SEMANTICS_INVALID"
		return sealJudgment(judgment)
	}
	judgment.SemanticDigest = semanticDigest
	if reason := receiptShape(sourcePath, headSHA, source, semanticDigest, receipt); reason != "" {
		judgment.Reason = reason
		return sealJudgment(judgment)
	}
	judgment.Cases = make([]JudgmentCase, len(receipt.Cases))
	for index, item := range receipt.Cases {
		if item.ID != declared[index].ID {
			judgment.Reason = "CASE_MATRIX_ORDER_INVALID"
			return sealJudgment(judgment)
		}
		if !judgeCaseBindsSource(item, declared[index]) {
			judgment.Reason = "SEMANTIC_CAUSALITY_INVALID"
			return sealJudgment(judgment)
		}
		if reason := validateCaseProvenance(item); reason != "" {
			judgment.Reason = reason
			return sealJudgment(judgment)
		}
		byteStatus, byteReason := judgeEvidence(item.Byte.Reference, item.Byte.Candidate)
		meaningStatus, meaningReason := judgeEvidence(item.Meaning.Expected, item.Meaning.Observed)
		status, reason := judgeCompose(byteStatus, meaningStatus)
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
	judgment.Summary = summarizeJudgment(judgment.Cases, true, true)
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

func validEvidenceDigest(value string) bool {
	return value == "" || shaPattern.MatchString(value)
}
