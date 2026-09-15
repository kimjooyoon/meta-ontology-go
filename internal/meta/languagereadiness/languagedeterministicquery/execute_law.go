package languagedeterministicquery

import queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"

func executeLawCase(definition Definition, fixture graphFixture, graph, permuted *queryengine.Graph) CaseResult {
	switch definition.ID {
	case "law-candidate-non-authority":
		return candidateAuthorityLaw(definition, fixture, graph)
	case "law-unknown-layer":
		return unknownLayerLaw(definition, fixture, graph)
	case "law-unknown-endpoint":
		return unknownEndpointLaw(definition, fixture, graph)
	case "law-read-only-graph":
		return readOnlyGraphLaw(definition, graph, permuted)
	default:
		return finishCase(definition, Evidence{UnknownAccepted: true, Error: "unknown law"}, false)
	}
}

func candidateAuthorityLaw(definition Definition, fixture graphFixture, graph *queryengine.Graph) CaseResult {
	deterministicRequest := exactRequest(
		fixture.CandidateEntityID, fixture.CandidateActionID, queryengine.WasGeneratedBy, queryengine.LayerDeterministic,
	)
	candidateRequest := deterministicRequest
	candidateRequest.Layer = queryengine.LayerCandidate
	deterministic, deterministicErr := graph.Execute(deterministicRequest)
	candidate, candidateErr := graph.Execute(candidateRequest)
	evidence := Evidence{
		DeterministicRows: len(deterministic.Result.DeterministicMatches),
		CandidateRows:     len(candidate.Result.CandidateMatches),
		CandidatePromoted: len(deterministic.Result.DeterministicMatches) != 0,
		Error:             errorMessage(deterministicErr) + errorMessage(candidateErr),
	}
	satisfied := deterministicErr == nil && candidateErr == nil &&
		evidence.DeterministicRows == 0 && evidence.CandidateRows == 1 && !evidence.CandidatePromoted
	return finishCase(definition, evidence, satisfied)
}

func unknownLayerLaw(definition Definition, fixture graphFixture, graph *queryengine.Graph) CaseResult {
	request := exactRequest(fixture.ConceptID, fixture.OperationID, queryengine.WasGeneratedBy, queryengine.Layer("unknown"))
	response, err := graph.Execute(request)
	accepted := err == nil || response.Status != queryengine.ResponseError
	evidence := Evidence{Rejected: !accepted, UnknownAccepted: accepted, Error: errorMessage(err)}
	return finishCase(definition, evidence, !accepted)
}

func unknownEndpointLaw(definition Definition, fixture graphFixture, graph *queryengine.Graph) CaseResult {
	unknown := stableID("unknown-endpoint", ConceptID)
	request := exactRequest(unknown, fixture.OperationID, queryengine.Used, queryengine.LayerDeterministic)
	response, err := graph.Execute(request)
	accepted := err == nil || response.Status != queryengine.ResponseError
	evidence := Evidence{Rejected: !accepted, UnknownAccepted: accepted, Error: errorMessage(err)}
	return finishCase(definition, evidence, !accepted)
}

func readOnlyGraphLaw(definition Definition, graph, permuted *queryengine.Graph) CaseResult {
	before := graph.StableHash()
	after := graph.StableHash()
	mutated := before != after || before != permuted.StableHash()
	evidence := Evidence{GraphBefore: before, GraphAfter: after, GraphMutated: mutated}
	return finishCase(definition, evidence, !mutated)
}
