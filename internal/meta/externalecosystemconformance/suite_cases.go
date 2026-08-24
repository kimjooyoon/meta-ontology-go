package externalecosystemconformance

func RunCase(subject, caseID string, source Capsule, sourceEvidence Evidence) Report {
	capsule := cloneCapsule(source)
	evidence := cloneEvidence(sourceEvidence)
	switch caseID {
	case "exact":
	case "readme-unavailable":
		evidence.Readme = nil
	case "gomod-unavailable":
		evidence.GoMod = nil
	case "unknown-relation":
		capsule.Capabilities[0].Relation = "UNRECOGNIZED"
	case "readme-digest-mismatch":
		evidence.Readme = append(evidence.Readme, 0)
	case "gomod-digest-mismatch":
		evidence.GoMod = append(evidence.GoMod, 0)
	case "commit-mismatch":
		capsule.CommitSHA = "0000000000000000000000000000000000000000"
	case "license-mismatch":
		capsule.LicenseSPDX = "UNKNOWN"
	case "external-execution":
		evidence.ExternalExecutions = 1
	case "observed-write":
		evidence.RepositoryWrites = 1
	default:
		return fail(baseReport(subject, capsule.ReferenceID, evidence), ResolutionUnknown, ReasonCaseUnknown)
	}
	return Evaluate(subject, capsule, evidence)
}

func expectedCase(caseID string) (string, string) {
	switch caseID {
	case "exact":
		return DecisionReferenceBound, ResolutionExact
	case "readme-unavailable", "gomod-unavailable", "unknown-relation":
		return DecisionFailClosed, ResolutionUnknown
	default:
		return DecisionFailClosed, ResolutionInvariant
	}
}
