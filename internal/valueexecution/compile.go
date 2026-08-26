package valueexecution

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Compile(filename string, source []byte, activityName string) (Program, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return Program{}, fail(ReasonSourceParseFailed, diagnostics.Error().Error())
	}
	document, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return Program{}, fail(ReasonSemanticBindingFailed, err.Error())
	}
	model, err := bidir.Get(document)
	if err != nil {
		return Program{}, fail(ReasonSemanticBindingFailed, err.Error())
	}
	declaration, ok := activityDeclaration(document, activityName)
	if !ok {
		return Program{}, fail(ReasonActivityNotFound, activityName)
	}
	if len(declaration.Inputs) != 1 || len(declaration.Outputs) != 1 {
		return Program{}, fail(ReasonSignatureArityUnsupported, fmt.Sprintf("inputs=%d outputs=%d", len(declaration.Inputs), len(declaration.Outputs)))
	}
	programText, present := declaration.Attributes[bidir.ActivityValueProgramAttribute]
	if !present || programText == "" {
		return Program{}, fail(ReasonProgramMissing, activityName)
	}
	modelProgram, ok := modelActivityProgram(model, activityName)
	if !ok || modelProgram != programText {
		return Program{}, fail(ReasonSemanticBindingFailed, "activity value program was not preserved in the bidir model")
	}
	operationID, operandText, found := strings.Cut(programText, ":")
	if !found || strings.Contains(operandText, ":") {
		return Program{}, fail(ReasonOperandInvalid, programText)
	}
	operation, known := operationByID(operationID)
	if !known {
		return Program{}, fail(ReasonProgramUnknown, operationID)
	}
	operand, err := strconv.ParseInt(operandText, 10, 64)
	if err != nil {
		return Program{}, fail(ReasonOperandInvalid, operandText)
	}
	return Program{
		Activity: activityName, Text: programText, OperationID: operationID, Operand: operand,
		SourceDigest: digestBytes(source), SemanticFingerprint: bidir.SemanticFingerprint(model),
		ModelProgram: modelProgram, operation: operation, document: document,
	}, nil
}
