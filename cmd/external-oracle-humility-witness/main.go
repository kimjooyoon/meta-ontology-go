package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	oracle "github.com/kimjooyoon/meta-ontology-go/internal/meta/externaloraclehumility"
)

type options struct {
	head, source, contract, references, output string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	options, err := parseOptions(args, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	var contract oracle.Contract
	var evidence oracle.ReferenceEvidenceSet
	contractRaw, err := os.ReadFile(options.contract)
	if err == nil {
		err = json.Unmarshal(contractRaw, &contract)
	}
	if err == nil {
		evidenceRaw, readErr := os.ReadFile(options.references)
		err = readErr
		if err == nil {
			err = json.Unmarshal(evidenceRaw, &evidence)
		}
	}
	source, sourceErr := os.ReadFile(options.source)
	if err == nil {
		err = sourceErr
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	receipt := oracle.ProduceSourceReceipt(options.head, "examples/external-oracle-humility/main.gooo", source, contract)
	input := oracle.Input{Contract: contract, Evidence: evidence, Receipt: receipt, Source: source, Subject: options.head}
	reports := map[string]oracle.Report{}
	for _, caseID := range []string{"reference-agreement", "reference-mismatch", "reference-absence"} {
		reports[caseID] = oracle.RunCase(caseID, input)
	}
	suite := oracle.RunSuite(contract, input)
	values := map[string]any{"source-receipt.json": receipt, "agreement-report.json": reports["reference-agreement"],
		"mismatch-report.json": reports["reference-mismatch"], "absence-report.json": reports["reference-absence"], "suite.json": suite}
	if err := os.MkdirAll(options.output, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	for name, value := range values {
		raw, encodeErr := oracle.Encode(value)
		if encodeErr != nil {
			fmt.Fprintln(stderr, encodeErr)
			return 2
		}
		if writeErr := os.WriteFile(filepath.Join(options.output, name), raw, 0o644); writeErr != nil {
			fmt.Fprintln(stderr, writeErr)
			return 2
		}
	}
	fmt.Fprintf(stdout, "external oracle humility: %s %d/%d cases=%d/%d agreement=%s authority=%s\n",
		suite.Decision, reports["reference-agreement"].Completed, reports["reference-agreement"].Total,
		suite.CasesSatisfied, suite.CasesTotal, reports["reference-agreement"].ReferenceAgreement,
		reports["reference-agreement"].SemanticAuthority)
	if suite.Decision != "HUMILITY_MODEL_BOUND" || suite.Resolution != "EXACT" {
		return 1
	}
	return 0
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	var result options
	flags := flag.NewFlagSet("external-oracle-humility-witness", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&result.head, "head", "", "exact subject SHA")
	flags.StringVar(&result.source, "source", "", "Gooo source path")
	flags.StringVar(&result.contract, "contract", "", "humility contract path")
	flags.StringVar(&result.references, "references", "", "reference evidence path")
	flags.StringVar(&result.output, "output", "", "output directory")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if result.head == "" || result.source == "" || result.contract == "" || result.references == "" || result.output == "" {
		return options{}, fmt.Errorf("head, source, contract, references, and output are required")
	}
	return result, nil
}
