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
	authority           resultAuthority
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

// ExecuteResult preserves the existing scalar Execute API while issuing a
// typed result only after the compiled operation has successfully applied.
// The result authority is retained privately by Compile and is never derived
// from the public Program fields at issuance time.
func (program Program) ExecuteResult(inputs []int64) (ProducedResult, error) {
	if err := program.validateResultAuthority(); err != nil {
		return ProducedResult{}, err
	}
	value, err := program.Execute(inputs)
	if err != nil {
		return ProducedResult{}, err
	}
	return issueProducedResult(program.authority, value), nil
}

// ValidateProducedResult accepts only a result issued by this exact compiled
// producer authority. The result itself remains opaque to callers.
func (program Program) ValidateProducedResult(result ProducedResult) error {
	if err := program.validateResultAuthority(); err != nil {
		return err
	}
	if err := result.validate(); err != nil {
		return err
	}
	if result.authority != program.authority {
		return failAt(ReasonResultProducerMismatch, "RESULT", "bind-produced-result", "result was issued by a different compiled producer")
	}
	return nil
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
