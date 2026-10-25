package sourcecompletion

var completionCases = []struct {
	name                               string
	counts                             [4]int64
	reason, decision, resolution, next string
}{
	{"empty", [4]int64{}, "NO_PROJECTABLE_SUBJECTS", "FIXED_POINT", "EXACT", "none"},
	{"applied", [4]int64{2, 2, 0, 0}, "PROJECTABLE_SUBJECTS_SPLIT", "PASS", "EXACT", "none"},
	{"pending without selected paths", [4]int64{0, 0, 1, 0},
		"SUCCESSOR_PROJECTION_REQUIRED", "PROGRESS", "LOGICAL_ONLY", "project-generated-paths"},
	{"pending after application", [4]int64{2, 2, 3, 0},
		"SUCCESSOR_PROJECTION_REQUIRED", "PROGRESS", "LOGICAL_ONLY", "project-generated-paths"},
	{"unhandled without selected paths", [4]int64{0, 0, 0, 2},
		"SOURCE_OBLIGATIONS_UNHANDLED", "PROGRESS", "UNRESOLVED_SUBJECTS", "resolve-unhandled-source-subjects"},
	{"unhandled after application", [4]int64{2, 2, 0, 1},
		"SOURCE_OBLIGATIONS_UNHANDLED", "PROGRESS", "UNRESOLVED_SUBJECTS", "resolve-unhandled-source-subjects"},
	{"pending with other obligations", [4]int64{2, 2, 1, 1},
		"SUCCESSOR_PROJECTION_REQUIRED", "PROGRESS", "LOGICAL_ONLY", "project-generated-paths"},
	{"partial", [4]int64{2, 1, 0, 0}, "APPLICATION_INCOMPLETE", "PROGRESS", "PARTIAL", "complete-logical-split"},
	{"partial with other obligations", [4]int64{2, 1, 2, 1},
		"APPLICATION_INCOMPLETE", "PROGRESS", "PARTIAL", "complete-logical-split"},
	{"extra application", [4]int64{1, 2, 0, 0},
		"APPLICATION_COUNT_REFUTED", "FAIL", "REFUTED", "report-counterexample"},
	{"negative selected", [4]int64{-1, 0, 0, 0}, "INVALID_SPLIT_COUNTS", "FAIL", "REFUTED", "report-counterexample"},
	{"negative applied", [4]int64{1, -1, 0, 0}, "INVALID_SPLIT_COUNTS", "FAIL", "REFUTED", "report-counterexample"},
	{"negative deferred", [4]int64{1, 1, -1, 0}, "INVALID_SPLIT_COUNTS", "FAIL", "REFUTED", "report-counterexample"},
	{"negative unhandled", [4]int64{1, 1, 0, -1}, "INVALID_SPLIT_COUNTS", "FAIL", "REFUTED", "report-counterexample"},
}
