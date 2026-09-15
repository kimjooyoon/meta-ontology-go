package languagereadiness

func artifactResolutionReason(artifact conceptArtifact) string {
	switch {
	case artifact.Schema != conceptArtifactSchema:
		return "CONCEPT_ARTIFACT_SCHEMA_UNKNOWN"
	case artifact.Decision != "PASS":
		return "CONCEPT_ARTIFACT_DECISION_NOT_PASS"
	case artifact.CatalogDigest == "" || artifact.ArtifactDigest == "":
		return "CONCEPT_ARTIFACT_DIGEST_MISSING"
	case !artifact.ReplayEqual:
		return "CONCEPT_ARTIFACT_REPLAY_MISMATCH"
	case artifact.Report.ReportDigest == "" || artifact.Report.ReportDigest != artifact.ReplayReportDigest:
		return "CONCEPT_REPORT_DIGEST_MISMATCH"
	case artifact.Bindings.Missing != 0 || artifact.Bindings.Unsupported != 0:
		return "CONCEPT_BINDING_EVIDENCE_INCOMPLETE"
	case artifact.RepositoryWrites != 0:
		return "CONCEPT_OBSERVER_WRITE_DETECTED"
	default:
		return ""
	}
}

func summarize(snapshot Snapshot) Snapshot {
	snapshot.Summary.Total = len(snapshot.Obligations)
	for _, result := range snapshot.Obligations {
		switch result.Status {
		case "SATISFIED":
			snapshot.Summary.Completed++
		case "NOT_SATISFIED":
			snapshot.Summary.NotSatisfied++
		default:
			snapshot.Summary.Unresolved++
		}
	}
	snapshot.Summary.RatioNumerator = snapshot.Summary.Completed
	snapshot.Summary.RatioDenominator = snapshot.Summary.Total
	if snapshot.Summary.Total > 0 {
		snapshot.Summary.ReadinessBPS = snapshot.Summary.Completed * 10000 / snapshot.Summary.Total
	}
	snapshot.Indicators = buildIndicators(snapshot)
	if snapshot.Summary.Unresolved > 0 {
		snapshot.Decision, snapshot.Reason = "LOWER_RESOLUTION", "READINESS_EVIDENCE_UNRESOLVED"
	} else {
		snapshot.Decision, snapshot.Reason = "PASS", "READINESS_EXACTLY_COUNTED"
	}
	return finalize(snapshot)
}
