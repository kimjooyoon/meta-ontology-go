package main

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	projectionextractor "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
)

// These IDs describe the archived fixture's six-stage extraction contract.
// Other operations may add relations; adapting this decoder fixture still needs
// the same extraction signature and chain. This is not reuse of execution proof.
func assertHistoricalExtractionShape(t *testing.T, evidence projectionextractor.StrategyEvidence,
	contract generation.OperationInputContractEvidence) {
	t.Helper()
	roles := map[string]string{
		"FunctionInput": "function-input", "OperationResult": "result",
		"ReturnShapeObligation": "return-shape-obligation", "ControlFlowObligation": "control-flow-obligation",
		"FreeBindingsObligation": "free-bindings-obligation", "CalleeEffectsObligation": "callee-effects-obligation",
		"RenderedCapacityObligation":     "rendered-capacity-obligation",
		"ProjectedConformanceObligation": "projected-conformance-obligation",
	}
	id := func(role string) string {
		if slug := roles[role]; slug != "" {
			return "gooo://meta-operation-input-contract/entity/" + slug
		}
		return ""
	}
	if evidence.ContractSemanticDigest != "b9a780ed3ba67f58bd91a86e83280ef8b98fd3b38893dc11ddd69169c85e012e" ||
		evidence.Operation != string(contract.Operation) || evidence.ContractActivity != contract.Activity ||
		evidence.ContractInputSubjectKind != string(contract.InputSubjectKind) ||
		id(evidence.ContractInputEntity) != contract.InputEntityID || id(evidence.ContractOutputEntity) != contract.OutputEntityID ||
		len(contract.Obligations) != 6 || len(evidence.ContractObligations) != 6 || len(evidence.ProofStages) != 6 {
		t.Fatal("extraction fixture signature changed; review its schema before adapting it")
	}
	for i, current := range contract.Obligations {
		old, stage := evidence.ContractObligations[i], evidence.ProofStages[i]
		if old.Name != current.Name || old.Activity != current.Activity ||
			old.UsedInputFact != current.UsedInputFact || old.GeneratedOutputFact != current.GeneratedOutputFact ||
			id(old.InputEntity) != current.InputEntityID || id(old.OutputEntity) != current.OutputEntityID ||
			stage.Name != current.Name || stage.Activity != current.Activity ||
			id(stage.InputEntity) != current.InputEntityID || id(stage.OutputEntity) != current.OutputEntityID {
			t.Fatal("extraction fixture chain changed; review its schema before adapting it", current.Name)
		}
	}
}
