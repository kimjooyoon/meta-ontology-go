package extractor

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

// StrategyEvidence records the proof obligations consumed by a native
// extraction strategy. It is intentionally attached to the existing
// ExtractFunction result rather than creating a second receipt system.
type StrategyEvidence struct {
	Strategy                      string                         `json:"strategy"`
	Operation                     string                         `json:"operation"`
	ContractActivity              string                         `json:"contract_activity"`
	ContractInputEntity           string                         `json:"contract_input_entity"`
	ContractOutputEntity          string                         `json:"contract_output_entity"`
	ContractInputSubjectKind      string                         `json:"contract_input_subject_kind"`
	ContractSourceDigest          string                         `json:"contract_source_digest"`
	ContractSemanticDigest        string                         `json:"contract_semantic_digest"`
	UsedInputFact                 bool                           `json:"used_input_fact"`
	GeneratedOutputFact           bool                           `json:"generated_output_fact"`
	Subject                       string                         `json:"subject"`
	Helper                        string                         `json:"helper"`
	BeforeBytes                   int                            `json:"before_bytes"`
	AfterBytes                    int                            `json:"after_bytes"`
	BeforeFunctionLines           int                            `json:"before_function_lines"`
	AfterFunctionLines            int                            `json:"after_function_lines"`
	BeforeRenderedCapacityOverage int                            `json:"before_rendered_capacity_overage"`
	AfterRenderedCapacityOverage  int                            `json:"after_rendered_capacity_overage"`
	RenderedHelperBytes           int                            `json:"rendered_helper_bytes"`
	RenderedHelperLines           int                            `json:"rendered_helper_lines"`
	RenderedOuterHelperBytes      int                            `json:"rendered_outer_helper_bytes"`
	RenderedOuterHelperLines      int                            `json:"rendered_outer_helper_lines"`
	Obligations                   []ObligationEvidence           `json:"obligations"`
	ContractObligations           []ContractObligationEvidence   `json:"contract_obligations"`
	ProofStages                   []ProofStageEvidence           `json:"proof_stages"`
	PreflightObservations         []PreflightObservationEvidence `json:"preflight_observations,omitempty"`
	PreparationProgress           *PreparationProgressEvidence  `json:"preparation_progress,omitempty"`
	FinalRenderedCapacity         *FinalRenderedCapacityEvidence `json:"final_rendered_capacity,omitempty"`
	FinalGeneratedBytes           int                            `json:"final_generated_bytes"`
	FinalGeneratedEvidenceBytes   int                            `json:"final_generated_evidence_bytes"`
	FinalGeneratedUnits           int                            `json:"final_generated_units"`
}

// PreparationProgressEvidence records a strict capacity decrease while the
// extractor is still preparing a source. It is deliberately separate from the
// final proof chain: an intermediate candidate is not a bounded final result.
type PreparationProgressEvidence struct {
	Operation              string `json:"operation"`
	Activity               string `json:"activity"`
	InputEntity            string `json:"input_entity"`
	InputSubjectKind       string `json:"input_subject_kind"`
	ContractSourceDigest   string `json:"contract_source_digest"`
	ContractSemanticDigest string `json:"contract_semantic_digest"`
	Subject                string `json:"subject"`
	SourceDigest           string `json:"source_digest"`
	BeforeOverage          int    `json:"before_overage"`
	AfterOverage           int    `json:"after_overage"`
	Status                 string `json:"status"`
}

// FinalRenderedCapacityEvidence records the bounded measurement of the
// actual final generated package. It is not a preparation-progress receipt.
type FinalRenderedCapacityEvidence struct {
	Scope        string `json:"scope"`
	PayloadDigest string `json:"payload_digest"`
	Bytes        int    `json:"bytes"`
	Lines        int    `json:"lines"`
	Overage      int    `json:"overage"`
	Status       string `json:"status"`
}

type PreflightObservationEvidence struct {
	Operation                       string                    `json:"operation"`
	Activity                        string                    `json:"activity"`
	InputEntity                     string                    `json:"input_entity"`
	InputSubjectKind                string                    `json:"input_subject_kind"`
	Metric                          sourcepolicy.Dimension    `json:"metric"`
	HelperMeasurementScope          string                    `json:"helper_measurement_scope"`
	Subject                         string                    `json:"subject"`
	Receiver                        string                    `json:"receiver,omitempty"`
	FunctionStart                   string                    `json:"function_start"`
	FunctionEnd                     string                    `json:"function_end"`
	DeclarationStart                string                    `json:"declaration_start"`
	DeclarationEnd                  string                    `json:"declaration_end"`
	SourceDigest                    string                    `json:"source_digest"`
	ContractSourceDigest            string                    `json:"contract_source_digest"`
	ContractSemanticDigest          string                    `json:"contract_semantic_digest"`
	FunctionLines                   int                       `json:"function_lines"`
	FunctionRenderedCapacityOverage int                       `json:"function_rendered_capacity_overage"`
	FunctionStatus                  string                    `json:"function_status"`
	HelperLines                     *int                      `json:"helper_lines,omitempty"`
	HelperRenderedCapacityOverage   *int                      `json:"helper_rendered_capacity_overage,omitempty"`
	HelperStatus                    string                    `json:"helper_status"`
	FailureReason                   string                    `json:"failure_reason,omitempty"`
	Failure                         *PreflightFailureEvidence `json:"failure,omitempty"`
}

type PreflightFailureEvidence struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
	Diagnostics   []string `json:"diagnostics"`
}

type ObligationEvidence struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type ContractObligationEvidence struct {
	Name                string `json:"name"`
	Activity            string `json:"activity"`
	InputEntity         string `json:"input_entity"`
	OutputEntity        string `json:"output_entity"`
	UsedInputFact       bool   `json:"used_input_fact"`
	GeneratedOutputFact bool   `json:"generated_output_fact"`
}

type ProofStageEvidence struct {
	Name             string `json:"name"`
	Activity         string `json:"activity"`
	InputEntity      string `json:"input_entity"`
	OutputEntity     string `json:"output_entity"`
	Status           string `json:"status"`
	SourceDigest     string `json:"source_digest"`
	CandidateDigest  string `json:"candidate_digest"`
	InputEvidenceID  string `json:"input_evidence_id"`
	OutputEvidenceID string `json:"output_evidence_id"`
	PayloadDigest    string `json:"payload_digest"`
	PayloadBytes     int    `json:"payload_bytes"`
	Detail           string `json:"detail,omitempty"`
}

const (
	obligationReturnShape      = "return-shape"
	obligationControlFlow      = "control-flow"
	obligationFreeBindings     = "free-bindings"
	obligationCalleeEffects    = "callee-effects"
	obligationRenderedCapacity = "rendered-capacity"
	obligationProjectedConform = "projected-conformance"
)

var returnTailObligations = []string{
	obligationReturnShape,
	obligationControlFlow,
	obligationFreeBindings,
	obligationCalleeEffects,
	obligationRenderedCapacity,
	obligationProjectedConform,
}
