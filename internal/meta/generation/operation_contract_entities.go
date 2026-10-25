package generation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// Roles belong to the native ABI. Source labels may change; the IDs bind the
// admitted Gooo entities to these existing host roles.
var nativeContractEntities = [...]struct{ Role, ID string }{
	{"FunctionInput", "gooo://meta-operation-input-contract/entity/function-input"},
	{"FileInput", "gooo://meta-operation-input-contract/entity/file-input"},
	{"OperationResult", "gooo://meta-operation-input-contract/entity/result"},
	{"ImportNormalizationPolicy", "gooo://meta-operation-input-contract/entity/import-normalization-policy"},
	{"ReturnShapeObligation", "gooo://meta-operation-input-contract/entity/return-shape-obligation"},
	{"ControlFlowObligation", "gooo://meta-operation-input-contract/entity/control-flow-obligation"},
	{"FreeBindingsObligation", "gooo://meta-operation-input-contract/entity/free-bindings-obligation"},
	{"CalleeEffectsObligation", "gooo://meta-operation-input-contract/entity/callee-effects-obligation"},
	{"RenderedCapacityObligation", "gooo://meta-operation-input-contract/entity/rendered-capacity-obligation"},
	{"ProjectedConformanceObligation", "gooo://meta-operation-input-contract/entity/projected-conformance-obligation"},
}

type contractEntityReference struct{ ID, Name string }

// NativeContractEntityID resolves a host ABI role, not an authored display name.
// Unknown roles have no identity and cannot introduce an input-kind alias.
func NativeContractEntityID(role string) string {
	for _, entity := range nativeContractEntities {
		if entity.Role == role {
			return entity.ID
		}
	}
	return ""
}

func operationContractEntities(ir semantic.IR) ([len(nativeContractEntities)]contractEntityReference, error) {
	var entities [len(nativeContractEntities)]contractEntityReference
	for index, native := range nativeContractEntities {
		node, ok := ir.Graph.Node(semantic.ID(native.ID))
		if !ok || node.Kind != semantic.Entity {
			return [len(nativeContractEntities)]contractEntityReference{},
				fmt.Errorf("operation input contract entity identity %q is not exact", native.ID)
		}
		entities[index] = contractEntityReference{ID: node.ID.String(), Name: node.Name}
	}
	return entities, nil
}

func (contract operationInputContract) entity(role string) contractEntityReference {
	for index, native := range nativeContractEntities {
		if native.Role == role {
			return contract.Entities[index]
		}
	}
	return contractEntityReference{}
}
