package proposalpredecessor

import (
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

func BuildResolution(repository, currentHead, predecessorSHA, reason string, selection *Report) (ResolutionReceipt, error) {
	if !KnownFailureReason(reason) || reason == ReasonSelected || !validSHA(currentHead) || !validSHA(predecessorSHA) || repository == "" {
		return ResolutionReceipt{}, fmt.Errorf("proposal predecessor resolution identity is invalid")
	}
	if selection != nil {
		if err := Validate(*selection); err != nil {
			return ResolutionReceipt{}, fmt.Errorf("proposal predecessor resolution selection is invalid: %w", err)
		}
		if selection.Repository != repository || selection.CurrentSubjectSHA != currentHead || selection.PredecessorSHA != predecessorSHA || selection.Reason != reason || selection.Ready() {
			return ResolutionReceipt{}, fmt.Errorf("proposal predecessor resolution selection linkage is invalid")
		}
	}
	receipt := ResolutionReceipt{
		Schema: ResolutionSchema, Repository: repository, CurrentHeadSHA: currentHead,
		PredecessorSHA: predecessorSHA, Conformance: ResolutionConformancePass,
		Decision: ResolutionFailClosed, Reason: reason, Resolution: ResolutionLower,
		Stage: ResolutionStage, Step: ResolutionStep, PromotionAuthority: false,
		Selection: selection,
	}
	return sealResolution(receipt)
}

func sealResolution(receipt ResolutionReceipt) (ResolutionReceipt, error) {
	receipt.ReportDigest = ""
	digest, err := artifact.Digest(receipt)
	receipt.ReportDigest = digest
	return receipt, err
}

func ValidateResolution(receipt ResolutionReceipt) error {
	if receipt.Schema != ResolutionSchema || receipt.Repository == "" || !validSHA(receipt.CurrentHeadSHA) || !validSHA(receipt.PredecessorSHA) {
		return fmt.Errorf("proposal predecessor resolution identity is invalid")
	}
	if receipt.Conformance != ResolutionConformancePass || receipt.Decision != ResolutionFailClosed ||
		receipt.Resolution != ResolutionLower || receipt.Stage != ResolutionStage || receipt.Step != ResolutionStep ||
		receipt.PromotionAuthority || receipt.ReadinessDelta != nil || !KnownFailureReason(receipt.Reason) || receipt.Reason == ReasonSelected {
		return fmt.Errorf("proposal predecessor resolution is not fail-closed")
	}
	if receipt.Selection != nil {
		if err := Validate(*receipt.Selection); err != nil {
			return fmt.Errorf("proposal predecessor resolution selection is invalid: %w", err)
		}
		if receipt.Selection.Repository != receipt.Repository || receipt.Selection.CurrentSubjectSHA != receipt.CurrentHeadSHA ||
			receipt.Selection.PredecessorSHA != receipt.PredecessorSHA || receipt.Selection.Reason != receipt.Reason || receipt.Selection.Ready() {
			return fmt.Errorf("proposal predecessor resolution selection linkage is invalid")
		}
	}
	digest := receipt.ReportDigest
	sealed, err := sealResolution(receipt)
	if err != nil || digest == "" || digest != sealed.ReportDigest {
		return fmt.Errorf("proposal predecessor resolution digest diverged")
	}
	return nil
}
