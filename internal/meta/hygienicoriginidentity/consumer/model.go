package consumer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const (
	SchemaVersion       = "gooo/hygienic-origin-identity/v2"
	Producer            = "gooo://hygienic-origin-identity/producer/name-generator"
	Consumer            = "gooo://hygienic-origin-identity/consumer/binding-site"
	MetaOperation       = "generate-name-preserving-origin-and-scope"
	ProofChoice         = "ORIGIN_SCOPE_EQUIVALENCE"
	DecisionPass        = "PASS"
	DecisionUnknown     = "UNKNOWN"
	DecisionRefuted     = "REFUTED"
	ResolutionExact     = "EXACT"
	ResolutionLower     = "LOWER_RESOLUTION"
	StatusOpen          = "OPEN"
	StatusDischarged    = "DISCHARGED"
	StatusRefuted       = "REFUTED"
	StatusUnclassified  = "UNCLASSIFIED"
	ConsumerBinding     = "gooo://hygienic-origin-identity/consumer/binding-site"
	ProducerExpansion   = "gooo://hygienic-origin-identity/producer/expansion-1"
	ConsumerCallSite    = "gooo://hygienic-origin-identity/scope/consumer-call-site"
	FreshProducerScope  = "gooo://hygienic-origin-identity/scope/fresh-producer-expansion-1"
	ExpectedCaseTotal   = 2
	ExpectedClaimTotal  = 4
	ExpectedTargetTotal = 2
)

type Report struct {
	SchemaVersion string       `json:"schema"`
	Decision      string       `json:"decision"`
	Resolution    string       `json:"resolution"`
	Producer      string       `json:"producer"`
	Consumer      string       `json:"consumer"`
	MetaOperation string       `json:"meta_operation"`
	ProofChoice   string       `json:"proof_choice"`
	Source        Subject      `json:"source"`
	Cases         []Case       `json:"cases"`
	Claims        []Claim      `json:"claims"`
	Transitions   []Transition `json:"claim_transitions"`
	Unknowns      []Unknown    `json:"unknowns"`
	Metrics       Metrics      `json:"metrics"`
	Authority     Authority    `json:"authority"`
	ReceiptDigest string       `json:"receipt_digest"`
}

type Subject struct {
	Path           string `json:"path"`
	HeadSHA        string `json:"head_sha"`
	RawDigest      string `json:"raw_digest"`
	SemanticDigest string `json:"semantic_digest"`
}

type Case struct {
	ID                       string   `json:"id"`
	Label                    string   `json:"label"`
	Spelling                 string   `json:"spelling"`
	OriginIdentity           string   `json:"origin_identity"`
	ScopeProvenance          string   `json:"scope_provenance"`
	ResolvedIdentity         string   `json:"resolved_identity"`
	SameSpelling             bool     `json:"same_spelling"`
	Captured                 bool     `json:"captured"`
	Control                  bool     `json:"control"`
	Target                   bool     `json:"target"`
	OriginIdentityPreserved  bool     `json:"origin_identity_preserved"`
	ScopeProvenancePreserved bool     `json:"scope_provenance_preserved"`
	ClaimIDs                 []string `json:"claim_ids"`
}

type Claim struct {
	ID             string `json:"id"`
	CaseID         string `json:"case_id"`
	Proposition    string `json:"proposition"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
}

type Transition struct {
	Sequence       int    `json:"sequence"`
	ClaimID        string `json:"claim_id"`
	Before         string `json:"before"`
	After          string `json:"after"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
}

type Unknown struct {
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
}

type Metrics struct {
	FixedCaseDenominator               int `json:"fixed_case_denominator"`
	FixedClaimDenominator              int `json:"fixed_claim_denominator"`
	FixedTargetPreservationDenominator int `json:"fixed_target_preservation_denominator"`
	ObservedCaseTotal                  int `json:"observed_case_total"`
	ObservedClaimTotal                 int `json:"observed_claim_total"`
	SameSpellingCaseTotal              int `json:"same_spelling_case_total"`
	CapturedCaseTotal                  int `json:"captured_case_total"`
	NonCapturedCaseTotal               int `json:"non_captured_case_total"`
	ClassifiedClaimTotal               int `json:"classified_claim_total"`
	DischargedClaimTotal               int `json:"discharged_claim_total"`
	RefutedClaimTotal                  int `json:"refuted_claim_total"`
	OpenClaimTotal                     int `json:"open_claim_total"`
	ClassificationCoverageBPS          int `json:"classification_coverage_bps"`
	PreservationSatisfactionBPS        int `json:"preservation_satisfaction_bps"`
	UnknownPathTotal                   int `json:"unknown_path_total"`

	SourceCasesObserved          int `json:"source_cases_observed"`
	SourceCasesExpected          int `json:"source_cases_expected"`
	ProducerImportsObserved      int `json:"producer_imports_observed"`
	ProducerImportsExpected      int `json:"producer_imports_expected"`
	SemanticCausalityObserved    int `json:"semantic_causality_observed"`
	SemanticCausalityExpected    int `json:"semantic_causality_expected"`
	CommentInvarianceObserved    int `json:"comment_invariance_observed"`
	CommentInvarianceExpected    int `json:"comment_invariance_expected"`
	ControlCaptureObserved       int `json:"control_capture_observed"`
	ControlCaptureExpected       int `json:"control_capture_expected"`
	HygienicNonCaptureObserved   int `json:"hygienic_noncapture_observed"`
	HygienicNonCaptureExpected   int `json:"hygienic_noncapture_expected"`
	TargetPreservationObserved   int `json:"target_preservation_observed"`
	TargetPreservationExpected   int `json:"target_preservation_expected"`
	TargetPreservationDischarged int `json:"target_preservation_discharged"`
	TargetPreservationRefuted    int `json:"target_preservation_refuted"`
	TargetPreservationOpen       int `json:"target_preservation_open"`
	TargetPreservationBPS        int `json:"target_preservation_bps"`
}

type Authority struct {
	RepositoryWrites             int    `json:"repository_writes"`
	RepositoryMutationAuthorized bool   `json:"repository_mutation_authorized"`
	BeforeSnapshotDigest         string `json:"before_snapshot_digest"`
	AfterSnapshotDigest          string `json:"after_snapshot_digest"`
	SnapshotsEqual               bool   `json:"snapshots_equal"`
}

type SnapshotPair struct {
	Before []byte
	After  []byte
}

func Seal(report Report) Report {
	report.ReceiptDigest = ""
	report.ReceiptDigest = digestJSON(report)
	return report
}

func Encode(report Report) ([]byte, error) {
	value, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(value, '\n'), nil
}

// ContentDigest is a reconstruction key. It deliberately excludes the receipt
// seal and CI authority metadata so a validator can compare fresh source
// reconstruction with a report resealed by an attacker.
func ContentDigest(report Report) string {
	report.ReceiptDigest = ""
	report.Authority = Authority{}
	report.Source.Path = ""
	report.Source.HeadSHA = ""
	report.Source.RawDigest = ""
	return digestJSON(report)
}

func SealedDigest(report Report) string {
	report.ReceiptDigest = ""
	return digestJSON(report)
}

func digestJSON(value any) string {
	encoded, _ := json.Marshal(value)
	return digestBytes(encoded)
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}
