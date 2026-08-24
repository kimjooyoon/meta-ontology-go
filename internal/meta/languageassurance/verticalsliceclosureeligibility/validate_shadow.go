package verticalsliceclosureeligibility

func validBoundaries(values []boundary) bool {
	expected := map[string][3]int{
		"syntax": {17, 17, 1}, "semantics": {20, 20, 2}, "binding": {12, 12, 2},
		"use-cases": {3, 3, 1}, "toolchain": {156, 156, 3}, "release": {20, 20, 3},
	}
	if len(values) != len(expected) {
		return false
	}
	for _, value := range values {
		target, ok := expected[value.ID]
		if !ok || value.Value != target[0] || value.Target != target[1] ||
			value.LinksSatisfied != target[2] || value.LinksTotal != target[2] ||
			value.Status != "SATISFIED" || value.Resolution != ResolutionExact ||
			value.HeadSHA != "64a529d71d2fc76000e345b4dd86ad982ebb679e" ||
			!value.EvidenceAvailable || value.UnknownTopDecision || value.KnownFailure ||
			value.RepositoryWrites != 0 {
			return false
		}
	}
	return true
}

func validSourceIndicators(values []sourceIndicator) bool {
	classes, proofs := map[string]int{}, map[string]int{}
	for _, value := range values {
		if !value.Satisfied {
			return false
		}
		classes[value.Class]++
		proofs[value.ProofChoice]++
	}
	return len(values) == 6 && classes["OUTCOME"] == 1 && classes["DRIVER"] == 2 &&
		classes["GUARDRAIL"] == 3 && proofs["FOUNDATION"] == 3 &&
		proofs["COHERENCE"] == 2 && proofs["REGRESSION"] == 1
}
