package policy

//gooo:generated:start id="gooo://self-improvement/consumer/bind-target-run" kind="activity"
func BindTargetRun(parent int64, expected int64, observed int64) string {

	if parent > 0 && expected > 0 && parent == expected {
		return "TARGET_RUN_REFUTED"
	}
	if expected > 0 && observed > 0 && expected != observed {
		return "TARGET_RUN_REFUTED"
	}
	if parent <= 0 || expected <= 0 || observed <= 0 {
		return "TARGET_RUN_UNKNOWN"
	}
	return "TARGET_RUN_BOUND"

}

//gooo:generated:end id="gooo://self-improvement/consumer/bind-target-run" kind="activity"
