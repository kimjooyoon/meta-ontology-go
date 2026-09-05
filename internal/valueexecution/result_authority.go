package valueexecution

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
)

// resultAuthority is compiled once and contains only immutable scalar
// bindings. In particular, it does not retain any public OperationIR slices.
type resultAuthority struct {
	activityID          string
	activityName        string
	outputEntityID      string
	outputEntityName    string
	sourceDigest        string
	semanticFingerprint string
	valueProgramDigest  string
	modelProgramDigest  string
	operationDigest     string
	operationSpecDigest string
}

func newResultAuthority(model bidir.Model, activityName string, source []byte, programText, modelProgram string, operation OperationIR) (resultAuthority, error) {
	var activity bidir.Node
	for _, node := range model.Nodes {
		if node.Kind == bidir.ActivityKind && node.Name == activityName {
			activity = node
			break
		}
	}
	if activity.ID == "" {
		return resultAuthority{}, failAt(ReasonProgramAuthorityInvalid, "LOWER", "bind-result-authority", fmt.Sprintf("activity %q is absent from the lowered model", activityName))
	}

	var output bidir.Node
	outputCount := 0
	for _, relation := range model.Relations {
		if relation.Kind != bidir.PredicateWasGeneratedBy || relation.Target != activity.ID {
			continue
		}
		for _, node := range model.Nodes {
			if node.ID == relation.Source {
				output = node
				outputCount++
				break
			}
		}
	}
	if outputCount != 1 || output.ID == "" || output.Kind != bidir.EntityKind {
		return resultAuthority{}, failAt(ReasonProgramAuthorityInvalid, "LOWER", "bind-result-authority", fmt.Sprintf("activity %q has %d lowered output entities", activityName, outputCount))
	}
	if output.Name != operation.OutputEntity {
		return resultAuthority{}, failAt(ReasonProgramAuthorityInvalid, "TYPECHECK", "bind-result-authority", fmt.Sprintf("lowered output %q does not match operation output %q", output.Name, operation.OutputEntity))
	}

	return resultAuthority{
		activityID:          string(activity.ID),
		activityName:        activity.Name,
		outputEntityID:      string(output.ID),
		outputEntityName:    output.Name,
		sourceDigest:        digestBytes(source),
		semanticFingerprint: bidir.SemanticFingerprint(model),
		valueProgramDigest:  digestBytes([]byte(programText)),
		modelProgramDigest:  digestBytes([]byte(modelProgram)),
		operationDigest:     digestValue(operation),
		operationSpecDigest: operation.SpecDigest,
	}, nil
}

func (program Program) validateResultAuthority() error {
	if !program.authority.valid() {
		return failAt(ReasonProgramAuthorityInvalid, "EXECUTE", "validate-result-authority", "program has no Compile-issued private authority")
	}
	checks := []struct {
		name string
		ok   bool
	}{
		{"activity", program.Activity == program.authority.activityName},
		{"value program", digestBytes([]byte(program.Text)) == program.authority.valueProgramDigest},
		{"source digest", program.SourceDigest == program.authority.sourceDigest},
		{"semantic fingerprint", program.SemanticFingerprint == program.authority.semanticFingerprint},
		{"model program", digestBytes([]byte(program.ModelProgram)) == program.authority.modelProgramDigest},
		{"operation", digestValue(program.Operation) == program.authority.operationDigest},
	}
	for _, check := range checks {
		if !check.ok {
			return failAt(ReasonProgramAuthorityMismatch, "EXECUTE", "validate-result-authority", check.name+" binding changed after compile")
		}
	}
	return nil
}

func (authority resultAuthority) valid() bool {
	return authority.activityID != "" && authority.activityName != "" && authority.outputEntityID != "" &&
		authority.outputEntityName != "" && validDigest(authority.sourceDigest) && authority.semanticFingerprint != "" &&
		validDigest(authority.valueProgramDigest) && validDigest(authority.modelProgramDigest) &&
		validDigest(authority.operationDigest) && validDigest(authority.operationSpecDigest)
}
