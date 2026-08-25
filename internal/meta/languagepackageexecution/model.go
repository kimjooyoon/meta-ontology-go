package languagepackageexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/packageexecution"

const (
	ContractSchema = "gooo/language-package-execution-contract/v1"
	ReportSchema   = "gooo/language-package-execution-report/v1"
)

type CaseSpec struct {
	ID               string `json:"id"`
	ExpectedDecision string `json:"expected_decision"`
	ExpectedReason   string `json:"expected_reason"`
	ProofChoice      string `json:"proof_choice"`
}

type Contract struct {
	Schema  string     `json:"schema"`
	Version int        `json:"version"`
	Cases   []CaseSpec `json:"cases"`
}

type CaseEvidence struct {
	ID      string                   `json:"id"`
	Receipt packageexecution.Receipt `json:"receipt"`
}

type Input struct {
	HeadSHA  string         `json:"head_sha"`
	Contract Contract       `json:"contract"`
	Cases    []CaseEvidence `json:"cases"`
}

type CaseResult struct {
	ID            string `json:"id"`
	Decision      string `json:"decision"`
	Reason        string `json:"reason"`
	Resolution    string `json:"resolution"`
	ReceiptDigest string `json:"receipt_digest"`
	Satisfied     bool   `json:"satisfied"`
}

type Summary struct {
	CasesSatisfied       int `json:"cases_satisfied"`
	CasesTotal           int `json:"cases_total"`
	SourceFilesLoaded    int `json:"source_files_loaded"`
	PackageExecutions    int `json:"package_executions"`
	DeterministicReplays int `json:"deterministic_replays"`
	DiagnosticRejections int `json:"diagnostic_rejections"`
	EventsObserved       int `json:"events_observed"`
	UnknownDecisions     int `json:"unknown_decisions"`
	RepositoryWrites     int `json:"repository_writes"`
	MutationAuthorities  int `json:"mutation_authorities"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}
