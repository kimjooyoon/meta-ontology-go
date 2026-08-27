package selfimprovementtransport

func closeLifecycle(receipt *LifecycleReceipt) {
	receipt.Metrics = LifecycleMetrics{FixedStepTotal: lifecycleFixedStepTotal}
	receipt.Claims = nil
	firstUnknown := -1
	for index, indicator := range receipt.Indicators {
		claim := LifecycleClaim{
			ClaimID: lifecycleDefinitions[index].ClaimID,
			Stage: indicator.Coordinate.Stage + "/" + indicator.Coordinate.Step,
			Statement: lifecycleDefinitions[index].Statement,
		}
		if indicator.Status == StatusVerified {
			receipt.Metrics.VerifiedTotal++
			claim.Status, claim.EvidenceDigest = LifecycleClaimDischarged, indicator.EvidenceDigest
		} else {
			receipt.Metrics.UnknownTotal++
			claim.Status = LifecycleClaimOpen
			if firstUnknown < 0 {
				firstUnknown = index
			}
		}
		receipt.Claims = append(receipt.Claims, claim)
	}
	receipt.Metrics.OpenTotal = receipt.Metrics.UnknownTotal
	receipt.Metrics.DischargedTotal = receipt.Metrics.VerifiedTotal
	receipt.Metrics.CoverageBasisPoints = receipt.Metrics.VerifiedTotal * 10000 / lifecycleFixedStepTotal
	if firstUnknown >= 0 {
		receipt.Metrics.UnknownPathCount = 1
		receipt.Decision, receipt.Resolution = DecisionFailClosed, LifecycleResolutionUnknown
		receipt.Reason = receipt.Indicators[firstUnknown].Reason
		receipt.Coordinate = receipt.Indicators[firstUnknown].Coordinate
	} else {
		receipt.Decision, receipt.Resolution = DecisionPass, LifecycleResolutionExact
		receipt.Reason = "GAL5_COMPLETE"
		receipt.Coordinate = Coordinate{Stage: "REDUCE", Step: "close-gal5"}
	}
	receipt.EnforcementEffect = LifecycleEffectNoEffect
	receipt.Digest = lifecycleReceiptDigest(*receipt)
}

func lifecycleReceiptDigest(receipt LifecycleReceipt) string {
	receipt.Digest = ""
	return digestJSON(receipt)
}
