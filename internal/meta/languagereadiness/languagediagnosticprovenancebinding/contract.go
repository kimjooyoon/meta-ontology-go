package languagediagnosticprovenancebinding

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagediagnosticprovenance"
)

const (
	Schema             = "gooo/language-diagnostic-provenance-readiness-binding/v1"
	FixedCoordinates   = 12
	ExpectedConcepts   = 18
	ExpectedCompleted  = 18
	ExpectedTotal      = 24
	ExpectedBPS        = 7500
	ExpectedCases      = 18
	ExpectedIndicators = 18
)

type Input struct {
	ExpectedHeadSHA string
	Concept         languageconcept.Artifact
	Readiness       languagereadiness.Snapshot
	Provenance      languagediagnosticprovenance.Report
}

type Coordinate struct {
	ID       string `json:"id"`
	Expected string `json:"expected"`
	Observed string `json:"observed"`
	Bound    bool   `json:"bound"`
}

type Summary struct {
	Coordinates         int `json:"coordinates"`
	BoundCoordinates    int `json:"bound_coordinates"`
	Unresolved          int `json:"unresolved"`
	ReadinessCompleted  int `json:"readiness_completed"`
	ReadinessTotal      int `json:"readiness_total"`
	ReadinessBPS        int `json:"readiness_bps"`
	ProvenanceSatisfied int `json:"provenance_satisfied"`
	ProvenanceTotal     int `json:"provenance_total"`
	Concepts            int `json:"concepts"`
	MetricBindings      int `json:"metric_bindings"`
	EffectfulStages     int `json:"effectful_stages"`
	RepositoryWrites    int `json:"repository_writes"`
	MutationAuthorities int `json:"mutation_authorities"`
}
