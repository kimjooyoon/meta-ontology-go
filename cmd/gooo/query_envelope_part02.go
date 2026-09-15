package main

import (
	"errors"
	"fmt"
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
