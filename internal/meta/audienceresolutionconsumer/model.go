package audienceresolutionconsumer

type RawLedger struct {
	Schema          string              `json:"schema"`
	ID              string              `json:"id"`
	Subject         string              `json:"subject"`
	Source          RawSource           `json:"source"`
	Records         []RawRecord         `json:"records"`
	Counterexamples []RawCounterexample `json:"counterexamples"`
}

type RawSource struct {
	Path             string `json:"path"`
	Kind             string `json:"kind"`
	Digest           string `json:"digest"`
	SemanticDigest   string `json:"semantic_digest,omitempty"`
	DeclarationCount int    `json:"declaration_count,omitempty"`
	Reconstructed    bool   `json:"reconstructed,omitempty"`
}

type RawRecord struct {
	ID                string `json:"id"`
	Coordinate        string `json:"coordinate"`
	Audience          string `json:"audience"`
	ClaimID           string `json:"claim_id"`
	Proposition       string `json:"proposition"`
	PropositionDigest string `json:"proposition_digest"`
	TargetAddress     string `json:"target_address"`
	Provider          string `json:"provider"`
	ArtifactPath      string `json:"artifact_path"`
	ContentDigest     string `json:"content_digest"`
	ObservedPredicate string `json:"observed_predicate"`
	ObservedValue     string `json:"observed_value"`
	EvidenceStatus    string `json:"evidence_status"`
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	MetaOperation     string `json:"meta_operation"`
	ProofChoice       string `json:"proof_choice"`
	Stage             string `json:"stage"`
	Step              string `json:"step"`
	Reason            string `json:"reason"`
	PriorClaim        string `json:"prior_claim"`
}

type RawCounterexample struct {
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	Trigger          string `json:"trigger"`
	Mutation         string `json:"mutation"`
	TargetCoordinate string `json:"target_coordinate"`
}

type Receipt struct {
	Schema           string                  `json:"schema"`
	ContractID       string                  `json:"contract_id"`
	Subject          string                  `json:"subject"`
	Decision         string                  `json:"decision"`
	Resolution       string                  `json:"resolution"`
	Reason           string                  `json:"reason"`
	Provisional      bool                    `json:"provisional"`
	Source           ReceiptSource           `json:"source"`
	Summary          ReceiptSummary          `json:"summary"`
	Indicators       []Indicator             `json:"indicators"`
	CurrentEvidence  []EvidenceRecord        `json:"current_evidence"`
	Views            []ReceiptView           `json:"views"`
	Replay           ReplayVerification      `json:"replay"`
	Counterexamples  []ReceiptCounterexample `json:"counterexamples"`
	ClaimTransitions []ReceiptTransition     `json:"claim_transitions"`
	FactsDigest      string                  `json:"facts_digest"`
	Digest           string                  `json:"digest"`
}

type ReceiptSource struct {
	Path             string `json:"path"`
	Kind             string `json:"kind"`
	Digest           string `json:"digest"`
	SemanticDigest   string `json:"semantic_digest"`
	DeclarationCount int    `json:"declaration_count"`
	Reconstructed    bool   `json:"reconstructed"`
}

