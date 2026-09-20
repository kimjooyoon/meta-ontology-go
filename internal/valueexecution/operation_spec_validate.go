package valueexecution

import (
	"fmt"
	"slices"
)

func ValidateOperationSpec(spec OperationSpec) error {
	if spec.Schema != OperationSpecSchema || spec.Version != 1 {
		return fmt.Errorf("operation identity is not closed")
	}
	if spec.ID != "int.add" && spec.ID != "int.sub" && spec.ID != "int.mul" && spec.ID != "int.div" && spec.ID != "int.mod" && spec.ID != "int.neg" && spec.ID != "int.abs" && spec.ID != "int.sign" && spec.ID != "int.max" && spec.ID != "int.min" && spec.ID != "int.iszero" && spec.ID != "int.eq" && spec.ID != "bool.not" && spec.ID != "bool.and" {
		return fmt.Errorf("operation identity is not registered")
	}
	expectedInput := IntegerEntity
	expectedOutput := IntegerEntity
	if spec.ID == "int.iszero" || spec.ID == "int.eq" || spec.ID == "bool.not" || spec.ID == "bool.and" {
		expectedOutput = BooleanEntity
	}
	if spec.ID == "bool.not" || spec.ID == "bool.and" {
		expectedInput = BooleanEntity
	}
	if spec.Arity != 1 || !slices.Equal(spec.InputEntities, []string{expectedInput}) || spec.OutputEntity != expectedOutput {
		return fmt.Errorf("operation signature is not closed")
	}
	if spec.OperandKind != OperandInt64Literal || spec.Effect != EffectPureValue || spec.Determinism != Deterministic {
		return fmt.Errorf("operation semantics are not closed")
	}
	expectedFailures := []string{ReasonInputArityMismatch, ReasonIntegerOverflow}
	if spec.ID == "int.iszero" || spec.ID == "bool.not" || spec.ID == "bool.and" {
		expectedFailures = []string{ReasonInputArityMismatch, ReasonOperationIRInvalid}
	} else if spec.ID == "int.eq" || spec.ID == "int.max" || spec.ID == "int.min" {
		expectedFailures = []string{ReasonInputArityMismatch}
	} else if spec.ID == "int.sign" {
		expectedFailures = []string{ReasonInputArityMismatch, ReasonOperationIRInvalid}
	} else if spec.ID == "int.div" {
		expectedFailures = append(expectedFailures, ReasonIntegerDivisionByZero)
	}
	if spec.ID == "int.mod" {
		expectedFailures = []string{ReasonInputArityMismatch, ReasonIntegerModuloByZero}
	}
	if !slices.Equal(spec.FailureReasons, expectedFailures) {
		return fmt.Errorf("operation failure set is not closed")
	}
	if spec.Authority.RepositoryWrite || spec.Authority.ExternalCall || spec.Authority.Promotion {
		return fmt.Errorf("operation authority is not zero")
	}
	return nil
}

func ValidateOperationIR(ir OperationIR) error {
	if ir.Schema != OperationIRSchema || ir.Activity == "" || ValidateOperationSpec(ir.Spec) != nil {
		return fmt.Errorf("operation IR identity is invalid")
	}
	if ir.SpecDigest != digestValue(ir.Spec) || !slices.Equal(ir.InputEntities, ir.Spec.InputEntities) {
		return fmt.Errorf("operation IR spec binding is invalid")
	}
	if ir.OutputEntity != ir.Spec.OutputEntity || ir.Operand.Kind != ir.Spec.OperandKind {
		return fmt.Errorf("operation IR type binding is invalid")
	}
	if ir.Program != fmt.Sprintf("%s:%d", ir.Spec.ID, ir.Operand.Int64) {
		return fmt.Errorf("operation IR program binding is invalid")
	}
	if (ir.Spec.ID == "int.neg" || ir.Spec.ID == "int.abs" || ir.Spec.ID == "int.sign" || ir.Spec.ID == "int.iszero" || ir.Spec.ID == "bool.not") && ir.Operand.Int64 != 0 {
		return fmt.Errorf("%s requires a zero sentinel operand", ir.Spec.ID)
	}
	if ir.Spec.ID == "bool.and" && ir.Operand.Int64 != 0 && ir.Operand.Int64 != 1 {
		return fmt.Errorf("bool.and requires a canonical Boolean operand")
	}
	return nil
}

func newOperationIR(activity, program string, spec OperationSpec, operand int64) OperationIR {
	spec = cloneOperationSpec(spec)
	return OperationIR{
		Schema: OperationIRSchema, Activity: activity, Program: program, Spec: spec,
		SpecDigest: digestValue(spec), InputEntities: append([]string(nil), spec.InputEntities...),
		OutputEntity: spec.OutputEntity, Operand: OperandIR{Kind: spec.OperandKind, Int64: operand},
	}
}

func cloneOperationSpec(spec OperationSpec) OperationSpec {
	spec.InputEntities = append([]string(nil), spec.InputEntities...)
	spec.FailureReasons = append([]string(nil), spec.FailureReasons...)
	return spec
}

func slicesEqual(left, right []string) bool { return slices.Equal(left, right) }
