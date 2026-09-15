package languagedeterministicquery

import queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"

func Evaluate(input Input) Report {
	registry := Registry()
	registryDigest := digestJSON(registry)
	source := newSource(input, registryDigest)
	if input.ExpectedHeadSHA == "" {
		return failureReport(source, "EXPECTED_HEAD_UNKNOWN")
	}
	if err := registry.Validate(); err != nil {
		return failureReport(source, "QUERY_REGISTRY_INVALID")
	}
	observation, err := observeConcept(input.ConceptArtifact)
	if err != nil {
		return failureReport(source, "CONCEPT_EVIDENCE_UNKNOWN")
	}
	source.ConceptBound = observation.Drift == 0
	fixture, err := buildFixture(observation.Concept)
	if err != nil {
		return failureReport(source, "META_GRAPH_PROJECTION_FAILED")
	}
	graph, err := instantiateGraph(fixture, false)
	if err != nil {
		return failureReport(source, "META_GRAPH_INVALID")
	}
	permuted, err := instantiateGraph(fixture, true)
	if err != nil {
		return failureReport(source, "PERMUTED_META_GRAPH_INVALID")
	}
	results := executeCases(registry.Cases, fixture, graph, permuted)
	summary := summarize(registry.Cases, results, observation.Drift)
	return successOrFailureReport(source, summary, results)
}

func executeCases(
	definitions []Definition,
	fixture graphFixture,
	graph, permuted *queryengine.Graph,
) []CaseResult {
	results := make([]CaseResult, 0, len(definitions))
	for _, definition := range definitions {
		if definition.Kind == CaseBinding {
			results = append(results, executeBindingCase(definition, fixture, graph, permuted))
		} else {
			results = append(results, executeLawCase(definition, fixture, graph, permuted))
		}
	}
	return results
}
