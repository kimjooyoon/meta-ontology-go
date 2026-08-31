package foundationseed

func proofs(source Source, authority Authority) []Proof {
	foundation := source.ExactExhaustion && source.SelectedAncestors == 0 &&
		source.ValidCandidates == 0
	coherence := source.ResolutionValid && source.HeadBound &&
		source.SearchComplete && source.MissingComplete && source.Contiguous
	regression := source.AmbiguousCandidates == 0 && source.RepositoryWrites == 0 &&
		source.ReadinessDeltaClaims != nil && *source.ReadinessDeltaClaims == 0 &&
		authorityDenied(authority)
	return []Proof{
		{Choice: "FOUNDATION", Claim: "all bounded ancestors are exactly absent",
			EvidenceDigest: source.ResolutionDigest, Passed: foundation},
		{Choice: "COHERENCE", Claim: "the exhausted chain is contiguous and head-bound",
			EvidenceDigest: source.ResolutionDigest, Passed: coherence},
		{Choice: "REGRESSION", Claim: "the seed grants no mutation, delta, or promotion authority",
			EvidenceDigest: source.ResolutionDigest, Passed: regression},
	}
}

func authorityDenied(value Authority) bool {
	return !value.RepositoryMutationAuthorized && !value.ReadinessDeltaAuthorized &&
		!value.PromotionAuthorized && !value.AutomaticAdoptionAuthorized
}
