package main

import (
	"flag"
	"fmt"
	"io"
)

type options struct {
	SubjectSHA string
	CaseID     string
	ReadmePath string
	GoModPath  string
	OutputPath string
}

func parseOptions(arguments []string) (options, error) {
	var result options
	flags := flag.NewFlagSet("external-ecosystem-conformance-witness", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&result.SubjectSHA, "subject-sha", "", "exact Gooo subject SHA")
	flags.StringVar(&result.CaseID, "case", "suite", "fixed conformance case")
	flags.StringVar(&result.ReadmePath, "readme", "", "pinned upstream README")
	flags.StringVar(&result.GoModPath, "go-mod", "", "pinned upstream go.mod")
	flags.StringVar(&result.OutputPath, "output", "", "receipt path")
	if err := flags.Parse(arguments); err != nil {
		return options{}, err
	}
	if result.SubjectSHA == "" || result.ReadmePath == "" || result.GoModPath == "" ||
		result.OutputPath == "" {
		return options{}, fmt.Errorf("subject-sha, readme, go-mod, and output are required")
	}
	return result, nil
}
