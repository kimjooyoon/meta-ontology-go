package bodycodegen

import "github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"

type CriteriaQuestionReceipt struct {
	Schema                string                       `json:"schema"`
	Status                string                       `json:"status"`
	SourceSHA256          string                       `json:"source_sha256"`
	ContractSHA256        string                       `json:"contract_sha256"`
	Options               assemblyspec.Observation     `json:"options"`
	InitialInputs         []string                     `json:"initial_inputs"`
	AcquiredInputs        []string                     `json:"acquired_inputs"`
	EffectiveInputsSHA256 string                       `json:"effective_inputs_sha256"`
	EffectiveInputs       int                          `json:"effective_inputs"`
	Bindings              []RecordCriterionBinding     `json:"bindings"`
	CandidateEvaluations  int                          `json:"candidate_evaluations"`
	PredicateEvaluations  int                          `json:"predicate_evaluations"`
	ReusedOutputs         int                          `json:"reused_outputs"`
	ReusedCriteriaResults int                          `json:"reused_criteria_results"`
	Outputs               []CriteriaQuestionOutput     `json:"outputs"`
	Evaluations           []CriteriaQuestionEvaluation `json:"evaluations"`
	Rounds                []CriteriaQuestionRound      `json:"rounds"`
	ResolvedMask          *uint16                      `json:"resolved_mask,omitempty"`
	Scope                 string                       `json:"scope"`
}

type CriteriaQuestionOutput struct {
	Mask       uint16                `json:"mask"`
	Input      string                `json:"input"`
	Value      string                `json:"value,omitempty"`
	Error      string                `json:"error,omitempty"`
	LocalCalls *LocalCallObservation `json:"local_calls,omitempty"`
}

type CriteriaQuestionEvaluation struct {
	Mask        uint16              `json:"mask"`
	OutputIndex int                 `json:"output_index"`
	Result      RecordCriteriaInput `json:"result"`
}

type CriteriaQuestionRound struct {
	RecommendationReason string                      `json:"recommendation_reason,omitempty"`
	Candidates           int                         `json:"candidates"`
	Unobserved           int                         `json:"unobserved"`
	SurvivingMasks       []uint16                    `json:"surviving_masks"`
	UnavailableMasks     []uint16                    `json:"unavailable_masks"`
	RejectedMasks        []uint16                    `json:"rejected_masks"`
	Evaluations          []CriteriaQuestionCandidate `json:"evaluations"`
	Probes               []CriteriaQuestionProbe     `json:"probes"`
	RecommendedIndex     *int                        `json:"recommended_index,omitempty"`
	AcquiredInput        string                      `json:"acquired_input,omitempty"`
}

type CriteriaQuestionCandidate struct {
	Mask             uint16 `json:"mask"`
	CriteriaIndices  []int  `json:"criteria_indices"`
	ExactCaseOutputs []int  `json:"exact_case_outputs,omitempty"`
	Status           string `json:"status"`
}

type CriteriaQuestionProbe struct {
	OutputEvidence *CriteriaQuestionOutputEvidence `json:"output_evidence,omitempty"`
	Input          string                          `json:"input"`
	AlreadyTested  bool                            `json:"already_tested"`
	SeparatedPairs int                             `json:"separated_pairs"`
	OutputIndices  []int                           `json:"output_indices,omitempty"`
}
