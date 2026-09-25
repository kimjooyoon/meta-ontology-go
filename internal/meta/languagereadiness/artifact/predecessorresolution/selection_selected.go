package predecessorresolution

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"

func exactSelected(report predecessorselection.Report) bool {
	summary := report.Summary
	workflowSuccess := 0
	if report.Selected != nil && report.Selected.WorkflowConclusion == "success" {
		workflowSuccess = 1
	}
	return report.Decision == predecessorselection.DecisionSelected &&
		report.Reason == predecessorselection.ReasonSelected &&
		report.Selected != nil && summary.ObservedCandidates == 1 &&
		summary.ExactHeadCandidates == 1 && summary.CanonicalCandidates == 1 &&
		summary.SuccessfulCandidates == workflowSuccess &&
		summary.ProducerConformantCandidates == 1 && summary.AvailableCandidates == 1 &&
		summary.ValidCandidates == 1 && summary.AmbiguousCandidates == 0 &&
		summary.RepositoryWrites == 0 && report.Selected.ProducerJobID > 0 &&
		report.Selected.ProducerJobRunAttempt == report.Selected.RunAttempt &&
		report.Selected.ProducerJobName == predecessorselection.ProducerJobName &&
		report.Selected.ProducerJobConclusion == "success" &&
		allSelectionProofsPassed(report)
}

func allSelectionProofsPassed(report predecessorselection.Report) bool {
	if len(report.Proofs) != 5 {
		return false
	}
	for _, proof := range report.Proofs {
		if !proof.Passed {
			return false
		}
	}
	return true
}
