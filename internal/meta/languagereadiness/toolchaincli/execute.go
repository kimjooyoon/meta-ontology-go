package toolchaincli

import (
	cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

func executeCase(executor cliruntime.Executor, definition Definition) CaseResult {
	arguments, known := argumentsFor(definition.Operation)
	result := CaseResult{Definition: definition, Arguments: arguments, Reason: "CLI_CASE_UNKNOWN"}
	if !known {
		return unresolvedCase(result)
	}
	first, err := executor.Invoke(arguments)
	if err != nil {
		return unresolvedCase(result)
	}
	result.First, result.Invocations = first, 1
	replay, err := executor.Invoke(arguments)
	if err != nil {
		return unresolvedCase(result)
	}
	result.Replay, result.Invocations = replay, 2
	if first.Failure != "" || replay.Failure != "" {
		return unresolvedCase(result)
	}
	firstOutput, replayOutput := inspectOutput(definition, first), inspectOutput(definition, replay)
	result.ExitMatched = first.ExitCode == definition.ExpectedExit && replay.ExitCode == definition.ExpectedExit
	result.StdoutMatched = firstOutput.stdoutOK && replayOutput.stdoutOK
	result.StderrMatched = firstOutput.stderrOK && replayOutput.stderrOK
	result.ReplayMatched = deterministicReplayEqual(first, replay)
	result.StructuredOutputs = firstOutput.structuredOutputs
	result.LanguageOperations = firstOutput.languageOperations
	result.DeclaredCommands = firstOutput.declaredCommands
	result.RepositoryWrites = first.RepositoryWrites + replay.RepositoryWrites
	result.Status, result.Reason = "NOT_SATISFIED", "CLI_CASE_MISMATCH"
	if result.ExitMatched && result.StdoutMatched && result.StderrMatched &&
		result.ReplayMatched && result.RepositoryWrites == 0 {
		result.Status, result.Reason = "SATISFIED", "CLI_CASE_OBSERVED_EXACTLY"
	}
	result.EvidenceDigest = caseDigest(result)
	return result
}

func unresolvedCase(result CaseResult) CaseResult {
	result.Status, result.Reason = "UNRESOLVED", "CLI_OBSERVATION_UNKNOWN"
	result.EvidenceDigest = caseDigest(result)
	return result
}
