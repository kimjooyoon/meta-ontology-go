package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
	"time"
)

func runAnalyzeWithLowerer(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer, lower func(*syntax.File) (semantic.IR, error)) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: gooo analyze <file.gooo>")
		return exitUsage
	}
	filename := args[0]
	deadline := time.Now().Add(commandDeadline)
	source, err := readSourceWithDeadline(reader, filename, remainingDeadline(deadline))
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: read error: %v\n", filename, err)
		return exitFailure
	}
	file, syntaxDiagnostics, err := parseWithDeadline(parser, filename, string(source), remainingDeadline(deadline))
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: parse error: %v\n", filename, err)
		return exitFailure
	}
	plan := newFixPlan(source, syntaxDiagnostics, file)
	if syntaxDiagnostics.HasErrors() {
		if hasCLIEntityFieldsDeferredDiagnostic(syntaxDiagnostics) {
			if !reportDiagnostics(syntaxDiagnostics, stderr) {
				return exitFailure
			}
			return exitFailure
		}
		plan.Status = fixPlanSyntaxInvalid
	} else {
		ir, lowerErr := lowerInspectIRWith(file, remainingDeadline(deadline), lower)
		if lowerErr != nil {
			if isCLIEntityFieldsDeferredError(lowerErr) {
				if !reportSemanticDiagnostic(filename, file, lowerErr, stderr) {
					return exitFailure
				}
				return exitFailure
			}
			plan.Status = fixPlanSemanticInvalid
			plan.Diagnostics = append(plan.Diagnostics, semanticFixDiagnostics(lowerErr, fileSpan(file))...)
		} else {
			applyFixPlanIR(&plan, ir)
		}
	}
	plan.Diagnostics = canonicalFixPlanDiagnostics(plan.Diagnostics)
	payload, err := marshalFixPlan(plan)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: fix plan failed: %v\n", filename, err)
		return exitFailure
	}
	if err := writeInspectOutput(stdout, payload, deadline); err != nil {
		fmt.Fprintf(stderr, "gooo: fix plan output: %v\n", err)
		return exitFailure
	}
	if plan.Status != fixPlanReady || hasErrorFixDiagnostic(plan.Diagnostics) {
		return exitFailure
	}
	return exitOK
}
