package artifactfeedback

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/semanticresolution"

type ResolutionInput struct {
	Feedback          Input                         `json:"feedback"`
	CurrentResolution semanticresolution.Resolution `json:"current_resolution"`
	Descents          int                           `json:"descents"`
}

type ResolutionReport struct {
	Schema           string                        `json:"schema"`
	Feedback         Report                        `json:"feedback"`
	SourceDecision   string                        `json:"source_decision"`
	Decision         string                        `json:"decision"`
	Reason           string                        `json:"reason"`
	FromResolution   semanticresolution.Resolution `json:"from_resolution"`
	ToResolution     semanticresolution.Resolution `json:"to_resolution"`
	PreviousDescents int                           `json:"previous_descents"`
	Descents         int                           `json:"descents"`
	NextOperation    string                        `json:"next_operation,omitempty"`
	RepositoryWrites int                           `json:"repository_writes"`
	Indicators       []ResolutionIndicator         `json:"indicators"`
	ReportDigest     string                        `json:"report_digest"`
}

type ResolutionIndicator struct {
	MetricID      string                            `json:"metric_id"`
	Class         semanticresolution.IndicatorClass `json:"class"`
	Target        int                               `json:"target"`
	Unit          string                            `json:"unit"`
	Relation      semanticresolution.Relation       `json:"relation"`
	ProofChoice   semanticresolution.ProofChoice    `json:"proof_choice"`
	Producer      string                            `json:"producer"`
	Consumer      string                            `json:"consumer"`
	MetaOperation string                            `json:"meta_operation"`
	Activity      string                            `json:"activity"`
	Value         int                               `json:"value"`
	Satisfied     bool                              `json:"satisfied"`
}
