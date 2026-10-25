package completion

//gooo:generated:start id="gooo://source-split/completion/resolution" kind="activity"
func Resolution(reason string) string {

	if reason == "INVALID_SPLIT_COUNTS" || reason == "APPLICATION_COUNT_REFUTED" {
		return "REFUTED"
	}
	if reason == "APPLICATION_INCOMPLETE" {
		return "PARTIAL"
	}
	if reason == "SUCCESSOR_PROJECTION_REQUIRED" {
		return "LOGICAL_ONLY"
	}
	if reason == "SOURCE_OBLIGATIONS_UNHANDLED" {
		return "UNRESOLVED_SUBJECTS"
	}
	if reason == "NO_PROJECTABLE_SUBJECTS" || reason == "PROJECTABLE_SUBJECTS_SPLIT" {
		return "EXACT"
	}
	return "UNKNOWN"

}

//gooo:generated:end id="gooo://source-split/completion/resolution" kind="activity"
