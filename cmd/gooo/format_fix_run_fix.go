package main

import (
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/formatfix"
)

func runFix(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, ok := parseFixOptions(args)
	if !ok {
		fmt.Fprintln(stderr, fixUsage)
		return exitUsage
	}
	raw, err := reader.ReadFile(options.filename)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: %v\n", options.filename, err)
		return exitFailure
	}
	plan := formatfix.Build(options.filename, string(raw))
	if plan.Decision == formatfix.DecisionFailClosed {
		return reportFormatFailure(plan, options.json, stdout, stderr)
	}
	if options.json {
		if writeFormatJSON(stdout, plan) != nil {
			return exitFailure
		}
		return exitOK
	}
	fmt.Fprintf(stdout, "%s: %s edits=%d writes=%d digest=%s\n",
		plan.Decision, plan.File, len(plan.Edits), plan.DirectWrites, plan.PlanDigest)
	return exitOK
}
