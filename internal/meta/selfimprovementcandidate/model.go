package selfimprovementcandidate

import valuewitnessinput "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementvaluewitnessinput"

type Coordinate struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type Authority struct {
	RepositoryWrites            int  `json:"repository_writes"`
	MutationAuthorized          bool `json:"mutation_authorized"`
	ExecutionAuthorized         bool `json:"execution_authorized"`
	PromotionAuthorized         bool `json:"promotion_authorized"`
	AutomaticAdoptionAuthorized bool `json:"automatic_adoption_authorized"`
}

type Summary struct {
	Coordinates       Coordinate `json:"coordinates"`
	SourceCoordinates Coordinate `json:"source_coordinates"`
	EligibleGaps      int        `json:"eligible_gaps"`
	CandidateCount    int        `json:"candidate_count"`
	AchievedDelta     int        `json:"achieved_delta"`
	TargetDelta       int        `json:"target_delta"`
	Unknowns          int        `json:"unknowns"`
}

type Report struct {
	Schema                  string                            `json:"schema"`
	Metaprogram             string                            `json:"metaprogram"`
	SubjectSHA              string                            `json:"subject_sha"`
	SourceWorkflowRunID     int64                             `json:"source_workflow_run_id"`
	SourceObservationDigest string                            `json:"source_observation_digest"`
	SourceFileDigest        string                            `json:"source_file_digest"`
	Contract                ContractEvidence                  `json:"contract"`
	PolicyVersion           string                            `json:"policy_version"`
	Decision                string                            `json:"decision"`
	Resolution              string                            `json:"resolution"`
	Reason                  string                            `json:"reason"`
	Summary                 Summary                           `json:"summary"`
	Candidates              []Candidate                       `json:"candidates"`
	ExecutionInput          *valuewitnessinput.ExecutionInput `json:"execution_input,omitempty"`
	Authority               Authority                         `json:"authority"`
	Indicators              []Indicator                       `json:"indicators"`
	Views                   []View                            `json:"views"`
	Proofs                  []Proof                           `json:"proofs"`
	NotClaimed              []string                          `json:"not_claimed"`
	Digest                  string                            `json:"digest"`
}
