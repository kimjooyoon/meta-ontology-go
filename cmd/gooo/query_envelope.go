package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const defaultCLIQueryLimit = 100

var errIncompleteCLIQuery = errors.New("query result is incomplete")

func parseQueryInteger(raw string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.New("must be an integer")
	}
	return value, nil
}

func runQueryEngine(options queryOptions, ir semantic.IR, filename string, jsonMode bool, stdout, stderr io.Writer) int {
	response, err := executeCLIQuery(ir, options)
	if response.Schema == "" {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "query.engine", err.Error(), syntax.Span{})
	}
	if jsonMode {
		if writeQueryResponse(stdout, response) != nil {
			return exitFailure
		}
		if err != nil || response.Status != queryengine.ResponseOK {
			return exitFailure
		}
		return exitOK
	}
	if err != nil || response.Status != queryengine.ResponseOK {
		if response.Error != nil {
			fmt.Fprintf(stderr, "gooo: %s: query: %s\n", filename, response.Error.Message)
		} else {
			fmt.Fprintf(stderr, "gooo: %s: query: %v\n", filename, err)
		}
		return exitFailure
	}
	printQueryResponse(stdout, response)
	return exitOK
}

func executeCLIQuery(ir semantic.IR, options queryOptions) (queryengine.Response, error) {
	graph, err := queryengine.FromSemanticIR(ir)
	if err != nil {
		return queryengine.Response{}, err
	}
	request := queryRequest(options)
	response, err := graph.Execute(request)
	if err != nil || response.Status != queryengine.ResponseOK {
		return response, err
	}
	if request.Operation == queryengine.OperationExact {
		return response, nil
	}
	complete, reason := queryResponseCompleteness(*graph, request, response)
	if complete {
		return response, nil
	}
	response.Status = queryengine.StatusDeferred
	response.Error = &queryengine.EnvelopeError{Code: "incomplete_result", Message: reason}
	response.Hash, err = response.CanonicalDigest()
	if err != nil {
		return queryengine.Response{}, err
	}
	return response, errIncompleteCLIQuery
}

func queryRequest(options queryOptions) queryengine.Request {
	operation := queryengine.Operation(options.operation)
	if operation == "" {
		operation = queryengine.OperationTraversal
	}
	root := options.root
	if root == "" {
		root = options.ID
	}
	relation := options.relation
	if relation == "" && options.Predicate != "" {
		relation = options.Predicate.String()
	}
	layer := queryengine.Layer(options.layer)
	if layer == "" {
		layer = queryengine.LayerDeterministic
	}
	maxDepth := options.maxDepth
	if !options.maxDepthSet {
		maxDepth = 1
	}
	limit := options.limit
	if !options.limitSet {
		limit = defaultCLIQueryLimit
	}
	direction := options.direction
	if direction == "" {
		direction = "outgoing"
	}
	return queryengine.Request{
		Schema: queryengine.QueryEnvelopeSchema, Operation: operation,
		Root: queryengine.ID(root), Target: queryengine.ID(options.target),
		Relation: queryengine.Relation(relation), Rule: queryengine.DerivedRuleID(canonicalCLIRule(options.rule)),
		Layer: layer, Direction: direction, MaxDepth: maxDepth, Limit: limit,
	}
}

func canonicalCLIRule(raw string) string {
	switch raw {
	case "usedBy":
		return string(queryengine.RuleUsedBy)
	case "generatedBy":
		return string(queryengine.RuleGeneratedBy)
	case "derivedTo":
		return string(queryengine.RuleDerivedTo)
	case "dependsOn":
		return string(queryengine.RuleDependsOn)
	default:
		return raw
	}
}

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
	if request.Limit < queryengine.MaxEnvelopeLimit && len(graph.AllFacts()) > request.Limit {
		return false, "query work budget was exhausted before completeness was proven"
	}
	if request.Limit == queryengine.MaxEnvelopeLimit && queryResultSize(response.Result) == request.Limit {
		return false, "query result reached the maximum work budget"
	}
	return true, ""
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

func printQueryResponse(writer io.Writer, response queryengine.Response) {
	fmt.Fprintf(writer, "query: status=%s graph=%s\n", response.Status, response.Metadata.GraphHash)
	for _, fact := range response.Result.DeterministicMatches {
		fmt.Fprintf(writer, "fact: %s %s %s (%s)\n", fact.Subject, fact.Predicate, fact.Object, fact.Status)
	}
	for _, fact := range response.Result.CandidateMatches {
		fmt.Fprintf(writer, "fact: %s %s %s (%s)\n", fact.Subject, fact.Predicate, fact.Object, fact.Status)
	}
	for _, path := range append(append([]queryengine.Path{}, response.Result.DeterministicPaths...), response.Result.CandidatePaths...) {
		fmt.Fprintf(writer, "path: %s\n", strings.Join(idsAsStrings(path.IDs), " -> "))
	}
	for _, fact := range append(append([]queryengine.DerivedFact{}, response.Result.DerivedDeterministic...), response.Result.DerivedCandidates...) {
		fmt.Fprintf(writer, "derived: %s %s %s (%s)\n", fact.Subject, fact.Predicate, fact.Object, fact.Status)
	}
}

func idsAsStrings(ids []queryengine.ID) []string {
	result := make([]string, len(ids))
	for index, id := range ids {
		result[index] = id.String()
	}
	return result
}
