package languagediagnosticprovenance

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

var fixedCodeBindings = []string{
	"internal/formatter",
	"internal/generator",
	"internal/lsp",
	"internal/meta/languagereadiness/languagediagnosticprovenance",
	"internal/meta/languagereadiness/languagediagnosticprovenancebinding",
	"cmd/language-diagnostic-provenance-witness",
	"cmd/language-diagnostic-provenance-readiness-binding",
	"examples/language-diagnostic-provenance",
}

var fixedUseCases = []string{
	"trace-generated-diagnostic",
	"trace-line-directed-diagnostic",
	"reject-ambiguous-or-unknown-provenance",
}

func observeConcept(artifact languageconcept.Artifact) (int, error) {
	if artifact.Decision != "PASS" || !artifact.Ready() || !artifact.ReplayEqual {
		return 0, fmt.Errorf("concept artifact is not an explicit replayed PASS")
	}
	for _, concept := range artifact.Report.Concepts {
		if concept.ID == ConceptID {
			return conceptDrift(concept), nil
		}
	}
	return 0, fmt.Errorf("concept %q is not registered", ConceptID)
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
