package valueexecution

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
)

type Program struct {
	Activity            string
	Text                string
	Operation           OperationIR
	SourceDigest        string
	SemanticFingerprint string
	ModelProgram        string
	implementation      registeredOperation
	document            bidir.Document
}

func (program Program) Execute(inputs []int64) (int64, error) {
	if err := ValidateOperationIR(program.Operation); err != nil {
		return 0, failAt(ReasonOperationIRInvalid, "EXECUTE", "validate-operation-ir", err.Error())
	}
	if len(inputs) != program.Operation.Spec.Arity {
		detail := fmt.Sprintf("got=%d want=%d", len(inputs), program.Operation.Spec.Arity)
		return 0, failAt(ReasonInputArityMismatch, "EXECUTE", "validate-input-arity", detail)
	}
	return program.implementation.Apply(inputs[0], program.Operation.Operand.Int64)
}

func activityDeclaration(document bidir.Document, name string) (bidir.Declaration, bool) {
	for _, declaration := range document.Declarations {
		if declaration.Kind == bidir.ActivityKind && declaration.Name == name {
			return declaration, true
		}
	}
	return bidir.Declaration{}, false
}

func modelActivityProgram(model bidir.Model, name string) (string, bool) {
	for _, node := range model.Nodes {
		if node.Kind == bidir.ActivityKind && node.Name == name {
			program, present := node.Attributes[bidir.ActivityValueProgramAttribute]
			return program, present
		}
	}
	return "", false
}
