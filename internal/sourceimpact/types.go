// Package sourceimpact compares source-owned declarations and follows potential
// dependencies independently in each revision. It never executes a body.
package sourceimpact

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

const Schema = "gooo/source-change-impact/v1"
const maxDeclarations = 4096
const maxEdges = 32768
const maxSourceBytes = 2 << 20

type Edge struct {
	Dependency string `json:"dependency"`
	Dependent  string `json:"dependent"`
	Kind       string `json:"kind"`
	Surface    string `json:"surface"`
}
type Unknown struct {
	Revision   string `json:"revision"`
	ActivityID string `json:"activity_id"`
	Surface    string `json:"surface"`
	Reference  string `json:"reference,omitempty"`
	Reason     string `json:"reason"`
}
type Change struct {
	ID         string   `json:"id"`
	BeforeName string   `json:"before_name,omitempty"`
	AfterName  string   `json:"after_name,omitempty"`
	Facets     []string `json:"facets"`
}
type Cause struct {
	Revision  string `json:"revision"`
	ChangedID string `json:"changed_id"`
	Depth     int    `json:"depth"`
	Via       *Edge  `json:"via,omitempty"`
}
type Obligations struct {
	Cases                   int    `json:"selection_cases"`
	Conditions              int    `json:"condition_cases"`
	Holdouts                int    `json:"holdout_cases"`
	ObservationInputs       int    `json:"observation_inputs"`
	Oracle                  string `json:"oracle,omitempty"`
	MaxAttempts             int    `json:"max_attempts,omitempty"`
	FillCandidates          int    `json:"declared_fill_candidates,omitempty"`
	GeneratedCandidateLimit int    `json:"generated_candidate_limit,omitempty"`
}
type Affected struct {
	ID                string       `json:"id"`
	Name              string       `json:"name"`
	PresentAfter      bool         `json:"present_after"`
	Causes            []Cause      `json:"causes"`
	BeforeObligations *Obligations `json:"before_obligations,omitempty"`
	AfterObligations  *Obligations `json:"after_obligations,omitempty"`
	State             string       `json:"state"`
}
type ScopeChange struct {
	Field        string `json:"field"`
	BeforeSHA256 string `json:"before_sha256"`
	AfterSHA256  string `json:"after_sha256"`
}

type Report struct {
	ScopeChanges        []ScopeChange `json:"scope_changes"`
	Schema              string        `json:"schema"`
	Status              string        `json:"status"`
	BeforeSourceSHA256  string        `json:"before_source_sha256"`
	AfterSourceSHA256   string        `json:"after_source_sha256"`
	Changes             []Change      `json:"changes"`
	Affected            []Affected    `json:"affected"`
	Unknowns            []Unknown     `json:"unknowns"`
	BeforeDeclarations  int           `json:"before_declarations"`
	AfterDeclarations   int           `json:"after_declarations"`
	ModelCalls          int           `json:"model_calls"`
	CandidateExecutions int           `json:"candidate_executions"`
	Scope               string        `json:"scope"`
	ReportSHA256        string        `json:"report_sha256,omitempty"`
}
type snapshot struct {
	scope    map[string]string
	nodes    map[string]semantic.Node
	edges    []Edge
	unknowns []Unknown
	facets   map[string]map[string]string
}
