package operationconformance

func observePackage(evidence SplitGoEvidence) Decision {
	_, source, err := parseEvidence(evidence.Source)
	if err != nil || len(evidence.Candidates) == 0 {
		return DecisionFail
	}
	packages := map[string]bool{}
	for _, candidate := range evidence.Candidates {
		_, parsed, parseErr := parseEvidence(candidate)
		if parseErr != nil {
			return DecisionFail
		}
		packages[parsed.Name.Name] = true
	}
	if len(packages) != 1 || !packages[source.Name.Name] {
		return DecisionFail
	}
	return DecisionPass
}
