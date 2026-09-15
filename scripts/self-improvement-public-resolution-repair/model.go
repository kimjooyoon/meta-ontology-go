package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/publicresolutionrepair"

type runInput struct {
	Source              string
	TestContract        string
	Gooo                string
	OrchestrationReport string
	V15Report           string
	V15Hidden           string
	RepoRoot            string
	Out                 string
}

type executionResult struct {
	Success      bool   `json:"successful"`
	ExitCode     int    `json:"exit_code"`
	ResultDigest string `json:"result_digest"`
	WallMS       int64  `json:"wall_ms"`
	PeakRSSKib   int64  `json:"peak_rss_kib"`
}

type upstreamReport struct {
	Schema    string `json:"schema"`
	Decision  string `json:"decision"`
	Operation string `json:"operation"`
}

type report struct {
	Schema                       string                                       `json:"schema"`
	Decision                     string                                       `json:"decision"`
	Reason                       string                                       `json:"reason"`
	Operation                    string                                       `json:"operation"`
	UpstreamReportDigest         string                                       `json:"upstream_orchestration_report_digest"`
	UpstreamOperation            string                                       `json:"upstream_orchestration_operation"`
	CompilerDigest               string                                       `json:"compiler_digest"`
	OriginalCounterexample       publicresolutionrepair.Counterexample        `json:"original_counterexample"`
	Proposal                     publicresolutionrepair.Proposal              `json:"repair_proposal"`
	Authorization                publicresolutionrepair.AuthorizationArtifact `json:"repair_authorization"`
	Overlay                      publicresolutionrepair.GraphOverlay          `json:"graph_overlay"`
	Policy                       publicresolutionrepair.Policy                `json:"policy"`
	Cases                        []publicresolutionrepair.CaseReport          `json:"cases"`
	CaseDenominator              int                                          `json:"case_denominator"`
	ClosedCases                  int                                          `json:"closed_cases"`
	UnknownCases                 int                                          `json:"unknown_cases"`
	RefutedCases                 int                                          `json:"refuted_cases"`
	ResolutionLevelCount         int                                          `json:"resolution_level_count"`
	ProofModeObservationCount    int                                          `json:"proof_mode_observation_count"`
	ProofFoundationCount         int                                          `json:"proof_foundation_count"`
	ProofCoherenceCount          int                                          `json:"proof_coherence_count"`
	ProofRegressionCount         int                                          `json:"proof_regression_count"`
	RepairProposalCount          int                                          `json:"repair_proposal_count"`
	AuthorizationDecisionCount   int                                          `json:"authorization_decision_count"`
	GraphEdgesBefore             int                                          `json:"graph_edges_before"`
	GraphEdgesAfter              int                                          `json:"graph_edges_after"`
	CanonicalGraphEdgeCount      int                                          `json:"canonical_graph_edge_count"`
	TestUnitsTotal               int                                          `json:"test_units_total"`
	FallbackTestUnitsExecuted    int                                          `json:"fallback_test_units_executed"`
	FallbackTestUnitsReused      int                                          `json:"fallback_test_units_reused"`
	OverlayTestUnitsExecuted     int                                          `json:"overlay_test_units_executed"`
	OverlayTestUnitsReused       int                                          `json:"overlay_test_units_reused"`
	SelectivityTestUnitsExecuted int                                          `json:"selectivity_test_units_executed"`
	SelectivityTestUnitsReused   int                                          `json:"selectivity_test_units_reused"`
	ContinuityEdgeCount          int                                          `json:"continuity_edge_count"`
	GeneratedArtifactCount       int                                          `json:"generated_artifact_count"`
	EvidenceArtifactCount        int                                          `json:"evidence_artifact_count"`
	Fallback                     publicresolutionrepair.Metrics               `json:"fallback"`
	OverlayReplay                publicresolutionrepair.Metrics               `json:"overlay_replay"`
	UnchangedSelectivity         publicresolutionrepair.Metrics               `json:"unchanged_partition_selectivity"`
	GeneratedBytesEqual          bool                                         `json:"generated_bytes_equal"`
	SemanticEqual                bool                                         `json:"semantic_equal"`
	TestContractEqual            bool                                         `json:"test_contract_equal"`
	FullTestOutcomeEqual         bool                                         `json:"full_test_outcome_equal"`
	OverlayBindingEqual          bool                                         `json:"overlay_binding_equal"`
	SafetyImprovement            bool                                         `json:"safety_improvement"`
	RuntimeComparable            bool                                         `json:"runtime_comparable"`
	RuntimeUnknown               string                                       `json:"runtime_unknown"`
	RepositoryWrites             int                                          `json:"repository_writes"`
	LocalTestExecutions          int                                          `json:"local_test_executions"`
	PublishedArtifacts           []string                                     `json:"published_artifacts"`
}

var publicationNames = []string{
	"canonical-source.gooo", "canonical-test.go", "upstream-orchestration-report.json", "original-hidden-dependency.json", "original-partial-reuse-report.json",
	"v15-counterexample-provenance.json", "resolution-repair-policy.json", "fallback-generated.go", "fallback-generated.manifest.jsonl", "fallback-baseline.json",
	"fallback-result.json", "fallback-counterexample-preserved.json", "repair-proposal.json", "repair-authorization.json", "authorized-graph-overlay.json",
	"overlay-generated.go", "overlay-generated.manifest.jsonl", "overlay-replay.json", "overlay-selectivity.json", "overlay-outcome-comparison.json",
	"ambiguous-repair-evidence.json", "unsupported-repair-proof-mode.json", "tampered-counterexample.json", "unauthorized-repair.json", "resolution-repair-report.json",
	"resolution-repair-human.txt", "runtime-measurements.json", "repository-status.json", "resolution-repair-verification-input.json", "resolution-repair-case-table.json",
}
