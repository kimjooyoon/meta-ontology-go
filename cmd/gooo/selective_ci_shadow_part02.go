package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

func parseSelectiveCIShadowArguments(args []string) (selectiveCIShadowOptions, error) {
	var options selectiveCIShadowOptions
	seen := map[string]bool{}
	for index := 0; index < len(args); index++ {
		flag := args[index]
		var target *string
		switch flag {
		case "--base-snapshot":
			target = &options.baseSnapshot
		case "--head-snapshot":
			target = &options.headSnapshot
		case "--plan-input":
			target = &options.planInput
		case "--evidence-input":
			target = &options.evidenceInput
		case "--lane-input":
			target = &options.laneInput
		default:
			return selectiveCIShadowOptions{}, errors.New(selectiveCIShadowUsage)
		}
		if seen[flag] || index+1 >= len(args) || args[index+1] == "" || strings.HasPrefix(args[index+1], "-") {
			return selectiveCIShadowOptions{}, errors.New(selectiveCIShadowUsage)
		}
		seen[flag] = true
		*target = args[index+1]
		index++
	}
	if options.baseSnapshot == "" || options.headSnapshot == "" || options.planInput == "" || options.evidenceInput == "" || options.laneInput == "" {
		return selectiveCIShadowOptions{}, errors.New(selectiveCIShadowUsage)
	}
	return options, nil
}
func runSelectiveCI(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "shadow" {
		fmt.Fprintln(stderr, selectiveCIShadowUsage)
		return exitUsage
	}
	options, err := parseSelectiveCIShadowArguments(args[1:])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	files, missing := readSelectiveCIShadowFiles(options, reader)
	if missing != "" {

		fmt.Fprintf(stderr, "gooo: cli.usage: missing %s input file\n%s\n", missing, selectiveCIShadowUsage)
		return exitUsage
	}
	output := evaluateSelectiveCIShadow(files)
	data, err := encodeSelectiveCIShadowOutput(output)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: selective-ci shadow: output encoding failed: %v\n", err)
		return exitFailure
	}
	if _, err := stdout.Write(data); err != nil {
		return exitFailure
	}
	return exitOK
}
