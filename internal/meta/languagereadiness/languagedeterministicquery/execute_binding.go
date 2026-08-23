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

func bindResponseEvidence(
	evidence Evidence,
	requestDigest, replayRequest string,
	first, replay, permutation queryengine.Response,
	graph, permuted *queryengine.Graph,
) (Evidence, error) {
	firstDigest, err := first.CanonicalDigest()
	if err != nil {
		return evidence, err
	}
	replayDigest, err := replay.CanonicalDigest()
	if err != nil {
		return evidence, err
	}
	permutationDigest, err := permutation.CanonicalDigest()
	if err != nil {
		return evidence, err
	}
	evidence.RequestDigest, evidence.ReplayRequest = requestDigest, replayRequest
	evidence.ResultDigest, evidence.ReplayResult = firstDigest, replayDigest
	evidence.PermutationResult = permutationDigest
	evidence.DeterministicRows = len(first.Result.DeterministicMatches)
	evidence.CandidateRows = len(first.Result.CandidateMatches)
	evidence.GraphAfter = graph.StableHash()
	evidence.GraphMutated = evidence.GraphBefore != evidence.GraphAfter || evidence.GraphBefore != permuted.StableHash()
	return evidence, nil
}

func bindingEvidenceSatisfied(evidence Evidence) bool {
	return evidence.RequestDigest == evidence.ReplayRequest &&
		evidence.ResultDigest == evidence.ReplayResult &&
		evidence.ResultDigest == evidence.PermutationResult &&
		evidence.DeterministicRows == 1 && evidence.CandidateRows == 0 &&
		!evidence.GraphMutated
}
