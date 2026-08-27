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
	Path           string `json:"path"`
	Kind           string `json:"kind"`
	Digest         string `json:"digest"`
	SemanticDigest string `json:"semantic_digest"`
}

type RawRecord struct {
	ID            string `json:"id"`
	Coordinate    string `json:"coordinate"`
	Audience      string `json:"audience"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	PriorClaim    string `json:"prior_claim"`
	Observation   string `json:"observation"`
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
	Source           ReceiptSource           `json:"source"`
	Summary          ReceiptSummary          `json:"summary"`
	Views            []ReceiptView           `json:"views"`
	Counterexamples  []ReceiptCounterexample `json:"counterexamples"`
	ClaimTransitions []ReceiptTransition     `json:"claim_transitions"`
	Digest           string                  `json:"digest"`
	FactsDigest      string                  `json:"facts_digest"`
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
	RecordsObserved         int `json:"records_observed"`
	CounterexamplesExecuted int `json:"counterexamples_executed"`
	SourceDenominator       int `json:"source_denominator"`
}

type ReceiptView struct {
	Audience               string            `json:"audience"`
	Resolution             string            `json:"resolution"`
	GlobalDecision         string            `json:"global_decision"`
	InheritedStatus        string            `json:"inherited_status"`
	LocalDecision          string            `json:"local_decision"`
	LocalResolution        string            `json:"local_resolution"`
	LocalReason            string            `json:"local_reason"`
	OmittedEvidence        []OmittedEvidence `json:"omitted_evidence"`
	Satisfied              int               `json:"satisfied"`
	Total                  int               `json:"total"`
	Visible                int               `json:"visible"`
	Required               int               `json:"required"`
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

type ReceiptCounterexample struct {
	ID     string `json:"id"`
	Global string `json:"global_decision"`
	Reason string `json:"reason"`
	Views  []struct {
		Audience        string `json:"audience"`
		LocalDecision   string `json:"local_decision"`
		LocalResolution string `json:"local_resolution"`
	} `json:"views"`
}

type ReceiptTransition struct {
	IndicatorID    string `json:"indicator_id"`
	Audience       string `json:"audience"`
	Before         string `json:"before"`
	After          string `json:"after"`
	Visibility     string `json:"visibility"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Input struct {
	SourcePath   string
	Source       []byte
	Ledger       RawLedger
	LedgerBytes  []byte
	Receipt      Receipt
	ReceiptBytes []byte
	RepoRoot     string
}

type AudienceCheck struct {
	Audience         string `json:"audience"`
	Visible          int    `json:"visible"`
	Required         int    `json:"required"`
	Decision         string `json:"local_decision"`
	Resolution       string `json:"local_resolution"`
	ClaimTransitions string `json:"claim_transitions"`
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

type Report struct {
	Schema                     string               `json:"schema"`
	Decision                   string               `json:"decision"`
	Reason                     string               `json:"reason"`
	RawLedgerFinalFieldsAbsent bool                 `json:"raw_ledger_final_fields_absent"`
	ReceiptDigestMatch         bool                 `json:"receipt_digest_match"`
	SourceReconstruction       SourceReconstruction `json:"source_reconstruction"`
	Audiences                  []AudienceCheck      `json:"audiences"`
	ProducerImports            ImportAudit          `json:"producer_imports"`
	ClaimTransitionsChecked    int                  `json:"claim_transitions_checked"`
	Digest                     string               `json:"digest"`
}
