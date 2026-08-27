package phaseseparation

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type ParsedFile struct {
	Filename     string
	Source       []byte
	File         *syntax.File
	IR           semantic.IR
	SemanticHash string
	Activities   []SourceRecord
	EntityIDs    map[string]string
}

type SourceRecord struct {
	ActivityName     string
	ActivityID       string
	CaseKey          string
	TransferID       string
	ValueID          string
	FromValueID      string
	ToValueID        string
	LiteralClass     string
	FromLiteralClass string
	ToLiteralClass   string
	FromPhase        string
	ToPhase          string
	PayloadClass     string
	ClaimDigest      string
	TargetDigest     string
	Provenance       string
	Program          string
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type CaseResult struct {
	Name            string   `json:"name"`
	Class           string   `json:"class"`
	Outcome         string   `json:"outcome"`
	Reason          string   `json:"reason"`
	Stage           string   `json:"stage"`
	Step            string   `json:"step"`
	ClaimState      string   `json:"claim_state"`
	TransferCount   int      `json:"transfer_count"`
	TransferIDs     []string `json:"transfer_ids"`
	ValueIDs        []string `json:"value_ids"`
	PayloadClasses  []string `json:"payload_classes"`
	EvidenceDigests []string `json:"evidence_digests"`
	Provenances     []string `json:"provenances"`
	PreviousDigests []string `json:"previous_digests"`
	Passed          bool     `json:"passed"`
}

type ClaimTransition struct {
	ID             string `json:"id"`
	FromPhase      string `json:"from_phase"`
	ToPhase        string `json:"to_phase"`
	FromClaim      string `json:"from_claim"`
	ToClaim        string `json:"to_claim"`
	FromState      string `json:"from_state"`
	ToState        string `json:"to_state"`
	ClaimDigest    string `json:"claim_digest"`
	TargetDigest   string `json:"target_digest"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
	PreviousDigest string `json:"previous_digest"`
	MetaOperation  string `json:"meta_operation"`
	ProofChoice    string `json:"proof_choice"`
	Preserved      bool   `json:"preserved"`
}

type Indicator struct {
	ID            string `json:"id"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Numerator     int    `json:"numerator"`
	Denominator   int    `json:"denominator"`
	Satisfied     bool   `json:"satisfied"`
}

type View struct {
	Audience      string `json:"audience"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Satisfied     int    `json:"satisfied"`
	Total         int    `json:"total"`
	BasisPoints   int    `json:"basis_points"`
}

type Proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
	Passed         bool   `json:"passed"`
}

type Intervention struct {
	Kind                    string `json:"kind"`
	BaseSemanticDigest      string `json:"base_semantic_digest"`
	VariantSemanticDigest   string `json:"variant_semantic_digest"`
	BaseDecision            string `json:"base_decision"`
	VariantDecision         string `json:"variant_decision"`
	BaseTransitionDigest    string `json:"base_transition_digest"`
	VariantTransitionDigest string `json:"variant_transition_digest"`
	Changed                 bool   `json:"changed"`
	Preserved               bool   `json:"preserved"`
	Passed                  bool   `json:"passed"`
	Numerator               int    `json:"numerator"`
	Denominator             int    `json:"denominator"`
}

type Summary struct {
	SourceCasesProcessed         int `json:"source_cases_processed"`
	SourceCasesTotal             int `json:"source_cases_total"`
	CleanCasesPassed             int `json:"clean_cases_passed"`
	CleanCasesTotal              int `json:"clean_cases_total"`
	LeakageRejections            int `json:"leakage_rejections"`
	LeakageRejectionsTotal       int `json:"leakage_rejections_total"`
	ClaimTransitionsPreserved    int `json:"claim_transitions_preserved"`
	ClaimTransitionsTotal        int `json:"claim_transitions_total"`
	ExplicitClaimTransfers       int `json:"explicit_claim_transfers"`
	ExplicitClaimTransfersTotal  int `json:"explicit_claim_transfers_total"`
	IndicatorsSatisfied          int `json:"indicators_satisfied"`
	IndicatorsTotal              int `json:"indicators_total"`
	SemanticCausality            int `json:"semantic_causality"`
	SemanticCausalityTotal       int `json:"semantic_causality_total"`
	NonsemanticPreservation      int `json:"nonsemantic_preservation"`
	NonsemanticPreservationTotal int `json:"nonsemantic_preservation_total"`
	UnknownCases                 int `json:"unknown_cases"`
	RepositoryWrites             int `json:"repository_writes"`
}

type Authority struct {
	Execution bool `json:"execution"`
	Mutation  bool `json:"mutation"`
	Promotion bool `json:"promotion"`
}

type CISnapshot struct {
	RepositoryWrites   int    `json:"repository_writes"`
	MutationAuthority  bool   `json:"mutation_authority"`
	PromotionAuthority bool   `json:"promotion_authority"`
	ExecutionAuthority bool   `json:"execution_authority"`
	Permissions        string `json:"permissions"`
	BeforeStatusDigest string `json:"before_status_digest"`
	AfterStatusDigest  string `json:"after_status_digest"`
}

type UnknownResult struct {
	Decision       string     `json:"decision"`
	Resolution     string     `json:"resolution"`
	Coordinate     Coordinate `json:"coordinate"`
	ClaimState     string     `json:"claim_state"`
	EvidenceDigest string     `json:"evidence_digest"`
	Provenance     string     `json:"provenance"`
	PreviousDigest string     `json:"previous_digest"`
}

type Report struct {
	Schema                  string            `json:"schema"`
	Decision                string            `json:"decision"`
	Reason                  string            `json:"reason"`
	Resolution              string            `json:"resolution"`
	HeadSHA                 string            `json:"head_sha"`
	Toolchain               string            `json:"toolchain"`
	SourcePath              string            `json:"source_path"`
	SourceDigest            string            `json:"source_digest"`
	LeakSourcePath          string            `json:"leak_source_path"`
	LeakSourceDigest        string            `json:"leak_source_digest"`
	UnknownSourcePath       string            `json:"unknown_source_path"`
	UnknownSourceDigest     string            `json:"unknown_source_digest"`
	Producer                string            `json:"producer"`
	Consumer                string            `json:"consumer"`
	MetaOperation           string            `json:"meta_operation"`
	ProofChoice             string            `json:"proof_choice"`
	Cases                   []CaseResult      `json:"cases"`
	Transitions             []ClaimTransition `json:"claim_transitions"`
	Indicators              []Indicator       `json:"indicators"`
	Views                   []View            `json:"views"`
	Proofs                  []Proof           `json:"proofs"`
	SemanticIntervention    Intervention      `json:"semantic_intervention"`
	NonsemanticIntervention Intervention      `json:"nonsemantic_intervention"`
	Summary                 Summary           `json:"summary"`
	Authority               Authority         `json:"authority"`
	CISnapshot              CISnapshot        `json:"ci_snapshot"`
	Unknown                 UnknownResult     `json:"unknown"`
	Coordinate              Coordinate        `json:"coordinate"`
	Digest                  string            `json:"digest"`
}
