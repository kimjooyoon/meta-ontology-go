package main

import (
	"flag"
	"fmt"
	"os"
)

func parse(args []string) (options, error) {
	settings := options{}
	flags := flag.NewFlagSet("ci-plan-scorecard", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&settings.contract, "contract", "", "fixed contract")
	flags.StringVar(&settings.source, "source", "", "Gooo source")
	flags.StringVar(&settings.generatedA, "generated-a", "", "first generated directory")
	flags.StringVar(&settings.generatedB, "generated-b", "", "second generated directory")
	flags.StringVar(&settings.reports, "reports", "", "invocation reports directory")
	flags.StringVar(&settings.replays, "replays", "", "replayed reports directory")
	flags.StringVar(&settings.golden, "golden", "", "golden plans directory")
	flags.StringVar(&settings.profile, "profile", "", "resource profile")
	flags.StringVar(&settings.output, "output", "", "scorecard output")
	flags.BoolVar(&settings.check, "check", false, "validate an existing scorecard")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 || settings.output == "" {
		return options{}, fmt.Errorf("usage: ci-plan-scorecard --output <report.json> [--check | evidence flags]")
	}
	if !settings.check && missingEvidenceFlag(settings) {
		return options{}, fmt.Errorf("all evidence flags are required")
	}
	return settings, nil
}

func missingEvidenceFlag(settings options) bool {
	return settings.contract == "" || settings.source == "" || settings.generatedA == "" || settings.generatedB == "" ||
		settings.reports == "" || settings.replays == "" || settings.golden == "" || settings.profile == ""
}
