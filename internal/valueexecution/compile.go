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
	if len(document.RuntimeBindings) != 0 {
		return Program{}, failAt(ReasonPlanRequired, "PLAN", "runtime-binding-plan-required", "runtime binding execution requires an explicit plan")
	}
	model, err := bidir.Get(document)
	if err != nil {
		return Program{}, failAt(ReasonSemanticBindingFailed, "LOWER", "read-bidir-model", err.Error())
	}
	return compileDocumentProgram(filename, source, document, model, activityName)
}
