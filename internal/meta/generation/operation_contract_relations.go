package generation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func operationContractRelation(ir semantic.IR, activity *syntax.ActivityDecl, name, inputRole, outputRole string,
	seen map[semantic.FactKey]struct{}) (operationInputContractFacts, error) {
	if activity == nil || len(activity.Inputs) != 1 {
		return operationInputContractFacts{}, fmt.Errorf("operation input contract activity %q signature is not exact", name)
	}
	input, inputOK := ir.Graph.NodeByName(ir.Namespace, activity.Inputs[0].Name)
	output, outputOK := ir.Graph.NodeByName(ir.Namespace, activity.Output)
	node, nodeOK := ir.Graph.NodeByName(ir.Namespace, name)
	if !inputOK || !outputOK || !nodeOK || input.Kind != semantic.Entity ||
		output.Kind != semantic.Entity || node.Kind != semantic.Activity ||
		input.ID.String() != NativeContractEntityID(inputRole) || output.ID.String() != NativeContractEntityID(outputRole) {
		return operationInputContractFacts{}, fmt.Errorf("operation input contract activity %q type identities are not exact", name)
	}
	used := semantic.FactKey{Subject: node.ID, Predicate: semantic.Used, Object: input.ID}
	generated := semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: node.ID}
	if !ir.Graph.HasFact(used) || !ir.Graph.HasFact(generated) {
		return operationInputContractFacts{}, fmt.Errorf("operation input contract facts are incomplete for %q", name)
	}
	if _, duplicate := seen[used]; duplicate {
		return operationInputContractFacts{}, fmt.Errorf("operation input contract activity %q is duplicated in the native ABI", name)
	}
	seen[used], seen[generated] = struct{}{}, struct{}{}
	return operationInputContractFacts{UsedInput: true, GeneratedOutput: true}, nil
}
