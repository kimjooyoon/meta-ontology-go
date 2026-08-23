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
	Coordinates        int `json:"coordinates"`
	BoundCoordinates   int `json:"bound_coordinates"`
	Unresolved         int `json:"unresolved"`
	ReadinessCompleted int `json:"readiness_completed"`
	ReadinessTotal     int `json:"readiness_total"`
	ReadinessBPS       int `json:"readiness_bps"`
	QuerySatisfied     int `json:"query_satisfied"`
	QueryTotal         int `json:"query_total"`
	Concepts           int `json:"concepts"`
	MetricBindings     int `json:"metric_bindings"`
	EffectfulStages    int `json:"effectful_stages"`
	RepositoryWrites   int `json:"repository_writes"`
	MutationAuthorities int `json:"mutation_authorities"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Artifact struct {
	Schema             string       `json:"schema"`
	Decision           string       `json:"decision"`
	Resolution         string       `json:"resolution"`
	ReasonCode         string       `json:"reason_code"`
	ExpectedHeadSHA    string       `json:"expected_head_sha"`
	ConceptDigest      string       `json:"concept_digest"`
	ReadinessDigest    string       `json:"readiness_digest"`
	QueryDigest        string       `json:"query_digest"`
	Coordinates        []Coordinate `json:"coordinates"`
	Summary            Summary      `json:"summary"`
	Indicators         []Indicator  `json:"indicators"`
	Proofs             []Proof      `json:"proofs"`
	RepositoryWrites   int          `json:"repository_writes"`
	MutationAuthorized bool         `json:"mutation_authorized"`
	ArtifactDigest     string       `json:"artifact_digest"`
}
