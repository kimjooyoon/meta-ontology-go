package sourceauthority

type Contract struct {
	Schema           string      `json:"schema"`
	MetricID         string      `json:"metric_id"`
	MetaOperation    string      `json:"meta_operation"`
	ProofChoice      string      `json:"proof_choice"`
	DenominatorID    string      `json:"denominator_id"`
	AdoptionState    string      `json:"adoption_state"`
	ReadinessCredit  int         `json:"readiness_credit"`
	Measurement      Measurement `json:"measurement"`
	States           StateAxes   `json:"states"`
	Rules            []string    `json:"rules"`
	UnknownEvidence  FailureMode `json:"unknown_evidence"`
	EmptyDenominator FailureMode `json:"empty_denominator"`
	Scope            Scope       `json:"scope"`
}

type Measurement struct {
	Numerator   string `json:"numerator"`
	Denominator string `json:"denominator"`
	Unit        string `json:"unit"`
	Target      int    `json:"target"`
}

type StateAxes struct {
	Observation []string `json:"observation"`
	Resolution  []string `json:"resolution"`
	Enforcement []string `json:"enforcement"`
}

type FailureMode struct {
	Observation string `json:"observation"`
	Resolution  string `json:"resolution"`
	Enforcement string `json:"enforcement"`
	Reason      string `json:"reason"`
}

type Scope struct {
	AcceptedFactMode       string `json:"accepted_fact_mode"`
	SemanticInterpretation string `json:"semantic_interpretation"`
	LiveURLWithoutSnapshot string `json:"live_url_without_snapshot"`
}
