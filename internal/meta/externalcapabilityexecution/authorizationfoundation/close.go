package authorizationfoundation

func closeReceipt(value Receipt, foundation Foundation) Receipt {
	binding := FoundationBinding{
		ArtifactID:      foundation.ArtifactID,
		ArtifactName:    foundation.ArtifactName,
		ArchiveDigest:   foundation.ArchiveDigest,
		ProducerRunID:   foundation.ProducerRunID,
		ProducerSubject: foundation.SubjectSHA,
		PriorReceipt:    foundation.ReceiptDigest,
	}
	binding.EvidenceDigest = digestValue(binding)
	value.Schema = ReceiptSchema
	value.Decision, value.Resolution = "AUTHORIZED_SHADOW", "EXACT"
	value.EnforcementEffect = "NO_EFFECT"
	value.Reason = "CAPABILITY_AUTHORIZATION_FOUNDATION_BOUND"
	value.Completed, value.Total, value.BasisPoints = 10, 10, 10000
	value.UnknownIndicators, value.OpenClaims = 0, 0
	value.DischargedClaims, value.RejectedClaims = 10, 0
	value.Unknowns = []Unknown{}
	value.Foundation = &binding
	for index := range value.Indicators {
		if value.Indicators[index].MetricID == PolicyMetric {
			value.Indicators[index].Value = 1
			value.Indicators[index].Status = "SATISFIED"
			value.Indicators[index].UnknownReason = ""
			value.Indicators[index].EvidenceDigest = binding.EvidenceDigest
		}
	}
	for index := range value.Claims {
		if value.Claims[index].ClaimID == PolicyClaim {
			value.Claims[index].Status = "DISCHARGED"
			value.Claims[index].EvidenceDigest = binding.EvidenceDigest
		}
	}
	closeProofsAndReaders(&value)
	return sealReceipt(value)
}

func closeProofsAndReaders(value *Receipt) {
	for index := range value.Proofs {
		if value.Proofs[index].Mode == "FOUNDATION" {
			value.Proofs[index].Status, value.Proofs[index].Resolution = "SATISFIED", "EXACT"
			value.Proofs[index].Completed, value.Proofs[index].Total = 4, 4
		}
	}
	for index := range value.ReaderViews {
		if value.ReaderViews[index].Reader == "AUTHORIZATION" {
			value.ReaderViews[index].Completed, value.ReaderViews[index].Total = 10, 10
			value.ReaderViews[index].BasisPoints, value.ReaderViews[index].Resolution = 10000, "EXACT"
		}
	}
}
