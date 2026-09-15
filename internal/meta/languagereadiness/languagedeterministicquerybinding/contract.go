package languagedeterministicquerybinding

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagedeterministicquery"
)

const (
	Schema             = "gooo/language-deterministic-query-readiness-binding/v1"
	FixedCoordinates   = 12
	ExpectedConcepts   = 16
	ExpectedCompleted  = 16
	ExpectedTotal      = 24
	ExpectedBPS        = 6666
	ExpectedQueryCases = 32
	ExpectedIndicators = 18
)

type Input struct {
	ExpectedHeadSHA string
	Concept         languageconcept.Artifact
	Readiness       languagereadiness.Snapshot
	Query           languagedeterministicquery.Report
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
	QuerySatisfied      int `json:"query_satisfied"`
	QueryTotal          int `json:"query_total"`
	Concepts            int `json:"concepts"`
	MetricBindings      int `json:"metric_bindings"`
	EffectfulStages     int `json:"effectful_stages"`
	RepositoryWrites    int `json:"repository_writes"`
	MutationAuthorities int `json:"mutation_authorities"`
}
