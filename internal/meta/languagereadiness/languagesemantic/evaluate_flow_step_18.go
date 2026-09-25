package languagesemantic

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic/replay"
)

func evaluateFlowStep16(flow *evaluateFlowState) {
	for _, definition := range flow.slot01.Cases {
		if definition.Kind != CaseSource {
			continue
		}
		observation, observeErr := replay.Observe(flow.slot06, definition.Path)
		result := CaseResult{Definition: definition}
		if observeErr != nil {
			result.Status = StatusUnresolved
			result.Evidence.Error = observeErr.Error()
		} else {
			flow.slot12[definition.ID] = observation
			copyOf := observation
			result.Evidence.Source = &copyOf
			if sourceSatisfied(observation) {
				result.Status = StatusSatisfied
			} else {
				result.Status = StatusNotSatisfied
			}
			if flow.slot13 == nil && observation.DeterministicFacts > 0 {
				candidate := observation
				flow.slot13 = &candidate
			}
		}
		result.Digest = caseDigest(result)
		flow.slot07 = append(flow.slot07, result)
	}
}
