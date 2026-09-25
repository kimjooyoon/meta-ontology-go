package languagedeterministicquery

import (
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
)

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
