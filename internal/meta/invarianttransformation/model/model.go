package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
)

const (
	ContractSchema      = "gooo/invariant-transformation-contract/v1"
	ReceiptSchema       = "gooo/invariant-transformation-receipt/v1"
	ReportSchema        = "gooo/invariant-transformation-report/v1"
	DenominatorID       = "gooo/invariant-transformation-denominator/v1"
	SourcePath          = "examples/invariant-transformation/main.gooo"
	ProducerID          = "invarianttransformation.producer"
	ConsumerID          = "invarianttransformation.independent-judge"
	AuthorityOp         = "authorize-invariant-preserving-transformation"
	EffectNoWrite       = "NO_EFFECT"
	EffectApproved      = "APPROVED_ARTIFACT_RECORDED"
	DecisionAllowed     = "AUTHORIZED"
	DecisionBlocked     = "BLOCKED"
	DecisionRefuted     = "REFUTED"
	DecisionPass        = "PASS"
	ResolutionExact     = "EXACT"
	ResolutionLower     = "LOWER_RESOLUTION"
	ResolutionInvariant = "INVARIANT_ONLY"
	StatusOpen          = "OPEN"
	StatusDischarged    = "DISCHARGED"
	StatusRefuted       = "REFUTED"
	ProofFoundation     = "FOUNDATION"
	ProofCoherence      = "COHERENCE"
	ProofRegression     = "REGRESSION"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type ValueSpec struct {
	ID            string     `json:"id"`
	Kind          string     `json:"kind"`
	Producer      string     `json:"producer"`
	Consumer      string     `json:"consumer"`
	MetaOperation string     `json:"meta_operation"`
	ProofChoice   string     `json:"proof_choice"`
	Coordinate    Coordinate `json:"coordinate"`
}

type CaseSpec struct {
	ID                 string `json:"id"`
	Kind               string `json:"kind"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedReason     string `json:"expected_reason"`
	ExpectedStatus     string `json:"expected_status"`
	ExpectedEffects    int    `json:"expected_effects"`
}

type Contract struct {
	Schema      string      `json:"schema"`
	Version     int         `json:"version"`
	Denominator string      `json:"denominator_id"`
	Source      string      `json:"source"`
	Values      []ValueSpec `json:"values"`
	Cases       []CaseSpec  `json:"cases"`
}

type Transition struct {
	From       string     `json:"from"`
	To         string     `json:"to"`
	Coordinate Coordinate `json:"coordinate"`
}

type MetaValue struct {
	ID             string     `json:"id"`
	Kind           string     `json:"kind"`
	Value          string     `json:"value"`
	EvidenceDigest string     `json:"evidence_digest"`
	Producer       string     `json:"producer"`
	Consumer       string     `json:"consumer"`
	MetaOperation  string     `json:"meta_operation"`
	ProofChoice    string     `json:"proof_choice"`
	Coordinate     Coordinate `json:"coordinate"`
}

type Claim struct {
	ID              string       `json:"id"`
	Status          string       `json:"status"`
	Reason          string       `json:"reason"`
	Coordinate      Coordinate   `json:"coordinate"`
	EvidenceDigests []string     `json:"evidence_digests"`
	Transitions     []Transition `json:"transitions"`
}

type TransformationEvidence struct {
	SourceDigest             string `json:"source_digest"`
	CandidateDigest          string `json:"candidate_digest"`
	SemanticBeforeDigest     string `json:"semantic_before_digest"`
	SemanticAfterDigest      string `json:"semantic_after_digest"`
	RegressionWitnessPresent bool   `json:"regression_witness_present"`
	ReplayBeforeDigest       string `json:"replay_before_digest,omitempty"`
	ReplayAfterDigest        string `json:"replay_after_digest,omitempty"`
	ReplayCount              int    `json:"replay_count"`
}

type Effect struct {
	Kind              string `json:"kind"`
	ArtifactID        string `json:"artifact_id,omitempty"`
	ArtifactDigest    string `json:"artifact_digest,omitempty"`
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	MetaOperation     string `json:"meta_operation"`
	RepositoryWrites  int    `json:"repository_writes"`
	MutationAuthority bool   `json:"mutation_authority"`
}

type Receipt struct {
	Schema            string                 `json:"schema"`
	CaseID            string                 `json:"case_id"`
	CaseKind          string                 `json:"case_kind"`
	HeadSHA           string                 `json:"head_sha"`
	SourcePath        string                 `json:"source_path"`
	SourceDigest      string                 `json:"source_digest"`
	ContractDigest    string                 `json:"contract_digest"`
	Producer          string                 `json:"producer"`
	Consumer          string                 `json:"consumer"`
	MetaOperation     string                 `json:"meta_operation"`
	ProofChoice       string                 `json:"proof_choice"`
	Values            []MetaValue            `json:"values"`
	Claims            []Claim                `json:"claims"`
	Evidence          TransformationEvidence `json:"evidence"`
	Decision          string                 `json:"decision"`
	Resolution        string                 `json:"resolution"`
	Reason            string                 `json:"reason"`
	Effects           []Effect               `json:"effects"`
	RepositoryWrites  int                    `json:"repository_writes"`
	MutationAuthority bool                   `json:"mutation_authority"`
	Digest            string                 `json:"digest"`
}

type Judgment struct {
	Decision         string `json:"decision"`
	Resolution       string `json:"resolution"`
	Reason           string `json:"reason"`
	Status           string `json:"status"`
	Independent      bool   `json:"independent"`
	CheckedClaims    int    `json:"checked_claims"`
	DischargedClaims int    `json:"discharged_claims"`
	OpenClaims       int    `json:"open_claims"`
	RefutedClaims    int    `json:"refuted_claims"`
	Effects          int    `json:"effects"`
}

type CaseResult struct {
	Spec      CaseSpec `json:"spec"`
	Receipt   Receipt  `json:"receipt"`
	Judgment  Judgment `json:"judgment"`
	Satisfied bool     `json:"satisfied"`
}

type Summary struct {
	CasesTotal              int `json:"cases_total"`
	CasesSatisfied          int `json:"cases_satisfied"`
	AuthorizedCases         int `json:"authorized_cases"`
	RefutedCases            int `json:"refuted_cases"`
	OpenCases               int `json:"open_cases"`
	ClaimsTotal             int `json:"claims_total"`
	DischargedClaims        int `json:"discharged_claims"`
	RefutedClaims           int `json:"refuted_claims"`
	OpenClaims              int `json:"open_claims"`
	TransitionEvents        int `json:"transition_events"`
	ApprovedArtifactEffects int `json:"approved_artifact_effects"`
	RepositoryWrites        int `json:"repository_writes"`
	MutationAuthority       int `json:"mutation_authority"`
	CoverageBPS             int `json:"coverage_bps"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Relation      string `json:"relation"`
	Satisfied     bool   `json:"satisfied"`
}

type Report struct {
	Schema           string       `json:"schema"`
	HeadSHA          string       `json:"head_sha"`
	SourcePath       string       `json:"source_path"`
	SourceDigest     string       `json:"source_digest"`
	ContractDigest   string       `json:"contract_digest"`
	DenominatorID    string       `json:"denominator_id"`
	DenominatorTotal int          `json:"denominator_total"`
	Decision         string       `json:"decision"`
	Resolution       string       `json:"resolution"`
	Reason           string       `json:"reason"`
	Cases            []CaseResult `json:"cases"`
	Indicators       []Indicator  `json:"indicators"`
	Summary          Summary      `json:"summary"`
	NotClaimed       []string     `json:"not_claimed"`
	Digest           string       `json:"digest"`
}

func CanonicalContract() Contract {
	return Contract{
		Schema: ContractSchema, Version: 1, Denominator: DenominatorID, Source: SourcePath,
		Values: []ValueSpec{
			{ID: "precondition", Kind: "PRECONDITION", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: "bind-transformation-precondition", ProofChoice: ProofFoundation, Coordinate: Coordinate{"PRECONDITION", "bind-source-snapshot", "EXACT_SOURCE_SNAPSHOT"}},
			{ID: "transformation", Kind: "TRANSFORMATION", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: "observe-transformation-result", ProofChoice: ProofCoherence, Coordinate: Coordinate{"TRANSFORMATION", "observe-candidate", "TRANSFORMATION_OBSERVED"}},
			{ID: "postcondition", Kind: "POSTCONDITION", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: "compare-postcondition", ProofChoice: ProofCoherence, Coordinate: Coordinate{"POSTCONDITION", "compare-semantic-value", "SEMANTIC_POSTCONDITION"}},
			{ID: "regression-witness", Kind: "REGRESSION_WITNESS", Producer: ProducerID, Consumer: ConsumerID, MetaOperation: "replay-regression-witness", ProofChoice: ProofRegression, Coordinate: Coordinate{"REGRESSION", "replay-before-after", "REGRESSION_WITNESS"}},
		},
		Cases: []CaseSpec{
			{ID: "preserved-translation", Kind: "PRESERVED", ExpectedDecision: DecisionAllowed, ExpectedResolution: ResolutionExact, ExpectedReason: "ALL_INVARIANTS_DISCHARGED", ExpectedStatus: StatusDischarged, ExpectedEffects: 0},
			{ID: "semantic-violation", Kind: "VIOLATION", ExpectedDecision: DecisionRefuted, ExpectedResolution: ResolutionInvariant, ExpectedReason: "SEMANTIC_POSTCONDITION_REFUTED", ExpectedStatus: StatusRefuted, ExpectedEffects: 0},
			{ID: "missing-regression-witness", Kind: "EVIDENCE_MISSING", ExpectedDecision: DecisionBlocked, ExpectedResolution: ResolutionLower, ExpectedReason: "REGRESSION_WITNESS_MISSING", ExpectedStatus: StatusOpen, ExpectedEffects: 0},
			{ID: "approved-artifact", Kind: "APPROVED_ARTIFACT", ExpectedDecision: DecisionAllowed, ExpectedResolution: ResolutionExact, ExpectedReason: "ALL_INVARIANTS_DISCHARGED", ExpectedStatus: StatusDischarged, ExpectedEffects: 1},
		},
	}
}

func DigestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func Digest(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return DigestBytes(raw)
}

func SealReceipt(receipt Receipt) Receipt {
	receipt.Digest = ""
	receipt.Digest = Digest(receipt)
	return receipt
}

func SealReport(report Report) Report {
	report.Digest = ""
	report.Digest = Digest(report)
	return report
}

func ValidDigest(value string) bool { return digestPattern.MatchString(value) }
func ValidHead(value string) bool   { return headPattern.MatchString(value) }

func ValidateContract(contract Contract) error {
	if Digest(contract) != Digest(CanonicalContract()) {
		return fmt.Errorf("invariant transformation contract drift")
	}
	return nil
}
