package languageassurance

const (
	RawReconstructionSchema = "gooo/language-assurance-raw-reconstruction/v1"
	RawVerifierID           = "gooo-independent-json-reconstructor-v1"
	MetricRawReconstruction = "gooo.metric.evidence.raw-reconstruction.v1"
	ReasonRawMismatch       = "ASSURANCE_RAW_RECONSTRUCTION_MISMATCH"
)

type RawObservation struct {
	EvidenceGroupsObserved   int        `json:"evidence_groups_observed"`
	EvidenceGroupsTotal      int        `json:"evidence_groups_total"`
	SelfMintingPaths         *int       `json:"self_minting_paths"`
	RoleConflictPaths        *int       `json:"role_conflict_paths"`
	UnknownLaunderingPaths   *int       `json:"unknown_laundering_paths"`
	UnknownTopDecisions      *int       `json:"unknown_top_decisions"`
	SnapshotBindingsObserved int        `json:"snapshot_bindings_observed"`
	SnapshotBindingsRequired int        `json:"snapshot_bindings_required"`
	ExactSnapshotBindingBPS  *int       `json:"exact_snapshot_binding_bps"`
	SnapshotMismatchPaths    *int       `json:"snapshot_mismatch_paths"`
	CandidateDecision        string     `json:"candidate_decision"`
	CandidateReason          string     `json:"candidate_reason"`
	CandidateResolution      Resolution `json:"candidate_resolution"`
}

type RawReconstructionReceipt struct {
	Schema               string         `json:"schema"`
	VerifierID           string         `json:"verifier_id"`
	SubjectSHA           string         `json:"subject_sha"`
	DenominatorDigest    string         `json:"denominator_digest"`
	RawTransactionDigest string         `json:"raw_transaction_digest"`
	Observation          RawObservation `json:"observation"`
}

type RawReconstructionSummary struct {
	RawReconstructionsObserved     int  `json:"raw_reconstructions_observed"`
	RawReconstructionsRequired     int  `json:"raw_reconstructions_required"`
	RawReconstructionBPS           *int `json:"raw_reconstruction_bps"`
	RawReconstructionMismatchPaths *int `json:"raw_reconstruction_mismatch_paths"`
}
