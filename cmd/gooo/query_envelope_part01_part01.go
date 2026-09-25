package main

import (
	"errors"
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
	"strconv"
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
	if err := rejectCLIEntityFieldsIR(ir); err != nil {
		return queryengine.Response{}, err
	}
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
