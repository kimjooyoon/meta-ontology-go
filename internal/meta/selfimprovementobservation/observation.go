package selfimprovementobservation

const (
	observationSchema = "gooo/self-improvement-language-observation/v1"
	metaprogram       = "internal/meta/selfimprovementobservation"
)

type Options struct {
	HeadSHA     string
	SourceRunID int64
}

type Observation struct {
	Schema              string             `json:"schema"`
	Metaprogram         string             `json:"metaprogram"`
	SubjectSHA          string             `json:"subject_sha"`
	SourceWorkflowRunID int64              `json:"source_workflow_run_id"`
	ContractID          string             `json:"contract_id"`
	Decision            string             `json:"decision"`
	Resolution          string             `json:"resolution"`
	Reason              string             `json:"reason"`
	Interpretation      string             `json:"interpretation"`
	Summary             ObservationSummary `json:"summary"`
	Authority           Authority          `json:"authority"`
	Artifacts           []ArtifactRef      `json:"artifacts"`
	InputDigest         string             `json:"input_digest"`
	Indicators          []Indicator        `json:"indicators"`
	Views               []View             `json:"views"`
	Proofs              []Proof            `json:"proofs"`
	NotClaimed          []string           `json:"not_claimed"`
	Digest              string             `json:"digest"`
}

type ObservationSummary struct {
	Coordinates         CountSummary `json:"coordinates"`
	SourceCoordinates   CountSummary `json:"source_coordinates"`
	Counterexamples     CountSummary `json:"counterexamples"`
	GoooDefinitionFiles int          `json:"gooo_definition_files"`
	GoDefinitionFiles   int          `json:"go_definition_files"`
	ResourceSamples     int          `json:"resource_samples"`
	MaxWallMS           int64        `json:"max_wall_ms"`
	MaxRSSKiB           int64        `json:"max_rss_kib"`
	BinaryBytes         int64        `json:"binary_bytes"`
	CandidateCount      int          `json:"candidate_count"`
	Unknowns            int          `json:"unknowns"`
}
