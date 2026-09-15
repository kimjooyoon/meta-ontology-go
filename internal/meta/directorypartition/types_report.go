package directorypartition

import "encoding/json"

const ReportSchema = "gooo/directory-partition-report/v1"

type Report struct {
	Schema              string      `json:"schema"`
	Repository          string      `json:"repository"`
	SubjectSHA          string      `json:"subject_sha"`
	SourceMetricsDigest string      `json:"source_metrics_digest"`
	OntologyDigest      string      `json:"ontology_digest"`
	Decision            string      `json:"decision"`
	Reason              string      `json:"reason"`
	RootPolicy          RootPolicy  `json:"root_policy"`
	Summary             Summary     `json:"summary"`
	Indicators          []Indicator `json:"indicators"`
	Candidates          []Candidate `json:"candidates"`
	Proofs              []Proof     `json:"proofs"`
	PlanDigest          string      `json:"plan_digest"`
	Digest              string      `json:"digest"`
}

type RootPolicy struct {
	CountsApplicability   string `json:"counts_applicability"`
	TopologyApplicability string `json:"topology_applicability"`
	TopologyReason        string `json:"topology_reason"`
	READMERequirement     string `json:"readme_requirement"`
}

type Summary struct {
	ApplicableIndicators  int  `json:"applicable_indicators"`
	ViolatingIndicators   int  `json:"violating_indicators"`
	PlannedDirectories    int  `json:"planned_directories"`
	RequiredEntries       int  `json:"required_entries"`
	PlannedEntries        int  `json:"planned_entries"`
	ProjectRootExemptions int  `json:"project_root_exemptions"`
	RepositoryWrites      int  `json:"repository_writes"`
	ReplayVerified        bool `json:"replay_verified"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Satisfied     bool   `json:"satisfied"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Activity      string `json:"activity"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	Activity       string `json:"activity"`
	Satisfied      bool   `json:"satisfied"`
	EvidenceDigest string `json:"evidence_digest"`
}

func (report Report) JSON() ([]byte, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	return append(data, '\n'), err
}
