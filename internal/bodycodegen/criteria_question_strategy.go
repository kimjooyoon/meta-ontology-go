package bodycodegen

import "github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"

type CriteriaQuestionOutputEvidence struct {
	Available   int `json:"available"`
	Unavailable int `json:"unavailable"`
}

func criteriaQuestionSchema(o assemblyspec.Observation) string {
	if o.CriteriaStrategy != "" {
		return "gooo/criteria-questions/v2"
	}
	return "gooo/criteria-questions/v1"
}

// This reads already observed output availability, never predicate results.
func bindCriteriaQuestionOutputEvidence(round *CriteriaQuestionRound, outputs []CriteriaQuestionOutput) {
	for i := range round.Probes {
		probe := &round.Probes[i]
		if probe.AlreadyTested {
			continue
		}
		probe.OutputEvidence = &CriteriaQuestionOutputEvidence{}
		for _, index := range probe.OutputIndices {
			if outputs[index].Error == "" {
				probe.OutputEvidence.Available++
			} else {
				probe.OutputEvidence.Unavailable++
			}
		}
	}
}

func criteriaQuestionRecommendation(round CriteriaQuestionRound) (*int, string) {
	best, index := 0, -1
	for i, probe := range round.Probes {
		if !probe.AlreadyTested && probe.SeparatedPairs > best {
			best, index = probe.SeparatedPairs, i
		}
	}
	if index >= 0 {
		return new(index), "SEPARATE_OUTPUTS"
	}
	for i, probe := range round.Probes {
		if !probe.AlreadyTested && probe.OutputEvidence != nil && probe.OutputEvidence.Available > 0 {
			return new(i), "CHECK_CRITERIA"
		}
	}
	return nil, ""
}

func validCriteriaQuestionReason(o assemblyspec.Observation, round CriteriaQuestionRound) bool {
	if o.CriteriaStrategy == "" {
		return round.RecommendationReason == "" && round.Probes[*round.RecommendedIndex].SeparatedPairs > 0
	}
	index, reason := criteriaQuestionRecommendation(round)
	return index != nil && *index == *round.RecommendedIndex && reason == round.RecommendationReason
}

func criteriaQuestionStop(o assemblyspec.Observation, round CriteriaQuestionRound) string {
	status := recordObservationStop(RecordObservationRound{Unobserved: round.Unobserved,
		SurvivingMasks: round.SurvivingMasks, UnavailableMasks: round.UnavailableMasks})
	if o.CriteriaStrategy == "" || round.Unobserved > 0 || len(round.SurvivingMasks)+len(round.UnavailableMasks) == 0 {
		return status
	}
	for _, probe := range round.Probes {
		if !probe.AlreadyTested && probe.OutputEvidence != nil && probe.OutputEvidence.Unavailable > 0 {
			return "QUESTION_OUTPUTS_UNAVAILABLE"
		}
	}
	return status
}
