package languagedelivery

func evaluateObligation(item Obligation, sources []SourceObservation, decoded decodedEvidence) ObligationResult {
	result := ObligationResult{ID: item.ID, Audience: item.Audience, Class: item.Class,
		MetaOperation: item.MetaOperation, ProofChoice: item.ProofChoice,
		Source: item.Evidence.Source, Expected: item.Evidence.Target}
	if item.Evidence.Kind == EvidenceUnimplemented {
		result.Status, result.Reason = StatusNotImplemented, "NO_RECEIPT_PRODUCER_REGISTERED"
		result.EvidenceDigest = digestValue(item)
		return result
	}
	state := observationFor(sources, item.Evidence.Source)
	if state.State != "PASS" {
		result.Status, result.Reason = StatusUnknown, state.Reason
		if state.State == "FAIL" {
			result.Status = StatusNotSatisfied
		}
		result.EvidenceDigest = digestValue(state)
		return result
	}
	result.Observed, result.Reason = observeRule(item.Evidence, decoded)
	if result.Observed >= result.Expected {
		result.Status = StatusSatisfied
	} else {
		result.Status = StatusNotSatisfied
	}
	result.EvidenceDigest = digestValue(struct {
		Item string
		State SourceObservation
		Observed int
	}{item.ID, state, result.Observed})
	return result
}

func observationFor(sources []SourceObservation, name SourceName) SourceObservation {
	for _, source := range sources {
		if source.Source == name {
			return source
		}
	}
	return SourceObservation{Source: name, State: "UNKNOWN", Reason: "SOURCE_OBSERVATION_MISSING"}
}
