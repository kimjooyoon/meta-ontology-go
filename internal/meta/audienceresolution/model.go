package audienceresolution

type Input struct {
	Contract     Contract
	Ledger       Ledger
	SourcePath   string
	Source       []byte
	LedgerBytes  []byte
	ArtifactRoot string
}

// Ledger is deliberately a raw recipe ledger. Its records describe claims
// that a CI provider must observe; they are not observations themselves.
type Ledger struct {
	Schema          string           `json:"schema"`
	ID              string           `json:"id"`
	Subject         string           `json:"subject"`
	Source          SourceBinding    `json:"source"`
	Records         []EvidenceRecord `json:"records"`
	Counterexamples []Counterexample `json:"counterexamples"`
}

type SourceBinding struct {
	Path             string `json:"path"`
	Kind             string `json:"kind"`
	Digest           string `json:"digest"`
	SemanticDigest   string `json:"semantic_digest,omitempty"`
	DeclarationCount int    `json:"declaration_count,omitempty"`
	Reconstructed    bool   `json:"reconstructed,omitempty"`
}

// EvidenceRecord is shared by raw recipes and derived evidence. Raw ledger
// records use EvidenceStatus=HISTORICAL_FIXTURE, while provider output uses
// CURRENT_EVIDENCE or UNKNOWN and fills artifact/content digests.
type EvidenceRecord struct {
	ID                string   `json:"id"`
	Coordinate        string   `json:"coordinate"`
	Audience          string   `json:"audience"`
	ClaimID           string   `json:"claim_id"`
	Proposition       string   `json:"proposition"`
	PropositionDigest string   `json:"proposition_digest"`
	TargetAddress     string   `json:"target_address"`
	Provider          string   `json:"provider"`
	ArtifactPath      string   `json:"artifact_path"`
	ArtifactPaths     []string `json:"artifact_paths,omitempty"`
	ContentDigest     string   `json:"content_digest"`
	ContentDigests    []string `json:"content_digests,omitempty"`
	ObservedPredicate string   `json:"observed_predicate"`
	ObservedValue     string   `json:"observed_value"`
	EvidenceStatus    string   `json:"evidence_status"`
	Producer          string   `json:"producer"`
	Consumer          string   `json:"consumer"`
	MetaOperation     string   `json:"meta_operation"`
	ProofChoice       string   `json:"proof_choice"`
	Stage             string   `json:"stage"`
	Step              string   `json:"step"`
	Reason            string   `json:"reason"`
	PriorClaim        string   `json:"prior_claim"`
}

type Counterexample struct {
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	Trigger          string `json:"trigger"`
	Mutation         string `json:"mutation"`
	TargetCoordinate string `json:"target_coordinate"`
}

type Indicator struct {
	ID                string `json:"id"`
	Class             string `json:"class"`
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	MetaOperation     string `json:"meta_operation"`
	ProofChoice       string `json:"proof_choice"`
	Stage             string `json:"stage"`
	Step              string `json:"step"`
	Reason            string `json:"reason"`
	ClaimBefore       string `json:"claim_before"`
	ClaimAfter        string `json:"claim_after"`
	Observed          int    `json:"observed"`
	Expected          int    `json:"expected"`
	Satisfied         bool   `json:"satisfied"`
	EvidenceStatus    string `json:"evidence_status"`
	PropositionDigest string `json:"proposition_digest"`
	TargetAddress     string `json:"target_address"`
	ArtifactPath      string `json:"artifact_path"`
	ContentDigest     string `json:"content_digest"`
}

type AudienceView struct {
	Audience               string            `json:"audience"`
	Resolution             string            `json:"resolution"`
	GlobalDecision         string            `json:"global_decision"`
	InheritedStatus        string            `json:"inherited_status"`
	LocalDecision          string            `json:"local_decision"`
	LocalResolution        string            `json:"local_resolution"`
	LocalReason            string            `json:"local_reason"`
	OmittedEvidence        []OmittedEvidence `json:"omitted_evidence,omitempty"`
	Satisfied              int               `json:"satisfied"`
	Total                  int               `json:"total"`
	Visible                int               `json:"visible"`
	Required               int               `json:"required"`
	SubjectSatisfied       int               `json:"subject_satisfied"`
	SubjectRequired        int               `json:"subject_required"`
	BasisPoints            int               `json:"basis_points"`
	CoordinateIDs          []string          `json:"coordinate_ids"`
	OmittedCoordinateCount int               `json:"omitted_coordinate_count"`
}

type OmittedEvidence struct {
	Coordinate string `json:"coordinate"`
	Stage      string `json:"stage"`
	Step       string `json:"step"`
	Reason     string `json:"reason"`
}

