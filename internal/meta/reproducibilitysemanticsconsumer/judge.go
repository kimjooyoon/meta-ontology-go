package reproducibilitysemanticsconsumer

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func Judge(sourcePath, headSHA string, source, receiptRaw []byte) Judgment {
	var receipt Receipt
	judgment := Judgment{Schema: JudgmentSchema, Version: 1, ContractID: ContractID, SourcePath: sourcePath,
		SourceDigest: digestBytes(source), HeadSHA: headSHA, Producer: ProducerID, Consumer: ConsumerID,
		MetaOperation: MetaOperation, ProofChoice: ProofComposition, Stage: "judge", Step: "independent-replay",
		Reason: "JUDGMENT_NOT_STARTED", Decision: StatusRefuted, ConformanceDecision: StatusRefuted,
		ConformanceResolution: "LOWER_RESOLUTION", SubjectDecision: StatusRefuted, SubjectResolution: "LOWER_RESOLUTION",
		SubjectReason: "JUDGMENT_NOT_STARTED", Authority: Authority{}}
	if err := json.Unmarshal(receiptRaw, &receipt); err != nil {
		return reject(judgment, "RECEIPT_DECODE_INVALID")
	}
	judgment.ReceiptDigest = receipt.ReceiptDigest
	declared, semanticDigest, err := deriveCases(sourcePath, source)
	if err != nil {
		return reject(judgment, "GOOO_SOURCE_SEMANTICS_INVALID")
	}
	judgment.SemanticDigest = semanticDigest
	if reason := receiptShape(sourcePath, headSHA, source, semanticDigest, receipt); reason != "" {
		return reject(judgment, reason)
	}
	judgment.Cases = make([]JudgmentCase, len(receipt.Cases))
	for index, item := range receipt.Cases {
		if item.ID != declared[index].ID {
			return reject(judgment, "CASE_MATRIX_ORDER_INVALID")
		}
		if !bindsSource(item, declared[index]) {
			return reject(judgment, "SEMANTIC_CAUSALITY_INVALID")
		}
		if reason := validateCaseProvenance(item); reason != "" {
			return reject(judgment, reason)
		}
		byteStatus, byteReason := judgeEvidence(item.Byte.Reference, item.Byte.Candidate)
		meaningStatus, meaningReason := judgeEvidence(item.Meaning.Expected, item.Meaning.Observed)
		if byteStatus == StatusOpen {
			byteReason = "BYTE_EVIDENCE_MISSING"
		}
		if meaningStatus == StatusOpen {
			meaningReason = "MEANING_EVIDENCE_MISSING"
		}
		status, reason := judgeCompose(byteStatus, meaningStatus)
		expected := Case{ID: item.ID, Byte: item.Byte, Meaning: item.Meaning, Status: status, Stage: item.Stage, Step: item.Step, Reason: reason}
		expected.ByteTransition, expected.MeaningTransition, expected.JointTransition = claimTransitions(expected)
		if item.Byte.Status != byteStatus || item.Meaning.Status != meaningStatus || item.Byte.Reason != byteReason || item.Meaning.Reason != meaningReason || item.Status != status || item.Reason != reason {
			return reject(judgment, "RECEIPT_REASON_DRIFT")
		}
		if item.ByteTransition != expected.ByteTransition || item.MeaningTransition != expected.MeaningTransition || item.JointTransition != expected.JointTransition {
			return reject(judgment, "RECEIPT_TRANSITION_DRIFT")
		}
		judgment.Cases[index] = JudgmentCase{ID: item.ID, ByteStatus: byteStatus, MeaningStatus: meaningStatus,
			Status: status, Reason: reason, ByteTransition: expected.ByteTransition,
			MeaningTransition: expected.MeaningTransition, JointTransition: expected.JointTransition}
	}
	judgment.Summary = summarize(judgment.Cases, true, true)
	if judgment.Summary != receipt.Summary {
		return reject(judgment, "RECEIPT_SUMMARY_DRIFT")
	}
	if reason := validateReceiptProofs(receipt); reason != "" {
		return reject(judgment, reason)
	}
	judgment.Proofs = judgeProofs(receipt, judgment)
	judgment.Decision = StatusDischarged
	judgment.ConformanceDecision = StatusDischarged
	judgment.ConformanceResolution = "EXACT"
	judgment.Reason = "NON_IDENTITY_EXHIBITED"
	judgment.SubjectDecision, judgment.SubjectResolution, judgment.SubjectReason = subjectOutcome(judgment.Cases)
	return sealJudgment(judgment)
}

func reject(judgment Judgment, reason string) Judgment {
	judgment.Decision = StatusRefuted
	judgment.ConformanceDecision = StatusRefuted
	judgment.ConformanceResolution = "LOWER_RESOLUTION"
	judgment.SubjectDecision = StatusRefuted
	judgment.SubjectResolution = "LOWER_RESOLUTION"
	judgment.Reason = reason
	judgment.SubjectReason = reason
	return sealJudgment(judgment)
}

func ValidateJudgment(sourcePath, headSHA string, source, receiptRaw []byte, judgment Judgment) error {
	want := Judge(sourcePath, headSHA, source, receiptRaw)
	if judgment.ConformanceDecision != StatusDischarged || judgment.ConformanceResolution != "EXACT" || judgment.Reason != "NON_IDENTITY_EXHIBITED" {
		return fmt.Errorf("reproducibility semantics conformance failed: %s", judgment.Reason)
	}
	if judgment.SubjectDecision == "" || judgment.SubjectResolution == "" || judgment.SubjectReason == "" {
		return fmt.Errorf("reproducibility semantics subject resolution missing")
	}
	if !reflect.DeepEqual(judgment, want) {
		return fmt.Errorf("reproducibility semantics judgment is not deterministic")
	}
	return nil
}
