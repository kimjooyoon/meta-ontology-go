package coupling

type Output struct {
	Schema                string              `json:"schema"`
	FixtureID             string              `json:"fixture_id"`
	InputDigest           string              `json:"input_digest"`
	Decision              Decision            `json:"decision"`
	Reason                Reason              `json:"reason"`
	AcceptedSurfaces      []string            `json:"accepted_surfaces"`
	ChangedSurfaces       []string            `json:"changed_surfaces"`
	ReceiptSurfaces       []string            `json:"receipt_surfaces"`
	SemanticBeforeDigest  string              `json:"semantic_before_digest"`
	SemanticAfterDigest   string              `json:"semantic_after_digest"`
	SemanticDeltaDigest   string              `json:"semantic_delta_digest,omitempty"`
	PathClosureDigest     string              `json:"path_closure_digest,omitempty"`
	ObservationCounts     ObservationCounts   `json:"observation_counts"`
	Resources             ResourceObservation `json:"resources"`
	CanonicalOutputDigest string              `json:"canonical_output_digest"`
	ReplayDigest          string              `json:"replay_digest"`
}

// BaselineResult is a separately implemented, typed-config/full-suite
// comparison. It intentionally does not reuse oracle validation helpers.
type BaselineResult struct {
	Decision          Decision            `json:"decision"`
	Reason            Reason              `json:"reason"`
	LocalizedSurfaces []string            `json:"localized_surfaces"`
	FullSuite         bool                `json:"full_suite"`
	ObservationCounts ObservationCounts   `json:"observation_counts"`
	Resources         ResourceObservation `json:"resources"`
	WorkUnits         uint64              `json:"work_units"`
}
type Comparison struct {
	Oracle            Output         `json:"oracle"`
	Baseline          BaselineResult `json:"baseline"`
	OutcomeMatch      bool           `json:"outcome_match"`
	ReasonMatch       bool           `json:"reason_match"`
	LocalizationMatch bool           `json:"localization_match"`
	Finding           string         `json:"finding"`
}
type FixtureExpectation struct {
	Decision          Decision            `json:"decision"`
	Reason            Reason              `json:"reason"`
	AcceptedSurfaces  []string            `json:"accepted_surfaces"`
	ChangedSurfaces   []string            `json:"changed_surfaces"`
	ReceiptSurfaces   []string            `json:"receipt_surfaces"`
	ObservationCounts ObservationCounts   `json:"observation_counts"`
	Resources         ResourceObservation `json:"resources"`
}
type CorpusCase struct {
	Name     string             `json:"name"`
	Input    Input              `json:"input"`
	Expected FixtureExpectation `json:"expected"`
}
type Corpus struct {
	Schema          string       `json:"schema"`
	CanonicalDigest string       `json:"canonical_digest"`
	Cases           []CorpusCase `json:"cases"`
}
