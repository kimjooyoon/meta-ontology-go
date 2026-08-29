package metrictransition

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect"
)

const (
	effectOutcomeFixedPoint   = "FIXED_POINT"
	effectOutcomeClosed       = "CLOSED"
	effectOutcomeMixedRefuted = "MIXED_CLOSED_REFUTED"
)

func validateEffectOutcome(inputs inputSet) (string, error) {
	ledger := inputs.effectLedger
	if ledger.Status != "BOUND" || !ledger.SourceWorkspaceUnchanged ||
		ledger.PromotionAuthorized || ledger.UnboundExecutorOperations != 0 ||
		ledger.BoundExecutorOperations != ledger.SelectedPlanOperations {
		return "", fmt.Errorf("metric transition effect binding is not non-promoting")
	}
	if len(ledger.Effects) != ledger.SelectedPlanOperations ||
		len(inputs.receiptReport.Receipts)+len(inputs.receiptReport.Failures) != len(ledger.Effects) {
		return "", fmt.Errorf("metric transition operation evidence counts diverged")
	}
	effects := make(map[string]transformationeffect.Effect, len(ledger.Effects))
	for _, effect := range ledger.Effects {
		if effect.ActionIndicatorID == "" || effects[effect.ActionIndicatorID].ActionIndicatorID != "" {
			return "", fmt.Errorf("metric transition operation identity is not unique")
		}
		if effect.Status != "APPLIED" && effect.Status != "REFUTED" {
			return "", fmt.Errorf("metric transition operation status is unknown")
		}
		effects[effect.ActionIndicatorID] = effect
	}
	consumed := make(map[string]string, len(ledger.Effects))
	for _, receipt := range inputs.receiptReport.Receipts {
		if receipt.ActionIndicatorID == "" || consumed[receipt.ActionIndicatorID] != "" {
			return "", fmt.Errorf("metric transition receipt identity is not disjoint")
		}
		if effect, ok := effects[receipt.ActionIndicatorID]; !ok || effect.Status != "APPLIED" {
			return "", fmt.Errorf("metric transition receipt is not bound to an applied effect")
		}
		consumed[receipt.ActionIndicatorID] = "receipt"
	}
	for _, failure := range inputs.receiptReport.Failures {
		if failure.ActionIndicatorID == "" || consumed[failure.ActionIndicatorID] != "" {
			return "", fmt.Errorf("metric transition failure identity is not disjoint")
		}
		if failure.Decision != "REFUTED" {
			return "", fmt.Errorf("metric transition failure decision is not typed REFUTED")
		}
		if effect, ok := effects[failure.ActionIndicatorID]; !ok || effect.Status != "REFUTED" {
			return "", fmt.Errorf("metric transition failure is not bound to its effect")
		}
		consumed[failure.ActionIndicatorID] = "failure"
	}
	if len(consumed) != len(effects) {
		return "", fmt.Errorf("metric transition effect identities are not fully consumed")
	}
	outcome := classifyEffectOutcome(ledger, inputs.receiptReport)
	if outcome == "" {
		return "", fmt.Errorf("metric transition operation outcome is unknown")
	}
	if ledger.OperationOutcome != outcome || ledger.ReceiptDecision != string(inputs.receiptReport.Decision) ||
		ledger.ReceiptCount != len(inputs.receiptReport.Receipts) ||
		ledger.FailureCount != len(inputs.receiptReport.Failures) ||
		ledger.UnknownCount != len(inputs.receiptReport.Unknowns) {
		return "", fmt.Errorf("metric transition operation outcome evidence diverged")
	}
	if err := validateCausalUnknowns(ledger, inputs.receiptReport); err != nil {
		return "", err
	}
	if inputs.provenanceReport.Decision != generation.ArtifactProvenanceDecisionBound ||
		inputs.provenanceReport.HeadSHA != ledger.HeadSHA ||
		inputs.provenanceReport.ReceiptReportDigest != inputs.receiptReport.ReportDigest {
		return "", fmt.Errorf("metric transition provenance is not bound to operation evidence")
	}
	return outcome, nil
}

func classifyEffectOutcome(ledger transformationeffect.Ledger, report generation.ReceiptReport) string {
	if len(ledger.Effects) == 0 && ledger.Decision == "FIXED_POINT" &&
		ledger.Reason == "EXACT_FIXED_POINT" &&
		ledger.OperationOutcome == effectOutcomeFixedPoint &&
		ledger.ReceiptDecision == string(generation.ReceiptDecisionFixedPoint) &&
		ledger.ReceiptCount == 0 && ledger.FailureCount == 0 && ledger.UnknownCount == 0 &&
		report.Decision == generation.ReceiptDecisionFixedPoint &&
		report.Reason == generation.ReceiptReasonExactFixedPoint &&
		len(report.Receipts) == 0 && len(report.Failures) == 0 && len(report.Unknowns) == 0 {
		return effectOutcomeFixedPoint
	}
	if ledger.Decision == "APPLIED" && ledger.Reason == "SANDBOX_EFFECTS_VERIFIED" &&
		ledger.OperationOutcome == effectOutcomeClosed &&
		ledger.ReceiptDecision == string(generation.ReceiptDecisionConformant) &&
		report.Decision == generation.ReceiptDecisionConformant &&
		report.Reason == generation.ReceiptReasonVerified && len(report.Failures) == 0 &&
		len(report.Unknowns) == 0 {
		return effectOutcomeClosed
	}
	if ledger.Decision == "APPLIED" && ledger.Reason == "SANDBOX_EFFECTS_VERIFIED" &&
		ledger.OperationOutcome == effectOutcomeMixedRefuted &&
		ledger.ReceiptDecision == string(generation.ReceiptDecisionRefuted) &&
		report.Decision == generation.ReceiptDecisionRefuted &&
		report.Reason == generation.ReceiptReasonRefutedOperation &&
		len(report.Receipts) > 0 && len(report.Failures) > 0 {
		return effectOutcomeMixedRefuted
	}
	return ""
}
