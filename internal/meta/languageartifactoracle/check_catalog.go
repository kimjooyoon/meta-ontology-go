package languageartifactoracle

type checkSpec struct {
	id        string
	proof     string
	operation string
}

var fixedChecks = []checkSpec{
	{"receipt-digest", "REGRESSION", "verify-independent-artifact-digest"},
	{"source-digest", "FOUNDATION", "bind-artifact-to-source-bytes"},
	{"receipt-identity", "FOUNDATION", "bind-source-execution-contract"},
	{"entry-header", "FOUNDATION", "project-source-headers"},
	{"input-bindings", "COHERENCE", "project-activity-inputs"},
	{"output-binding", "COHERENCE", "project-activity-output"},
	{"event-sequence", "COHERENCE", "compare-independent-event-projection"},
	{"semantic-event-coherence", "COHERENCE", "bind-semantic-event-subject"},
	{"zero-effects", "REGRESSION", "deny-oracle-observed-effects"},
}

func unknownChecks() []CheckResult {
	result := make([]CheckResult, len(fixedChecks))
	for index, spec := range fixedChecks {
		result[index] = CheckResult{ID: spec.id, Status: "UNKNOWN", ProofChoice: spec.proof,
			MetaOperation: spec.operation, Expected: "true", Observed: "unknown"}
	}
	return result
}
