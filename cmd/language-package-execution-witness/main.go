package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagepackageexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageexecution"
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

func buildCases(root string) ([]languagepackageexecution.CaseEvidence, error) {
	sources, err := packageexecution.LoadDirectory(filepath.Join(root, "examples", "billing-package"))
	if err != nil {
		return nil, err
	}
	request := packageexecution.Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: sources}
	positive := packageexecution.Execute(request)
	mismatch := append([]packageexecution.Source(nil), sources...)
	mismatch[0].Content = strings.Replace(mismatch[0].Content, "package billing", "package other", 1)
	if mismatch[0].Content == sources[0].Content {
		return nil, fmt.Errorf("witness: package header fixture not found")
	}
	duplicate := append(append([]packageexecution.Source(nil), sources...), packageexecution.Source{Filename: "duplicate.gooo", Content: sources[1].Content})
	return []languagepackageexecution.CaseEvidence{
		{ID: "positive-package-execution", Receipt: positive},
		{ID: "deterministic-replay", Receipt: packageexecution.Execute(request)},
		{ID: "header-mismatch-rejection", Receipt: packageexecution.Execute(packageexecution.Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: mismatch})},
		{ID: "duplicate-declaration-rejection", Receipt: packageexecution.Execute(packageexecution.Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: duplicate})},
		{ID: "source-count-rejection", Receipt: packageexecution.Execute(packageexecution.Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: sources[:1]})},
	}, nil
}
