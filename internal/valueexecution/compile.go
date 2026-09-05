package valueexecution

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Compile(filename string, source []byte, activityName string) (Program, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return Program{}, failAt(ReasonSourceParseFailed, "PARSE", "parse-source", diagnostics.Error().Error())
	}
	document, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return Program{}, failAt(ReasonSemanticBindingFailed, "LOWER", "bind-bidir-document", err.Error())
	}
	if len(document.RuntimeBindings) != 0 {
		return Program{}, failAt(ReasonPlanRequired, "PLAN", "runtime-binding-plan-required", "runtime binding execution requires an explicit plan")
	}
	model, err := bidir.Get(document)
	if err != nil {
		return Program{}, failAt(ReasonSemanticBindingFailed, "LOWER", "read-bidir-model", err.Error())
	}
	declaration, ok := activityDeclaration(document, activityName)
	if !ok {
		return Program{}, failAt(ReasonActivityNotFound, "LOWER", "resolve-activity", activityName)
	}
	if len(declaration.Inputs) != 1 || len(declaration.Outputs) != 1 {
		detail := fmt.Sprintf("inputs=%d outputs=%d", len(declaration.Inputs), len(declaration.Outputs))
		return Program{}, failAt(ReasonSignatureArityUnsupported, "TYPECHECK", "bind-operation-signature", detail)
	}
	programText, present := declaration.Attributes[bidir.ActivityValueProgramAttribute]
	if !present || programText == "" {
		return Program{}, failAt(ReasonProgramMissing, "LOWER", "read-computes-program", activityName)
	}
	modelProgram, ok := modelActivityProgram(model, activityName)
	if !ok || modelProgram != programText {
		return Program{}, failAt(ReasonSemanticBindingFailed, "LOWER", "preserve-computes-program", "activity value program was not preserved in the bidir model")
	}
	operationIR, implementation, err := compileOperation(activityName, programText, declaration)
	if err != nil {
		return Program{}, err
	}
	return Program{
		Activity: activityName, Text: programText, Operation: operationIR,
		SourceDigest: digestBytes(source), SemanticFingerprint: bidir.SemanticFingerprint(model),
		ModelProgram: modelProgram, implementation: implementation, document: document,
	}, nil
}
