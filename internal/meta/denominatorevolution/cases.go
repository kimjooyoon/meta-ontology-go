package denominatorevolution

func CanonicalCases() []CaseInput {
	contract := CanonicalContract()
	base := Denominator{Version: contract.Denominator.Version, Obligations: cloneObligations(contract.Denominator.Obligations)}
	base.Digest = denominatorDigest(base)

	legalSuccessor := Denominator{Version: "gooo/measurement-denominator/v2", Obligations: []Obligation{
		base.Obligations[0], base.Obligations[1], base.Obligations[2], base.Obligations[3],
		{ID: "governance/replay-receipt", Claim: "The migration receipt replays to the same successor digest.", Class: "OUTCOME", ProofChoice: "COHERENCE", MetaOperation: "replay-migration-receipt", Stage: "REPLAY", Step: "compare-successor-digest", Reason: "replay is an exact check of the accepted transition"},
	}}
	legalSuccessor.Digest = denominatorDigest(legalSuccessor)
	legalReceipt := MigrationReceipt{
		Schema: ReceiptSchema, ID: "receipt-legal-advance", Producer: "denominatorevolution.Evaluate", Consumer: "denominatorevolutionverify.Verify",
		Predecessor: DenominatorRef{Version: base.Version, Digest: base.Digest}, Successor: DenominatorRef{Version: legalSuccessor.Version, Digest: legalSuccessor.Digest},
		Additions: []Change{{ObligationID: "governance/replay-receipt", Reason: "NEW_MEASURABLE_OBLIGATION"}},
		Deletions: []Change{{ObligationID: "governance/legacy-summary", Reason: "DEPRECATED_DUPLICATE"}},
		Decision:  "ADVANCE", Reason: "DENOMINATOR_ADVANCE_AUTHORIZED", Coordinate: Coordinate{Stage: "MIGRATE", Step: "issue-receipt", Reason: "DENOMINATOR_ADVANCE_AUTHORIZED"},
		RepositoryWrites: 0, MutationAuthority: false,
	}
	legalReceipt.Digest = receiptDigest(legalReceipt)

	unauthorizedSuccessor := Denominator{Version: "gooo/measurement-denominator/v2", Obligations: append(cloneObligations(base.Obligations), Obligation{
		ID: "governance/improvement-rate", Claim: "Infer an improvement rate from the changed denominator.", Class: "OUTCOME", ProofChoice: "FOUNDATION", MetaOperation: "estimate-improvement-rate", Stage: "GUARD", Step: "infer-rate", Reason: "forbidden estimate",
	})}
	unauthorizedSuccessor.Digest = denominatorDigest(unauthorizedSuccessor)

	unknownPredecessor := Denominator{Version: "gooo/measurement-denominator/v9", Obligations: cloneObligations(base.Obligations)}
	unknownPredecessor.Digest = denominatorDigest(unknownPredecessor)
	unknownSuccessor := Denominator{Version: "gooo/measurement-denominator/v10", Obligations: cloneObligations(base.Obligations)}
	unknownSuccessor.Digest = denominatorDigest(unknownSuccessor)

	return []CaseInput{
		{Spec: contract.Cases[0], Predecessor: base, Successor: legalSuccessor, Receipt: &legalReceipt},
		{Spec: contract.Cases[1], Predecessor: base, Successor: unauthorizedSuccessor},
		{Spec: contract.Cases[2], Predecessor: unknownPredecessor, Successor: unknownSuccessor},
	}
}

func cloneObligations(value []Obligation) []Obligation {
	return append([]Obligation(nil), value...)
}
