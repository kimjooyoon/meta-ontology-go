package policy

//gooo:generated:start id="gooo://self-improvement/consumer/next-observation" kind="activity"
func NextObservation(dependency string) string {

	if dependency == "producer_run" {
		return "observe_exact_source_run"
	}
	if dependency == "subject_blob" {
		return "observe_exact_subject_blob"
	}
	if dependency == "policy_source" {
		return "restore_policy_source"
	}
	if dependency == "expired_source_artifact" {
		return "produce_fresh_source_artifact"
	}
	if dependency == "source_artifact" {
		return "observe_exact_source_artifact"
	}
	if dependency == "dispatch_outcome" {
		return "observe_dispatch_outcome"
	}
	if dependency == "target_run_reference" {
		return "observe_dispatch_response"
	}
	if dependency == "reservation_history" {
		return "restore_retained_reservation_history"
	}
	return "observe_returned_target_run"

}

//gooo:generated:end id="gooo://self-improvement/consumer/next-observation" kind="activity"
