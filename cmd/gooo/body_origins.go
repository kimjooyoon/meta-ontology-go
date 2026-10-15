package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const bodyOriginsUsage = "usage: gooo body-origins --activity <name> [--format json|text] [--fill-candidates | --fill-candidate <id>] [--inputs-for <package/path:Activity> [--inputs-shape ancestry-columns|source-control-columns] [--field <stable-field-id>] [--include-control]] <file.gooo>"

func runBodyOrigins(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, err := parseBodyOriginOptions(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	source, err := reader.ReadFile(options.file)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	if options.list {
		return writeBodyFillOriginCandidates(options.file, source, options.flags["--activity"], stdout, stderr)
	}
	if options.flags["--fill-candidate"] != "" {
		return writeBodyFillCandidateOrigins(options.file, source, options.flags, stdout, stderr)
	}
	return writeTypedBodyOrigins(options, source, stdout, stderr)
}

func writeTypedBodyOrigins(options bodyOriginOptions, source []byte, stdout, stderr io.Writer) int {
	flags := options.flags
	report, err := bodycodegen.InspectTypedValueOrigins(options.file, source, flags["--activity"])
	if err == nil && flags["--inputs-for"] != "" {
		var raw []byte
		if flags["--inputs-shape"] == "source-control-columns" {
			raw, err = bodycodegen.TypedSourceControlInputs(report, flags["--inputs-for"])
		} else {
			raw, err = bodycodegen.TypedValueOriginInputs(report, flags["--inputs-for"], flags["--field"], options.control)
		}
		if err == nil {
			_, err = fmt.Fprintln(stdout, string(raw))
		}
	} else if err == nil && options.format == "text" {
		err = bodycodegen.WriteTypedOriginText(stdout, report, source)
	} else if err == nil {
		err = json.NewEncoder(stdout).Encode(report)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	if report.Status != "RESOLVED" {
		return exitFailure
	}
	return exitOK
}
