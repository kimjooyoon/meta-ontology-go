package originpolicy

//gooo:generated:start id="gooo://self-improvement/transport/candidate-artifact-role" kind="activity"
func CandidateArtifactRole(name string) string {

	var prefix = "self-improvement-nonexecuting-candidate-"
	if len(name) == len(prefix)+40 && name[:len(prefix)] == prefix {
		return "gooo://self-improvement/non-executing-candidate"
	}
	return ""

}

//gooo:generated:end id="gooo://self-improvement/transport/candidate-artifact-role" kind="activity"
