package languageproofartifactverifier

func summarize(cases []CaseResult, independence IndependenceEvidence, writeSet WriteSetObservation, interventions []InterventionResult) Summary {
	summary := Summary{CasesTotal: len(cases), TransitionTotal: TransitionTotal,
		ProducerDependencies:    independence.ProducerDependencies,
		ProducerImportNumerator: independence.ProducerImportNumerator, ProducerImportDenominator: independence.ProducerImportDenominator,
		CoreParserDependencies: independence.CoreParserDependencies,
		RepositoryWrites:       writeSet.RepositoryWrites, MutationAuthorities: boolToInt(writeSet.MutationAuthority),
		PromotionAuthorities: 0, SemanticAuthorities: 0}
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
				summary.PreservedTransitions = EvidenceTotal
				summary.GeneratedAuthority = 0
				summary.ReadOnlyAuthorities = 1
			}
		case "tampered-evidence":
			if item.ObservedReason == "PROOF_EVIDENCE_DIGEST_MISMATCH" {
				summary.TamperedRejections = 1
			}
		case "coherent-tamper-reconstruction":
			if item.ObservedReason == "OPERATION_RECONSTRUCTION_MISMATCH" {
				summary.CoherentTamperRejections = 1
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
		for _, claim := range item.Claims {
			switch claim.Status {
			case "DISCHARGED":
				summary.LedgerDischargedClaims++
			case "OPEN":
				summary.LedgerOpenClaims++
			case "REFUTED":
				summary.LedgerRefutedClaims++
			}
		}
	}
	for _, item := range interventions {
		if item.Status != "SATISFIED" {
			continue
		}
		switch item.Kind {
		case "SEMANTIC":
			summary.SemanticInterventions++
		case "NONSEMANTIC":
			summary.NonsemanticInterventions++
		}
	}
	return summary
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
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
		result = append(result, ClaimTransition{ClaimID: claim.ID, Capability: "ARTIFACT_TRANSPORT", From: "CARRIED", To: "PRESERVED",
			Producer: ProducerID, Consumer: ConsumerID, ProofChoice: claim.ProofChoice, MetaOperation: claim.MetaOperation,
			Coordinate: Coordinate{"CONSUME", "recheck-claim", "CLAIM_PRESERVED"}, Reason: "INDEPENDENT_RECHECK_PASSED",
			EvidenceDigest: []string{claim.EvidenceDigest}})
	}
	result = append(result, ClaimTransition{ClaimID: "consumer-authority", Capability: "ARTIFACT_USE", From: "NONE", To: "READ_ONLY_CONSUMPTION",
		Producer: ProducerID, Consumer: ConsumerID, ProofChoice: "COHERENCE", MetaOperation: "grant-read-only-consumption",
		Coordinate: Coordinate{"CONSUME", "grant-read-only-consumption", "ALL_PROOFS_DISCHARGED"}, Reason: "CONSUMER_ONLY_READ_ONLY_AUTHORITY",
		EvidenceDigest: []string{claims[0].EvidenceDigest, claims[1].EvidenceDigest, claims[2].EvidenceDigest}})
	return result
}
