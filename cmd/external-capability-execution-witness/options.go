package main

import (
	"flag"
	"fmt"
	"io"
)

type options struct {
	Mode         string
	SubjectSHA   string
	SourceRoot   string
	ExternalRoot string
	ParentReport string
	Observation  string
	Report       string
	Suite        string
}

func parseOptions(arguments []string) (options, error) {
	var result options
	flags := flag.NewFlagSet("external-capability-execution-witness", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&result.Mode, "mode", "observe", "observe or replay")
	flags.StringVar(&result.SubjectSHA, "subject-sha", "", "exact Gooo subject SHA")
	flags.StringVar(&result.SourceRoot, "source-root", "", "Gooo repository root")
	flags.StringVar(&result.ExternalRoot, "external-root", "", "pinned gomacro root")
	flags.StringVar(&result.ParentReport, "parent-report", "", "whole-project compatibility report")
	flags.StringVar(&result.Observation, "observation", "", "observation path")
	flags.StringVar(&result.Report, "report", "", "report path")
	flags.StringVar(&result.Suite, "suite", "", "suite path")
	if err := flags.Parse(arguments); err != nil {
		return options{}, err
	}
	if result.Observation == "" || result.Report == "" || result.Suite == "" {
		return options{}, fmt.Errorf("observation, report, and suite are required")
	}
	if result.Mode == "observe" && (result.SubjectSHA == "" || result.SourceRoot == "" ||
		result.ExternalRoot == "" || result.ParentReport == "") {
		return options{}, fmt.Errorf("observe mode requires subject, roots, and parent report")
	}
	if result.Mode != "observe" && result.Mode != "replay" {
		return options{}, fmt.Errorf("unknown mode %q", result.Mode)
	}
	return result, nil
}
