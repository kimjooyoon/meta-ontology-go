package languagegointeroperation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

var fixedCodeBindings = []string{
	"internal/generator",
	"internal/meta/languagereadiness/languagegointeroperation",
	"internal/meta/languagereadiness/languagegointeroperationbinding",
	"cmd/language-go-interoperation-witness",
	"cmd/language-go-interoperation-readiness-binding",
	"examples/language-go-interoperation",
}

var fixedUseCases = []string{"project-gooo-to-go-api", "consume-go-1.27-boundary", "reject-go-boundary-unknowns"}

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
	drift := boolInt(concept.MetaOperation != ExpectedMetaOperation)
	drift += setDrift(concept.CodeBindings, fixedCodeBindings)
	drift += setDrift(concept.MetricBindings, fixedMetricBindings)
	useCases := make([]string, 0, len(concept.UseCases))
	for _, useCase := range concept.UseCases {
		useCases = append(useCases, useCase.ID)
	}
	return drift + setDrift(useCases, fixedUseCases)
}

func setDrift(actual, expected []string) int {
	left, right := stringSet(actual), stringSet(expected)
	drift := 0
	for value := range left {
		_, found := right[value]
		drift += boolInt(!found)
	}
	for value := range right {
		_, found := left[value]
		drift += boolInt(!found)
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

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
