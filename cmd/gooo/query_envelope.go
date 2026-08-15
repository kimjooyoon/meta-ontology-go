package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

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
		message := "canonical query engine returned no response"
		if err != nil {
			message = err.Error()
		}
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "query.engine", message, syntax.Span{})
	}
	if writeQueryResponse(stdout, response) != nil {
		return exitFailure
	}
	if err != nil || response.Status != queryengine.ResponseOK {
		return exitFailure
	}
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
	if err := validateCLIQueryKind(*graph, response.Request, options); err != nil {
		return rejectCLIQueryResponse(response, err)
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

type cliQueryValidationError struct {
	code    string
	message string
}

func (err cliQueryValidationError) Error() string {
	return err.code + ": " + err.message
}

func validateCLIQueryKind(graph queryengine.Graph, request queryengine.Request, options queryOptions) error {
	if !options.KindSet {
		return nil
	}
	if options.Kind != semantic.Entity && options.Kind != semantic.Activity && options.Kind != semantic.Agent {
		return cliQueryValidationError{code: "invalid_kind", message: "kind must be entity, activity, or agent"}
	}
	node, ok := graph.Node(request.Root)
	if !ok {
		return cliQueryValidationError{code: "unknown_endpoint", message: fmt.Sprintf("root %q is not present in the query graph", request.Root)}
	}
	want := queryengine.UnknownNodeKind
	switch options.Kind {
	case semantic.Entity:
		want = queryengine.EntityNodeKind
	case semantic.Activity:
		want = queryengine.ActivityNodeKind
	case semantic.Agent:
		want = queryengine.AgentNodeKind
	}
	if node.Kind != want {
		return cliQueryValidationError{
			code:    "kind_mismatch",
			message: fmt.Sprintf("root %q has kind %s, want %s", request.Root, node.Kind, want),
		}
	}
	return nil
}

func rejectCLIQueryResponse(response queryengine.Response, validationErr error) (queryengine.Response, error) {
	var typed cliQueryValidationError
	if !errors.As(validationErr, &typed) {
		return queryengine.Response{}, validationErr
	}
	response.Status = queryengine.ResponseError
	response.Error = &queryengine.EnvelopeError{Code: typed.code, Message: typed.message}
	var err error
	response.Hash, err = response.CanonicalDigest()
	if err != nil {
		return queryengine.Response{}, err
	}
	return response, validationErr
}

func queryRequest(options queryOptions) queryengine.Request {
	operation := queryengine.Operation(options.operation)
	if operation == "" {
		operation = queryengine.OperationTraversal
	}
	root := options.root
	relation := options.relation
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
		if options.IDSelector && operation != queryengine.OperationExact {
			direction = "both"
		}
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
