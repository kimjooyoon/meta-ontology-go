package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"io"
)

func queryResponseCompleteness(graph queryengine.Graph, request queryengine.Request, response queryengine.Response) (bool, string) {
	probe := request
	probe.Limit = queryengine.MaxEnvelopeLimit
	if probe.Operation == queryengine.OperationTraversal || (probe.Operation == queryengine.OperationDerived &&
		probe.Rule != queryengine.RuleUsedBy && probe.Rule != queryengine.RuleGeneratedBy &&
		probe.Rule != queryengine.RuleDerivedTo) {
		probe.MaxDepth = queryengine.MaxEnvelopeDepth
	}
	expanded, err := graph.Execute(probe)
	if err != nil || expanded.Status != queryengine.ResponseOK {
		return false, "query bounds prevented a complete result"
	}
	if queryResultSize(expanded.Result) > queryResultSize(response.Result) {
		return false, "query result exceeded the requested budget or depth bound"
	}
	if !bytes.Equal(mustQueryJSON(expanded.Result), mustQueryJSON(response.Result)) {
		return false, "query bounds produced an incomplete result"
	}
	if request.Limit < queryengine.MaxEnvelopeLimit && queryFactCount(graph, request.Layer) > request.Limit {
		return false, "query work budget was exhausted before completeness was proven"
	}
	if request.Limit == queryengine.MaxEnvelopeLimit && queryResultSize(response.Result) == request.Limit {
		return false, "query result reached the maximum work budget"
	}
	return true, ""
}
func queryFactCount(graph queryengine.Graph, layer queryengine.Layer) int {
	switch layer {
	case queryengine.LayerCandidate:
		return len(graph.CandidateFacts())
	case queryengine.LayerAll:
		return len(graph.AllFacts())
	default:
		return len(graph.DeterministicFacts())
	}
}
func queryResultSize(result queryengine.QueryResult) int {
	return len(result.DeterministicMatches) + len(result.CandidateMatches) +
		len(result.DeterministicPaths) + len(result.CandidatePaths) +
		len(result.DerivedDeterministic) + len(result.DerivedCandidates)
}
func mustQueryJSON(value queryengine.QueryResult) []byte {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return encoded
}
func writeQueryResponse(writer io.Writer, response queryengine.Response) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	if len(payload)+1 > maxDiagnosticBytes {
		return errDiagnosticLimit
	}
	_, err = fmt.Fprintf(writer, "%s\n", payload)
	return err
}
