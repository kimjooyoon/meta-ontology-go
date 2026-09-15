package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagedebug"
)

type debugOptions struct {
	jsonOutput bool
	entry      string
	breakEvent string
	filename   string
}

func parseDebugOptions(args []string, stderr io.Writer) (debugOptions, error) {
	flags := flag.NewFlagSet("debug", flag.ContinueOnError)
	flags.SetOutput(stderr)
	jsonOutput := flags.Bool("json", false, "emit a JSON debug receipt")
	entry := flags.String("entry", "", "activity to execute")
	breakEvent := flags.String("break-event", "ACTIVITY_INVOKED", "event kind to pause after")
	if err := flags.Parse(args); err != nil {
		return debugOptions{}, err
	}
	if flags.NArg() != 1 || *entry == "" || *breakEvent == "" {
		return debugOptions{}, fmt.Errorf("usage: gooo debug [--json] --entry <activity> --break-event <kind> <file.gooo>")
	}
	return debugOptions{*jsonOutput, *entry, *breakEvent, flags.Arg(0)}, nil
}

func runDebug(args []string, stdout, stderr io.Writer) int {
	options, err := parseDebugOptions(args, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	var execution bytes.Buffer
	runArgs := []string{"--json", "--entry", options.entry, options.filename}
	if code := runSource(runArgs, OSFileReader{}, &execution, stderr); code != exitOK {
		_, _ = stdout.Write(execution.Bytes())
		return code
	}
	receipt := languagedebug.Observe(execution.Bytes(), options.breakEvent)
	return writeDebugReceipt(receipt, options.jsonOutput, stdout, stderr)
}
