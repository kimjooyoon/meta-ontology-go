package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

// These three native capabilities admit the authored Gooo prerequisite chain.
// They do not turn a readonly capacity observation into projected conformance.
var nativeGoSplitObligations = [...]nativeOperationObligation{
	{Name: "rendered-capacity", Activity: "ProveGoSplitRenderedCapacity",
		ActivityID:  "gooo://meta-operation-input-contract/activity/go-split-rendered-capacity",
		InputEntity: "FileInput", OutputEntity: "RenderedCapacityObligation"},
	{Name: "directory-capacity", Activity: "ProveGoSplitDirectoryCapacity",
		ActivityID:  "gooo://meta-operation-input-contract/activity/go-split-directory-capacity",
		InputEntity: "RenderedCapacityObligation", OutputEntity: "DirectoryCapacityObligation"},
	{Name: "projected-conformance", Activity: "ProveGoSplitProjectedConformance",
		ActivityID:  "gooo://meta-operation-input-contract/activity/go-split-projected-conformance",
		InputEntity: "DirectoryCapacityObligation", OutputEntity: "ProjectedConformanceObligation"},
}

func allNativeOperationObligations() []nativeOperationObligation {
	return append(append([]nativeOperationObligation{}, nativeOperationObligations...), nativeGoSplitObligations[:]...)
}

// GoSplitInputContractEvidence retains the same source/IR/origin envelope as
// extraction while exposing only the three split obligations in source order.
func GoSplitInputContractEvidence() (OperationInputContractEvidence, error) {
	return operationInputContractEvidence(sourcepolicy.OperationSplitGo, nativeGoSplitObligations[:])
}
