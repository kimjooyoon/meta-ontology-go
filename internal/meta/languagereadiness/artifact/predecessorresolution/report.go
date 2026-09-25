package predecessorresolution

import (
	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

type AttemptReceipt struct {
	Depth       int                         `json:"depth"`
	AncestorSHA string                      `json:"ancestor_sha"`
	ParentSHA   string                      `json:"parent_sha,omitempty"`
	Selection   predecessorselection.Report `json:"selection"`
}

type Resolution struct {
	Depth           int                                 `json:"depth"`
	AncestorSHA     string                              `json:"ancestor_sha"`
	SelectionDigest string                              `json:"selection_digest"`
	Baseline        readinessartifact.BaselineReference `json:"baseline"`
}

type Summary struct {
	ObservedAttempts     int  `json:"observed_attempts"`
	MissingAttempts      int  `json:"missing_attempts"`
	SelectedAncestors    int  `json:"selected_ancestors"`
	SelectedDepth        int  `json:"selected_depth"`
	SearchLimit          int  `json:"search_limit"`
	ValidCandidates      int  `json:"valid_candidates"`
	AmbiguousCandidates  int  `json:"ambiguous_candidates"`
	RepositoryWrites     int  `json:"repository_writes"`
	ReadinessDeltaClaims *int `json:"readiness_delta_claims"`
	CoordinatesCompleted int  `json:"coordinates_completed"`
	CoordinatesTotal     int  `json:"coordinates_total"`
	BasisPoints          int  `json:"basis_points"`
}

type Indicator struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Value      int    `json:"value"`
	Comparator string `json:"comparator"`
	Target     int    `json:"target"`
	Passed     bool   `json:"passed"`
}

type Proof struct {
	ID     string `json:"id"`
	Choice string `json:"choice"`
	Passed bool   `json:"passed"`
}

type Report struct {
	Schema                  string           `json:"schema"`
	Repository              string           `json:"repository"`
	CurrentHeadSHA          string           `json:"current_head_sha"`
	ImmediatePredecessorSHA string           `json:"immediate_predecessor_sha"`
	Decision                string           `json:"decision"`
	Reason                  string           `json:"reason"`
	BlockingSelectionReason string           `json:"blocking_selection_reason,omitempty"`
	Attempts                []AttemptReceipt `json:"attempts"`
	Selected                *Resolution      `json:"selected,omitempty"`
	// Decision describes readiness/predecessor availability. Conformance is
	// intentionally separate: a well-formed fail-closed report can conform.
	Conformance  string      `json:"conformance"`
	Resolution   string      `json:"resolution"`
	Summary      Summary     `json:"summary"`
	Indicators   []Indicator `json:"indicators"`
	Proofs       []Proof     `json:"proofs"`
	ReportDigest string      `json:"report_digest"`
}
