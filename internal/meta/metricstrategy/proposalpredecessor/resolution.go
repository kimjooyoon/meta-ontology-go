package proposalpredecessor

import (
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

func BuildResolution(repository, currentHead, predecessorSHA, requestedRoute, reason string, selection *Report, observationEvidence ObservationEvidence) (ResolutionReceipt, error) {
	if !KnownFailureReason(reason) || reason == ReasonSelected || !validSHA(currentHead) || !validSHA(predecessorSHA) || repository == "" || !validRoute(requestedRoute) {
		return ResolutionReceipt{}, fmt.Errorf("proposal predecessor resolution identity is invalid")
	}
	observationDecision, observationResolution := DecisionUnknown, ResolutionLower
	var unknown *Unknown
	if reason == ReasonRouteContradiction {
		observationDecision, observationResolution = DecisionRefuted, ResolutionExact
	} else {
		unknown = routeEvidenceUnknown(reason, requestedRoute)
	}
	if err := ValidateObservationEvidence(observationEvidence); err != nil {
		return ResolutionReceipt{}, err
	}
	if selection != nil {
		if err := Validate(*selection); err != nil {
			return ResolutionReceipt{}, fmt.Errorf("proposal predecessor resolution selection is invalid: %w", err)
		}
		if selection.Repository != repository || selection.CurrentSubjectSHA != currentHead || selection.PredecessorSHA != predecessorSHA || selection.RequestedRoute != requestedRoute || selection.Reason != reason || selection.Ready() {
			return ResolutionReceipt{}, fmt.Errorf("proposal predecessor resolution selection linkage is invalid")
		}
		observationDecision, observationResolution, unknown = selection.ObservationDecision, selection.ObservationResolution, selection.Unknown
	}
	receipt := ResolutionReceipt{
		Schema: ResolutionSchema, Repository: repository, CurrentHeadSHA: currentHead,
		PredecessorSHA: predecessorSHA, RequestedRoute: requestedRoute, Conformance: ResolutionConformancePass,
		Decision: ResolutionFailClosed, Reason: reason, Resolution: ResolutionLower,
		Stage: ResolutionStage, Step: ResolutionStep, ObservationDecision: observationDecision,
		ObservationResolution: observationResolution, Unknown: unknown, PromotionAuthority: false,
		Selection: selection, ObservationEvidence: observationEvidence,
	}
	return sealResolution(receipt)
}

func sealResolution(receipt ResolutionReceipt) (ResolutionReceipt, error) {
	receipt.ReportDigest = ""
	digest, err := artifact.Digest(receipt)
	receipt.ReportDigest = digest
	return receipt, err
}

func ValidateResolution(receipt ResolutionReceipt, expectedRepository, expectedCurrentHead, expectedPredecessorSHA, expectedRoute string) error {
	if receipt.Schema != ResolutionSchema || receipt.Repository == "" || !validSHA(receipt.CurrentHeadSHA) || !validSHA(receipt.PredecessorSHA) || !validRoute(receipt.RequestedRoute) || !validRoute(expectedRoute) {
		return fmt.Errorf("proposal predecessor resolution identity is invalid")
	}
	if expectedRepository == "" || !validSHA(expectedCurrentHead) || !validSHA(expectedPredecessorSHA) ||
		receipt.Repository != expectedRepository || receipt.CurrentHeadSHA != expectedCurrentHead || receipt.PredecessorSHA != expectedPredecessorSHA || receipt.RequestedRoute != expectedRoute {
		return fmt.Errorf("FAIL_CLOSED: proposal predecessor resolution context mismatch")
	}
	if receipt.Conformance != ResolutionConformancePass || receipt.Decision != ResolutionFailClosed ||
		receipt.Resolution != ResolutionLower || receipt.Stage != ResolutionStage || receipt.Step != ResolutionStep ||
		receipt.PromotionAuthority || receipt.ReadinessDelta != nil || !KnownFailureReason(receipt.Reason) || receipt.Reason == ReasonSelected {
		return fmt.Errorf("proposal predecessor resolution is not fail-closed")
	}
	if receipt.Reason == ReasonRouteContradiction {
		if receipt.ObservationDecision != DecisionRefuted || receipt.ObservationResolution != ResolutionExact || receipt.Unknown != nil {
			return fmt.Errorf("proposal predecessor resolution refutation diverged")
		}
	} else if receipt.ObservationDecision != DecisionUnknown || receipt.ObservationResolution != ResolutionLower || !validUnknown(receipt.Unknown, receipt.Reason) {
		return fmt.Errorf("proposal predecessor resolution unknown observation diverged")
	}
	if err := ValidateObservationEvidence(receipt.ObservationEvidence); err != nil {
		return err
	}
	if receipt.Selection != nil {
		if err := Validate(*receipt.Selection); err != nil {
			return fmt.Errorf("proposal predecessor resolution selection is invalid: %w", err)
		}
		if receipt.Selection.Repository != receipt.Repository || receipt.Selection.CurrentSubjectSHA != receipt.CurrentHeadSHA ||
			receipt.Selection.PredecessorSHA != receipt.PredecessorSHA || receipt.Selection.RequestedRoute != receipt.RequestedRoute || receipt.Selection.Reason != receipt.Reason || receipt.Selection.Ready() {
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
