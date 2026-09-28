package main

import (
	"strings"
)

func parseQueryArguments(args []string) (queryOptions, string, string) {
	usage := "usage: gooo query [--json] <file.gooo> [--id <stable-id>] [--kind <kind>] [--predicate <relation>] [--story <stable-id> --provenance-ledger <path>]"
	if len(args) == 0 {
		return queryOptions{}, "", usage
	}
	filename := args[0]
	if filename == "" || strings.HasPrefix(filename, "-") {
		return queryOptions{}, "", usage
	}
	var options queryOptions
	for index := 1; index < len(args); index++ {
		arg := args[index]
		if arg == "--exact" || arg == "--traverse" || arg == "--derived" {
			if options.operation != "" {
				options.operation = "invalid"
				continue
			}
			options.operation = strings.TrimPrefix(arg, "--")
			continue
		}
		if index+1 >= len(args) {
			return queryOptions{}, "", usage
		}
		var optionUsage string
		if arg == "--id" || arg == "--kind" || arg == "--predicate" {
			options, optionUsage = parseLegacyQueryArgument(options, arg, args[index+1])
		} else {
			options, optionUsage = parseEngineQueryArgument(options, arg, args[index+1])
		}
		if optionUsage != "" {
			return queryOptions{}, "", optionUsage
		}
		index++
	}
	if options.storyID != "" {
		if options.ledgerPath == "" || options.root != "" || options.target != "" || options.relation != "" || options.rule != "" || options.operation != "" && options.operation != "story" {
			return queryOptions{}, "", usage
		}
		options.operation = "story"
	} else if options.ledgerPath != "" {
		return queryOptions{}, "", usage
	}
	if options.operation == "" {
		switch {
		case options.rule != "":
			options.operation = "derived"
		case options.target != "":
			options.operation = "exact"
		default:
			options.operation = "traverse"
		}
	}
	return options, filename, ""
}
