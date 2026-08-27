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
	if options.contract == "" || options.source == "" || options.head == "" || options.independence == "" || options.writeSet == "" || options.output == "" {
		return fmt.Errorf("contract, source, head, independence, write-set, and output are required")
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
	writeSet, err := freshness.LoadWriteSet(options.writeSet)
	if err != nil {
		return err
	}
	report := freshness.Evaluate(freshness.Input{Contract: contract, HeadSHA: options.head, Source: source, Independence: independence, WriteSet: writeSet})
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
		for _, item := range report.Cases[1:] {
			if item.ID == "synthetic-comment-only" || item.ID == "synthetic-semantic-change" || item.ID == "synthetic-source-unavailable" {
				if err := freshness.WriteJSON(filepath.Join(options.emitDir, item.ID+"-context.json"), item.Context); err != nil {
					return err
				}
			}
		}
	}
	if err := freshness.Validate(report); err != nil {
		return err
	}
	fmt.Printf("evidence freshness: %s cases=%d current=%d synthetic=%d raw=%d/%d/%d semantic=%d/%d/%d\n",
		report.Decision, report.Summary.CasesObserved, report.Summary.CurrentEvidenceCases, report.Summary.SyntheticCounterexamples,
		report.Summary.RawFreshCases, report.Summary.RawStaleCases, report.Summary.RawUnknownCases,
		report.Summary.SemanticFreshCases, report.Summary.SemanticStaleCases, report.Summary.SemanticUnknownCases)
	return nil
}
