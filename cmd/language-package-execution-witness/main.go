package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagepackageexecution"
)

type options struct {
	head, root, contract, output string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	contractRaw, err := os.ReadFile(filepath.Join(options.root, options.contract))
	if err != nil {
		return err
	}
	contract, err := languagepackageexecution.DecodeContract(contractRaw)
	if err != nil {
		return err
	}
	cases, err := buildCases(options.root)
	if err != nil {
		return err
	}
	report := languagepackageexecution.Evaluate(languagepackageexecution.Input{HeadSHA: options.head, Contract: contract, Cases: cases})
	data, err := languagepackageexecution.Marshal(report)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(options.output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(options.output, data, 0o644); err != nil {
		return err
	}
	fmt.Printf("package execution: cases=%d/%d files=%d replay=%d rejections=%d unknown=%d writes=%d\n", report.Summary.CasesSatisfied, report.Summary.CasesTotal, report.Summary.SourceFilesLoaded, report.Summary.DeterministicReplays, report.Summary.DiagnosticRejections, report.Summary.UnknownDecisions, report.Summary.RepositoryWrites)
	if report.Decision != "PASS" {
		return fmt.Errorf("language-package-execution-witness: %s", report.Reason)
	}
	return nil
}

func parseOptions(args []string) (options, error) {
	set := flag.NewFlagSet("language-package-execution-witness", flag.ContinueOnError)
	var value options
	set.StringVar(&value.head, "head", "", "subject commit SHA")
	set.StringVar(&value.root, "root", ".", "repository root")
	set.StringVar(&value.contract, "contract", "examples/language-package-execution/contract.json", "fixed contract")
	set.StringVar(&value.output, "out", "language-package-execution-report.json", "report output")
	if err := set.Parse(args); err != nil {
		return options{}, err
	}
	if value.head == "" || len(set.Args()) != 0 {
		return options{}, fmt.Errorf("usage: language-package-execution-witness --head SHA [--root DIR] [--out FILE]")
	}
	return value, nil
}
