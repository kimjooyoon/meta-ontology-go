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
	meta, metaErr := ReadMetaContract()
	return produceBytes(input, before, after, meta, metaErr), nil
}

func produceBytes(input Input, beforeRaw, afterRaw []byte, meta MetaContract, metaErr error) Receipt {
	beforeSource, beforeErr := projectSourceSide(input.BeforePath, beforeRaw, true)
	afterSource, afterErr := projectSourceSide(input.AfterPath, afterRaw, false)
	if beforeSource.path == "" {
		beforeSource = sourceEnvelope(input.BeforePath, beforeRaw)
	}
	if afterSource.path == "" {
		afterSource = sourceEnvelope(input.AfterPath, afterRaw)
	}
	receipt := Receipt{Schema: ReceiptSchema, CaseID: input.CaseID, SubjectSHA: input.SubjectSHA, ExpectedSubjectSHA: input.SubjectSHA, ObservedCheckoutSHA: input.ObservedCheckoutSHA, SubjectBinding: subjectBinding(input.SubjectSHA, input.ObservedCheckoutSHA), Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, ProofChoice: "FOUNDATION", Stage: "produce", Step: "separate-delta-layers", MetaSourcePath: MetaSourcePath, MetaContractDigest: meta.Digest, DenominatorVersion: meta.Version, DenominatorCases: meta.DenominatorCases, ModeledSemanticComponents: ModeledComponentCount, TotalSemanticComponents: TotalComponentCount, DeclaredProjectionComponentKindCoverageBPS: semanticCoverageBPS(ModeledComponentCount, TotalComponentCount), SemanticEquivalenceClaim: SemanticEquivalenceNotClaimed, Before: snapshot(beforeRaw, beforeSource, beforeErr), After: snapshot(afterRaw, afterSource, afterErr), TextualDelta: textualDelta(beforeRaw, afterRaw), Effects: observeEffects(input)}
	if metaErr != nil {
		return unknownReceipt(receipt, beforeSource, afterSource, nil, nil, "meta-source", "parse-lower", ReasonMeta)
	}
	if beforeErr != nil || afterErr != nil {
		return unknownReceipt(receipt, beforeSource, afterSource, beforeErr, afterErr)
	}
	if receipt.SubjectBinding != "EXACT" {
		return subjectUnknown(receipt, beforeSource, afterSource, receipt.SubjectBinding)
	}
	receipt.StructuralDelta = structuralDelta(beforeSource, afterSource)
	receipt.SemanticComponentDelta = semanticComponentDelta(beforeSource, afterSource)
	receipt.SemanticClaimDelta = claimDelta(beforeSource, afterSource)
	if receipt.SemanticClaimDelta.Status == "UNKNOWN" {
		return ambiguousReceipt(receipt, beforeSource, afterSource)
	}
	receipt.RawDecision = receipt.TextualDelta.Decision
	receipt.SemanticDecision = SemanticPreserved
	receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = DecisionFixedPoint, ResolutionExact, ClassPreserved, ReasonTextualOnly
	if beforeSource.semanticDigest != afterSource.semanticDigest && len(receipt.SemanticComponentDelta.Added)+len(receipt.SemanticComponentDelta.Removed)+len(receipt.SemanticComponentDelta.Changed) == 0 {
		return unmodeledReceipt(receipt, beforeSource, afterSource)
	}
	if hasSemanticDelta(receipt.StructuralDelta, receipt.SemanticComponentDelta, receipt.SemanticClaimDelta) {
		receipt.SemanticDecision = SemanticChanged
		receipt.Decision, receipt.Classification = DecisionDelta, ClassChanged
		if len(receipt.SemanticComponentDelta.Added)+len(receipt.SemanticComponentDelta.Removed)+len(receipt.SemanticComponentDelta.Changed) > 0 && len(receipt.StructuralDelta.AddedNodes)+len(receipt.StructuralDelta.RemovedNodes)+len(receipt.StructuralDelta.AddedFacts)+len(receipt.StructuralDelta.RemovedFacts)+len(receipt.SemanticClaimDelta.Added)+len(receipt.SemanticClaimDelta.Removed)+len(receipt.SemanticClaimDelta.Changed) == 0 {
			receipt.Reason = ReasonComponentDelta
		} else {
			receipt.Reason = ReasonMeaning
		}
	}
	receipt.Stage, receipt.Step = "produce", "classify"
	receipt.ClaimLedger, receipt.ClaimTransitions = claimLedger(beforeSource, afterSource, receipt.Classification, receipt.Reason)
	finishReceiptClaims(&receipt)
	sealReceipt(&receipt)
	return receipt
}
