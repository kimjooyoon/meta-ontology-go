package extractor

// StrategyEvidence records the proof obligations consumed by a native
// extraction strategy. It is intentionally attached to the existing
// ExtractFunction result rather than creating a second receipt system.
type StrategyEvidence struct {
	Strategy                 string                 `json:"strategy"`
	Operation                string                 `json:"operation"`
	ContractActivity         string                 `json:"contract_activity"`
	ContractInputEntity      string                 `json:"contract_input_entity"`
	ContractOutputEntity     string                 `json:"contract_output_entity"`
	ContractInputSubjectKind string                 `json:"contract_input_subject_kind"`
	ContractSourceDigest     string                 `json:"contract_source_digest"`
	ContractSemanticDigest   string                 `json:"contract_semantic_digest"`
	UsedInputFact            bool                   `json:"used_input_fact"`
	GeneratedOutputFact      bool                   `json:"generated_output_fact"`
	Subject                  string                 `json:"subject"`
	Helper                   string                 `json:"helper"`
	BeforeBytes              int                    `json:"before_bytes"`
	AfterBytes               int                    `json:"after_bytes"`
	BeforeFunctionLines      int                    `json:"before_function_lines"`
	AfterFunctionLines       int                    `json:"after_function_lines"`
	RenderedHelperBytes      int                    `json:"rendered_helper_bytes"`
	RenderedHelperLines      int                    `json:"rendered_helper_lines"`
	RenderedOuterHelperBytes  int                    `json:"rendered_outer_helper_bytes"`
	RenderedOuterHelperLines  int                    `json:"rendered_outer_helper_lines"`
	Obligations              []ObligationEvidence  `json:"obligations"`
	ContractObligations      []ContractObligationEvidence `json:"contract_obligations"`
}

type ObligationEvidence struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type ContractObligationEvidence struct {
	Name               string `json:"name"`
	Activity           string `json:"activity"`
	InputEntity        string `json:"input_entity"`
	OutputEntity       string `json:"output_entity"`
	UsedInputFact      bool   `json:"used_input_fact"`
	GeneratedOutputFact bool  `json:"generated_output_fact"`
}

const (
	obligationReturnShape       = "return-shape"
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
