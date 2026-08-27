package metainvocation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Compile(sourcePath string, source []byte, registry Registry) (Program, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if diagnostics.HasErrors() {
		return Program{}, diagnostics.Error()
	}
	if file == nil || file.Package == nil || file.Namespace == nil {
		return Program{}, fmt.Errorf("source headers are incomplete")
	}
	program := Program{
		SourcePath: sourcePath, Package: file.Package.Name, Namespace: file.Namespace.Name,
		SourceDigest: bytesDigest(source), Entities: map[string]string{}, Operations: map[string]BoundOperation{},
	}
	for _, declaration := range file.Decls {
		switch item := declaration.(type) {
		case *syntax.EntityDecl:
			if _, exists := program.Entities[item.Name]; exists {
				return Program{}, fmt.Errorf("entity %q is declared twice", item.Name)
			}
			program.Entities[item.Name] = item.ID
		case *syntax.ActivityDecl:
			if !item.ValueProgramPresent {
				continue
			}
			spec, ok := registry.lookup(item.ValueProgram)
			if !ok {
				return Program{}, fmt.Errorf("activity %q references unknown meta operation %q", item.Name, item.ValueProgram)
			}
			if len(item.Inputs) != 1 || item.Inputs[0].Name != spec.InputEntity || item.Output != spec.OutputEntity {
				return Program{}, fmt.Errorf("activity %q signature does not conform to operation %q", item.Name, spec.ID)
			}
			if _, exists := program.Operations[item.Name]; exists {
				return Program{}, fmt.Errorf("activity %q is declared twice", item.Name)
			}
			program.Operations[item.Name] = BoundOperation{
				Activity: item.Name, Program: item.ValueProgram, Input: item.Inputs[0].Name, Output: item.Output,
				SpecDigest: digest(spec), Source: coordinate(sourcePath, item.ValueProgramSpan),
			}
		}
	}
	if err := validateProgram(program); err != nil {
		return Program{}, err
	}
	return program, nil
}

func coordinate(path string, span syntax.Span) SourceCoordinate {
	return SourceCoordinate{Path: path, StartLine: span.Start.Line, StartColumn: span.Start.Column, EndLine: span.End.Line, EndColumn: span.End.Column}
}

func validateProgram(program Program) error {
	requiredEntities := map[string]string{
		"ChangeSet": "gooo://meta/ci-plan/entity/change-set",
		"CheckPlan": "gooo://meta/ci-plan/entity/check-plan",
		"VerificationReceipt": "gooo://meta/ci-plan/entity/verification-receipt",
	}
	for name, id := range requiredEntities {
		if program.Entities[name] != id {
			return fmt.Errorf("entity %q must bind id %q", name, id)
		}
	}
	requiredOperations := map[string]string{
		"SelectGoCheck": operationGoRule,
		"SelectDocsCheck": operationDocsRule,
		"SelectYAMLCheck": operationYAMLRule,
		"PlanCI": operationPlan,
		"VerifyCIPlan": operationVerify,
	}
	for activity, operation := range requiredOperations {
		if program.Operations[activity].Program != operation {
			return fmt.Errorf("activity %q must bind operation %q", activity, operation)
		}
	}
	return nil
}
