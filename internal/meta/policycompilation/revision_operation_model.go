package policycompilation

const PolicyRevisionOperationSchema = "gooo/meta-policy-revision-operation/v1"
const PolicyRevisionOperationProgram = "policy.revision.observe:v1"

// PolicyRevisionOperationBinding describes source/IR/native ABI agreement only.
// It does not authorize repository mutation or adoption of the observed policy.
type PolicyRevisionOperationBinding struct {
	ContractSourceDigest       string `json:"contract_source_digest"`
	ContractSemanticDigest     string `json:"contract_semantic_digest"`
	NativeContractSourceDigest string `json:"native_contract_source_digest"`
	ActivityID                 string `json:"activity_id"`
	Program                    string `json:"program"`
	PolicySourceEntityID       string `json:"policy_source_entity_id"`
	RequestEntityID            string `json:"request_entity_id"`
	ObservationEntityID        string `json:"observation_entity_id"`
	UsedPolicySource           bool   `json:"used_policy_source"`
	UsedRevisionRequest        bool   `json:"used_revision_request"`
	GeneratedObservation       bool   `json:"generated_observation"`
}

// NativeWorkerInvocations counts calls to the bounded native API, not child
// processes or successful results. The existing observation owns those facts.
type PolicyRevisionOperationObservation struct {
	Schema                  string                         `json:"schema"`
	Binding                 PolicyRevisionOperationBinding `json:"binding"`
	PolicySourceDigest      string                         `json:"policy_source_digest"`
	RequestArtifactDigest   string                         `json:"request_artifact_digest"`
	NativeWorkerInvocations int                            `json:"native_worker_invocations"`
	Observation             *PolicyRevisionObservation     `json:"observation,omitempty"`
	MutationAuthority       int                            `json:"mutation_authority"`
	PromotionAuthority      int                            `json:"promotion_authority"`
}
