package valueexecution

import (
	"fmt"
	"slices"
)

func ValidateOperationSpec(spec OperationSpec) error {
	if spec.Schema != OperationSpecSchema || spec.ID != "int.add" || spec.Version != 1 {
		return fmt.Errorf("operation identity is not closed")
	}
	if spec.Arity != 1 || !slices.Equal(spec.InputEntities, []string{IntegerEntity}) || spec.OutputEntity != IntegerEntity {
		return fmt.Errorf("operation signature is not closed")
	}
	if spec.OperandKind != OperandInt64Literal || spec.Effect != EffectPureValue || spec.Determinism != Deterministic {
		return fmt.Errorf("operation semantics are not closed")
	}
	expectedFailures := []string{ReasonInputArityMismatch, ReasonIntegerOverflow}
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
