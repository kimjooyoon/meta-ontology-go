package predecessorresolution

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

type attemptKind int

const (
	attemptBlocked attemptKind = iota
	attemptMissing
	attemptSelected
)

func classifyAttempt(input Input, attempt Attempt) (attemptKind, error) {
	report := attempt.Selection.Report
	if report.Repository != input.Repository ||
		report.CurrentHeadSHA != input.CurrentHeadSHA ||
		report.PredecessorSHA != attempt.AncestorSHA {
		return attemptBlocked, fmt.Errorf("selection identity mismatch at depth %d", attempt.Depth)
	}
	if err := validateSelectionReport(report); err != nil {
		return attemptBlocked, err
	}
	if exactMissing(report) {
		return attemptMissing, nil
	}
	if exactSelected(report) {
		if len(attempt.Selection.BaselineRaw) == 0 ||
			len(attempt.Selection.BindingRaw) == 0 {
			return attemptBlocked, fmt.Errorf("selected payload missing")
		}
		return attemptSelected, nil
	}
	return attemptBlocked, nil
}

func exactMissing(report predecessorselection.Report) bool {
	summary := report.Summary
	return report.Decision == predecessorselection.DecisionFailClosed &&
		report.Reason == predecessorselection.ReasonNotFound &&
		report.Selected == nil && summary.ObservedCandidates == 0 &&
		summary.ExactHeadCandidates == 0 && summary.CanonicalCandidates == 0 &&
		summary.SuccessfulCandidates == 0 && summary.AvailableCandidates == 0 &&
		summary.ProducerConformantCandidates == 0 &&
		summary.ValidCandidates == 0 && summary.AmbiguousCandidates == 0 &&
		summary.RepositoryWrites == 0
}
