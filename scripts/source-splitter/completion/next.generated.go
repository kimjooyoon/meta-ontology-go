package completion

//gooo:generated:start id="gooo://source-split/completion/next-operation" kind="activity"
func NextOperation(reason string) string {

	if reason == "INVALID_SPLIT_COUNTS" || reason == "APPLICATION_COUNT_REFUTED" {
		return "report-counterexample"
	}
	if reason == "APPLICATION_INCOMPLETE" {
		return "complete-logical-split"
	}
	if reason == "SUCCESSOR_PROJECTION_REQUIRED" {
		return "project-generated-paths"
	}
	if reason == "SOURCE_OBLIGATIONS_UNHANDLED" {
		return "resolve-unhandled-source-subjects"
	}
	if reason == "NO_PROJECTABLE_SUBJECTS" || reason == "PROJECTABLE_SUBJECTS_SPLIT" {
		return "none"
	}
	return "restore-split-observation"

}

//gooo:generated:end id="gooo://source-split/completion/next-operation" kind="activity"
