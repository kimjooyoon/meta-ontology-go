package languagedeterministicquery

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

type conceptObservation struct {
	Concept languageconcept.Concept
	Drift   int
}

func observeConcept(artifact languageconcept.Artifact) (conceptObservation, error) {
	if artifact.Decision != "PASS" || !artifact.Ready() || !artifact.ReplayEqual {
		return conceptObservation{}, fmt.Errorf("concept artifact is not an explicit replayed PASS")
	}
	for _, concept := range artifact.Report.Concepts {
		if concept.ID == ConceptID {
			return conceptObservation{Concept: concept, Drift: conceptDrift(concept)}, nil
		}
	}
	return conceptObservation{}, fmt.Errorf("concept %q is not registered", ConceptID)
}

func conceptDrift(concept languageconcept.Concept) int {
	drift := 0
	if concept.MetaOperation != ExpectedMetaOperation {
		drift++
	}
	drift += setDrift(concept.CodeBindings, fixedCodeBindings)
	drift += setDrift(concept.MetricBindings, fixedMetricBindings)
	actualUseCases := make([]string, 0, len(concept.UseCases))
	for _, useCase := range concept.UseCases {
		actualUseCases = append(actualUseCases, useCase.ID)
	}
	return drift + setDrift(actualUseCases, fixedUseCases)
}

func setDrift(actual, expected []string) int {
	actualSet := stringSet(actual)
	expectedSet := stringSet(expected)
	drift := 0
	for value := range actualSet {
		if _, found := expectedSet[value]; !found {
			drift++
		}
	}
	for value := range expectedSet {
		if _, found := actualSet[value]; !found {
			drift++
		}
	return drift
}

func stringSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
