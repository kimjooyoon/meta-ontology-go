package predecessorresolution

import (
	"strings"

	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func fixtureSHA(value string) string { return strings.Repeat(value, 40) }

func missingSelection(current, ancestor string) predecessorselection.Result {
	report := predecessorselection.Report{Schema: predecessorselection.Schema,
		Repository: "owner/repo", CurrentHeadSHA: current,
		PredecessorSHA: ancestor, Decision: predecessorselection.DecisionFailClosed,
		Reason: predecessorselection.ReasonNotFound,
		Proofs: []predecessorselection.Proof{
			{ID: "exact-predecessor-head", Choice: "FOUNDATION", Passed: false},
			{ID: "canonical-artifact-pair", Choice: "COHERENCE", Passed: false},
			{ID: "unambiguous-selection", Choice: "REGRESSION", Passed: true},
			{ID: "read-only-selection", Choice: "FOUNDATION", Passed: true},
		}}
	report.ReportDigest = digestJSON(report)
	return predecessorselection.Result{Report: report}
}

func selectedSelection(current, ancestor string) predecessorselection.Result {
	baseline := readinessartifact.BaselineReference{Completed: 10, Total: 24,
		BasisPoints: 4166}
	report := predecessorselection.Report{Schema: predecessorselection.Schema,
		Repository: "owner/repo", CurrentHeadSHA: current,
		PredecessorSHA: ancestor, Decision: predecessorselection.DecisionSelected,
		Reason: predecessorselection.ReasonSelected,
		Selected: &predecessorselection.Selection{RunID: 1, RunAttempt: 1,
			ReadinessArtifactID: 2, BindingArtifactID: 3, Baseline: baseline},
		Summary: predecessorselection.Summary{ObservedCandidates: 1,
			ExactHeadCandidates: 1, CanonicalCandidates: 1,
			SuccessfulCandidates: 1, AvailableCandidates: 1, ValidCandidates: 1},
		Proofs: []predecessorselection.Proof{
			{ID: "exact-predecessor-head", Choice: "FOUNDATION", Passed: true},
			{ID: "canonical-artifact-pair", Choice: "COHERENCE", Passed: true},
			{ID: "unambiguous-selection", Choice: "REGRESSION", Passed: true},
			{ID: "read-only-selection", Choice: "FOUNDATION", Passed: true},
		}}
	report.ReportDigest = digestJSON(report)
	return predecessorselection.Result{Report: report,
		BaselineRaw: []byte("baseline"), BindingRaw: []byte("binding")}
}
