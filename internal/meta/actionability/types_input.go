package actionability

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metabinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

const (
	Schema        = "gooo/meta-actionability-report/v1"
	AuthorityPath = "examples/meta-actionability/main.gooo"
)

type input struct {
	metrics       metricsDocument
	binding       metabinding.Report
	metricsDigest string
	bindingDigest string
}

type metricsDocument struct {
	CommitSHA  string              `json:"commit_sha"`
	Repository string              `json:"repository"`
	Meta       sourcepolicy.Report `json:"meta"`
}

type Executor struct {
	Operation   string `json:"operation"`
	Activity    string `json:"activity"`
	ProofChoice string `json:"proof_choice"`
	Registry    string `json:"registry"`
	Executor    string `json:"executor"`
	Evaluator   string `json:"evaluator"`
}
