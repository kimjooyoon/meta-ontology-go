package valueexecution

import "math"

type registeredOperation struct {
	Spec  OperationSpec
	Apply func(int64, int64) (int64, error)
}

var integerOperationRegistry = []registeredOperation{
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "int.add", Version: 1, Arity: 1,
		InputEntities: []string{IntegerEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: IntegerEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch, ReasonIntegerOverflow},
	}, Apply: checkedAdd},
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "int.sub", Version: 1, Arity: 1,
		InputEntities: []string{IntegerEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: IntegerEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch, ReasonIntegerOverflow},
	}, Apply: checkedSubtract},
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "int.mul", Version: 1, Arity: 1,
		InputEntities: []string{IntegerEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: IntegerEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch, ReasonIntegerOverflow},
	}, Apply: checkedMultiply},
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "int.div", Version: 1, Arity: 1,
		InputEntities: []string{IntegerEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: IntegerEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch, ReasonIntegerOverflow, ReasonIntegerDivisionByZero},
	}, Apply: checkedDivide},
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "int.mod", Version: 1, Arity: 1,
		InputEntities: []string{IntegerEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: IntegerEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch, ReasonIntegerModuloByZero},
	}, Apply: checkedModulo},
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "int.neg", Version: 1, Arity: 1,
		InputEntities: []string{IntegerEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: IntegerEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch, ReasonIntegerOverflow},
	}, Apply: checkedNegate},
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "int.abs", Version: 1, Arity: 1,
		InputEntities: []string{IntegerEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: IntegerEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch, ReasonIntegerOverflow},
	}, Apply: checkedAbsolute},
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "int.sign", Version: 1, Arity: 1,
		InputEntities: []string{IntegerEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: IntegerEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch, ReasonOperationIRInvalid},
	}, Apply: checkedSign},
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "int.max", Version: 1, Arity: 1,
		InputEntities: []string{IntegerEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: IntegerEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch},
	}, Apply: checkedMaximum},
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "int.min", Version: 1, Arity: 1,
		InputEntities: []string{IntegerEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: IntegerEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch},
	}, Apply: checkedMinimum},
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "int.iszero", Version: 1, Arity: 1,
		InputEntities: []string{IntegerEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: BooleanEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch, ReasonOperationIRInvalid},
	}, Apply: checkedIsZero},
}

var booleanOperationRegistry = []registeredOperation{
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "bool.not", Version: 1, Arity: 1,
		InputEntities: []string{BooleanEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: BooleanEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch, ReasonOperationIRInvalid},
	}, Apply: checkedBooleanNot},
	{Spec: OperationSpec{
		Schema: OperationSpecSchema, ID: "bool.and", Version: 1, Arity: 1,
		InputEntities: []string{BooleanEntity}, OperandKind: OperandInt64Literal,
		OutputEntity: BooleanEntity, Effect: EffectPureValue, Determinism: Deterministic,
		FailureReasons: []string{ReasonInputArityMismatch, ReasonOperationIRInvalid},
	}, Apply: checkedBooleanAnd},
}

var intEqualOperation = registeredOperation{Spec: OperationSpec{
	Schema: OperationSpecSchema, ID: "int.eq", Version: 1, Arity: 1,
	InputEntities: []string{IntegerEntity}, OperandKind: OperandInt64Literal,
	OutputEntity: BooleanEntity, Effect: EffectPureValue, Determinism: Deterministic,
	FailureReasons: []string{ReasonInputArityMismatch},
}, Apply: checkedEqual}

var operationRegistry = append(append([]registeredOperation{}, integerOperationRegistry...), append(booleanOperationRegistry, intEqualOperation)...)

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

func checkedSubtract(input, operand int64) (int64, error) {
	if operand > 0 && input < math.MinInt64+operand {
		return 0, failAt(ReasonIntegerOverflow, "EXECUTE", "apply-int-sub", "negative int64 overflow")
	}
	if operand < 0 && input > math.MaxInt64+operand {
		return 0, failAt(ReasonIntegerOverflow, "EXECUTE", "apply-int-sub", "positive int64 overflow")
	}
	return input - operand, nil
}

