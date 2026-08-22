package proposalpromotion

func evaluate(currentHead, evidenceHead string, source Source) Receipt {
	coordinates := buildCoordinates(currentHead, evidenceHead, source)
	satisfied := 0
	for _, coordinate := range coordinates {
		if coordinate.Status == "SATISFIED" {
			satisfied++
		}
	}
	summary := Summary{
		Satisfied: satisfied, Total: totalCoordinates,
		NotSatisfied: totalCoordinates - satisfied,
		Unresolved:   totalCoordinates - satisfied,
		ReadinessBPS: satisfied * 10_000 / totalCoordinates,
		ValidPredecessors: source.Selection.ValidCandidates,
		AmbiguousCandidates: source.Selection.AmbiguousCandidates,
		RepositoryWrites: source.Selection.RepositoryWrites +
			source.Selection.SelectedRepositoryWrites + source.Contract.RepositoryWrites,
	}
	decision, reason := DecisionPass, ReasonReady
	if satisfied != totalCoordinates {
		decision, reason = DecisionFailClosed, ReasonUnresolved
	}
	receipt := Receipt{
		Schema: Schema, Repository: source.Selection.Repository,
		CurrentHeadSHA: currentHead, EvidenceHeadSHA: evidenceHead,
		Decision: decision, Reason: reason,
		MetaOperation: "promote-verified-change-proposal",
		Source: source, Summary: summary, Coordinates: coordinates,
		RepositoryWrites: summary.RepositoryWrites,
		RepositoryMutationAuthorized: false,
	}
	receipt.Indicators = buildIndicators(summary, source)
	receipt.Proofs = buildProofs(coordinates)
	return seal(receipt)
}
