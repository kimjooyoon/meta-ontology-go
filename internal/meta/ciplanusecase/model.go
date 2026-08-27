package ciplanusecase

import "github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"

const (
	ContractSchema = "gooo/ci-plan-usecase-contract/v1"
	ReportSchema   = "gooo/ci-plan-usecase-scorecard/v1"
)

type CaseSpec struct {
	ID               string `json:"id"`
	ExpectedDecision string `json:"expected_decision"`
	ProofChoice      string `json:"proof_choice"`
}

type Limits struct {
	MaxWallMS       int64 `json:"max_wall_ms"`
	MaxPeakRSSKiB   int64 `json:"max_peak_rss_kib"`
	MaxReceiptBytes int64 `json:"max_receipt_bytes"`
}

type Contract struct {
	Schema      string     `json:"schema"`
	Denominator int        `json:"denominator"`
	Cases       []CaseSpec `json:"cases"`
	Limits      Limits     `json:"limits"`
	NotClaimed  []string   `json:"not_claimed"`
}

type ProfileSample struct {
	CaseID       string `json:"case_id"`
	WallMS       int64  `json:"wall_ms"`
	PeakRSSKiB   int64  `json:"peak_rss_kib"`
	ReceiptBytes int64  `json:"receipt_bytes"`
}

type Profile struct {
	Schema  string          `json:"schema"`
	Samples []ProfileSample `json:"samples"`
}

type SourceProfile struct {
	GoooFiles int `json:"gooo_files"`
	GoFiles   int `json:"go_files"`
	GoooLines int `json:"gooo_lines"`
	GoLines   int `json:"go_lines"`
}

type GoldenReason struct {
	Operation  string `json:"operation"`
	File       string `json:"file"`
	SourcePath string `json:"source_path"`
	SourceLine int    `json:"source_line"`
}

type GoldenCheck struct {
	ID      string         `json:"id"`
	Command string         `json:"command"`
	Files   []string       `json:"files"`
	Reasons []GoldenReason `json:"reasons"`
}

type GoldenPlan struct {
	Schema string        `json:"schema"`
	CaseID string        `json:"case_id"`
	Checks []GoldenCheck `json:"checks"`
}

type Input struct {
	Contract        Contract
	Reports         map[string]metainvocation.Report
	Replays         map[string]metainvocation.Report
	Goldens         map[string]GoldenPlan
	Profile         Profile
	Source          SourceProfile
	GeneratedReplay bool
}

type CaseResult struct {
	ID               string                        `json:"id"`
	ExpectedDecision string                        `json:"expected_decision"`
	ObservedDecision string                        `json:"observed_decision"`
	ProofChoice      string                        `json:"proof_choice"`
	Status           string                        `json:"status"`
	Unknowns         []metainvocation.UnknownCause `json:"unknowns"`
	ClaimStatuses    map[string]string             `json:"claim_statuses"`
	EvidenceDigest   string                        `json:"evidence_digest"`
}

type Summary struct {
	CasesSatisfied       int   `json:"cases_satisfied"`
	PassDecisions        int   `json:"pass_decisions"`
	FailClosedDecisions  int   `json:"fail_closed_decisions"`
	UnknownDecisions     int   `json:"unknown_decisions"`
	DeterministicReplays int   `json:"deterministic_replays"`
	GoldenPlans          int   `json:"golden_plans"`
	RuleEvidenceRefs     int   `json:"rule_evidence_refs"`
	DirectUnknownClaims  int   `json:"direct_unknown_claims"`
	DependencyBlocked    int   `json:"dependency_blocked_claims"`
	RefutedClaims        int   `json:"refuted_claims"`
	PersistentClaims     int   `json:"persistent_claims"`
	GeneratedReplays     int   `json:"generated_source_replays"`
	ResourceSamples      int   `json:"resource_samples"`
	GoooFiles            int   `json:"gooo_files"`
	GoFiles              int   `json:"go_files"`
	GoooLines            int   `json:"gooo_lines"`
	GoLines              int   `json:"go_lines"`
	MaxWallMS            int64 `json:"max_wall_ms"`
	MaxPeakRSSKiB        int64 `json:"max_peak_rss_kib"`
	MaxReceiptBytes      int64 `json:"max_receipt_bytes"`
	RepositoryWrites     int   `json:"repository_writes"`
	MutationAuthority    int   `json:"mutation_authority"`
}

type Indicator struct {
	ID            string `json:"id"`
	Reader        string `json:"reader"`
	Observed      int64  `json:"observed"`
	Comparator    string `json:"comparator"`
	Target        int64  `json:"target"`
	Status        string `json:"status"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
}

type ReaderView struct {
	Reader       string   `json:"reader"`
	Resolution   string   `json:"resolution"`
	IndicatorIDs []string `json:"indicator_ids"`
}

type Proof struct {
	Choice   string   `json:"choice"`
	Status   string   `json:"status"`
	Evidence []string `json:"evidence"`
}

type Report struct {
	Schema         string       `json:"schema"`
	Decision       string       `json:"decision"`
	Resolution     string       `json:"resolution"`
	Interpretation string       `json:"interpretation"`
	ContractDigest string       `json:"contract_digest"`
	Cases          []CaseResult `json:"cases"`
	Summary        Summary      `json:"summary"`
	Indicators     []Indicator  `json:"indicators"`
	ReaderViews    []ReaderView `json:"reader_views"`
	Proofs         []Proof      `json:"proofs"`
	NotClaimed     []string     `json:"not_claimed"`
	ReportDigest   string       `json:"report_digest"`
}
