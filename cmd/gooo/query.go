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
	var options queryOptions
	for index := 1; index < len(args); index++ {
		if index+1 >= len(args) {
			return queryOptions{}, "", usage
		}
		value := args[index+1]
		switch args[index] {
		case "--id":
			if options.ID != "" {
				return queryOptions{}, "", "usage: gooo query: --id may be specified once"
			}
			id, err := semantic.ParseIdentity(value)
			if err != nil {
				return queryOptions{}, "", fmt.Sprintf("usage: gooo query: invalid --id: %v", err)
			}
			options.ID = id.String()
		case "--kind":
			if options.Kind != "" {
				return queryOptions{}, "", "usage: gooo query: --kind may be specified once"
			}
			switch strings.ToLower(value) {
			case "entity":
				options.Kind = semantic.Entity
			case "activity":
				options.Kind = semantic.Activity
			case "agent":
				options.Kind = semantic.Agent
			default:
				return queryOptions{}, "", "usage: gooo query: --kind must be entity, activity, or agent"
			}
		case "--predicate":
			if options.Predicate != "" {
				return queryOptions{}, "", "usage: gooo query: --predicate may be specified once"
			}
			predicate, ok := normalizePredicate(value)
			if !ok {
				return queryOptions{}, "", "usage: gooo query: --predicate is not a supported PROV relation"
			}
			options.Predicate = semantic.Relation(predicate)
		default:
			return queryOptions{}, "", usage
		}
		index++
	}
	return options, filename, ""
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
