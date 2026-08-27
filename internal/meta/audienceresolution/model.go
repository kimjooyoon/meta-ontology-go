package audienceresolution

type Input struct {
	Contract   Contract
	Ledger     Ledger
	Replay     Ledger
	SourcePath string
	Source     []byte
}

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
	SemanticDigest   string `json:"semantic_digest"`
	DeclarationCount int    `json:"declaration_count,omitempty"`
	Reconstructed    bool   `json:"reconstructed,omitempty"`
}

type EvidenceRecord struct {
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

type Counterexample struct {
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	Trigger          string `json:"trigger"`
	Mutation         string `json:"mutation"`
	TargetCoordinate string `json:"target_coordinate"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	ClaimBefore   string `json:"claim_before"`
	ClaimAfter    string `json:"claim_after"`
	Observed      int    `json:"observed"`
	Expected      int    `json:"expected"`
	Satisfied     bool   `json:"satisfied"`
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

type ClaimTransition struct {
	IndicatorID    string `json:"indicator_id"`
	Audience       string `json:"audience"`
	Before         string `json:"before"`
	After          string `json:"after"`
	Visibility     string `json:"visibility"`
	EvidenceDigest string `json:"evidence_digest"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
}

type Coordinates struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type Summary struct {
	Coordinates              Coordinates `json:"coordinates"`
	RecordsObserved          int         `json:"records_observed"`
	CounterexamplesExecuted  int         `json:"counterexamples_executed"`
	MissingCoordinates       int         `json:"missing_coordinates"`
	ContradictoryCoordinates int         `json:"contradictory_coordinates"`
	SourceDenominator        int         `json:"source_denominator"`
}

type CounterexampleResult struct {
	ID         string               `json:"id"`
	Kind       string               `json:"kind"`
	Trigger    string               `json:"trigger"`
	Mutation   string               `json:"mutation"`
	Global     string               `json:"global_decision"`
	Views      []CounterexampleView `json:"views"`
	Transition string               `json:"transition"`
	Reason     string               `json:"reason"`
}

type CounterexampleView struct {
	Audience        string `json:"audience"`
	Before          string `json:"before"`
	After           string `json:"after"`
	LocalDecision   string `json:"local_decision"`
	LocalResolution string `json:"local_resolution"`
}

type Receipt struct {
	Schema           string                 `json:"schema"`
	ContractID       string                 `json:"contract_id"`
	Subject          string                 `json:"subject"`
	Decision         string                 `json:"decision"`
	Resolution       string                 `json:"resolution"`
	Reason           string                 `json:"reason"`
	Source           SourceBinding          `json:"source"`
	Summary          Summary                `json:"summary"`
	Indicators       []Indicator            `json:"indicators"`
	Views            []AudienceView         `json:"views"`
	Counterexamples  []CounterexampleResult `json:"counterexamples"`
	ClaimTransitions []ClaimTransition      `json:"claim_transitions"`
	NotClaimed       []string               `json:"not_claimed"`
	FactsDigest      string                 `json:"facts_digest"`
	Digest           string                 `json:"digest"`
}
