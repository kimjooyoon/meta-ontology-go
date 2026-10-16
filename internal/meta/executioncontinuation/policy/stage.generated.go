package policy

//gooo:generated:start id="gooo://self-improvement/consumer/observe-stage" kind="activity"
func ObserveStage(dispatched bool, observed bool, executed bool) string {

	if executed && !observed {
		return "CONTINUATION_REFUTED"
	}
	if observed && !dispatched {
		return "CONTINUATION_REFUTED"
	}
	if executed {
		return "EXECUTION_OBSERVED"
	}
	if observed {
		return "TARGET_RUN_OBSERVED"
	}
	if dispatched {
		return "DISPATCH_SENT"
	}
	return "DISPATCH_UNKNOWN"

}

//gooo:generated:end id="gooo://self-improvement/consumer/observe-stage" kind="activity"
