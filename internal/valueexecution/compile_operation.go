package valueexecution

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
)

func compileOperation(activity, text string, declaration bidir.Declaration) (OperationIR, registeredOperation, error) {
	id, operandText, found := strings.Cut(text, ":")
	if !found || strings.Contains(operandText, ":") {
		return OperationIR{}, registeredOperation{}, failAt(ReasonOperandInvalid, "PARSE", "split-operation-program", text)
	}
	implementation, known := operationByID(id)
	if !known {
		return OperationIR{}, registeredOperation{}, failAt(ReasonProgramUnknown, "RESOLVE", "resolve-operation-spec", id)
	}
	if err := ValidateOperationSpec(implementation.Spec); err != nil {
		return OperationIR{}, registeredOperation{}, failAt(ReasonOperationSpecInvalid, "RESOLVE", "validate-operation-spec", err.Error())
	}
	if !signatureMatches(declaration, implementation.Spec) {
		detail := fmt.Sprintf("inputs=%v output=%v", referenceNames(declaration.Inputs), referenceNames(declaration.Outputs))
		return OperationIR{}, registeredOperation{}, failAt(ReasonSignatureTypeMismatch, "TYPECHECK", "bind-operation-signature", detail)
	}
	operand, err := strconv.ParseInt(operandText, 10, 64)
	if err != nil {
		return OperationIR{}, registeredOperation{}, failAt(ReasonOperandInvalid, "TYPECHECK", "parse-operation-operand", operandText)
	}
	ir := newOperationIR(activity, text, implementation.Spec, operand)
	if err := ValidateOperationIR(ir); err != nil {
		return OperationIR{}, registeredOperation{}, failAt(ReasonOperationIRInvalid, "LOWER", "validate-operation-ir", err.Error())
	}
	return ir, implementation, nil
}

func signatureMatches(declaration bidir.Declaration, spec OperationSpec) bool {
	inputs, outputs := referenceNames(declaration.Inputs), referenceNames(declaration.Outputs)
	return len(outputs) == 1 && slicesEqual(inputs, spec.InputEntities) && outputs[0] == spec.OutputEntity
}

func referenceNames(references []bidir.Reference) []string {
	result := make([]string, len(references))
	for index, reference := range references {
		result[index] = reference.Name
	}
	return result
}
