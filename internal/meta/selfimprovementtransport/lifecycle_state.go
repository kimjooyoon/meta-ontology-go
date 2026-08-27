package selfimprovementtransport

import "strings"

func newLifecycleReceipt(contract ContractEvidence, input ArtifactSelectionInput) LifecycleReceipt {
	receipt := LifecycleReceipt{
		Schema: LifecycleReceiptSchema, MetricID: LifecycleMetricID,
		DenominatorID: LifecycleDenominatorID, Contract: contract,
		Repository: input.Repository, ExpectedRunID: input.ExpectedRunID,
		ExpectedRunAttempt: input.ExpectedRunAttempt, ArtifactName: input.ArtifactName,
		EnforcementEffect: LifecycleEffectNoEffect,
	}
	for index, definition := range lifecycleDefinitions {
		receipt.Indicators = append(receipt.Indicators, LifecycleIndicator{
			Ordinal: index + 1, MetricID: LifecycleMetricID, Class: definition.Class,
			ProofChoice:   definition.ProofChoice,
			Coordinate:    Coordinate{Stage: definition.Stage, Step: definition.Step},
			MetaOperation: definition.MetaOperation, Target: 1,
			Status: StatusUnknown, Reason: "LIFECYCLE_STEP_NOT_RUN",
		})
	}
	return receipt
}

func verifyLifecycleStep(receipt *LifecycleReceipt, index int, evidenceDigest string) {
	indicator := &receipt.Indicators[index]
	indicator.Status, indicator.Value = StatusVerified, 1
	indicator.Reason = lifecycleDefinitions[index].SuccessReason
	indicator.ObservationClass = "EVIDENCE"
	indicator.ExpectedDigest, indicator.ObservedDigest = "", ""
	indicator.EvidenceDigest = evidenceDigest
}

func failLifecycleAt(receipt *LifecycleReceipt, index int, reason, expected, observed, class string) {
	indicator := &receipt.Indicators[index]
	indicator.Status, indicator.Value, indicator.Reason = StatusUnknown, 0, reason
	indicator.ExpectedDigest, indicator.ObservedDigest = expected, observed
	indicator.ObservationClass, indicator.EvidenceDigest = class, ""
	for next := index + 1; next < len(receipt.Indicators); next++ {
		receipt.Indicators[next].Status = StatusUnknown
		receipt.Indicators[next].Value = 0
		receipt.Indicators[next].Reason = "UPSTREAM_ARTIFACT_STATE_UNKNOWN"
		receipt.Indicators[next].EvidenceDigest = ""
	}
}

func lifecycleSelectionReason(err error) string {
	coordinate, _, _ := strings.Cut(err.Error(), ":")
	parts := strings.Split(coordinate, "/")
	reason := parts[len(parts)-1]
	if strings.HasPrefix(reason, "SOURCE_RUN_") || reason == "SOURCE_WORKFLOW_PATH_UNKNOWN" {
		return "ARTIFACT_RUN_BINDING_MISMATCH"
	}
	return reason
}
