package originpolicy

//gooo:generated:start id="gooo://self-improvement/transport/continuation-artifact-name" kind="activity"
func ContinuationArtifactName(target string, source string, subject string) string {

	if target == ".github/workflows/self-improvement-execution-contract.yml" {
		return "self-improvement-candidate-authorization-" + source
	}
	if target == ".github/workflows/self-improvement-execution-grant.yml" {
		return "self-improvement-execution-contract-" + subject
	}
	return ""

}

//gooo:generated:end id="gooo://self-improvement/transport/continuation-artifact-name" kind="activity"
