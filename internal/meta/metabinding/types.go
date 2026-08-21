package metabinding

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

const (
	Schema       = "gooo/meta-binding-report/v1"
	OntologyPath = "examples/meta-binding-coverage/main.gooo"
	MetricID     = "gooo.metric.meta.unbound-indicators.v1"
)

type input struct {
	document     metricsDocument
	raw          map[string]json.RawMessage
	sourceDigest string
}

type metricsDocument struct {
	CommitSHA  string              `json:"commit_sha"`
	Repository string              `json:"repository"`
	Meta       sourcepolicy.Report `json:"meta"`
}

type Binding struct {
	Operation   string `json:"operation"`
	Activity    string `json:"activity"`
	ProofChoice string `json:"proof_choice"`
	Registry    string `json:"registry"`
	Executor    string `json:"executor,omitempty"`
	Evaluator   string `json:"evaluator,omitempty"`
}

type Witness struct {
	Binding
	IndicatorCount int      `json:"indicator_count"`
	Bound          bool     `json:"bound"`
	Reasons        []string `json:"reasons,omitempty"`
}

type Summary struct {
	SourceIndicators    int            `json:"source_indicators"`
	RecursiveIndicators int            `json:"recursive_indicators"`
	BoundIndicators     int            `json:"bound_indicators"`
	UnboundIndicators   int            `json:"unbound_indicators"`
	CoverageBasisPoints int            `json:"coverage_basis_points"`
	RegistryOperations  int            `json:"registry_operations"`
	UsedOperations      int            `json:"used_operations"`
}

type Report struct {
	Schema              string                 `json:"schema"`
	CommitSHA           string                 `json:"commit_sha"`
	Repository          string                 `json:"repository"`
	SourceMetricsDigest string                 `json:"source_metrics_digest"`
	RegistryDigest      string                 `json:"registry_digest"`
	OntologyPath        string                 `json:"ontology_path"`
	OntologyDigest      string                 `json:"ontology_digest"`
	Decision            string                 `json:"decision"`
	Reason              string                 `json:"reason"`
	Summary             Summary                `json:"summary"`
	SelfIndicator       sourcepolicy.Indicator `json:"self_indicator"`
	Witnesses           []Witness              `json:"witnesses"`
	ReportDigest        string                 `json:"report_digest"`
}
