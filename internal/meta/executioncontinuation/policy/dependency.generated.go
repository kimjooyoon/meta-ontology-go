package policy

//gooo:generated:start id="gooo://self-improvement/consumer/missing-dependency" kind="activity"
func MissingDependency(reason string) string {

	if reason == "SOURCE_RUN_UNAVAILABLE" || reason == "SOURCE_RUN_COMPLETION_UNKNOWN" || reason == "SOURCE_RUN_METADATA_UNKNOWN" || reason == "SOURCE_GRANT_JOB_UNAVAILABLE" {
		return "producer_run"
	}
	if reason == "SUBJECT_SOURCE_UNKNOWN" {
		return "subject_blob"
	}
	if reason == "GRANT_POLICY_UNAVAILABLE" {
		return "policy_source"
	}
	if reason == "SOURCE_ARTIFACT_EXPIRED" {
		return "expired_source_artifact"
	}
	if reason == "SOURCE_ARTIFACT_MISSING" || reason == "SOURCE_ARCHIVE_UNAVAILABLE" || reason == "SOURCE_MEMBER_MISSING" || reason == "ARTIFACT_POPULATION_UNKNOWN" || reason == "ARTIFACT_POPULATION_INCOMPLETE" {
		return "source_artifact"
	}
	if reason == "DISPATCH_RESPONSE_UNKNOWN" {
		return "dispatch_outcome"
	}
	if reason == "TARGET_REFERENCE_UNKNOWN" {
		return "target_run_reference"
	}
	if len(reason) >= 12 && reason[:12] == "RESERVATION_" {
		return "reservation_history"
	}
	return "target_run"

}

//gooo:generated:end id="gooo://self-improvement/consumer/missing-dependency" kind="activity"