type ReceiptSummary struct {
	Coordinates struct {
		Satisfied   int `json:"satisfied"`
		Total       int `json:"total"`
		BasisPoints int `json:"basis_points"`
	} `json:"coordinates"`
	DistinctPropositions     int                `json:"distinct_propositions"`
	RecordsObserved          int                `json:"records_observed"`
	CounterexamplesExecuted  int                `json:"counterexamples_executed"`
	MissingCoordinates       int                `json:"missing_coordinates"`
	ContradictoryCoordinates int                `json:"contradictory_coordinates"`
	SourceDenominator        int                `json:"source_denominator"`
	EvidenceCounts           EvidenceCounts     `json:"evidence_counts"`
	Conformance              ConformanceSummary `json:"conformance"`
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

type Indicator struct {
	ID                string `json:"id"`
	Satisfied         bool   `json:"satisfied"`
	ClaimAfter        string `json:"claim_after"`
	EvidenceStatus    string `json:"evidence_status"`
	PropositionDigest string `json:"proposition_digest"`
	TargetAddress     string `json:"target_address"`
	ArtifactPath      string `json:"artifact_path"`
	ContentDigest     string `json:"content_digest"`
}

type ReceiptView struct {
	Audience         string            `json:"audience"`
	Resolution       string            `json:"resolution"`
	GlobalDecision   string            `json:"global_decision"`
	InheritedStatus  string            `json:"inherited_status"`
	LocalDecision    string            `json:"local_decision"`
	LocalResolution  string            `json:"local_resolution"`
	LocalReason      string            `json:"local_reason"`
	OmittedEvidence  []OmittedEvidence `json:"omitted_evidence"`
	Satisfied        int               `json:"satisfied"`
	Total            int               `json:"total"`
	Visible          int               `json:"visible"`
	Required         int               `json:"required"`
	SubjectSatisfied int               `json:"subject_satisfied"`
	SubjectRequired  int               `json:"subject_required"`
	CoordinateIDs    []string          `json:"coordinate_ids"`
}

type OmittedEvidence struct {
	Coordinate string `json:"coordinate"`
	Stage      string `json:"stage"`
	Step       string `json:"step"`
	Reason     string `json:"reason"`
}

type ReplayVerification struct {
	RunAPath       string `json:"run_a_path"`
	RunADigest     string `json:"run_a_digest"`
	RunBPath       string `json:"run_b_path"`
	RunBDigest     string `json:"run_b_digest"`
	Equal          bool   `json:"equal"`
	CombinedDigest string `json:"combined_digest"`
}

type ReceiptCounterexample struct {
	ID                  string               `json:"id"`
	Kind                string               `json:"kind"`
	TargetCoordinate    string               `json:"target_coordinate"`
	TargetAddress       string               `json:"target_address"`
	PropositionDigest   string               `json:"proposition_digest"`
	Global              string               `json:"global_decision"`
	Resolution          string               `json:"resolution"`
	Stage               string               `json:"stage"`
	Step                string               `json:"step"`
	Reason              string               `json:"reason"`
	Views               []CounterexampleView `json:"views"`
	BeforeClaim         string               `json:"before_claim"`
	AfterClaim          string               `json:"after_claim"`
	ArtifactPath        string               `json:"artifact_path"`
	ContentDigest       string               `json:"content_digest"`
	MutatedLedgerPath   string               `json:"mutated_ledger_path"`
	MutatedLedgerDigest string               `json:"mutated_ledger_digest"`
	Replay              ReplayVerification   `json:"replay"`
	ExecutionValidated  bool                 `json:"execution_validated"`
}

type CounterexampleView struct {
	Audience        string `json:"audience"`
	Before          string `json:"before"`
	After           string `json:"after"`
	LocalDecision   string `json:"local_decision"`
	LocalResolution string `json:"local_resolution"`
}

type ReceiptTransition struct {
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

type Input struct {
	SourcePath   string
	Source       []byte
	Ledger       RawLedger
	LedgerBytes  []byte
	Receipt      Receipt
	ReceiptBytes []byte
	RepoRoot     string
	ArtifactRoot string
}

type AudienceCheck struct {
	Audience         string `json:"audience"`
	Visible          int    `json:"visible"`
	Required         int    `json:"required"`
	Decision         string `json:"local_decision"`
	Resolution       string `json:"local_resolution"`
	SubjectSatisfied int    `json:"subject_satisfied"`
	SubjectRequired  int    `json:"subject_required"`
}

type SourceReconstruction struct {
	ParsedAndLowered  bool   `json:"parsed_and_lowered"`
	DeclarationCount  int    `json:"declaration_count"`
	SemanticDigest    string `json:"semantic_digest"`
	CanonicalIRDigest string `json:"canonical_ir_digest"`
	ReceiptMatches    bool   `json:"receipt_matches"`
}

type ImportAudit struct {
	Numerator   int      `json:"producer_import_numerator"`
	Denominator int      `json:"producer_import_denominator"`
	Forbidden   []string `json:"forbidden_imports,omitempty"`
}

type Attestation struct {
	Schema               string            `json:"schema"`
	SubjectReceiptDigest string            `json:"subject_receipt_digest"`
	Decision             string            `json:"decision"`
	Resolution           string            `json:"resolution"`
	Stage                string            `json:"stage"`
	Step                 string            `json:"step"`
	Reason               string            `json:"reason"`
	EvidenceStatus       string            `json:"evidence_status"`
	Evidence             EvidenceRecord    `json:"evidence"`
	ClaimTransition      ReceiptTransition `json:"claim_transition"`
	Digest               string            `json:"digest"`
}

type CounterexampleExecution struct {
	ID                  string               `json:"id"`
	Kind                string               `json:"kind"`
	TargetCoordinate    string               `json:"target_coordinate"`
	PropositionDigest   string               `json:"proposition_digest"`
	EvidenceDigest      string               `json:"evidence_digest"`
	MutatedLedgerPath   string               `json:"mutated_ledger_path"`
	MutatedLedgerDigest string               `json:"mutated_ledger_digest"`
	Replay              ReplayVerification   `json:"replay"`
	GlobalDecision      string               `json:"global_decision"`
	Resolution          string               `json:"resolution"`
	Stage               string               `json:"stage"`
	Step                string               `json:"step"`
	ExecutionReason     string               `json:"execution_reason"`
	BeforeClaim         string               `json:"before_claim"`
	AfterClaim          string               `json:"after_claim"`
	Views               []CounterexampleView `json:"views"`
	Reexecuted          bool                 `json:"reexecuted"`
	Passed              bool                 `json:"passed"`
	Reason              string               `json:"reason"`
}

type CounterexampleTamperCheck struct {
	ID                   string `json:"id"`
	ProducerArtifactPath string `json:"producer_artifact_path"`
	OriginalDigest       string `json:"original_digest"`
	TamperedDigest       string `json:"tampered_digest"`
	Detected             bool   `json:"detected"`
}

type Report struct {
	Schema                     string                    `json:"schema"`
	Decision                   string                    `json:"decision"`
	Reason                     string                    `json:"reason"`
	RawLedgerFinalFieldsAbsent bool                      `json:"raw_ledger_final_fields_absent"`
	RawEvidenceHistoricalOnly  bool                      `json:"raw_evidence_historical_only"`
	ReceiptDigestMatch         bool                      `json:"receipt_digest_match"`
	SourceReconstruction       SourceReconstruction      `json:"source_reconstruction"`
	Audiences                  []AudienceCheck           `json:"audiences"`
	ProducerImports            ImportAudit               `json:"producer_imports"`
	CurrentEvidenceCounts      EvidenceCounts            `json:"current_evidence_counts"`
	DistinctPropositions       int                       `json:"distinct_propositions"`
	Replay                     ReplayVerification        `json:"replay"`
	CounterexamplesChecked     int                       `json:"counterexamples_checked"`
	CounterexamplesReexecuted  int                       `json:"counterexamples_reexecuted"`
	CounterexampleExecutions   []CounterexampleExecution `json:"counterexample_executions"`
	CounterexampleTamper       CounterexampleTamperCheck `json:"counterexample_tamper_check"`
	ClaimTransitionsChecked    int                       `json:"claim_transitions_checked"`
	Attestation                Attestation               `json:"verification_attestation"`
	Digest                     string                    `json:"digest"`
}
