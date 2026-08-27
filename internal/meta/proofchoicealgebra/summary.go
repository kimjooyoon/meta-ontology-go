package proofchoicealgebra

type Summary struct {
	Claims                int           `json:"claims"`
	Metrics               int           `json:"metrics"`
	Observations          int           `json:"observations"`
	Items                 int           `json:"items"`
	ChoiceCounts          map[Route]int `json:"choice_counts"`
	ExactChoices          int           `json:"exact_choices"`
	ChoiceCoverageBPS     int           `json:"choice_coverage_bps"`
	FixedDenominator      int           `json:"fixed_denominator"`
	MetricSlotNumerator   int           `json:"metric_slot_numerator"`
	MetricSlotDenominator int           `json:"metric_slot_denominator"`
	Transitions           int           `json:"transitions"`
	Discharged            int           `json:"discharged"`
	OpenPreserved         int           `json:"open_preserved"`
	Refuted               int           `json:"refuted"`
	CaseDenominator       int           `json:"case_denominator"`
	RouteDenominator      int           `json:"route_denominator"`
	ClaimDenominator      int           `json:"claim_denominator"`
	InterventionDenom     int           `json:"intervention_denominator"`
}

type Reconstruction struct {
	Numerator   int `json:"numerator"`
	Denominator int `json:"denominator"`
}

type Effects struct {
	Observed           bool   `json:"observed"`
	BeforeStatusDigest string `json:"before_status_digest"`
	AfterStatusDigest  string `json:"after_status_digest"`
	RepositoryWrites   int    `json:"repository_writes"`
	MutationAuthority  bool   `json:"mutation_authority"`
}
