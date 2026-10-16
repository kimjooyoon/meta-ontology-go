package policy

//gooo:generated:start id="gooo://self-improvement/consumer/artifact-label" kind="activity"
func ArtifactLabel(kind string, subject string) string {

	if kind == "grant" {
		return "self-improvement-execution-grant-" + subject
	}
	if kind == "contract" {
		return "self-improvement-execution-contract-" + subject
	}
	return ""

}

//gooo:generated:end id="gooo://self-improvement/consumer/artifact-label" kind="activity"
