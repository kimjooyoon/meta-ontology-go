package bodyexecution

func observedCompositionReadEvidence(prior Composition) (string, string, []CompositionFaultSite, []CompositionConditionSite, []CompositionReadSite, []CompositionReadInput, error) {
	projection, driver, faults, conditions, err := observedCompositionEvidence(prior)
	if err != nil {
		return "", "", nil, nil, nil, nil, err
	}
	reads, err := compositionReadSites(prior)
	if err != nil {
		return "", "", nil, nil, nil, nil, err
	}
	var inputs []CompositionReadInput
	if len(reads) > 0 {
		inputs = compositionReadInputs(prior)
		prefix := readObservationPrefix(prior.Source)
		projection, err = observedReadProjection(projection, prefix, prior, reads)
		if err == nil {
			driver, err = observedReadDriver(driver, prefix, prior)
		}
	}
	return projection, driver, faults, conditions, reads, inputs, err
}
