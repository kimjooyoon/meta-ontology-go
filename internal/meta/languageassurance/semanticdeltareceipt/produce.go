package semanticdeltareceipt

import (
	"fmt"
	"os"
)

// ProduceFiles is the only producer entry point. It reads the named files and
// passes their bytes to the canonical syntax/lowering projection.
func ProduceFiles(input Input) (Receipt, error) {
	before, err := os.ReadFile(input.BeforePath)
	if err != nil {
		return Receipt{}, fmt.Errorf("read before source %q: %w", input.BeforePath, err)
	}
	after, err := os.ReadFile(input.AfterPath)
	if err != nil {
		return Receipt{}, fmt.Errorf("read after source %q: %w", input.AfterPath, err)
	}
	return produceBytes(input, before, after), nil
}

func produceBytes(input Input, beforeRaw, afterRaw []byte) Receipt {
	beforeSource, beforeErr := projectSource(input.BeforePath, beforeRaw)
	afterSource, afterErr := projectSource(input.AfterPath, afterRaw)
	receipt := Receipt{Schema: ReceiptSchema, CaseID: input.CaseID, SubjectSHA: input.SubjectSHA, Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, ProofChoice: "FOUNDATION", Stage: "produce", Step: "separate-delta-layers", Before: snapshot(beforeRaw, beforeSource, beforeErr), After: snapshot(afterRaw, afterSource, afterErr), TextualDelta: textualDelta(beforeRaw, afterRaw), RepositoryWrites: 0}
	if beforeErr != nil || afterErr != nil {
		return unknownReceipt(receipt, beforeSource, afterSource, beforeErr, afterErr)
	}
	if !validSubject(input.SubjectSHA) {
		return subjectUnknown(receipt)
	}
	receipt.StructuralDelta = structuralDelta(beforeSource, afterSource)
	receipt.SemanticClaimDelta = claimDelta(beforeSource, afterSource)
	receipt.RawDecision = receipt.TextualDelta.Decision
	receipt.SemanticDecision = SemanticPreserved
	receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = DecisionFixedPoint, ResolutionExact, ClassPreserved, ReasonTextualOnly
	if hasSemanticDelta(receipt.StructuralDelta, receipt.SemanticClaimDelta) {
		receipt.SemanticDecision = SemanticChanged
		receipt.Decision, receipt.Classification, receipt.Reason = DecisionDelta, ClassChanged, ReasonMeaning
	}
	receipt.Stage, receipt.Step = "produce", "classify"
	receipt.ClaimLedger, receipt.ClaimTransitions = claimLedger(beforeSource, afterSource, receipt.Classification, receipt.Reason)
	sealReceipt(&receipt)
	return receipt
}
