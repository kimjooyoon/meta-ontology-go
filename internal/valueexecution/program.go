package valueexecution

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
)

type Program struct {
	Activity            string
	Text                string
	OperationID         string
	Operand             int64
	SourceDigest        string
	SemanticFingerprint string
	ModelProgram        string
	operation           operationSpec
	document            bidir.Document
}

func (program Program) Execute(inputs []int64) (int64, error) {
	if len(inputs) != program.operation.Arity {
		return 0, fail(ReasonInputArityMismatch, fmt.Sprintf("got=%d want=%d", len(inputs), program.operation.Arity))
	}
	return program.operation.Apply(inputs[0], program.Operand)
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