func checkedMultiply(input, operand int64) (int64, error) {
	if input != 0 && operand != 0 {
		if (input == math.MinInt64 && operand == -1) || (operand == math.MinInt64 && input == -1) {
			return 0, failAt(ReasonIntegerOverflow, "EXECUTE", "apply-int-mul", "int64 multiplication overflow")
		}
		result := input * operand
		if result/operand != input {
			return 0, failAt(ReasonIntegerOverflow, "EXECUTE", "apply-int-mul", "int64 multiplication overflow")
		}
		return result, nil
	}
	return 0, nil
}

func checkedDivide(input, operand int64) (int64, error) {
	if operand == 0 {
		return 0, failAt(ReasonIntegerDivisionByZero, "EXECUTE", "apply-int-div", "integer division by zero")
	}
	if input == math.MinInt64 && operand == -1 {
		return 0, failAt(ReasonIntegerOverflow, "EXECUTE", "apply-int-div", "int64 division overflow")
	}
	return input / operand, nil
}

func checkedModulo(input, operand int64) (int64, error) {
	if operand == 0 {
		return 0, failAt(ReasonIntegerModuloByZero, "EXECUTE", "apply-int-mod", "integer modulo by zero")
	}
	return input % operand, nil
}

func checkedNegate(input, operand int64) (int64, error) {
	if operand != 0 {
		return 0, failAt(ReasonOperationIRInvalid, "EXECUTE", "apply-int-neg", "int.neg requires a zero sentinel operand")
	}
	if input == math.MinInt64 {
		return 0, failAt(ReasonIntegerOverflow, "EXECUTE", "apply-int-neg", "int64 negation overflow")
	}
	return -input, nil
}

func checkedAbsolute(input, operand int64) (int64, error) {
	if operand != 0 {
		return 0, failAt(ReasonOperationIRInvalid, "EXECUTE", "apply-int-abs", "int.abs requires a zero sentinel operand")
	}
	if input == math.MinInt64 {
		return 0, failAt(ReasonIntegerOverflow, "EXECUTE", "apply-int-abs", "int64 absolute value overflow")
	}
	if input < 0 {
		return -input, nil
	}
	return input, nil
}

func checkedSign(input, operand int64) (int64, error) {
	if operand != 0 {
		return 0, failAt(ReasonOperationIRInvalid, "EXECUTE", "apply-int-sign", "int.sign requires a zero sentinel operand")
	}
	if input < 0 {
		return -1, nil
	}
	if input > 0 {
		return 1, nil
	}
	return 0, nil
}

func checkedMaximum(input, operand int64) (int64, error) {
	if input > operand {
		return input, nil
	}
	return operand, nil
}

func checkedMinimum(input, operand int64) (int64, error) {
	if input < operand {
		return input, nil
	}
	return operand, nil
}

func checkedIsZero(input, operand int64) (int64, error) {
	if operand != 0 {
		return 0, failAt(ReasonOperationIRInvalid, "EXECUTE", "apply-int-iszero", "int.iszero requires a zero sentinel operand")
	}
	if input == 0 {
		return 1, nil
	}
	return 0, nil
}

func checkedEqual(input, operand int64) (int64, error) {
	if input == operand {
		return 1, nil
	}
	return 0, nil
}

func checkedBooleanNot(input, operand int64) (int64, error) {
	if operand != 0 {
		return 0, failAt(ReasonOperationIRInvalid, "EXECUTE", "apply-bool-not", "bool.not requires a zero sentinel operand")
	}
	if input != 0 && input != 1 {
		return 0, failAt(ReasonOperationIRInvalid, "EXECUTE", "apply-bool-not", "bool.not requires canonical Boolean input")
	}
	return 1 - input, nil
}

func checkedBooleanAnd(input, operand int64) (int64, error) {
	if input != 0 && input != 1 {
		return 0, failAt(ReasonOperationIRInvalid, "EXECUTE", "apply-bool-and", "bool.and requires canonical Boolean input")
	}
	if operand != 0 && operand != 1 {
		return 0, failAt(ReasonOperationIRInvalid, "EXECUTE", "apply-bool-and", "bool.and requires a canonical Boolean operand")
	}
	if input == 1 && operand == 1 {
		return 1, nil
	}
	return 0, nil
}
