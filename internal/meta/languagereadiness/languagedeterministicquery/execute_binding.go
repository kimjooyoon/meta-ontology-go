package languagedeterministicquery

import queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"

func executeBindingCase(
	definition Definition,
	fixture graphFixture,
	graph, permuted *queryengine.Graph,
) CaseResult {
	request := requestFor(definition, fixture)
	evidence, err := collectBindingEvidence(request, graph, permuted)
	if err != nil {
		evidence.Error = err.Error()
		return finishCase(definition, evidence, false)
	}
	return finishCase(definition, evidence, bindingEvidenceSatisfied(evidence))
}

func collectBindingEvidence(
	request queryengine.Request,
	graph, permuted *queryengine.Graph,
) (Evidence, error) {
	evidence := Evidence{GraphBefore: graph.StableHash()}
	requestDigest, err := request.CanonicalDigest()
	if err != nil {
		return evidence, err
	}
	replayRequest, err := request.CanonicalDigest()
	if err != nil {
		return evidence, err
	}
	first, err := graph.Execute(request)
	if err != nil {
		return evidence, err
	}
	replay, err := graph.Execute(request)
	if err != nil {
		return evidence, err
	}
	permutation, err := permuted.Execute(request)
	if err != nil {
		return evidence, err
	}
	return bindResponseEvidence(evidence, requestDigest, replayRequest, first, replay, permutation, graph, permuted)
}

func bindingEvidenceSatisfied(evidence Evidence) bool {
	return evidence.RequestDigest == evidence.ReplayRequest &&
		evidence.ResultDigest == evidence.ReplayResult &&
		evidence.ResultDigest == evidence.PermutationResult &&
		evidence.DeterministicRows == 1 && evidence.CandidateRows == 0 &&
		!evidence.GraphMutated
}
