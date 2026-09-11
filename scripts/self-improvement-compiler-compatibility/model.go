package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compilercompatibility"
)

type caseReport struct {
	ID               string                                 `json:"id"`
	ExpectedDecision string                                 `json:"expected_decision"`
	ObservedDecision string                                 `json:"observed_decision"`
	Reason           string                                 `json:"reason"`
	Unknown          *compilercompatibility.UnknownState    `json:"unknown"`
	AxisComparisons  []compilercompatibility.AxisComparison `json:"axis_comparisons"`
	MismatchDetected bool                                   `json:"mismatch_detected"`
	FallbackRejected bool                                   `json:"fallback_rejected"`
}

type conformanceReport struct {
	Schema                        string                                  `json:"schema"`
	Decision                      string                                  `json:"decision"`
	Reason                        string                                  `json:"reason"`
	PolicySourceDigest            string                                  `json:"policy_source_digest"`
	PolicyEvaluatorDigest         string                                  `json:"policy_evaluator_digest"`
	CaseDenominator               int                                     `json:"case_denominator"`
	CaseIDs                       []string                                `json:"case_ids"`
	Cases                         []caseReport                            `json:"cases"`
	ClosedCases                   int                                     `json:"closed_cases"`
	UnknownCases                  int                                     `json:"unknown_cases"`
	RefutedCases                  int                                     `json:"refuted_cases"`
	IdentityAxisCount             int                                     `json:"identity_axis_count"`
	StrictPredecessorConsumption  compilercompatibility.StrictConsumption `json:"strict_predecessor_consumption"`
	ImplementationDigestEqual     bool                                    `json:"implementation_digest_equal"`
	ImplementationDigestDifferent bool                                    `json:"implementation_digest_different"`
	SemanticEqual                 bool                                    `json:"semantic_equal"`
	GeneratedBytesEqual           bool                                    `json:"generated_bytes_equal"`
	GeneratedManifestEqual        bool                                    `json:"generated_manifest_equal"`
	PolicyResultEqual             bool                                    `json:"policy_result_equal"`
	FullTestContractEqual         bool                                    `json:"full_test_contract_equal"`
	IndependentReplayExecutions   int                                     `json:"independent_replay_executions"`
	TestContractReplays           int                                     `json:"test_contract_replays"`
	CompatibilityScopeSubjects    int                                     `json:"compatibility_scope_subjects"`
	CertificateCount              int                                     `json:"certificate_count"`
	CertificateBytes              int                                     `json:"certificate_bytes"`
	CompatibilityHits             int                                     `json:"compatibility_hits"`
	CompatibilityMisses           int                                     `json:"compatibility_misses"`
	MismatchDetections            int                                     `json:"mismatch_detections"`
	CertificateTamperDetections   int                                     `json:"certificate_tamper_detections"`
	ScopeWideningDetections       int                                     `json:"scope_widening_detections"`
	FallbackAttempts              int                                     `json:"fallback_attempts"`
	FallbackAccepted              int                                     `json:"fallback_accepted"`
	FallbackRejected              int                                     `json:"fallback_rejected"`
	EvidenceArtifactCount         int                                     `json:"evidence_artifact_count"`
	ContinuityEdgeCount           int                                     `json:"continuity_edge_count"`
	Claim                         string                                  `json:"claim"`
	PerformanceClaim              bool                                    `json:"performance_claim"`
	GeneralCompatibilityClaim     bool                                    `json:"general_compiler_compatibility_claim"`
	UnsupportedFrontierDecision   string                                  `json:"unsupported_frontier_decision"`
	UnsupportedFrontierClaims     []string                                `json:"unsupported_frontier_claims"`
	RepositoryWrites              int                                     `json:"repository_writes"`
	LocalTestExecutions           int                                     `json:"local_test_executions"`
	WallMS                        int64                                   `json:"wall_ms"`
	PeakRSSKib                    int64                                   `json:"peak_rss_kib"`
	EvidenceArtifactNames         []string                                `json:"evidence_artifact_names"`
}

var publicationNames = []string{
	"canonical-policy.gooo",
	"compatibility-policy.json",
	"transition-table.json",
	"case-table.json",
	"predecessor-execution-receipt.json",
	"successor-execution-receipt.json",
	"compatibility-authorization.json",
	"compatibility-certificate.json",
	"compatibility-certificate-digest.txt",
	"actual-compatibility-consumption.json",
	"strict-replay.json",
	"bounded-implementation-successor-replay.json",
	"missing-successor-replay.json",
	"unbounded-compatibility-scope.json",
	"semantic-policy-output-mismatch.json",
	"tampered-widened-certificate.json",
	"axis-comparisons.json",
	"replay-evidence.json",
	"compatibility-scope.json",
	"continuity-edges.json",
	"compatibility-metrics.json",
	"compatibility-claim.json",
	"compiler-identity.json",
	"compatibility-report.json",
}
