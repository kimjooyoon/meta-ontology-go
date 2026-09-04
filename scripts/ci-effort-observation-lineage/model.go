package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/publicworkflowlineage"

type caseReport struct {
	CaseID              string                               `json:"case_id"`
	ExpectedDecision    string                               `json:"expected_decision"`
	Decision            string                               `json:"decision"`
	LineageState        string                               `json:"lineage_state"`
	Reason              string                               `json:"reason"`
	SourceRunID         int64                                `json:"source_run_id"`
	SourceSubjectSHA    string                               `json:"source_subject_sha"`
	ArtifactIdentity    string                               `json:"artifact_identity,omitempty"`
	ExactSubjectBinding bool                                 `json:"exact_subject_binding"`
	MismatchDetected    bool                                 `json:"mismatch_detected"`
	FallbackAttempted   bool                                 `json:"fallback_attempted"`
	FallbackRejected    bool                                 `json:"fallback_rejected"`
	ArtifactResolved    bool                                 `json:"artifact_resolved"`
	Unknown             *publicworkflowlineage.CausalUnknown `json:"unknown,omitempty"`
}

type report struct {
	Schema                          string                       `json:"schema"`
	Decision                        string                       `json:"decision"`
	Reason                          string                       `json:"reason"`
	Policy                          publicworkflowlineage.Policy `json:"policy"`
	Cases                           []caseReport                 `json:"cases"`
	CaseDenominator                 int                          `json:"case_denominator"`
	ClosedCases                     int                          `json:"closed_cases"`
	UnknownCases                    int                          `json:"unknown_cases"`
	RefutedCases                    int                          `json:"refuted_cases"`
	LineageEdgeCount                int                          `json:"lineage_edge_count"`
	SourceReceiptCount              int                          `json:"source_receipt_count"`
	ConsumerReceiptCount            int                          `json:"consumer_receipt_count"`
	EvidenceArtifactCount           int                          `json:"evidence_artifact_count"`
	StaleMisattributedBefore        int                          `json:"stale_misattributed_current_head_failures_before"`
	StaleMisattributedAfter         int                          `json:"stale_misattributed_current_head_failures_after"`
	ExactSubjectBindings            int                          `json:"exact_subject_bindings"`
	UnknownClassifications          int                          `json:"unknown_classifications"`
	StaleSourceStatesUnknown        int                          `json:"stale_source_states_unknown"`
	MismatchDetections              int                          `json:"mismatch_detections"`
	FallbackAttempts                int                          `json:"fallback_to_current_dev_attempts"`
	FallbackAccepted                int                          `json:"fallback_to_current_dev_accepted"`
	FallbackRejected                int                          `json:"fallback_to_current_dev_rejected"`
	SourceArtifactResolutions       int                          `json:"source_artifact_resolutions"`
	ActiveLineageRoots              int                          `json:"active_lineage_roots"`
	CasesSatisfied                  int                          `json:"cases_satisfied"`
	CasesTotal                      int                          `json:"cases_total"`
	UnknownSixFieldPreservations    int                          `json:"unknown_six_field_preservations"`
	ContradictionsRefuted           int                          `json:"contradictions_refuted"`
	ExactReplayComparisons          int                          `json:"exact_replay_comparisons"`
	ProvenanceState                 string                       `json:"provenance_state"`
	TrueProductFailuresNotRelabeled bool                         `json:"true_product_failures_not_relabeled"`
	WallMS                          int64                        `json:"wall_ms"`
	PeakRSSKib                      int64                        `json:"peak_rss_kib"`
	RuntimeComparable               bool                         `json:"runtime_comparable"`
	RuntimeUnknown                  string                       `json:"runtime_unknown"`
	RepositoryWrites                int                          `json:"repository_writes"`
	LocalTestExecutions             int                          `json:"local_test_executions"`
	PublishedArtifacts              []string                     `json:"published_artifacts"`
}

var publicationNames = []string{
	"canonical-source.gooo", "workflow-lineage-policy.json", "workflow-lineage-case-table.json", "source-run-receipts.json", "consumer-receipts.json", "lineage-edges.json",
	"exact-subject-binding.json", "stale-lineage-a5697c29.json", "source-api-missing.json", "artifact-lookup-missing.json", "mismatch-lineage.json", "tampered-artifact.json", "source-repository-mismatch.json", "product-failure-safety.json",
	"runtime-measurements.json", "repository-status.json", "workflow-lineage-report.json", "workflow-lineage-human.txt", "workflow-lineage-verification-input.json", "workflow-lineage-metrics.json",
}

type runInput struct {
	Source    string
	Out       string
	Trigger   string
	Run       string
	Artifacts string
}
