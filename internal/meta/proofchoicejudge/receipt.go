package proofchoicejudge

type summary struct {
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

type effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type receipt struct {
	Schema       string       `json:"schema"`
	Decision     string       `json:"decision"`
	Reason       string       `json:"reason"`
	Resolution   string       `json:"resolution"`
	SourcePath   string       `json:"source_path"`
	SourceDigest string       `json:"source_digest"`
	FixedDenom   int          `json:"fixed_denominator"`
	Items        []item       `json:"items"`
	Transitions  []transition `json:"transitions"`
	Indicators   []indicator  `json:"indicators"`
	Summary      summary      `json:"summary"`
	Effects      effects      `json:"effects"`
	Digest       string       `json:"digest"`
}

type Verdict struct {
	Schema         string `json:"schema"`
	Decision       string `json:"decision"`
	Reason         string `json:"reason"`
	ReceiptDigest  string `json:"receipt_digest"`
	ComputedDigest string `json:"computed_digest"`
	DigestMatch    bool   `json:"digest_match"`
	Items          int    `json:"items"`
	Transitions    int    `json:"transitions"`
	Independent    bool   `json:"independent"`
}
