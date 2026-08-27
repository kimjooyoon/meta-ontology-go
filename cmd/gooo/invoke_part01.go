package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"
)

const invokeUsage = "usage: gooo invoke [--json] --entry <activity> --input <input.json> <file.gooo>"

type invokeOptions struct {
	entry  string
	input  string
	source string
}

func runInvoke(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	options, err := parseInvokeArguments(args)
	if err != nil {
		if jsonMode {
			_ = json.NewEncoder(stdout).Encode(map[string]string{"schema": metainvocation.ReportSchema, "decision": metainvocation.DecisionClosed, "reason": "CLI_USAGE"})
		} else {
			fmt.Fprintln(stderr, invokeUsage)
		}
		return exitUsage
	}
	source, err := reader.ReadFile(options.source)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: invoke: read source: %v\n", err)
		return exitFailure
	}
	input, err := reader.ReadFile(options.input)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: invoke: read input: %v\n", err)
		return exitFailure
	}
	program, err := metainvocation.Compile(options.source, source, metainvocation.StandardRegistry())
	if err != nil {
		fmt.Fprintf(stderr, "gooo: invoke: compile: %v\n", err)
		return exitFailure
	}
	report, err := metainvocation.Invoke(program, options.entry, input)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: invoke: evaluate: %v\n", err)
		return exitFailure
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintf(stderr, "gooo: invoke: write report: %v\n", err)
		return exitFailure
	}
	if report.Decision != metainvocation.DecisionPass {
		return exitFailure
	}
	return exitOK
}

func parseInvokeArguments(args []string) (invokeOptions, error) {
	options := invokeOptions{}
	for len(args) != 0 {
		switch args[0] {
		case "--entry":
			if len(args) < 2 || options.entry != "" {
				return invokeOptions{}, fmt.Errorf("invalid entry option")
			}
			options.entry, args = args[1], args[2:]
		case "--input":
			if len(args) < 2 || options.input != "" {
				return invokeOptions{}, fmt.Errorf("invalid input option")
			}
			options.input, args = args[1], args[2:]
		default:
			if options.source != "" || len(args[0]) == 0 || args[0][0] == '-' {
				return invokeOptions{}, fmt.Errorf("invalid source argument")
			}
			options.source, args = args[0], args[1:]
		}
	}
	if options.entry == "" || options.input == "" || options.source == "" {
		return invokeOptions{}, fmt.Errorf("missing required argument")
	}
	return options, nil
}
