package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func parseLegacyQueryArgument(options queryOptions, arg, value string) (queryOptions, string) {
	switch arg {
	case "--id":
		if options.root != "" {
			options.operation = "invalid"
			return options, ""
		}
		options.root, options.IDSelector = value, true
	case "--kind":
		if options.KindSet {
			options.Kind = semantic.Kind("invalid")
			return options, ""
		}
		options.KindSet = true
		switch strings.ToLower(value) {
		case "entity":
			options.Kind = semantic.Entity
		case "activity":
			options.Kind = semantic.Activity
		case "agent":
			options.Kind = semantic.Agent
		default:
			options.Kind = semantic.Kind(strings.TrimSpace(value))
		}
	case "--predicate":
		if options.relation != "" {
			options.relation = "invalid"
			return options, ""
		}
		options.relation = value
	}
	return options, ""
}
func parseEngineQueryArgument(options queryOptions, arg, value string) (queryOptions, string) {
	if arg == "--operation" || arg == "--op" {
		if options.operation != "" {
			options.operation = "invalid"
			return options, ""
		}
		if value != "exact" && value != "traverse" && value != "derived" {
			options.operation = value
			return options, ""
		}
		options.operation = value
		return options, ""
	}
	if arg == "--max-depth" || arg == "--depth" || arg == "--limit" {
		return parseQueryBoundArgument(options, arg, value)
	}
	return parseQueryStringArgument(options, arg, value)
}
