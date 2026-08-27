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
	Decision        string           `json:"decision"`
	Resolution      string           `json:"resolution"`
	Reason          string           `json:"reason"`
	Source          SourceBinding    `json:"source"`
	Records         []EvidenceRecord `json:"records"`
	Counterexamples []Counterexample `json:"counterexamples"`
}

type SourceBinding struct {
	Path             string `json:"path"`
	Kind             string `json:"kind"`
	Digest           string `json:"digest"`
	DeclarationCount int    `json:"declaration_count"`
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
	ClaimBefore   string `json:"claim_before"`
	ClaimAfter    string `json:"claim_after"`
	Decision      string `json:"decision"`
	Satisfied     bool   `json:"satisfied"`
}

type Counterexample struct {
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	Trigger          string `json:"trigger"`
	ExpectedDecision string `json:"expected_decision"`
	ObservedDecision string `json:"observed_decision"`
	Reason           string `json:"reason"`
	Blocked          bool   `json:"blocked"`
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
	Audience               string   `json:"audience"`
	Resolution             string   `json:"resolution"`
	Decision               string   `json:"decision"`
	Reason                 string   `json:"reason"`
	Satisfied              int      `json:"satisfied"`
	Total                  int      `json:"total"`
	BasisPoints            int      `json:"basis_points"`
	CoordinateIDs          []string `json:"coordinate_ids"`
	OmittedCoordinateCount int      `json:"omitted_coordinate_count"`
}

type ClaimTransition struct {
	IndicatorID string `json:"indicator_id"`
	Before      string `json:"before"`
	After       string `json:"after"`
	Reason      string `json:"reason"`
}

type Coordinates struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type Summary struct {
	Coordinates              Coordinates `json:"coordinates"`
	RecordsObserved          int         `json:"records_observed"`
	CounterexamplesBlocked   int         `json:"counterexamples_blocked"`
	MissingCoordinates       int         `json:"missing_coordinates"`
	ContradictoryCoordinates int         `json:"contradictory_coordinates"`
}

type Receipt struct {
	Schema           string            `json:"schema"`
	ContractID       string            `json:"contract_id"`
	Subject          string            `json:"subject"`
	Decision         string            `json:"decision"`
	Resolution       string            `json:"resolution"`
	Reason           string            `json:"reason"`
	Source           SourceBinding     `json:"source"`
	Summary          Summary           `json:"summary"`
	Indicators       []Indicator       `json:"indicators"`
	Views            []AudienceView    `json:"views"`
	Counterexamples  []Counterexample  `json:"counterexamples"`
	ClaimTransitions []ClaimTransition `json:"claim_transitions"`
	NotClaimed       []string          `json:"not_claimed"`
	FactsDigest      string            `json:"facts_digest"`
	Digest           string            `json:"digest"`
}
