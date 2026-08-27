package languageproofartifactverifier

func summarize(cases []CaseResult, independence IndependenceEvidence, writeSet WriteSetObservation, interventions []InterventionResult, finalLedger ClaimLedger) Summary {
	summary := Summary{CasesTotal: len(cases), TransitionTotal: TransitionTotal, ClaimTemplates: ClaimTemplateTotal,
		ProducerDependencies: independence.ProducerDependencies, ProducerImportNumerator: independence.ProducerImportNumerator,
		ProducerImportDenominator: independence.ProducerImportDenominator, CoreParserDependencies: independence.CoreParserDependencies,
		RepositoryWrites: writeSet.RepositoryWrites, MutationAuthorities: boolToInt(writeSet.MutationAuthority),
		NetRepositoryStateUnchanged: boolToInt(writeSet.NetUnchanged)}
	for _, item := range cases {
		if item.Status == "SATISFIED" {
			summary.CasesSatisfied++
		}
		if item.ID == "valid-proof-carrying-artifact" && item.Status == "SATISFIED" {
			if item.ObservedDecision == "PASS" {
				summary.ValidArtifacts = 1
				summary.EvidenceKindsCarried = EvidenceTotal
				summary.ExactEvidenceLinks = EvidenceTotal
				summary.RecipeMatches = 1
				summary.PreservedTransitions = EvidenceTotal + 1
				summary.AcceptedTransitions = TransitionTotal
				summary.ReadOnlyAuthorities = 1
			}
		}
		switch item.ID {
		case "tampered-evidence":
			summary.TamperedRejections = boolToInt(item.ObservedReason == "PROOF_EVIDENCE_DIGEST_MISMATCH")
		case "coherent-tamper-reconstruction":
			summary.CoherentTamperRejections = boolToInt(item.ObservedReason == "OPERATION_RECONSTRUCTION_MISMATCH")
		case "missing-operation-evidence":
			summary.MissingEvidenceRejections = boolToInt(item.ObservedReason == "PROOF_EVIDENCE_MISSING")
		case "bytes-only-no-authority":
			summary.ByteOnlyDenials = boolToInt(item.ObservedReason == "ARTIFACT_BYTES_NOT_AUTHORITY")
		case "independent-recipe-mismatch":
			summary.RecipeRejections = boolToInt(item.ObservedReason == "INDEPENDENT_RECIPE_MISMATCH")
		case "recipe-only-mismatch":
			summary.RecipeOnlyRejections = boolToInt(item.ObservedReason == "RECIPE_CLAIM_ONLY_MISMATCH")
		case "missing-attachment":
			summary.MissingAttachmentRejections = boolToInt(item.ObservedReason == "ARTIFACT_ATTACHMENT_MISSING")
		case "wrong-attachment-digest":
			summary.WrongAttachmentRejections = boolToInt(item.ObservedReason == "OPERATION_ATTACHMENT_DIGEST_MISMATCH")
		case "unrelated-evidence-tamper":
			summary.UnrelatedEvidenceRejections = boolToInt(item.ObservedReason == "INVARIANT_EVIDENCE_NOT_PRESERVED")
		case "stale-head":
			summary.StaleHeadRejections = boolToInt(item.ObservedReason == "HEAD_BINDING_MISMATCH")
		case "unauthorized-consumer":
			summary.UnauthorizedConsumerDenials = boolToInt(item.ObservedReason == "UNAUTHORIZED_CONSUMER_NOT_ATTESTED")
		}
		for _, claim := range item.Claims {
			summary.ClaimInstances++
			switch claim.Status {
			case "DISCHARGED":
				summary.CaseDischargedClaims++
			case "OPEN":
				summary.CaseOpenClaims++
			case "REFUTED":
				summary.CaseRefutedClaims++
			}
		}
	}
	for _, entry := range finalLedger.Entries {
		switch entry.Status {
		case "OPEN":
			summary.FinalLedgerOpenClaims++
		case "DISCHARGED":
			summary.FinalLedgerDischargedClaims++
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
	previous := ""
	for _, claim := range claims {
		capability, from, to, reason := "ARTIFACT_TRANSPORT", "CARRIED", "PRESERVED", "INDEPENDENT_RECHECK_PASSED"
		if claim.ID == "consumer-authority" {
			capability, from, to, reason = "ARTIFACT_USE", "NONE", "READ_ONLY_CONSUMPTION", "CONSUMER_ONLY_READ_ONLY_AUTHORITY"
		}
		transition := ClaimTransition{ClaimID: claim.ID, Proposition: claim.Proposition, TargetDigest: claim.TargetDigest, StateDigest: claim.StateDigest,
			Dependencies: append([]string(nil), claim.Dependencies...), Capability: capability, From: from, To: to, Producer: ProducerID, Consumer: ConsumerID,
			PriorStateDigest: priorClaimStateDigest(claim, from),
			ProofChoice:      claim.ProofChoice, MetaOperation: claim.MetaOperation, Coordinate: Coordinate{"CONSUME", "recheck-" + claim.ID, "CLAIM_TRANSITION_ACCEPTED"},
			Reason: reason, EvidenceDigest: append([]string(nil), claim.EvidenceDigests...), PreviousDigest: previous}
		transition.Digest = transitionDigest(transition)
		result = append(result, transition)
		previous = transition.Digest
	}
	return result
}
