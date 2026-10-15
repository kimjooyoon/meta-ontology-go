package originpolicy

//gooo:generated:start id="gooo://self-improvement/transport/bind-candidate-subject" kind="activity"
func BindCandidateSubject(candidate string, transport string, declared string) string {

	if candidate == "" || transport == "" || declared == "" {
		return "CANDIDATE_SUBJECT_UNKNOWN"
	}
	if candidate != transport || candidate != declared {
		return "CANDIDATE_SUBJECT_MISMATCH"
	}
	return "CANDIDATE_SUBJECT_BOUND"

}

//gooo:generated:end id="gooo://self-improvement/transport/bind-candidate-subject" kind="activity"
