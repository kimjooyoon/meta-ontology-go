package proofchoicealgebra

type Indicator struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Choice   Choice `json:"choice"`
	Decision string `json:"decision"`
	Relation string `json:"relation"`
	Value    string `json:"value"`
	Limit    string `json:"limit"`
}

type Summary struct {
	Claims                int `json:"claims"`
	Metrics               int `json:"metrics"`
	Items                 int `json:"items"`
	Transitions           int `json:"transitions"`
	PersistentTransitions int `json:"persistent_transitions"`
	ChoicesExplicit       int `json:"choices_explicit"`
	ChoiceCoverageBPS     int `json:"choice_coverage_bps"`
	FixedDenominator      int `json:"fixed_denominator"`
	Unknowns              int `json:"unknowns"`
	Contradictions        int `json:"contradictions"`
	Compositions          int `json:"compositions"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Receipt struct {
	Schema       string       `json:"schema"`
	Decision     string       `json:"decision"`
	Reason       string       `json:"reason"`
	Resolution   string       `json:"resolution"`
	SourcePath   string       `json:"source_path"`
	SourceDigest string       `json:"source_digest"`
	FixedDenom   int          `json:"fixed_denominator"`
	Items        []Item       `json:"items"`
	Transitions  []Transition `json:"transitions"`
	Indicators   []Indicator  `json:"indicators"`
	Summary      Summary      `json:"summary"`
	Effects      Effects      `json:"effects"`
	Digest       string       `json:"digest"`
}
