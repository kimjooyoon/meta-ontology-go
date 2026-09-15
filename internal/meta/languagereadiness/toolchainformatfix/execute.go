package toolchainformatfix

import (
	"reflect"

	cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

func executeCase(executor cliruntime.Executor, definition Definition) CaseResult {
	arguments, ok := argumentsFor(definition.Operation)
	result := CaseResult{Definition: definition, Arguments: arguments}
	if !ok {
		return unresolvedCase(result, "FORMAT_FIX_OPERATION_UNKNOWN")
	}
	first, err := executor.Invoke(arguments)
	if err != nil {
		return unresolvedCase(result, "FORMAT_FIX_FIRST_INVOCATION_UNKNOWN")
	}
	replay, err := executor.Invoke(arguments)
	if err != nil {
		result.First, result.Invocations = first, 1
		return unresolvedCase(result, "FORMAT_FIX_REPLAY_INVOCATION_UNKNOWN")
	}
	result.First, result.Replay, result.Invocations = first, replay, 2
	result.ExitMatched = first.ExitCode == definition.ExpectedExit && replay.ExitCode == definition.ExpectedExit
	firstOK, firstStructured, firstPlan := inspectOutput(definition.Operation, first)
	replayOK, replayStructured, replayPlan := inspectOutput(definition.Operation, replay)
	result.OutputMatched = firstOK && replayOK
	result.StructuredOutput = firstStructured
	result.StructuredPlan = firstPlan
	result.ReplayMatched = reflect.DeepEqual(first, replay)
	result.RepositoryWrites = first.RepositoryWrites + replay.RepositoryWrites
	result.Status, result.Reason = "SATISFIED", "FORMAT_FIX_CASE_SATISFIED"
	if !result.ExitMatched || !result.OutputMatched || !result.ReplayMatched ||
		result.RepositoryWrites != 0 || firstStructured != replayStructured || firstPlan != replayPlan {
		result.Status, result.Reason = "NOT_SATISFIED", "FORMAT_FIX_CASE_MISMATCH"
	}
	result.EvidenceDigest = caseDigest(result)
	return result
}

func unresolvedCase(result CaseResult, reason string) CaseResult {
	result.Status, result.Reason = "UNRESOLVED", reason
	result.EvidenceDigest = caseDigest(result)
	return result
}
