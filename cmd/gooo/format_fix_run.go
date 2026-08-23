package main

import (
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/formatfix"
)

func runFormat(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, ok := parseFormatOptions(args)
	if !ok {
		fmt.Fprintln(stderr, formatUsage)
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
	formatted, err := formatfix.Apply(string(raw), plan)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: %v\n", options.filename, err)
		return exitFailure
	}
	if options.json {
		status := "formatted"
		if !plan.Changed {
			status = "canonical"
		}
		report := formatCommandReport{Schema: "gooo/format-report/v1", Command: "format",
			Status: status, File: options.filename, Changed: plan.Changed, Source: formatted,
			SourceDigest: plan.SourceDigest, FormattedDigest: plan.ResultDigest,
			Diagnostics: []string{}, DirectWrites: 0}
		if writeFormatJSON(stdout, report) != nil {
			return exitFailure
		}
		return exitOK
	}
	if options.check {
		if plan.Changed {
			fmt.Fprintf(stderr, "gooo: %s: format.check: source is not canonical\n", options.filename)
			return exitFailure
		}
		fmt.Fprintf(stdout, "canonical: %s\n", options.filename)
		return exitOK
	}
	fmt.Fprint(stdout, formatted)
	return exitOK
}

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
