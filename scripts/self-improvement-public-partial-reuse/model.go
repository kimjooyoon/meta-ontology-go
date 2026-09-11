package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/publicpartialreuse"

type runInput struct {
	Source              string
	TestContract        string
	Gooo                string
	OrchestrationReport string
	RepoRoot            string
	Out                 string
}

type executionResult struct {
	Success      bool
	ExitCode     int
	ResultDigest string
	WallMS       int64
	PeakRSSKib   int64
}

type caseArtifacts struct {
	Report            publicpartialreuse.CaseReport
	Source            string
	GeneratedGo       string
	GeneratedManifest string
	Baseline          string
	Selective         string
	Receipts          map[string]string
}

type report struct {
	Schema               string                          `json:"schema"`
	Decision             string                          `json:"decision"`
	Reason               string                          `json:"reason"`
	Operation            string                          `json:"operation"`
	UpstreamReportDigest string                          `json:"upstream_orchestration_report_digest"`
	UpstreamOperation    string                          `json:"upstream_orchestration_operation"`
	Policy               publicpartialreuse.Policy       `json:"policy"`
	Cases                []publicpartialreuse.CaseReport `json:"cases"`
	CaseDenominator      int                             `json:"case_denominator"`
	ClosedCases          int                             `json:"closed_cases"`
	UnknownCases         int                             `json:"unknown_cases"`
	RefutedCases         int                             `json:"refuted_cases"`
	TestUnitsTotal       int                             `json:"test_units_total"`
	GeneratedArtifacts   int                             `json:"generated_artifacts"`
	EvidenceArtifacts    int                             `json:"evidence_artifacts"`
	InputRegularFiles    int                             `json:"input_regular_files"`
	InputPhysicalLines   int                             `json:"input_physical_lines"`
	GeneratedGoFiles     int                             `json:"generated_go_files"`
	GeneratedGoBytes     int64                           `json:"generated_go_bytes"`
	GeneratedGoLines     int                             `json:"generated_go_lines"`
	TestContractBytes    int64                           `json:"test_contract_bytes"`
	TestContractLines    int                             `json:"test_contract_lines"`
	Before               publicpartialreuse.Metrics      `json:"before"`
	After                publicpartialreuse.Metrics      `json:"after"`
	NoChangeBefore       publicpartialreuse.Metrics      `json:"no_change_before"`
	NoChangeAfter        publicpartialreuse.Metrics      `json:"no_change_after"`
	SingleChangeBefore   publicpartialreuse.Metrics      `json:"single_change_before"`
	SingleChangeAfter    publicpartialreuse.Metrics      `json:"single_change_after"`
	GeneratedBytesEqual  bool                            `json:"generated_bytes_equal"`
	SemanticEqual        bool                            `json:"semantic_equal"`
	TestContractEqual    bool                            `json:"test_contract_equal"`
	ReceiptBindingEqual  bool                            `json:"receipt_binding_equal"`
	RuntimeComparable    bool                            `json:"runtime_comparable"`
	RuntimeUnknown       string                          `json:"runtime_unknown"`
	RepositoryWrites     int                             `json:"repository_writes"`
	LocalTestExecutions  int                             `json:"local_test_executions"`
	PublishedArtifacts   []string                        `json:"published_artifacts"`
}

var publicationNames = []string{
	"canonical-source.gooo", "canonical-test.go", "upstream-orchestration-report.json", "partial-reuse-policy.json",
	"no-change-generated.go", "no-change-generated.manifest.jsonl", "no-change-baseline.json", "no-change-selective.json",
	"no-change-orders.receipt.json", "no-change-inventory.receipt.json", "single-change-source.gooo", "single-change-generated.go",
	"single-change-generated.manifest.jsonl", "single-change-baseline.json", "single-change-selective.json", "single-change-orders.receipt.json",
	"single-change-inventory.receipt.json", "missing-dependency-edge.json", "unbounded-impact.json", "changed-hidden-dependency.json",
	"tampered-partition-receipt.json", "partial-reuse-report.json", "partial-reuse-human.txt", "runtime-measurements.json", "repository-status.json",
}
