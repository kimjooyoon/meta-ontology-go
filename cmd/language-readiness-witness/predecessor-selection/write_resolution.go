package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorresolution"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func writeResolutionResult(cfg config, result predecessorselection.Result,
	report predecessorresolution.Report) error {
	reportRaw, err := encodeJSON(report)
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
	selected := report.Selected
	fmt.Printf("readiness-ancestor: current=%s immediate=%s selected=%s depth=%d "+
		"missing=%d completed=%d total=%d writes=%d coordinates=%d/%d digest=%s\n",
		report.CurrentHeadSHA, report.ImmediatePredecessorSHA, selected.AncestorSHA,
		selected.Depth, report.Summary.MissingAttempts, selected.Baseline.Completed,
		selected.Baseline.Total, report.Summary.RepositoryWrites,
		report.Summary.CoordinatesCompleted, report.Summary.CoordinatesTotal,
		report.ReportDigest)
	return nil
}

func writeResolutionFailure(path string, report predecessorresolution.Report) error {
	reportRaw, err := encodeJSON(report)
	if err != nil {
		return err
	}
	return writeExclusive(path, reportRaw)
}
