package toolchaincli

import cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"

type outputEvidence struct {
	stdoutOK           bool
	stderrOK           bool
	structuredOutputs  int
	languageOperations int
	declaredCommands   int
}

func inspectOutput(definition Definition, observation cliruntime.Observation) outputEvidence {
	if definition.Kind == "POSITIVE" {
		return inspectPositive(definition.Operation, observation)
	}
	return inspectGuardrail(definition.Operation, observation)
}
