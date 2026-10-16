package policy

//gooo:generated:start id="gooo://self-improvement/consumer/bind-subjects" kind="activity"
func BindSubjects(declared string, grant string, contract string) string {

	if declared != "" && grant != "" && declared != grant {
		return "SUBJECT_REFUTED"
	}
	if declared != "" && contract != "" && declared != contract {
		return "SUBJECT_REFUTED"
	}
	if grant != "" && contract != "" && grant != contract {
		return "SUBJECT_REFUTED"
	}
	if declared == "" || grant == "" || contract == "" {
		return "SUBJECT_UNKNOWN"
	}
	return "SUBJECT_BOUND"

}

//gooo:generated:end id="gooo://self-improvement/consumer/bind-subjects" kind="activity"
