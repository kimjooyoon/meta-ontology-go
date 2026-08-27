package proofchoicealgebra

import "fmt"

const (
	Schema                  = "gooo/proof-choice-algebra-receipt/v1"
	FixedDenominator        = 3
	Claim            Kind   = "CLAIM"
	Metric           Kind   = "METRIC"
	Pass                    = "PASS"
	FailClosed              = "FAIL_CLOSED"
	Exact                   = "EXACT"
	Foundation       Choice = "FOUNDATION"
	Coherence        Choice = "COHERENCE"
	Regression       Choice = "REGRESSION"
)

type Kind string
type Choice string

func (c Choice) Valid() bool {
	return c == Foundation || c == Coherence || c == Regression
}

func (c Choice) String() string { return string(c) }

// Item is a claim or metric carrying one explicit proof-choice meta value.
// The fields are deliberately domain-neutral: they describe evidence routing,
// not a replicated theorem-prover type system.
type Item struct {
	Kind          Kind   `json:"kind"`
	ID            string `json:"id"`
	Statement     string `json:"statement"`
	Choice        Choice `json:"choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Numerator     int    `json:"numerator,omitempty"`
	Denominator   int    `json:"denominator,omitempty"`
	Line          int    `json:"line"`
}

// Transition records a persistent claim moving between lifecycle states. Its
// choice is immutable across the transition and is checked against ClaimID.
type Transition struct {
	ClaimID       string `json:"claim_id"`
	From          string `json:"from"`
	To            string `json:"to"`
	Choice        Choice `json:"choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Persistent    bool   `json:"persistent"`
	Line          int    `json:"line"`
}

type Bundle struct {
	Items       []Item
	Transitions []Transition
}

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

type issue struct {
	Reason string
	Line   int
}

func (i issue) Error() string {
	if i.Line > 0 {
		return fmt.Sprintf("%s at line %d", i.Reason, i.Line)
	}
	return i.Reason
}
