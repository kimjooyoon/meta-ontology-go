package fullsoundness

type Output struct {
	SchemaVersion                        string          `json:"schema_version"`
	SnapshotDigest                       string          `json:"snapshot_digest"`
	PolicyDigest                         string          `json:"policy_digest"`
	RegistryDigest                       string          `json:"registry_digest"`
	SelectionDigest                      string          `json:"selection_digest"`
	CommandCount                         uint64          `json:"command_count"`
	SelectedCommandCount                 uint64          `json:"selected_command_count"`
	ObligationCount                      uint64          `json:"obligation_count"`
	AuthoritativeImpactedObligationCount uint64          `json:"authoritative_impacted_obligation_count"`
	SemanticEvaluated                    bool            `json:"semantic_evaluated"`
	FullFailureCommandIDs                []string        `json:"full_failure_command_ids"`
	SelectedFailureCommandIDs            []string        `json:"selected_failure_command_ids"`
	OmittedCommandIDs                    []string        `json:"omitted_command_ids"`
	ResourceVector                       *ResourceVector `json:"resource_vector"`
	Decision                             Decision        `json:"decision"`
	Reason                               Reason          `json:"reason"`
	ExecutionAuthorized                  bool            `json:"execution_authorized"`
	CIAuthorized                         bool            `json:"ci_authorized"`
	DecisionDigest                       string          `json:"decision_digest"`
	CanonicalDigest                      string          `json:"canonical_digest"`
}