// ClaimTransition is an audience-specific event over one of the 12 distinct
// propositions. The 36 events are intentionally not the denominator: they
// are three visibility projections of 12 claims.
type ClaimTransition struct {
	ClaimID              string `json:"claim_id"`
	Audience             string `json:"audience"`
	IndicatorID          string `json:"indicator_id"`
	Proposition          string `json:"proposition"`
	PropositionDigest    string `json:"proposition_digest"`
	TargetAddress        string `json:"target_address"`
	Before               string `json:"before"`
	After                string `json:"after"`
	Visibility           string `json:"visibility"`
	EvidenceStatus       string `json:"evidence_status"`
	EvidenceDigest       string `json:"evidence_digest"`
	EventDigest          string `json:"event_digest"`
	PreviousEventDigest  string `json:"previous_event_digest"`
	SourceDigest         string `json:"source_digest"`
	SemanticSourceDigest string `json:"semantic_source_digest"`
	Producer             string `json:"producer"`
	Consumer             string `json:"consumer"`
	MetaOperation        string `json:"meta_operation"`
	ProofChoice          string `json:"proof_choice"`
	Stage                string `json:"stage"`
	Step                 string `json:"step"`
	Reason               string `json:"reason"`
}

type Coordinates struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type EvidenceCounts struct {
	Current    int `json:"current_evidence"`
	Historical int `json:"historical_fixture"`
	Unknown    int `json:"unknown"`
}

type ConformanceSummary struct {
	Decision           string `json:"decision"`
	Resolution         string `json:"resolution"`
	SealClaimBefore    string `json:"seal_claim_before"`
	SealClaimAfter     string `json:"seal_claim_after"`
	SealEvidenceStatus string `json:"seal_evidence_status"`
}

type Summary struct {
	Coordinates              Coordinates        `json:"coordinates"`
	DistinctPropositions     int                `json:"distinct_propositions"`
	RecordsObserved          int                `json:"records_observed"`
	CounterexamplesExecuted  int                `json:"counterexamples_executed"`
	MissingCoordinates       int                `json:"missing_coordinates"`
	ContradictoryCoordinates int                `json:"contradictory_coordinates"`
	SourceDenominator        int                `json:"source_denominator"`
	EvidenceCounts           EvidenceCounts     `json:"evidence_counts"`
	Conformance              ConformanceSummary `json:"conformance"`
}

type ReplayVerification struct {
	RunAPath       string `json:"run_a_path"`
	RunADigest     string `json:"run_a_digest"`
	RunBPath       string `json:"run_b_path"`
	RunBDigest     string `json:"run_b_digest"`
	Equal          bool   `json:"equal"`
	CombinedDigest string `json:"combined_digest"`
}

type CounterexampleView struct {
	Audience        string `json:"audience"`
	Before          string `json:"before"`
	After           string `json:"after"`
	LocalDecision   string `json:"local_decision"`
	LocalResolution string `json:"local_resolution"`
}

type CounterexampleResult struct {
	ID                 string               `json:"id"`
	Kind               string               `json:"kind"`
	Trigger            string               `json:"trigger"`
	Mutation           string               `json:"mutation"`
	TargetCoordinate   string               `json:"target_coordinate"`
	TargetAddress      string               `json:"target_address"`
	Proposition        string               `json:"proposition"`
	PropositionDigest  string               `json:"proposition_digest"`
	Global             string               `json:"global_decision"`
	Resolution         string               `json:"resolution"`
	Stage              string               `json:"stage"`
	Step               string               `json:"step"`
	Reason             string               `json:"reason"`
	Views              []CounterexampleView `json:"views"`
	BeforeClaim        string               `json:"before_claim"`
	AfterClaim         string               `json:"after_claim"`
	ArtifactPath       string               `json:"artifact_path"`
	ContentDigest      string               `json:"content_digest"`
	ExecutionValidated bool                 `json:"execution_validated"`
}

type VerificationAttestation struct {
	Schema               string          `json:"schema"`
	SubjectReceiptDigest string          `json:"subject_receipt_digest"`
	Decision             string          `json:"decision"`
	Resolution           string          `json:"resolution"`
	Stage                string          `json:"stage"`
	Step                 string          `json:"step"`
	Reason               string          `json:"reason"`
	EvidenceStatus       string          `json:"evidence_status"`
	Evidence             EvidenceRecord  `json:"evidence"`
	ClaimTransition      ClaimTransition `json:"claim_transition"`
	Digest               string          `json:"digest"`
}

type Receipt struct {
	Schema           string                 `json:"schema"`
	ContractID       string                 `json:"contract_id"`
	Subject          string                 `json:"subject"`
	Decision         string                 `json:"decision"`
	Resolution       string                 `json:"resolution"`
	Reason           string                 `json:"reason"`
	Provisional      bool                   `json:"provisional"`
	Source           SourceBinding          `json:"source"`
	Summary          Summary                `json:"summary"`
	Indicators       []Indicator            `json:"indicators"`
	CurrentEvidence  []EvidenceRecord       `json:"current_evidence"`
	Views            []AudienceView         `json:"views"`
	Replay           ReplayVerification     `json:"replay"`
	Counterexamples  []CounterexampleResult `json:"counterexamples"`
	ClaimTransitions []ClaimTransition      `json:"claim_transitions"`
	NotClaimed       []string               `json:"not_claimed"`
	FactsDigest      string                 `json:"facts_digest"`
	Digest           string                 `json:"digest"`
}
