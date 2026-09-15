package languagegointeroperationbinding

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagegointeroperation"
)

const (
	Schema               = "gooo/language-go-interoperation-readiness-binding/v1"
	FixedCoordinates     = 12
	ExpectedConcepts     = 17
	ExpectedCompleted    = 17
	ExpectedTotal        = 24
	ExpectedBPS          = 7083
	ExpectedInteropCases = 24
	ExpectedIndicators   = 18
)

type Input struct {
	ExpectedHeadSHA string
	Concept         languageconcept.Artifact
	Readiness       languagereadiness.Snapshot
	Interoperation  languagegointeroperation.Report
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
	InteropSatisfied    int `json:"interop_satisfied"`
	InteropTotal        int `json:"interop_total"`
	Concepts            int `json:"concepts"`
	MetricBindings      int `json:"metric_bindings"`
	EffectfulStages     int `json:"effectful_stages"`
	RepositoryWrites    int `json:"repository_writes"`
	MutationAuthorities int `json:"mutation_authorities"`
}
