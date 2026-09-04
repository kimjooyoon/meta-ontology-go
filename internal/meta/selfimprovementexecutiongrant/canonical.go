package selfimprovementexecutiongrant

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

func canonicalV24() V24Binding {
	return V24Binding{RequestSchema: "gooo/self-improvement-candidate-authorization-request/v1", RequestDigest: digestBytes([]byte("v24-request")), ResolutionSchema: "gooo/self-improvement-candidate-authorization-resolution/v1", ResolutionDigest: digestBytes([]byte("v24-resolution")), CandidateStableID: digestBytes([]byte("candidate-id")), CandidateDigest: digestBytes([]byte("candidate-digest")), SubjectSHA: "0123456789abcdef0123456789abcdef01234567", ObservationDigest: digestBytes([]byte("observation")), ContractDigest: digestBytes([]byte("v24-contract")), AuthorizationDecision: "ALLOW", AuthorizationResolution: "CLOSED", AuthorizationOutcome: "AUTHORIZED", RequestValid: true, ResolutionValid: true}
}

func canonicalV25() V25Binding {
	return V25Binding{Schema: "gooo/self-improvement-execution-contract/v1", ContractID: "gooo://self-improvement/candidate-execution-contract/v1", ContractDigest: digestBytes([]byte("v25-contract")), Decision: "CLOSED", Resolution: "DECLARED", CandidateStableID: digestBytes([]byte("candidate-id")), CandidateDigest: digestBytes([]byte("candidate-digest")), SubjectSHA: "0123456789abcdef0123456789abcdef01234567", ObservationDigest: digestBytes([]byte("observation")), CandidateInputDigest: digestBytes([]byte("candidate-input")), OperationID: "self-improvement.value-witness-experiment.v1", BoundedTarget: "VALUE_WITNESS_EXPERIMENT", EvaluatorRegistryDigest: digestBytes([]byte("evaluator-registry")), ToolchainTestContractIdentity: digestBytes([]byte("toolchain-test-contract")), MaxExecutions: 1, RepositoryWritesAllowed: false, ExecutionAuthorized: false, ExecutionGrantRequired: true, Valid: true}
}

func canonicalSource() SourceArtifact {
	return SourceArtifact{Repository: "kimjooyoon/meta-ontology-go", WorkflowRunID: 1, WorkflowRunAttempt: 1, ArtifactID: 1, ArtifactDigest: digestBytes([]byte("v25-artifact")), ArtifactExpired: false, ArtifactExpiryKnown: true}
}

func canonicalRequest(program PolicyProgram) GrantRequest {
	return BuildRequest(program, canonicalV24(), canonicalV25(), canonicalSource())
}

func fixtureDecision(request GrantRequest, decision string) GrantDecisionInput {
	input := GrantDecisionInput{Schema: GrantDecisionSchema, Decision: decision, RequestDigest: request.Digest, V24: request.V24, V25: request.V25, Source: request.Source, DecisionSource: DecisionSourceCanonical, ActorEvidence: ActorEvidence{Repository: "canonical-fixture", Actor: "canonical-fixture", WorkflowRunID: 1, WorkflowRunAttempt: 1, Event: "canonical-fixture", EvidenceLabel: CanonicalEvidenceLabel}}
	input.DecisionDigest = decisionDigest(input)
	return input
}

