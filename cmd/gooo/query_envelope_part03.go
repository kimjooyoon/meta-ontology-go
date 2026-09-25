package main

import (
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
)

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
