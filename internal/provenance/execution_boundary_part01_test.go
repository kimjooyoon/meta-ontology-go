package provenance

import (
	"strings"
	"testing"
)

func executionBoundaryTestDigest(value byte) string {
	return "sha256:" + strings.Repeat(string(value), 64)
}

func TestExecutionBoundaryBindsPipelineAndAXLifecycle(t *testing.T) {
	value := ObserveExecutionBoundary(ExecutionBoundaryInput{
		DeclarationDigest:        executionBoundaryTestDigest('a'),
		ContractDigest:           executionBoundaryTestDigest('0'),
		IRDigest:                 executionBoundaryTestDigest('b'),
		GeneratedDigest:          executionBoundaryTestDigest('c'),
		ReverseObservationDigest: executionBoundaryTestDigest('d'),
		TaskDigest:               executionBoundaryTestDigest('e'),
		WorkspaceDigest:          executionBoundaryTestDigest('f'),
		GatewayPolicyDigest:      executionBoundaryTestDigest('1'),
		ModelDigest:              executionBoundaryTestDigest('2'),
	}, ExecutionBoundarySuspended)
	if value.Decision != ExecutionBoundaryDecisionClosed || value.Reason != "EXECUTION_BOUNDARY_OBSERVED" ||
		value.MissingStageIndex != -1 || value.EvidencePrefixDigest == "" || !value.NonAuthorizing {
		t.Fatalf("incomplete execution boundary observation: %#v", value)
	}
	if err := ValidateExecutionBoundaryObservation(value); err != nil {
		t.Fatal(err)
	}
	changed := ObserveExecutionBoundary(ExecutionBoundaryInput{
		DeclarationDigest:        value.DeclarationDigest,
		IRDigest:                 value.IRDigest,
		GeneratedDigest:          value.GeneratedDigest,
		ReverseObservationDigest: value.ReverseObservationDigest,
		TaskDigest:               value.TaskDigest,
		WorkspaceDigest:          value.WorkspaceDigest,
		GatewayPolicyDigest:      value.GatewayPolicyDigest,
		ModelDigest:              executionBoundaryTestDigest('3'),
	}, ExecutionBoundaryRunning)
	if changed.ObservationDigest == value.ObservationDigest || changed.EvidencePrefixDigest == value.EvidencePrefixDigest {
		t.Fatal("execution boundary change was not reflected in evidence")
	}
	tampered := value
	tampered.EvidencePrefixDigest = executionBoundaryTestDigest('9')
	if ValidateExecutionBoundaryObservation(tampered) == nil {
		t.Fatal("tampered execution boundary prefix was accepted")
	}
}

func TestExecutionBoundaryPreservesFirstUnknownStage(t *testing.T) {
	value := ObserveExecutionBoundary(ExecutionBoundaryInput{
		DeclarationDigest: executionBoundaryTestDigest('a'),
		ContractDigest:    executionBoundaryTestDigest('0'),
		IRDigest:          executionBoundaryTestDigest('b'),
		TaskDigest:        executionBoundaryTestDigest('e'),
		WorkspaceDigest:   executionBoundaryTestDigest('f'),
	}, ExecutionBoundaryRunning)
	if value.Decision != ExecutionBoundaryDecisionUnknown || value.Reason != "EXECUTION_BOUNDARY_DIGEST_INCOMPLETE" ||
		value.MissingStageIndex != 3 || value.EvidencePrefixDigest == "" {
		t.Fatalf("unexpected first unresolved execution stage: %#v", value)
	}
	if err := ValidateExecutionBoundaryObservation(value); err != nil {
		t.Fatal(err)
	}
}

func TestExecutionBoundaryRequiresTerminalRecoveryAndPreservation(t *testing.T) {
	input := ExecutionBoundaryInput{
		DeclarationDigest:        executionBoundaryTestDigest('a'),
		ContractDigest:           executionBoundaryTestDigest('0'),
		IRDigest:                 executionBoundaryTestDigest('b'),
		GeneratedDigest:          executionBoundaryTestDigest('c'),
		ReverseObservationDigest: executionBoundaryTestDigest('d'),
		TaskDigest:               executionBoundaryTestDigest('e'),
		WorkspaceDigest:          executionBoundaryTestDigest('f'),
		GatewayPolicyDigest:      executionBoundaryTestDigest('1'),
		ModelDigest:              executionBoundaryTestDigest('2'),
	}
	value := ObserveExecutionBoundary(input, ExecutionBoundaryFailed)
	if value.Decision != ExecutionBoundaryDecisionUnknown || value.MissingStageIndex != 5 ||
		value.Reason != "EXECUTION_BOUNDARY_DIGEST_INCOMPLETE" {
		t.Fatalf("unexpected terminal evidence gap: %#v", value)
	}
	if err := ValidateExecutionBoundaryObservation(value); err != nil {
		t.Fatal(err)
	}
	input.CounterexampleRecoveryDigest = executionBoundaryTestDigest('3')
	input.ContractPreservationDigest = executionBoundaryTestDigest('4')
	value = ObserveExecutionBoundary(input, ExecutionBoundaryFailed)
	if value.Decision != ExecutionBoundaryDecisionClosed || value.MissingStageIndex != -1 {
		t.Fatalf("terminal evidence was not closed: %#v", value)
	}
	if err := ValidateExecutionBoundaryObservation(value); err != nil {
		t.Fatal(err)
	}
	tampered := value
	tampered.ContractPreservationDigest = executionBoundaryTestDigest('9')
	if ValidateExecutionBoundaryObservation(tampered) == nil {
		t.Fatal("tampered contract preservation evidence was accepted")
	}
}