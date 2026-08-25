package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

const runSourceUsage = "usage: gooo run [--json] --entry <activity> <file.gooo>"

func runSource(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	options, err := parseRunSourceArguments(args)
	if err != nil {
		return reportUsage(jsonMode, stdout, stderr, "run", runSourceUsage)
	}
	source, err := readSource(reader, options.filename)
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "run", options.filename,
			"read", err.Error(), sourceexecutionSpan())
	}
	receipt := sourceexecution.Execute(sourceexecution.Request{
		Filename: options.filename, Source: string(source), Entry: options.entry,
	})
	payload, err := sourceexecution.Marshal(receipt)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: run receipt: %v\n", err)
		return exitFailure
	}
	if jsonMode {
		if _, err := stdout.Write(payload); err != nil {
			return exitFailure
		}
	} else if receipt.Decision == "PASS" {
		fmt.Fprintf(stdout, "executed: %s.%s(%s) -> %s digest=%s\n", receipt.Entry.Package,
			receipt.Entry.Activity, inputNames(receipt.Entry.Inputs), receipt.Entry.Output.Name, receipt.Digest)
	} else {
		diagnostic := receipt.Diagnostics[0]
		fmt.Fprintf(stderr, "%s: %s: %s\n", options.filename, diagnostic.Code, diagnostic.Message)
	}
	if receipt.Decision == "PASS" {
		return exitOK
	}
	return exitFailure
}

func inputNames(bindings []sourceexecution.Binding) string {
	names := make([]string, len(bindings))
	for index, binding := range bindings {
		names[index] = binding.Name
	}
	return strings.Join(names, ", ")
}
