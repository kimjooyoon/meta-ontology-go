package generation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

// SplitGoooInputContractEvidence binds a workspace partition to the existing
// source-owned operation. This is its input/output ABI, not admission of every
// native consumer or a claim that the repository metric executor has a route.
func SplitGoooInputContractEvidence() (OperationInputContractEvidence, error) {
	contract, err := loadOperationInputContract()
	if err != nil {
		return OperationInputContractEvidence{}, err
	}
	native := nativeOperationInputFor(sourcepolicy.OperationSplitGooo)
	binding, bindingOK := contract.Bindings[native.Operation]
	facts, factsOK := contract.Facts[native.Operation]
	if !bindingOK || !factsOK || !facts.UsedInput || !facts.GeneratedOutput {
		return OperationInputContractEvidence{}, fmt.Errorf("SplitGoooSections input/output contract is incomplete")
	}
	input, output := contract.entity(native.InputEntity), contract.entity(native.OutputEntity)
	return OperationInputContractEvidence{
		Operation: native.Operation, Activity: native.Activity, ActivityID: binding.ActivityID,
		InputEntity: input.Name, OutputEntity: output.Name,
		InputEntityID: input.ID, OutputEntityID: output.ID,
		InputSubjectKind: binding.InputSubjectKind,
		SourceDigest:     contract.SourceDigest, SemanticDigest: contract.SemanticDigest,
		SourceOrigins: contract.SourceOrigins,
		UsedInputFact: facts.UsedInput, GeneratedOutputFact: facts.GeneratedOutput,
	}, nil
}
