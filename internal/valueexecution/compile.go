package valueexecution

import (
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
	model, err := bidir.Get(document)
	if err != nil {
		return Program{}, failAt(ReasonSemanticBindingFailed, "LOWER", "read-bidir-model", err.Error())
	}
	declaration, ok := activityDeclaration(document, activityName)
	if !ok {
		return Program{}, failAt(ReasonActivityNotFound, "LOWER", "resolve-activity", activityName)
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
