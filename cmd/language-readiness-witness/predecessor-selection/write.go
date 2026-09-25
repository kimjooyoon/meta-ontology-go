package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func writeResult(cfg config, result predecessorselection.Result) error {
	reportRaw, err := encodeJSON(result.Report)
	if err != nil {
		return err
	}
	referenceRaw, err := encodeJSON(result.Report.Selected.Baseline)
	if err != nil {
		return err
	}
	outputs := []struct {
		path string
		raw  []byte
	}{
		{cfg.baseline, result.BaselineRaw},
		{cfg.reference, referenceRaw},
		{cfg.bindingBaseline, result.BindingRaw},
		{cfg.receipt, reportRaw},
	}
	for _, output := range outputs {
		if err := writeExclusive(output.path, output.raw); err != nil {
			return err
		}
	}
	selected := result.Report.Selected
	fmt.Printf("readiness-predecessor: current=%s predecessor=%s run=%d attempt=%d "+
		"completed=%d total=%d writes=%d digest=%s\n", result.Report.CurrentHeadSHA,
		result.Report.PredecessorSHA, selected.RunID, selected.RunAttempt,
		selected.Baseline.Completed, selected.Baseline.Total,
		result.Report.Summary.RepositoryWrites, result.Report.ReportDigest)
	return nil
}