func canonicalCase(program PolicyProgram, id string, request GrantRequest, inputs []GrantDecisionInput, expectedDecision Decision, expectedResolution Resolution, expectedReason string) (CanonicalCase, error) {
	input := GrantInput{Request: request, DecisionInputs: inputs}
	resolution := Evaluate(program, input)
	verification := Verify(program, input, resolution)
	if resolution.Decision != expectedDecision || resolution.Resolution != expectedResolution || resolution.Reason != expectedReason || !verification.Verified {
		return CanonicalCase{}, fmt.Errorf("canonical case %s resolved %s/%s/%s, expected %s/%s/%s", id, resolution.Decision, resolution.Resolution, resolution.Reason, expectedDecision, expectedResolution, expectedReason)
	}
	return CanonicalCase{ID: id, ExpectedDecision: expectedDecision, ExpectedResolution: expectedResolution, ExpectedReason: expectedReason, ActualDecision: resolution.Decision, ActualResolution: resolution.Resolution, ActualReason: resolution.Reason, Unknown: resolution.Unknown, GrantAllowsExecution: resolution.GrantAllowsExecution, RemainingUses: resolution.RemainingUses, ConsumedUses: resolution.ConsumedUses, ExecutionCount: resolution.ExecutionCount, Pass: true}, nil
}

func BuildCanonicalCaseReport(program PolicyProgram) (CanonicalCaseReport, error) {
	request := canonicalRequest(program)
	allow := fixtureDecision(request, DecisionAllow)
	deny := fixtureDecision(request, DecisionDeny)
	missingArtifactRequest := request
	missingArtifactRequest.Source.ArtifactExpired = true
	missingArtifactRequest.Digest = requestDigest(missingArtifactRequest)
	incompleteRequest := request
	incompleteRequest.V25.OperationID = ""
	incompleteRequest.Digest = requestDigest(incompleteRequest)
	scopeRequest := request
	scopeRequest.V25.BoundedTarget = "UNBOUNDED"
	scopeRequest.Digest = requestDigest(scopeRequest)
	unsafeRequest := request
	unsafeRequest.V25.MaxExecutions = 2
	unsafeRequest.Digest = requestDigest(unsafeRequest)
	caseSpecs := []struct {
		id, reason string
		request    GrantRequest
		inputs     []GrantDecisionInput
		decision   Decision
		resolution Resolution
	}{
		{"explicit-allow", ReasonAllow, request, []GrantDecisionInput{allow}, DecisionClosed, ResolutionGrantedUnconsumed},
		{"explicit-deny", ReasonDeny, request, []GrantDecisionInput{deny}, DecisionClosed, ResolutionDenied},
		{"deterministic-replay", ReasonAllow, request, []GrantDecisionInput{allow}, DecisionClosed, ResolutionGrantedUnconsumed},
		{"missing-decision", ReasonMissingDecision, request, nil, DecisionUnknown, ResolutionLower},
		{"missing-or-expired-upstream-artifact", ReasonMissingArtifact, missingArtifactRequest, []GrantDecisionInput{fixtureDecision(missingArtifactRequest, DecisionAllow)}, DecisionUnknown, ResolutionLower},
		{"incomplete-grant-input", ReasonIncompleteInput, incompleteRequest, []GrantDecisionInput{fixtureDecision(incompleteRequest, DecisionAllow)}, DecisionUnknown, ResolutionLower},
		{"digest-or-scope-mismatch", ReasonScopeMismatch, scopeRequest, []GrantDecisionInput{fixtureDecision(scopeRequest, DecisionAllow)}, DecisionRefuted, ResolutionExact},
		{"conflicting-duplicate-grant", ReasonDuplicate, request, []GrantDecisionInput{allow, deny}, DecisionRefuted, ResolutionExact},
		{"unauthorized-or-unsafe-grant", ReasonUnsafe, unsafeRequest, []GrantDecisionInput{fixtureDecision(unsafeRequest, DecisionAllow)}, DecisionRefuted, ResolutionExact},
	}
	report := CanonicalCaseReport{Schema: CanonicalCasesSchema, Policy: program.Evidence, RequiredFields: RequiredBindingNames(), RequestDigest: request.Digest, CaseDenominator: 9, StructuralSeparateGrantEdgesBefore: 0, StructuralSeparateGrantEdgesAfter: 1, Counts: map[string]int{"CLOSED": 0, "UNKNOWN": 0, "REFUTED": 0}, ReplayEqual: true, LiveGrantRequests: 0, LiveGrants: 0, LiveExecutionCount: 0, CanonicalExecutionCount: 0, RepositoryWrites: 0, LocalTestExecutions: 0, FallbackAccepted: 0, PerformanceImprovement: PerformanceUnknown, Decision: DecisionClosed, Resolution: ResolutionExact, Reason: "NINE_CANONICAL_EXECUTION_GRANT_CASES", GoPhysicalLines: program.Inventory.GoPhysicalLines, GoooPhysicalLines: program.Inventory.GoooPhysicalLines}
	for _, spec := range caseSpecs {
		result, err := canonicalCase(program, spec.id, spec.request, spec.inputs, spec.decision, spec.resolution, spec.reason)
		if err != nil {
			return CanonicalCaseReport{}, err
		}
		report.Cases = append(report.Cases, result)
		report.Counts[string(result.ActualDecision)]++
		if result.GrantAllowsExecution {
			report.CanonicalGrantedCases++
		}
		if result.Unknown != nil {
			report.SixFieldUnknowns++
		}
		if result.ActualDecision == DecisionRefuted {
			report.RefutedContradictions++
		}
		first := Evaluate(program, GrantInput{Request: spec.request, DecisionInputs: spec.inputs})
		second := Evaluate(program, GrantInput{Request: spec.request, DecisionInputs: spec.inputs})
		if firstBytes, _ := json.Marshal(first); !bytes.Equal(firstBytes, mustJSON(second)) {
			report.ReplayEqual = false
		}
	}
	report.ClosedCases, report.UnknownCases, report.RefutedCases = report.Counts["CLOSED"], report.Counts["UNKNOWN"], report.Counts["REFUTED"]
	report.GrantRemainingUses = report.CanonicalGrantedCases
	report.IndependentReplayComparisons = 1
	report.ArtifactFiles, report.ArtifactTypes = 9, 3
	if report.ClosedCases != 3 || report.UnknownCases != 3 || report.RefutedCases != 3 || !report.ReplayEqual {
		report.Decision, report.Resolution, report.Reason = DecisionRefuted, ResolutionExact, "CANONICAL_CASE_PARTITION_FAILED"
	}
	report.Digest = canonicalDigest(report)
	return report, nil
}

