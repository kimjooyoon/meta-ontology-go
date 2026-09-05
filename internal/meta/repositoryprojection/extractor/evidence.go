package extractor

// StrategyEvidence records the proof obligations consumed by a native
// extraction strategy. It is intentionally attached to the existing
// ExtractFunction result rather than creating a second receipt system.
type StrategyEvidence struct {
	Strategy                    string                       `json:"strategy"`
	Operation                   string                       `json:"operation"`
	ContractActivity            string                       `json:"contract_activity"`
	ContractInputEntity         string                       `json:"contract_input_entity"`
	ContractOutputEntity        string                       `json:"contract_output_entity"`
	ContractInputSubjectKind    string                       `json:"contract_input_subject_kind"`
	ContractSourceDigest        string                       `json:"contract_source_digest"`
	ContractSemanticDigest      string                       `json:"contract_semantic_digest"`
	UsedInputFact               bool                         `json:"used_input_fact"`
	GeneratedOutputFact         bool                         `json:"generated_output_fact"`
	Subject                     string                       `json:"subject"`
	Helper                      string                       `json:"helper"`
	BeforeBytes                 int                          `json:"before_bytes"`
	AfterBytes                  int                          `json:"after_bytes"`
	BeforeFunctionLines         int                          `json:"before_function_lines"`
	AfterFunctionLines          int                          `json:"after_function_lines"`
	RenderedHelperBytes         int                          `json:"rendered_helper_bytes"`
	RenderedHelperLines         int                          `json:"rendered_helper_lines"`
	RenderedOuterHelperBytes    int                          `json:"rendered_outer_helper_bytes"`
	RenderedOuterHelperLines    int                          `json:"rendered_outer_helper_lines"`
	Obligations                 []ObligationEvidence         `json:"obligations"`
	ContractObligations         []ContractObligationEvidence `json:"contract_obligations"`
	ProofStages                 []ProofStageEvidence         `json:"proof_stages"`
	PreflightObservations       []PreflightObservationEvidence  `json:"preflight_observations,omitempty"`
	FinalGeneratedBytes         int                          `json:"final_generated_bytes"`
	FinalGeneratedEvidenceBytes int                          `json:"final_generated_evidence_bytes"`
	FinalGeneratedUnits         int                          `json:"final_generated_units"`
}

type PreflightObservationEvidence struct {
	Operation              string  `json:"operation"`
	Activity               string  `json:"activity"`
	InputEntity            string  `json:"input_entity"`
	InputSubjectKind       string  `json:"input_subject_kind"`
	Metric                 string  `json:"metric"`
	Subject                string  `json:"subject"`
	Receiver               string  `json:"receiver,omitempty"`
	FunctionStart          string  `json:"function_start"`
	FunctionEnd            string  `json:"function_end"`
	DeclarationStart       string  `json:"declaration_start"`
	DeclarationEnd         string  `json:"declaration_end"`
	SourceDigest           string  `json:"source_digest"`
	ContractSourceDigest   string  `json:"contract_source_digest"`
	ContractSemanticDigest string  `json:"contract_semantic_digest"`
	FunctionLines          int     `json:"function_lines"`
	FunctionStatus         string  `json:"function_status"`
	HelperLines            *int    `json:"helper_lines,omitempty"`
	HelperStatus           string  `json:"helper_status"`
	FailureReason          string  `json:"failure_reason,omitempty"`
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
