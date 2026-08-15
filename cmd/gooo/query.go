package main

import (
	"io"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type queryOptions struct {
	Kind        semantic.Kind
	KindSet     bool
	IDSelector  bool
	operation   string
	root        string
	target      string
	relation    string
	rule        string
	layer       string
	direction   string
	maxDepth    int
	maxDepthSet bool
	limit       int
	limitSet    bool
}

func runQuery(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	options, filename, usage := parseQueryArguments(args)
	if usage != "" {
		return reportUsage(jsonMode, stdout, stderr, "query", usage)
	}
	deadline := time.Now().Add(commandDeadline)
	source, err := readSourceWithDeadline(reader, filename, remainingDeadline(deadline))
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "io.read", err.Error(), syntax.Span{})
	}
	file, diagnostics, err := parseWithDeadline(parser, filename, string(source), remainingDeadline(deadline))
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "parse", err.Error(), syntaxFileSpan(file))
	}
	if diagnostics.HasErrors() {
		if jsonMode {
			if err := writeJSONReport(stdout, newJSONReport("query", "error", filename, syntaxCLIDiagnostics(diagnostics))); err != nil {
				return exitFailure
			}
			return exitFailure
		}
		if err := printSyntaxDiagnostics(stderr, diagnostics); err != nil {
			return exitFailure
		}
		return exitFailure
	}
	if !jsonMode {
		if err := printSyntaxDiagnostics(stderr, diagnostics); err != nil {
			return exitFailure
		}
	}
	ir, err := lowerInspectIRWith(file, remainingDeadline(deadline), bidir.Lower)
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "semantic.lowering", err.Error(), syntaxFileSpan(file))
	}
	return runQueryEngine(options, ir, filename, jsonMode, stdout, stderr)
}

func parseQueryArguments(args []string) (queryOptions, string, string) {
	usage := "usage: gooo query [--json] <file.gooo> [--id <stable-id>] [--kind <kind>] [--predicate <relation>]"
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

func parseQueryStringArgument(options queryOptions, arg, value string) (queryOptions, string) {
	var target *string
	switch arg {
	case "--root":
		target = &options.root
	case "--target":
		target = &options.target
	case "--relation":
		target = &options.relation
	case "--rule":
		target = &options.rule
	case "--layer":
		target = &options.layer
	case "--direction":
		target = &options.direction
	default:
		options.operation = "invalid"
		return options, ""
	}
	if *target != "" {
		options.operation = "invalid"
		return options, ""
	}
	*target = value
	return options, ""
}

func parseQueryBoundArgument(options queryOptions, arg, value string) (queryOptions, string) {
	if arg == "--limit" {
		if options.limitSet {
			options.limit = 0
			return options, ""
		}
		limit, err := parseQueryInteger(value)
		if err != nil {
			options.limit, options.limitSet = 0, true
			return options, ""
		}
		options.limit, options.limitSet = limit, true
		return options, ""
	}
	if options.maxDepthSet {
		options.maxDepth = 0
		return options, ""
	}
	depth, err := parseQueryInteger(value)
	if err != nil {
		options.maxDepth, options.maxDepthSet = 0, true
		return options, ""
	}
	options.maxDepth, options.maxDepthSet = depth, true
	return options, ""
}
