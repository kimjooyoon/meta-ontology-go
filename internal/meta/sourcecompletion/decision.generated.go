package sourcecompletion

//gooo:generated:start id="gooo://source-split/completion/decision" kind="activity"
func Decision(reason string) string {

	if reason == "INVALID_SPLIT_COUNTS" || reason == "APPLICATION_COUNT_REFUTED" {
		return "FAIL"
	}
	if reason == "APPLICATION_INCOMPLETE" || reason == "SUCCESSOR_PROJECTION_REQUIRED" ||
		reason == "SOURCE_OBLIGATIONS_UNHANDLED" {
		return "PROGRESS"
	}
	if reason == "NO_PROJECTABLE_SUBJECTS" {
		return "FIXED_POINT"
	}
	if reason == "PROJECTABLE_SUBJECTS_SPLIT" {
		return "PASS"
	}
	return "UNKNOWN"

}

//gooo:generated:end id="gooo://source-split/completion/decision" kind="activity"