func mustJSON(value GrantResolution) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}

func ValidateCanonicalCases(report CanonicalCaseReport) error {
	if report.Schema != CanonicalCasesSchema || report.CaseDenominator != 9 || report.StructuralSeparateGrantEdgesBefore != 0 || report.StructuralSeparateGrantEdgesAfter != 1 || report.ClosedCases != 3 || report.UnknownCases != 3 || report.RefutedCases != 3 || report.Counts["CLOSED"] != 3 || report.Counts["UNKNOWN"] != 3 || report.Counts["REFUTED"] != 3 || !report.ReplayEqual || report.LiveGrantRequests != 0 || report.LiveGrants != 0 || report.LiveExecutionCount != 0 || report.CanonicalExecutionCount != 0 || report.GrantConsumedUses != 0 || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || report.FallbackAccepted != 0 || report.Digest != canonicalDigest(report) {
		return errors.New("canonical execution grant cases are not exact")
	}
	if len(report.Cases) != report.CaseDenominator {
		return errors.New("canonical execution grant case denominator mismatch")
	}
	for _, item := range report.Cases {
		if !item.Pass || item.ExpectedDecision != item.ActualDecision || item.ExpectedResolution != item.ActualResolution || item.ExpectedReason != item.ActualReason {
			return errors.New("canonical execution grant case failed")
		}
		if item.ActualDecision == DecisionUnknown && item.Unknown == nil {
			return errors.New("canonical UNKNOWN grant case lacks six-field evidence")
		}
		if item.ActualDecision != DecisionUnknown && item.Unknown != nil {
			return errors.New("canonical non-UNKNOWN grant case contains unknown evidence")
		}
	}
	return nil
}
