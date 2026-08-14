package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type queryOptions struct {
	ID        string
	Kind      semantic.Kind
	Predicate semantic.Relation

	engine      bool
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
	if options.engine {
		return runQueryEngine(options, ir, filename, jsonMode, stdout, stderr)
	}
	nodes, facts, err := selectQueryResults(ir, options)
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "query.invalid", err.Error(), syntaxFileSpan(file))
	}
	if jsonMode {
		report := newJSONReport("query", "ok", filename, syntaxCLIDiagnostics(diagnostics))
		report.SemanticHash = ir.StableHash()
		report.Nodes, report.Facts = nodes, facts
		if err := writeJSONReport(stdout, report); err != nil {
			return exitFailure
		}
		return exitOK
	}
	fmt.Fprintf(stdout, "query: %s nodes=%d facts=%d\n", filename, len(nodes), len(facts))
	for _, node := range nodes {
		fmt.Fprintf(stdout, "node: %s %s %s/%s\n", node.ID, node.Kind, node.Namespace, node.Name)
	}
	for _, fact := range facts {
		fmt.Fprintf(stdout, "fact: %s %s %s (%s)\n", fact.Subject, fact.Predicate, fact.Object, fact.Status)
	}
	return exitOK
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
	options := queryOptions{engine: queryArgumentsUseEngine(args[1:])}
	for index := 1; index < len(args); index++ {
		arg := args[index]
		if arg == "--exact" || arg == "--traverse" || arg == "--derived" {
			if options.operation != "" {
				return queryOptions{}, "", "usage: gooo query: query operation may be specified once"
			}
			options.engine = true
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
	if options.engine && options.operation == "" {
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

func queryArgumentsUseEngine(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "--exact", "--traverse", "--derived", "--operation", "--op", "--root", "--target", "--relation", "--rule", "--layer", "--direction", "--max-depth", "--depth", "--limit":
			return true
		}
	}
	return false
}

func parseLegacyQueryArgument(options queryOptions, arg, value string) (queryOptions, string) {
	switch arg {
	case "--id":
		if options.ID != "" || options.root != "" {
			return queryOptions{}, "usage: gooo query: --id may be specified once"
		}
		if options.engine {
			options.root = value
			return options, ""
		}
		id, err := semantic.ParseIdentity(value)
		if err != nil {
			return queryOptions{}, fmt.Sprintf("usage: gooo query: invalid --id: %v", err)
		}
		options.ID = id.String()
	case "--kind":
		if options.engine {
			return queryOptions{}, "usage: gooo query: --kind is only available for the legacy projection"
		}
		if options.Kind != "" {
			return queryOptions{}, "usage: gooo query: --kind may be specified once"
		}
		switch strings.ToLower(value) {
		case "entity":
			options.Kind = semantic.Entity
		case "activity":
			options.Kind = semantic.Activity
		case "agent":
			options.Kind = semantic.Agent
		default:
			return queryOptions{}, "usage: gooo query: --kind must be entity, activity, or agent"
		}
	case "--predicate":
		if options.engine {
			if options.relation != "" {
				return queryOptions{}, "usage: gooo query: --predicate may be specified once"
			}
			options.relation = value
			return options, ""
		}
		if options.Predicate != "" {
			return queryOptions{}, "usage: gooo query: --predicate may be specified once"
		}
		predicate, ok := normalizePredicate(value)
		if !ok {
			return queryOptions{}, "usage: gooo query: --predicate is not a supported PROV relation"
		}
		options.Predicate = semantic.Relation(predicate)
	}
	return options, ""
}

func parseEngineQueryArgument(options queryOptions, arg, value string) (queryOptions, string) {
	if arg == "--operation" || arg == "--op" {
		if options.operation != "" {
			return queryOptions{}, "usage: gooo query: query operation may be specified once"
		}
		if value != "exact" && value != "traverse" && value != "derived" {
			return queryOptions{}, "usage: gooo query: --operation must be exact, traverse, or derived"
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
		return queryOptions{}, "usage: gooo query: unknown query option"
	}
	if *target != "" {
		return queryOptions{}, fmt.Sprintf("usage: gooo query: %s may be specified once", arg)
	}
	*target = value
	return options, ""
}

func parseQueryBoundArgument(options queryOptions, arg, value string) (queryOptions, string) {
	if arg == "--limit" {
		if options.limitSet {
			return queryOptions{}, "usage: gooo query: --limit may be specified once"
		}
		limit, err := parseQueryInteger(value)
		if err != nil {
			return queryOptions{}, fmt.Sprintf("usage: gooo query: invalid --limit: %v", err)
		}
		options.limit, options.limitSet = limit, true
		return options, ""
	}
	if options.maxDepthSet {
		return queryOptions{}, "usage: gooo query: --max-depth may be specified once"
	}
	depth, err := parseQueryInteger(value)
	if err != nil {
		return queryOptions{}, fmt.Sprintf("usage: gooo query: invalid --max-depth: %v", err)
	}
	options.maxDepth, options.maxDepthSet = depth, true
	return options, ""
}

func selectQueryResults(ir semantic.IR, options queryOptions) ([]jsonNode, []jsonFact, error) {
	if err := ir.Validate(); err != nil {
		return nil, nil, err
	}
	nodes := make([]jsonNode, 0)
	knownID := options.ID == ""
	for _, node := range ir.Graph.Nodes() {
		if options.ID != "" && node.ID.String() != options.ID {
			continue
		}
		if options.Kind != "" && node.Kind != options.Kind {
			continue
		}
		knownID = true
		nodes = append(nodes, jsonNode{ID: node.ID.String(), Kind: node.Kind.String(), Namespace: node.Namespace.String(), Name: node.Name})
	}
	if !knownID {
		return nil, nil, fmt.Errorf("stable ID %q was not found", options.ID)
	}
	facts := make([]jsonFact, 0)
	for _, fact := range ir.Graph.AllFacts() {
		if options.ID != "" && fact.Subject.String() != options.ID && fact.Object.String() != options.ID {
			continue
		}
		if options.Predicate != "" && fact.Predicate != options.Predicate {
			continue
		}
		facts = append(facts, jsonFact{Subject: fact.Subject.String(), Predicate: fact.Predicate.String(), Object: fact.Object.String(), Status: fact.Status.String()})
	}
	return nodes, facts, nil
}
