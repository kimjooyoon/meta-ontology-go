package languageartifactoracle

func evaluateArtifact(source, raw []byte, filename, entry string) oracleResult {
	sourceDigest := digestBytes(source)
	artifactEvidence := digestBytes(raw)
	artifact, err := decodeArtifact(raw)
	if err != nil {
		return oracleFailure("LOWER_RESOLUTION", "ARTIFACT_DECODE_UNKNOWN", "ORACLE_DECODE",
			"decode-artifact", unknownChecks(), sourceDigest, artifactEvidence)
	}
	want, err := projectSource(source, entry)
	if err != nil {
		return oracleFailure("LOWER_RESOLUTION", "ORACLE_SOURCE_PROJECTION_UNKNOWN", "ORACLE_PARSE",
			"project-source", unknownChecks(), sourceDigest, artifactEvidence)
	}
	checks := compareArtifact(want, artifact, filename, sourceDigest)
	if artifact.Decision != "PASS" {
		return oracleFailure("LOWER_RESOLUTION", "ARTIFACT_DECISION_UNKNOWN", "ORACLE_COMPARE",
			"receipt-identity", checks, sourceDigest, artifactEvidence)
	}
	if failed := firstFailedCheck(checks); failed != "" {
		reason := "ARTIFACT_SOURCE_PROJECTION_MISMATCH"
		if failed == "receipt-digest" { reason = "ARTIFACT_DIGEST_MISMATCH" }
		return oracleFailure("INVARIANT_ONLY", reason, "ORACLE_COMPARE", failed,
			checks, sourceDigest, artifactEvidence)
	}
	return oracleResult{Decision: "PASS", Resolution: "EXACT", Reason: "ARTIFACT_SOURCE_PROJECTION_EXACT",
		Coordinate: Coordinate{Stage: "ORACLE_COMPARE", Step: "complete", Reason: "ARTIFACT_SOURCE_PROJECTION_EXACT"},
		Checks: checks, SourceDigest: sourceDigest, ArtifactDigest: artifactEvidence}
}

func oracleFailure(resolution, reason, stage, step string, checks []CheckResult,
	sourceDigest, artifactDigest string) oracleResult {
	return oracleResult{Decision: "FAIL_CLOSED", Resolution: resolution, Reason: reason,
		Coordinate: Coordinate{Stage: stage, Step: step, Reason: reason}, Checks: checks,
		SourceDigest: sourceDigest, ArtifactDigest: artifactDigest}
}
