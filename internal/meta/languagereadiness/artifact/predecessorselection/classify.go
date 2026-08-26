package predecessorselection

func canonical(input Input, candidate Candidate) bool {
	return candidate.Workflow == input.Workflow && candidate.HeadBranch == input.Branch &&
		candidate.Event == "workflow_run" &&
		candidate.ReadinessArtifactName == "language-readiness-artifact-"+input.PredecessorSHA &&
		candidate.BindingArtifactName == "language-readiness-predecessor-binding-"+input.PredecessorSHA
}

func failureReason(summary Summary) string {
	switch {
	case summary.RepositoryWrites != 0:
		return ReasonWriteEffect
	case summary.ExactHeadCandidates == 0:
		return ReasonNotFound
	case summary.CanonicalCandidates == 0:
		return ReasonUnbound
	case summary.ProducerConformantCandidates == 0:
		return ReasonProducer
	case summary.AvailableCandidates == 0:
		return ReasonExpired
	case summary.ValidCandidates == 0:
		return ReasonInvalid
	case summary.AmbiguousCandidates != 0:
		return ReasonAmbiguous
	default:
		return ""
	}
}

func proofs(report Report) []Proof {
	return []Proof{
		{ID: "exact-predecessor-head", Choice: "FOUNDATION", Passed: report.Summary.ExactHeadCandidates > 0},
		{ID: "metric-scoped-producer", Choice: "COHERENCE", Passed: report.Summary.ProducerConformantCandidates == 1},
		{ID: "canonical-artifact-pair", Choice: "COHERENCE", Passed: report.Summary.ValidCandidates == 1},
		{ID: "unambiguous-selection", Choice: "REGRESSION", Passed: report.Summary.AmbiguousCandidates == 0},
		{ID: "read-only-selection", Choice: "FOUNDATION", Passed: report.Summary.RepositoryWrites == 0},
	}
}
