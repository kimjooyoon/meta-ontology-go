package main

import (
	"errors"
	"flag"
	"io"
)

type options struct {
	mode, sourceRoot, externalRoot string
	observation, report, suite     string
}

func parseOptions(args []string) (options, error) {
	var o options
	set := flag.NewFlagSet("external-ecosystem-execution-witness", flag.ContinueOnError)
	set.SetOutput(io.Discard)
	set.StringVar(&o.mode, "mode", "observe", "observe or replay")
	set.StringVar(&o.sourceRoot, "source-root", "", "Gooo repository root")
	set.StringVar(&o.externalRoot, "external-root", "", "pinned external checkout root")
	set.StringVar(&o.observation, "observation", "", "observation JSON path")
	set.StringVar(&o.report, "report", "", "report JSON path")
	set.StringVar(&o.suite, "suite", "", "suite JSON path")
	if err := set.Parse(args); err != nil { return o, err }
	if o.mode != "observe" && o.mode != "replay" { return o, errors.New("mode must be observe or replay") }
	if o.observation == "" || o.report == "" || o.suite == "" { return o, errors.New("observation, report, and suite are required") }
	if o.mode == "observe" && (o.sourceRoot == "" || o.externalRoot == "") { return o, errors.New("observe requires source-root and external-root") }
	return o, nil
}
