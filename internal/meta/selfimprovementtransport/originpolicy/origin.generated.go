package originpolicy

//gooo:generated:start id="gooo://self-improvement/transport/bind-producer-run" kind="activity"
func BindProducerRun(event string, subject string, workflow string, run string) string {

	if event == "" {
		return "PRODUCER_EVENT_UNKNOWN"
	}
	if run == "" {
		return "PRODUCER_RUN_ORIGIN_UNKNOWN"
	}
	if event == "workflow_run" {
		if workflow == "" {
			return "PRODUCER_WORKFLOW_ORIGIN_UNKNOWN"
		}
		if workflow != run {
			return "PRODUCER_WORKFLOW_RUN_MISMATCH"
		}
		return "WORKFLOW_ORIGIN_BOUND"
	}
	if event == "pull_request" || event == "push" {
		if subject == "" {
			return "PRODUCER_SUBJECT_ORIGIN_UNKNOWN"
		}
		if subject != run {
			return "PRODUCER_SUBJECT_RUN_MISMATCH"
		}
		return "SUBJECT_ORIGIN_BOUND"
	}
	return "PRODUCER_EVENT_UNSUPPORTED"

}

//gooo:generated:end id="gooo://self-improvement/transport/bind-producer-run" kind="activity"
