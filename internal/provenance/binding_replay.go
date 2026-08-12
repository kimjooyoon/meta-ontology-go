package provenance

func predecessorOwners(records []Evidence) map[string]string {
	owners := make(map[string]string)
	for _, record := range records {
		if record.Binding == nil {
			continue
		}
		for _, predecessor := range record.Binding.Predecessors {
			owners[predecessor] = record.ID
		}
	}
	return owners
}

func checkPredecessorClaims(owners map[string]string, record Evidence) error {
	if record.Binding == nil {
		return nil
	}
	for _, predecessor := range record.Binding.Predecessors {
		if previousID, exists := owners[predecessor]; exists {
			return &ReplayError{ID: record.ID, Predecessor: predecessor, PreviousID: previousID}
		}
		owners[predecessor] = record.ID
	}
	return nil
}
