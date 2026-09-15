package externalcapabilityexecution

func RunCase(subject, caseID string) Report {
	observation := cloneObservation(exactObservation(subject))
	switch caseID {
	case "exact":
	case "observation-unavailable":
		observation.Available = false
	case "run-status-unknown":
		observation.Runs[0].Status = "UNRECOGNIZED"
	case "parent-decision-unknown":
		observation.Parent.Decision = "UNRECOGNIZED"
	case "repository-mismatch":
		observation.Reference.RepositoryURL = "https://example.invalid/gomacro"
	case "commit-mismatch":
		observation.Reference.CommitSHA = "0000000000000000000000000000000000000000"
	case "tree-mismatch":
		observation.Reference.TreeSHA = "0000000000000000000000000000000000000000"
	case "toolchain-mismatch":
		observation.Reference.GoVersion = "go1.26.0"
	case "arithmetic-mismatch":
		observation.Runs[0].Arithmetic = "41"
	case "function-mismatch":
		observation.Runs[0].Function = "54"
	case "macro-mismatch":
		observation.Runs[0].MacroGeneratedSHA256 = "sha256:mismatch"
	case "replay-mismatch":
		observation.ReplayExact = false
	case "project-write":
		observation.RepositoryWrites = 1
	case "external-write":
		observation.ExternalRepositoryWrites = 1
	case "parent-promoted":
		observation.Parent.Decision = "EXECUTION_COMPATIBLE"
		observation.Parent.PromotionCount = 1
	default:
		observation.Available = false
	}
	return Evaluate(sealObservation(observation))
}

func expectedCase(caseID string) (string, string) {
	if caseID == "exact" {
		return DecisionExecutable, ResolutionExact
	}
	if caseID == "observation-unavailable" || caseID == "run-status-unknown" ||
		caseID == "parent-decision-unknown" {
		return DecisionFailClosed, ResolutionUnknown
	}
	return DecisionFailClosed, ResolutionInvariant
}
