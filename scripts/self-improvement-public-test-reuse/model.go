package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/publictestreuse"

const (
	testCommand  = "go test -tags generated_project -count=1 ."
	buildCommand = "go build ."
)

type executionInput struct {
	Policy       string
	Source       string
	Program      string
	Manifest     string
	TestContract string
	PackageDir   string
	OutputDir    string
	Receipt      string
	Authorize    bool
}

type identity struct {
	Policy        publictestreuse.Policy
	Binding       publictestreuse.Binding
	ProgramBytes  []byte
	ManifestBytes []byte
	TestBytes     []byte
}

type commandResult struct {
	Success      bool
	ExitCode     int
	ResultDigest string
	WallMS       int64
	PeakRSSKib   int64
}

type verificationInput struct {
	Schema               string   `json:"schema"`
	Policy               string   `json:"policy"`
	Source               string   `json:"source"`
	BaselineProgram      string   `json:"baseline_program"`
	ReplayProgram        string   `json:"replay_program"`
	Manifest             string   `json:"manifest"`
	TestContract         string   `json:"test_contract"`
	BaselineReport       string   `json:"baseline_report"`
	ReplayReport         string   `json:"replay_report"`
	MissingAuthorization string   `json:"missing_authorization_report"`
	StaleEvidence        string   `json:"stale_evidence_report"`
	TamperedReceipt      string   `json:"tampered_receipt_report"`
	PolicyMismatch       string   `json:"policy_mismatch_report"`
	BaselineReceipt      string   `json:"baseline_receipt"`
	PublishedRoot        string   `json:"published_root"`
	PublishedArtifacts   []string `json:"published_artifacts"`
}

type metricSnapshot struct {
	BuildExecutions      int   `json:"build_executions"`
	TestExecutions       int   `json:"test_executions"`
	ReusedTestExecutions int   `json:"reused_test_executions"`
	ReceiptHits          int   `json:"receipt_hits"`
	ReceiptMisses        int   `json:"receipt_misses"`
	BuildMS              int64 `json:"build_ms"`
	TestMS               int64 `json:"test_ms"`
	WallMS               int64 `json:"wall_ms"`
	PeakRSSKib           int64 `json:"peak_rss_kib"`
}

type comparisonSnapshot struct {
	GeneratedProgramBytesEqual bool `json:"generated_program_bytes_equal"`
	GeneratedSemanticEqual     bool `json:"generated_semantic_equal"`
	TestContractBytesEqual     bool `json:"test_contract_bytes_equal"`
	ReceiptBindingEqual        bool `json:"receipt_binding_equal"`
}

type verifiedCase struct {
	ID                   string `json:"id"`
	ExpectedDecision     string `json:"expected_decision"`
	ObservedDecision     string `json:"observed_decision"`
	Reason               string `json:"reason"`
	TestExecutions       int    `json:"test_executions"`
	ReusedTestExecutions int    `json:"reused_test_executions"`
	ReceiptHits          int    `json:"receipt_hits"`
	ReceiptMisses        int    `json:"receipt_misses"`
	RepositoryWrites     int    `json:"repository_writes"`
	LocalTestExecutions  int    `json:"local_test_executions"`
}

type verificationReport struct {
	Schema                string                  `json:"schema"`
	Decision              string                  `json:"decision"`
	Reason                string                  `json:"reason"`
	PolicySourceDigest    string                  `json:"policy_source_digest"`
	PolicySemanticDigest  string                  `json:"policy_semantic_digest"`
	PolicyEvaluatorDigest string                  `json:"policy_evaluator_digest"`
	Journey               []string                `json:"journey"`
	InputRegularFiles     int                     `json:"input_regular_files"`
	InputPhysicalLines    int                     `json:"input_physical_lines"`
	InputGoFiles          int                     `json:"input_go_files"`
	InputGoPhysicalLines  int                     `json:"input_go_physical_lines"`
	GeneratedFiles        int                     `json:"generated_files"`
	GeneratedGoFiles      int                     `json:"generated_go_files"`
	GeneratedProgramBytes int64                   `json:"generated_program_bytes"`
	GeneratedProgramLines int                     `json:"generated_program_lines"`
	TestContractBytes     int64                   `json:"test_contract_bytes"`
	TestContractLines     int                     `json:"test_contract_lines"`
	Before                metricSnapshot          `json:"before"`
	After                 metricSnapshot          `json:"after"`
	Comparisons           comparisonSnapshot      `json:"comparisons"`
	Cases                 []verifiedCase          `json:"cases"`
	CaseDenominator       int                     `json:"case_denominator"`
	ClosedCases           int                     `json:"closed_cases"`
	UnknownCases          int                     `json:"unknown_cases"`
	RefutedCases          int                     `json:"refuted_cases"`
	ArtifactDenominator   int                     `json:"artifact_denominator"`
	ArtifactCount         int                     `json:"artifact_count"`
	RepositoryWrites      int                     `json:"repository_writes"`
	LocalTestExecutions   int                     `json:"local_test_executions"`
	NoAggregateScore      bool                    `json:"no_aggregate_score"`
	Binding               publictestreuse.Binding `json:"binding"`
}
