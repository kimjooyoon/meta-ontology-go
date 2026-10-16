package policy

//gooo:generated:start id="gooo://self-improvement/consumer/workflow-path" kind="activity"
func WorkflowPath(kind string) string {

	if kind == "grant" {
		return ".github/workflows/self-improvement-execution-grant.yml"
	}
	if kind == "contract" {
		return ".github/workflows/self-improvement-execution-contract.yml"
	}
	if kind == "consumer" {
		return ".github/workflows/self-improvement-execution-consumer.yml"
	}
	return ""

}

//gooo:generated:end id="gooo://self-improvement/consumer/workflow-path" kind="activity"
