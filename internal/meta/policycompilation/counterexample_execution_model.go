package policycompilation

const PolicyCounterexampleExecutionSchema = "gooo/meta-policy-counterexample-execution/v1"

// Callers supply observations, never the condition/from/to derived by Gooo.
type PolicyCounterexampleExecutionInput struct {
	Counterexample PolicyRevisionCounterexample `json:"counterexample"`
	Cases          []PolicyRevisionCasePair     `json:"cases"`
}

// Proposal and execution are separate stages; neither grants adoption.
type PolicyCounterexampleExecution struct {
	Schema              string                              `json:"schema"`
	InputArtifactDigest string                              `json:"input_artifact_digest"`
	Proposal            PolicyCounterexampleProposal        `json:"proposal"`
	RevisionRequestJSON string                              `json:"revision_request_json,omitempty"`
	Operation           *PolicyRevisionOperationObservation `json:"operation,omitempty"`
	Pending             *PolicyRevisionPending              `json:"pending,omitempty"`
}
