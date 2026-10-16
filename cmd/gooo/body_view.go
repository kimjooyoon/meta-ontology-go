package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/viewlens"
)

const bodyViewUsage = "usage: gooo body-view --activity <name> [--actor <id>] [--audience agent|person] " +
	"[--view auto|full|inputs] [--prefer auto|full|inputs] [--hint <json>] [--format json|text] <file.gooo>"

type bodyViewOptions struct {
	activity, actor, audience, view, prefer, hint, format string
	capabilities                                          bool
}

func runBodyView(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options := bodyViewOptions{}
	flags := flag.NewFlagSet("body-view", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.activity, "activity", "", "source activity name")
	flags.StringVar(&options.actor, "actor", "agent", "opaque actor identity")
	flags.StringVar(&options.audience, "audience", "agent", "agent or person")
	flags.StringVar(&options.view, "view", "auto", "advisory presentation for this request")
	flags.StringVar(&options.prefer, "prefer", "", "preference to record separately; auto clears it")
	flags.StringVar(&options.hint, "hint", "", "preloaded advisory JSON, at most 4096 bytes")
	flags.StringVar(&options.format, "format", "json", "json or text")
	flags.BoolVar(&options.capabilities, "capabilities", false, "describe the machine interface as JSON")
	if err := flags.Parse(args); err != nil {
		writeBodyViewError(stdout, stderr, "REQUEST_SYNTAX_INVALID", err)
		return exitUsage
	}
	if options.capabilities {
		return writeBodyViewCapabilities(stdout, stderr)
	}
	if flags.NArg() != 1 || options.activity == "" || (options.format != "json" && options.format != "text") {
		writeBodyViewError(stdout, stderr, "REQUEST_ARGUMENTS_INVALID", fmt.Errorf("%s", bodyViewUsage))
		return exitUsage
	}
	return executeBodyView(flags.Arg(0), options, reader, stdout, stderr)
}

func executeBodyView(filename string, options bodyViewOptions, reader SourceReader, stdout, stderr io.Writer) int {
	source, err := reader.ReadFile(filename)
	if err != nil {
		return writeBodyViewError(stdout, stderr, "SOURCE_UNAVAILABLE", err)
	}
	report, err := bodycodegen.InspectTypedValueOrigins(filename, source, options.activity)
	if err != nil {
		return writeBodyViewError(stdout, stderr, "SOURCE_INSPECTION_FAILED", err)
	}
	selection := viewlens.Select(viewlens.Request{
		ActorID: options.actor, Audience: options.audience, View: options.view, Prefer: options.prefer,
		ActivityID: report.ActivityID, SourceDigest: report.SourceSHA256,
	}, []byte(options.hint))
	output := makeBodyViewOutput(report, selection, options, filename)
	if options.format == "text" {
		err = writeBodyViewText(stdout, output, source)
	} else {
		err = json.NewEncoder(stdout).Encode(output)
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

func writeBodyViewError(stdout, stderr io.Writer, code string, err error) int {
	if encodeErr := json.NewEncoder(stdout).Encode(map[string]any{
		"schema": "gooo/body-view-error/v1", "status": "ERROR", "code": code, "message": err.Error(),
	}); encodeErr != nil {
		fmt.Fprintln(stderr, encodeErr)
	}
	return exitFailure
}
