package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/audienceresolution"
)

type options struct {
	contract string
	ledger   string
	source   string
	out      string
}

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	options, err := parseOptions(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	var input audienceresolution.Input
	if err := readJSON(options.contract, &input.Contract); err != nil {
		return reportError(err)
	}
	if err := readJSON(options.ledger, &input.Ledger); err != nil {
		return reportError(err)
	}
	source, err := os.ReadFile(options.source)
	if err != nil {
		return reportError(fmt.Errorf("read source: %w", err))
	}
	input.Replay, input.SourcePath, input.Source = input.Ledger, options.source, source
	receipt := audienceresolution.Evaluate(input)
	if err := audienceresolution.ValidateReceipt(receipt); err != nil {
		return reportError(fmt.Errorf("independent receipt check: %w", err))
	}
	payload, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return reportError(err)
	}
	if err := os.WriteFile(options.out, append(payload, '\n'), 0o640); err != nil {
		return reportError(fmt.Errorf("write receipt: %w", err))
	}
	fmt.Printf("audience resolution: %s %d/%d USER=%d/4 TOOL_AUTHOR=%d/8 GOVERNOR=%d/12\n",
		receipt.Decision, receipt.Summary.Coordinates.Satisfied, receipt.Summary.Coordinates.Total,
		receipt.Views[0].Satisfied, receipt.Views[1].Satisfied, receipt.Views[2].Satisfied)
	if receipt.Decision != "PASS" {
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
		case "--contract":
			value.contract = args[index+1]
		case "--ledger":
			value.ledger = args[index+1]
		case "--source":
			value.source = args[index+1]
		case "--out":
			value.out = args[index+1]
		default:
			return options{}, errors.New("unknown option: " + args[index])
		}
		index++
	}
	if value.contract == "" || value.ledger == "" || value.source == "" || value.out == "" {
		return options{}, errors.New("--contract, --ledger, --source, and --out are required")
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
