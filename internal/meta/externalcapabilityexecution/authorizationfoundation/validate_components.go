package authorizationfoundation

func validateBootstrapComponents(value Receipt) error {
	indicatorOK, claimOK, proofOK, readerOK := false, false, false, false
	for _, item := range value.Indicators {
		if item.MetricID == PolicyMetric && item.Stage == PolicyStage && item.Value == 0 &&
			item.Target == 1 && item.Status == "UNKNOWN" && item.UnknownReason == "POLICY_FOUNDATION_UNAVAILABLE" {
			indicatorOK = true
		}
	}
	for _, item := range value.Claims {
		if item.ClaimID == PolicyClaim && item.Stage == PolicyStage && item.Status == "OPEN" {
			claimOK = true
		}
	}
	for _, item := range value.Proofs {
		if item.Mode == "FOUNDATION" && item.Status == "UNKNOWN" && item.Completed == 3 &&
			item.Total == 4 && item.Resolution == "UNKNOWN" {
			proofOK = true
		}
	}
	for _, item := range value.ReaderViews {
		if item.Reader == "AUTHORIZATION" && item.Completed == 9 && item.Total == 10 &&
			item.BasisPoints == 9000 && item.Resolution == "UNKNOWN" {
			readerOK = true
		}
	}
	unknownValue := value.Unknowns[0]
	unknownOK := unknownValue.Stage == PolicyStage && unknownValue.IndicatorID == PolicyMetric &&
		unknownValue.Reason == "POLICY_FOUNDATION_UNAVAILABLE"
	if !indicatorOK || !claimOK || !proofOK || !readerOK || !unknownOK {
		return denied("BOOTSTRAP_UNKNOWN_PROVENANCE_MISMATCH")
	}
	return nil
}
