package selfimprovementobservation

func failureClassification(check validation, source SourceReport) (string, string) {
	if !check.SourceSchema {
		return "EXACT", "SOURCE_SCHEMA_REJECTED"
	}
	if !check.ExactHead {
		return "EXACT", "SOURCE_HEAD_MISMATCH"
	}
	if !check.SourceDigest {
		return "EXACT", "SOURCE_REPORT_DIGEST_INVALID"
	}
	if source.Decision == "FAIL_CLOSED" {
		if source.Resolution == "LOWER_RESOLUTION" {
			return "LOWER_RESOLUTION", "SOURCE_RESOLUTION_LOWERED"
		}
		return "EXACT", "SOURCE_EXPLICITLY_REJECTED"
	}
	if source.Decision != "PASS" {
		return "LOWER_RESOLUTION", "SOURCE_DECISION_UNKNOWN"
	}
	if source.Resolution != "EXACT" {
		return "LOWER_RESOLUTION", "SOURCE_RESOLUTION_UNKNOWN"
	}
	if !check.Contract {
		return "EXACT", "GOOO_OBSERVATION_CONTRACT_REJECTED"
	}
	if !check.SourceEffects {
		return "EXACT", "SOURCE_EFFECTS_REJECTED"
	}
	if !check.Counterexamples {
		return "EXACT", "COUNTEREXAMPLE_COVERAGE_REJECTED"
	}
	return "EXACT", "READ_ONLY_OBSERVATION_REJECTED"
}
