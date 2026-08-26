package valueexecution

import "math"

type operationSpec struct {
	ID    string
	Arity int
	Apply func(int64, int64) (int64, error)
}

var operationRegistry = []operationSpec{
	{ID: "int.add", Arity: 1, Apply: checkedAdd},
}

func operationByID(id string) (operationSpec, bool) {
	for _, operation := range operationRegistry {
		if operation.ID == id {
			return operation, true
		}
	}
	return operationSpec{}, false
}

func operationIDs() []string {
	result := make([]string, len(operationRegistry))
	for index, operation := range operationRegistry {
		result[index] = operation.ID
	}
	return result
}

func checkedAdd(input, operand int64) (int64, error) {
	if operand > 0 && input > math.MaxInt64-operand {
		return 0, fail(ReasonIntegerOverflow, "positive int64 overflow")
	}
	if operand < 0 && input < math.MinInt64-operand {
		return 0, fail(ReasonIntegerOverflow, "negative int64 overflow")
	}
	return input + operand, nil
}
