package main

import (
	"fmt"
	"os"
	"path/filepath"

	freshness "github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness"
)

func run(args []string) error {
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	if options.check != "" {
		report, err := freshness.LoadReport(options.check)
		if err != nil {
			return err
		}
		return freshness.Validate(report)
	}
	if options.contract == "" || options.source == "" || options.head == "" || options.independence == "" || options.output == "" {
		return fmt.Errorf("contract, source, head, independence, and output are required")
	}
	contract, err := freshness.LoadContract(options.contract)
	if err != nil {
		return err
	}
	source, err := os.ReadFile(options.source)
	if err != nil {
		return err
	}
	independence, err := freshness.LoadIndependence(options.independence)
	if err != nil {
		return err
	}
	report := freshness.Evaluate(freshness.Input{Contract: contract, HeadSHA: options.head, Source: source, Independence: independence})
	if err := freshness.WriteJSON(options.output, report); err != nil {
		return err
	}
	if options.emitDir != "" {
		if err := os.MkdirAll(options.emitDir, 0o755); err != nil {
			return err
		}
		if err := freshness.WriteJSON(filepath.Join(options.emitDir, "receipt.json"), report.Receipt); err != nil {
			return err
		}
		if len(report.Cases) == 0 {
			return fmt.Errorf("freshness report did not emit cases")
		}
		if err := freshness.WriteJSON(filepath.Join(options.emitDir, "fresh-context.json"), report.Cases[0].Context); err != nil {
			return err
		}
	}
	if err := freshness.Validate(report); err != nil {
		return err
	}
	fmt.Printf("evidence freshness: %s %d/%d (fresh=%d stale=%d unknown=%d)\n",
		report.Decision, report.Summary.CasesSatisfied, report.Summary.CasesTotal,
		report.Summary.FreshCases, report.Summary.StaleCases, report.Summary.UnknownCases)
	return nil
}
