package selfimprovementtransport

type ProducerReceipt struct {
	Schema        string           `json:"schema"`
	Contract      ContractEvidence `json:"contract"`
	RepositoryURI string           `json:"repository_uri"`
	SubjectSHA    string           `json:"subject_sha"`
	CheckoutSHA   string           `json:"checkout_sha"`
	WorkflowRef   string           `json:"workflow_ref"`
	WorkflowSHA   string           `json:"workflow_sha"`
	RunID         int64            `json:"run_id"`
	RunAttempt    int              `json:"run_attempt"`
	Job           string           `json:"job"`
	ArtifactName  string           `json:"artifact_name"`
	Subject       LogicalSubject   `json:"logical_subject"`
	Decision      string           `json:"decision"`
	Resolution    string           `json:"resolution"`
	Reason        string           `json:"reason"`
	Digest        string           `json:"digest"`
}

type Attestation struct {
	Status           string `json:"status,omitempty"`
	Digest           string `json:"digest,omitempty"`
	ProducerIdentity string `json:"producer_identity,omitempty"`
}

type TransportMetadata struct {
	Schema                    string      `json:"schema"`
	Repository                string      `json:"repository"`
	ProducerRunID             int64       `json:"producer_run_id"`
	ProducerRunAttempt        int         `json:"producer_run_attempt"`
	OrchestrationHeadSHA      string      `json:"orchestration_head_sha"`
	WorkflowPath              string      `json:"workflow_path"`
	ArtifactID                int64       `json:"artifact_id"`
	ArtifactName              string      `json:"artifact_name"`
	ArtifactDigest            string      `json:"artifact_digest"`
	ArtifactSizeBytes         int64       `json:"artifact_size_bytes"`
	ArtifactInstanceCount     int         `json:"artifact_instance_count"`
	ArtifactTypeCount         int         `json:"artifact_type_count"`
	ProducerDeclarationCount  int         `json:"producer_declaration_count"`
	ProducerDeclarationDigest string      `json:"producer_declaration_digest,omitempty"`
	ProducerSubjectSHA        string      `json:"producer_subject_sha,omitempty"`
	ProducerPayloadCount      int         `json:"producer_payload_count"`
	ProducerPayloadName       string      `json:"producer_payload_name,omitempty"`
	ProducerPayloadDigest     string      `json:"producer_payload_digest,omitempty"`
	ProducerPayloadBytes      int         `json:"producer_payload_bytes,omitempty"`
	Attestation               Attestation `json:"attestation"`
}

type CausalUnknown struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type ResolutionMetrics struct {
	CaseDenominator                    int    `json:"case_denominator"`
	ClosedCases                        int    `json:"closed_cases"`
	UnknownCases                       int    `json:"unknown_cases"`
	RefutedCases                       int    `json:"refuted_cases"`
	ActiveRootBefore                   int    `json:"active_root_before"`
	ActiveRootAfter                    int    `json:"active_root_after"`
	ExactResolutionsBefore             int    `json:"exact_resolutions_before"`
	ExactResolutionsAfter              int    `json:"exact_resolutions_after"`
	UnknownSixFieldBefore              int    `json:"unknown_six_field_before"`
	UnknownSixFieldAfter               int    `json:"unknown_six_field_after"`
	RefutedContradictionsBefore        int    `json:"refuted_contradictions_before"`
	RefutedContradictionsAfter         int    `json:"refuted_contradictions_after"`
	FallbackAcceptedBefore             int    `json:"fallback_accepted_before"`
	FallbackAcceptedAfter              int    `json:"fallback_accepted_after"`
	ArtifactInstancesBefore            int    `json:"artifact_instances_before"`
	ArtifactInstancesAfter             int    `json:"artifact_instances_after"`
	ArtifactTypesBefore                int    `json:"artifact_types_before"`
	ArtifactTypesAfter                 int    `json:"artifact_types_after"`
	IndependentReplayComparisonsBefore int    `json:"independent_replay_comparisons_before"`
	IndependentReplayComparisonsAfter  int    `json:"independent_replay_comparisons_after"`
	ProvenanceStateBefore              string `json:"provenance_state_before"`
	ProvenanceStateAfter               string `json:"provenance_state_after"`
	CurrentExact                       int    `json:"current_exact"`
	CurrentUnknownSixField             int    `json:"current_unknown_six_field"`
	CurrentRefuted                     int    `json:"current_refuted"`
	FallbackAccepted                   int    `json:"fallback_accepted"`
}

type ProvenanceResolution struct {
	State                     string         `json:"state"`
	Stage                     string         `json:"stage"`
	Step                      string         `json:"step"`
	Reason                    string         `json:"reason"`
	Unknown                   *CausalUnknown `json:"unknown,omitempty"`
	ProducerDeclarationDigest string         `json:"producer_declaration_digest,omitempty"`
	ProducerSubjectSHA        string         `json:"producer_subject_sha,omitempty"`
	ProducerPayloadDigest     string         `json:"producer_payload_digest,omitempty"`
	ProducerPayloadBytes      int            `json:"producer_payload_bytes,omitempty"`
	ArtifactInstances         int            `json:"artifact_instances"`
	ArtifactTypes             int            `json:"artifact_types"`
	FallbackAttempted         bool           `json:"fallback_attempted"`
	FallbackAccepted          bool           `json:"fallback_accepted"`
}

type Obligation struct {
	ID             string     `json:"id"`
	ProofRoute     string     `json:"proof_route"`
	Coordinate     Coordinate `json:"coordinate"`
	Status         string     `json:"status"`
	Reason         string     `json:"reason"`
	EvidenceDigest string     `json:"evidence_digest,omitempty"`
}

type Metrics struct {
	FixedObligationTotal int `json:"fixed_obligation_total"`
	VerifiedTotal        int `json:"verified_total"`
	UnknownTotal         int `json:"unknown_total"`
	FalseTotal           int `json:"false_total"`
	OpenTotal            int `json:"open_total"`
	CoverageBasisPoints  int `json:"coverage_basis_points"`
	FalsePromotionCount  int `json:"false_promotion_count"`
}
