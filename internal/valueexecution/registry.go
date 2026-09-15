package valueexecution

import "math"

type registeredOperation struct {
	Spec  OperationSpec
	Apply func(int64, int64) (int64, error)
}

var operationRegistry = []registeredOperation{
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "int.add", Version: 1, Arity: 1,
		InputEntities: []string{IntegerEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: IntegerEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch, ReasonIntegerOverflow},
	}, Apply: checkedAdd},
}

func operationByID(id string) (registeredOperation, bool) {
	for _, operation := range operationRegistry {
		if operation.Spec.ID == id {
			operation.Spec = cloneOperationSpec(operation.Spec)
			return operation, true
		}
	}
	return registeredOperation{}, false
}

func operationIDs() []string {
	result := make([]string, len(operationRegistry))
	for index, operation := range operationRegistry {
		result[index] = operation.Spec.ID
	}
	return result
}

func CanonicalOperationSpecs() []OperationSpec {
	result := make([]OperationSpec, len(operationRegistry))
	for index, operation := range operationRegistry {
		result[index] = cloneOperationSpec(operation.Spec)
	}
	return result
}

func checkedAdd(input, operand int64) (int64, error) {
	if operand > 0 && input > math.MaxInt64-operand {
		return 0, failAt(ReasonIntegerOverflow, "EXECUTE", "apply-int-add", "positive int64 overflow")
	}
	if operand < 0 && input < math.MinInt64-operand {
		return 0, failAt(ReasonIntegerOverflow, "EXECUTE", "apply-int-add", "negative int64 overflow")
	}
	return input + operand, nil
}
