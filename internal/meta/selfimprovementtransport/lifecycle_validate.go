package selfimprovementtransport

import "fmt"

var lifecycleUnknownReasons = [][]string{
	{"LIFECYCLE_CONTRACT_INVALID", "ARTIFACT_LIFECYCLE_INPUT_INVALID",
		"ARTIFACT_METADATA_LOOKUP_FAILED", "ARTIFACT_METADATA_RESPONSE_INVALID"},
	{"ARTIFACT_RUN_BINDING_MISMATCH", "ARTIFACT_NOT_FOUND",
		"ARTIFACT_SELECTION_AMBIGUOUS", "UPSTREAM_ARTIFACT_STATE_UNKNOWN"},
	{"ARTIFACT_EXPIRED", "ARTIFACT_METADATA_INVALID", "UPSTREAM_ARTIFACT_STATE_UNKNOWN"},
	{"LIFECYCLE_STEP_NOT_RUN", "ARTIFACT_DOWNLOAD_FAILED", "UPSTREAM_ARTIFACT_STATE_UNKNOWN"},
	{"LIFECYCLE_STEP_NOT_RUN", "ARCHIVE_DIGEST_MISMATCH", "UPSTREAM_ARTIFACT_STATE_UNKNOWN"},
}

func ValidateArtifactLifecycleReceipt(receipt LifecycleReceipt) error {
	if receipt.Schema != LifecycleReceiptSchema || receipt.MetricID != LifecycleMetricID ||
		receipt.DenominatorID != LifecycleDenominatorID ||
		receipt.EnforcementEffect != LifecycleEffectNoEffect {
		return fmt.Errorf("artifact lifecycle identity mismatch")
	}
	if len(receipt.Indicators) != lifecycleFixedStepTotal ||
		len(receipt.Claims) != lifecycleFixedStepTotal {
		return fmt.Errorf("artifact lifecycle denominator mismatch")
	}
	verified, firstUnknown := 0, -1
	for index, definition := range lifecycleDefinitions {
		indicator, claim := receipt.Indicators[index], receipt.Claims[index]
		if indicator.Ordinal != index+1 || indicator.MetricID != LifecycleMetricID ||
			indicator.Class != definition.Class || indicator.ProofChoice != definition.ProofChoice ||
			indicator.Coordinate != (Coordinate{Stage: definition.Stage, Step: definition.Step}) ||
			indicator.MetaOperation != definition.MetaOperation || indicator.Target != 1 ||
			claim.ClaimID != definition.ClaimID ||
			claim.Stage != definition.Stage+"/"+definition.Step ||
			claim.Statement != definition.Statement {
			return fmt.Errorf("artifact lifecycle step %d mismatch", index+1)
		}
		switch indicator.Status {
		case StatusVerified:
			if indicator.Value != 1 || !validDigest(indicator.EvidenceDigest) ||
				indicator.Reason != definition.SuccessReason ||
				claim.Status != LifecycleClaimDischarged || claim.EvidenceDigest != indicator.EvidenceDigest {
				return fmt.Errorf("artifact lifecycle verified step %d invalid", index+1)
			}
			verified++
		case StatusUnknown:
			if indicator.Value != 0 || indicator.Reason == "" || claim.Status != LifecycleClaimOpen ||
				claim.EvidenceDigest != "" || !lifecycleReasonAllowed(index, indicator.Reason) {
				return fmt.Errorf("artifact lifecycle unknown step %d invalid", index+1)
			}
			if firstUnknown < 0 {
				firstUnknown = index
			}
		default:
			return fmt.Errorf("artifact lifecycle step %d status unknown", index+1)
		}
	}
	expected := LifecycleMetrics{
		FixedStepTotal: lifecycleFixedStepTotal, VerifiedTotal: verified,
		UnknownTotal: lifecycleFixedStepTotal - verified,
		OpenTotal: lifecycleFixedStepTotal - verified, DischargedTotal: verified,
		CoverageBasisPoints: verified * 10000 / lifecycleFixedStepTotal,
	}
	if firstUnknown >= 0 {
		expected.UnknownPathCount = 1
		if receipt.Decision != DecisionFailClosed || receipt.Resolution != LifecycleResolutionUnknown ||
			receipt.Reason != receipt.Indicators[firstUnknown].Reason ||
			receipt.Coordinate != receipt.Indicators[firstUnknown].Coordinate {
			return fmt.Errorf("artifact lifecycle fail-closed reduction mismatch")
		}
	} else if receipt.Decision != DecisionPass || receipt.Resolution != LifecycleResolutionExact ||
		receipt.Reason != "GAL5_COMPLETE" {
		return fmt.Errorf("artifact lifecycle exact reduction mismatch")
	}
	if receipt.Metrics != expected || receipt.Authority != (LifecycleAuthority{}) {
		return fmt.Errorf("artifact lifecycle metrics or authority mismatch")
	}
	if receipt.Reason != "LIFECYCLE_CONTRACT_INVALID" &&
		(receipt.Contract.ContractID != ContractID ||
			!validDigest(receipt.Contract.SourceDigest) ||
			!validDigest(receipt.Contract.CanonicalDigest) ||
			receipt.Contract.EntityCount != len(expectedEntities) ||
			receipt.Contract.ActivityCount != len(expectedActivities)) {
		return fmt.Errorf("artifact lifecycle metacode binding mismatch")
	}
	if receipt.Indicators[2].Status == StatusVerified &&
		(receipt.ArtifactID <= 0 || !validDigest(receipt.ArtifactDigest)) {
		return fmt.Errorf("artifact lifecycle immutable metadata missing")
	}
	if receipt.Reason == "ARCHIVE_DIGEST_MISMATCH" &&
		(receipt.Indicators[4].ExpectedDigest != receipt.ArtifactDigest ||
			receipt.Indicators[4].ObservedDigest != receipt.ActualArchiveDigest) {
		return fmt.Errorf("artifact lifecycle contradiction mismatch")
	}
	if receipt.Digest != lifecycleReceiptDigest(receipt) {
		return fmt.Errorf("artifact lifecycle digest mismatch")
	}
	return nil
}

func lifecycleReasonAllowed(index int, reason string) bool {
	for _, allowed := range lifecycleUnknownReasons[index] {
		if reason == allowed {
			return true
		}
	}
	return false
}
