package bodycodegen

import "github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"

// RecordObservationReceipt separates questions, acquired answers and the finite
// candidate population from the subsequent selection and held evaluation.
type RecordObservationReceipt struct {
	Schema               string                   `json:"schema"`
	Status               string                   `json:"status"`
	SourceSHA256         string                   `json:"source_sha256"`
	ContractSHA256       string                   `json:"contract_sha256"`
	Options              assemblyspec.Observation `json:"options"`
	OptionsSHA256        string                   `json:"options_sha256"`
	InitialCases         []assemblyspec.ValueCase `json:"initial_cases"`
	EffectiveCasesSHA256 string                   `json:"effective_cases_sha256"`
	EffectiveCases       int                      `json:"effective_cases"`
	OracleActivityID     string                   `json:"oracle_activity_id,omitempty"`
	OracleGoSHA256       string                   `json:"oracle_go_sha256,omitempty"`
	OracleInitialCases   []RecordAssemblyCase     `json:"oracle_initial_cases,omitempty"`
	OracleEvaluations    int                      `json:"oracle_evaluations"`
	CandidateEvaluations int                      `json:"candidate_evaluations"`
	ReusedOutputs        int                      `json:"reused_outputs"`
	Rounds               []RecordObservationRound `json:"rounds"`
	ResolvedMask         *uint16                  `json:"resolved_mask,omitempty"`
	Scope                string                   `json:"scope"`
}

type RecordObservationRound struct {
	Candidates       int                      `json:"candidates"`
	Unobserved       int                      `json:"unobserved"`
	SurvivingMasks   []uint16                 `json:"surviving_masks"`
	UnavailableMasks []uint16                 `json:"unavailable_masks,omitempty"`
	Unavailable      []RecordProbeOutput      `json:"unavailable,omitempty"`
	Probes           []RecordObservationProbe `json:"probes"`
	RecommendedIndex *int                     `json:"recommended_index,omitempty"`
	Observation      *assemblyspec.ValueCase  `json:"oracle_observation,omitempty"`
}

type RecordObservationProbe struct {
	Input          string              `json:"input"`
	AlreadyTested  bool                `json:"already_tested"`
	SeparatedPairs int                 `json:"separated_pairs"`
	Outputs        []RecordProbeOutput `json:"outputs,omitempty"`
}

type RecordProbeOutput struct {
	Mask  uint16 `json:"mask"`
	Value string `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}
