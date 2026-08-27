package languageproofartifactverifier

func summarize(cases []CaseResult, independence IndependenceEvidence) Summary {
	summary := Summary{CasesTotal: len(cases), TransitionTotal: TransitionTotal,
		ProducerDependencies: independence.ProducerDependencies}
	for _, item := range cases {
		if item.Status == "SATISFIED" {
			summary.CasesSatisfied++
		}
		switch item.ID {
		case "valid-proof-carrying-artifact":
			if item.ObservedDecision == "PASS" {
				summary.ValidArtifacts = 1
				summary.EvidenceKindsCarried = len(item.Claims)
				summary.ExactEvidenceLinks = len(item.Claims)
				summary.RecipeMatches = 1
				summary.PreservedTransitions = TransitionTotal
				summary.GeneratedAuthority = 0
			}
		case "tampered-evidence":
			if item.ObservedReason == "PROOF_EVIDENCE_DIGEST_MISMATCH" {
				summary.TamperedRejections = 1
			}
		case "missing-operation-evidence":
			if item.ObservedReason == "PROOF_EVIDENCE_MISSING" {
				summary.MissingEvidenceRejections = 1
			}
		case "bytes-only-no-authority":
			if item.ObservedReason == "ARTIFACT_BYTES_NOT_AUTHORITY" {
				summary.ByteOnlyDenials = 1
			}
		case "independent-recipe-mismatch":
			if item.ObservedReason == "INDEPENDENT_RECIPE_MISMATCH" {
				summary.RecipeRejections = 1
			}
		}
	}
	return summary
}

func transitions(cases []CaseResult) []ClaimTransition {
	for _, item := range cases {
		if item.ID == "valid-proof-carrying-artifact" && item.Status == "SATISFIED" {
			return claimTransitions(item.Claims)
		}
	}
	return []ClaimTransition{}
}

func claimTransitions(claims []ClaimResult) []ClaimTransition {
	result := make([]ClaimTransition, 0, TransitionTotal)
	for _, claim := range claims {
		result = append(result, ClaimTransition{ClaimID: claim.ID, From: "CARRIED", To: "PRESERVED",
			Producer: ProducerID, Consumer: ConsumerID, ProofChoice: claim.ProofChoice, MetaOperation: claim.MetaOperation,
			Coordinate: Coordinate{"CONSUME", "recheck-claim", "CLAIM_PRESERVED"}, Reason: "INDEPENDENT_RECHECK_PASSED",
			EvidenceDigest: []string{claim.EvidenceDigest}})
	}
	result = append(result, ClaimTransition{ClaimID: "consumer-authority", From: "NOT_GRANTED", To: "GRANTED",
		Producer: ProducerID, Consumer: ConsumerID, ProofChoice: "COHERENCE", MetaOperation: "grant-only-after-proof",
		Coordinate: Coordinate{"CONSUME", "grant-authority", "ALL_PROOFS_DISCHARGED"}, Reason: "CONSUMER_ONLY_AUTHORITY",
		EvidenceDigest: []string{claims[0].EvidenceDigest, claims[1].EvidenceDigest, claims[2].EvidenceDigest}})
	return result
}
