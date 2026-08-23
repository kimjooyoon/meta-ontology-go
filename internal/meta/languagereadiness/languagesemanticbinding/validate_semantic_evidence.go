package languagesemanticbinding

func validateSemanticEvidence(value semanticArtifact, metrics []string) error {
	if err := require(len(value.Cases) == 18, "semantic case denominator mismatch"); err != nil {
		return err
	}
	for _, item := range value.Cases {
		if err := require(item.Status == "SATISFIED", "semantic case not satisfied"); err != nil {
			return err
		}
	}
	classes := map[string]int{}
	metricIDs := make([]string, 0, len(value.Indicators))
	for _, indicator := range value.Indicators {
		classes[indicator.Class]++
		metricIDs = append(metricIDs, indicator.MetricID)
		if err := require(indicator.Satisfied, "semantic indicator not satisfied"); err != nil {
			return err
		}
		if indicator.Class == "GUARDRAIL" {
			if err := require(indicator.Value == 0 && indicator.Target == 0, "semantic guardrail nonzero"); err != nil {
				return err
			}
		}
	}
	validClasses := len(value.Indicators) == 19 && classes["OUTCOME"] == 1
	validClasses = validClasses && classes["DRIVER"] == 10 && classes["GUARDRAIL"] == 8
	if err := require(validClasses && sameSet(metricIDs, metrics), "semantic metric binding mismatch"); err != nil {
		return err
	}
	return validateProofChoices(value.Proofs)
}

func validateProofChoices(proofs []semanticProof) error {
	choices := map[string]int{}
	for _, proof := range proofs {
		choices[proof.Choice]++
		if err := require(proof.Passed && validDigest(proof.EvidenceDigest), "semantic proof rejected"); err != nil {
			return err
		}
	}
	valid := len(proofs) == 3 && choices["FOUNDATION"] == 1
	valid = valid && choices["COHERENCE"] == 1 && choices["REGRESSION"] == 1
	return require(valid, "semantic proof trilemma mismatch")
}
