package languagedelivery

func baseObservation(source SourceName, entry ManifestEntry, schema, decision, resolution string) SourceObservation {
	return SourceObservation{Source: source, Schema: schema, Decision: decision,
		Resolution: resolution, ArtifactID: entry.ArtifactID, ArtifactName: entry.ArtifactName,
		ArchiveDigest: entry.ArchiveDigest, ReportDigest: entry.ReportDigest}
}

func unknownObservation(source SourceName, entry ManifestEntry, reason string) SourceObservation {
	observation := baseObservation(source, entry, "UNKNOWN", "UNKNOWN", "LOWER_RESOLUTION")
	observation.State = "UNKNOWN"
	observation.Reason = reason
	return observation
}

func finalizeObservation(observation SourceObservation, schema, wantSchema string) SourceObservation {
	if schema != wantSchema {
		observation.State, observation.Reason = "UNKNOWN", "SOURCE_SCHEMA_UNKNOWN"
		return observation
	}
	if observation.RepositoryWrites != 0 || observation.MutationAuthority {
		observation.State, observation.Reason = "FAIL", "SOURCE_EFFECT_BOUNDARY_VIOLATED"
		return observation
	}
	if observation.Decision == "PASS" && (observation.Resolution == "EXACT" || observation.Resolution == "") {
		observation.State, observation.Reason = "PASS", "SOURCE_RECEIPT_EXACT"
		return observation
	}
	if observation.Decision == "FAIL_CLOSED" {
		observation.State, observation.Reason = "FAIL", "SOURCE_RECEIPT_FAIL_CLOSED"
		return observation
	}
	observation.State, observation.Reason = "UNKNOWN", "SOURCE_DECISION_UNKNOWN"
	return observation
}

func headUnknown(observation SourceObservation) SourceObservation {
	observation.State, observation.Reason = "UNKNOWN", "SOURCE_HEAD_UNKNOWN"
	return observation
}
