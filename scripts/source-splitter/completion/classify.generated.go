package completion

//gooo:generated:start id="gooo://source-split/completion/classify" kind="activity"
func Classify(selected int64, applied int64, deferred int64, unhandled int64) string {

	if selected < 0 || applied < 0 || deferred < 0 || unhandled < 0 {
		return "INVALID_SPLIT_COUNTS"
	}
	if applied > selected {
		return "APPLICATION_COUNT_REFUTED"
	}
	if applied < selected {
		return "APPLICATION_INCOMPLETE"
	}
	if deferred > 0 {
		return "SUCCESSOR_PROJECTION_REQUIRED"
	}
	if unhandled > 0 {
		return "SOURCE_OBLIGATIONS_UNHANDLED"
	}
	if selected == 0 {
		return "NO_PROJECTABLE_SUBJECTS"
	}
	return "PROJECTABLE_SUBJECTS_SPLIT"

}

//gooo:generated:end id="gooo://source-split/completion/classify" kind="activity"
