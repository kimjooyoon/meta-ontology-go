package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/audienceresolutionconsumer"
)

type options struct {
	Ledger  string
	Source  string
	Receipt string
	Out     string
	Root    string
}

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	options, err := parseOptions(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	var ledger audienceresolutionconsumer.RawLedger
	ledgerBytes, err := os.ReadFile(options.Ledger)
	if err != nil {
		return reportError(fmt.Errorf("read ledger: %w", err))
	}
	if err := json.Unmarshal(ledgerBytes, &ledger); err != nil {
		return reportError(err)
	}
	source, err := os.ReadFile(options.Source)
	if err != nil {
		return reportError(fmt.Errorf("read source: %w", err))
	}
	receiptBytes, err := os.ReadFile(options.Receipt)
	if err != nil {
		return reportError(fmt.Errorf("read receipt: %w", err))
	}
	var receipt audienceresolutionconsumer.Receipt
	if err := json.Unmarshal(receiptBytes, &receipt); err != nil {
		return reportError(fmt.Errorf("decode receipt: %w", err))
	}
	report := audienceresolutionconsumer.Check(audienceresolutionconsumer.Input{SourcePath: options.Source,
		Source: source, Ledger: ledger, LedgerBytes: ledgerBytes, Receipt: receipt, ReceiptBytes: receiptBytes, RepoRoot: options.Root})
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return reportError(err)
	}
	if err := os.WriteFile(options.Out, append(payload, '\n'), 0o640); err != nil {
		return reportError(fmt.Errorf("write report: %w", err))
	}
	fmt.Printf("audience consumer: %s (%s), imports=%d/%d, source=%d\n", report.Decision, report.Reason,
		report.ProducerImports.Numerator, report.ProducerImports.Denominator, report.SourceReconstruction.DeclarationCount)
	if report.Decision != "PASS" {
		return 1
	}
	return 0
}

func parseOptions(args []string) (options, error) {
	var value options
	for index := 0; index < len(args); index++ {
		if index+1 >= len(args) {
			return options{}, errors.New("every option requires a value")
		}
		switch args[index] {
		case "--ledger":
			value.Ledger = args[index+1]
		case "--source":
			value.Source = args[index+1]
		case "--receipt":
			value.Receipt = args[index+1]
		case "--out":
			value.Out = args[index+1]
		case "--root":
			value.Root = args[index+1]
		default:
			return options{}, errors.New("unknown option: " + args[index])
		}
		index++
	}
	if value.Ledger == "" || value.Source == "" || value.Receipt == "" || value.Out == "" || value.Root == "" {
		return options{}, errors.New("--ledger, --source, --receipt, --out, and --root are required")
	}
	return value, nil
}

func readJSON(path string, target any) error {
	payload, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func reportError(err error) int {
	fmt.Fprintln(os.Stderr, err)
	return 1
}
