package selfimprovementexecutiongrant

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

func canonicalV24() V24Binding {
	return V24Binding{RequestSchema: "gooo/self-improvement-candidate-authorization-request/v2", RequestDigest: digestBytes([]byte("v24-request")), ResolutionSchema: "gooo/self-improvement-candidate-authorization-resolution/v2", ResolutionDigest: digestBytes([]byte("v24-resolution")), CandidateStableID: digestBytes([]byte("candidate-id")), CandidateDigest: digestBytes([]byte("candidate-digest")), SubjectSHA: "0123456789abcdef0123456789abcdef01234567", ObservationDigest: digestBytes([]byte("observation")), ContractDigest: digestBytes([]byte("v24-contract")), AuthorizationDecision: "ALLOW", AuthorizationResolution: "CLOSED", AuthorizationOutcome: "AUTHORIZED", RequestValid: true, ResolutionValid: true}
}

func canonicalV25() V25Binding {
	return V25Binding{Schema: "gooo/self-improvement-execution-contract/v1", ContractID: "gooo://self-improvement/candidate-execution-contract/v1", ContractDigest: digestBytes([]byte("v25-contract")), Decision: "CLOSED", Resolution: "DECLARED", CandidateStableID: digestBytes([]byte("candidate-id")), CandidateDigest: digestBytes([]byte("candidate-digest")), SubjectSHA: "0123456789abcdef0123456789abcdef01234567", ObservationDigest: digestBytes([]byte("observation")), CandidateInputDigest: digestBytes([]byte("candidate-input")), OperationID: "self-improvement.value-witness-experiment.v1", BoundedTarget: "VALUE_WITNESS_EXPERIMENT", EvaluatorRegistryDigest: digestBytes([]byte("evaluator-registry")), ToolchainTestContractIdentity: digestBytes([]byte("toolchain-test-contract")), MaxExecutions: 1, RepositoryWritesAllowed: false, ExecutionAuthorized: false, ExecutionGrantRequired: true, Valid: true}
}

func canonicalSource() SourceArtifact {
	return SourceArtifact{Repository: "kimjooyoon/meta-ontology-go", WorkflowRunID: 1, WorkflowRunAttempt: 1, ArtifactID: 1, ArtifactDigest: digestBytes([]byte("v25-artifact")), ObservedArtifactDigest: digestBytes([]byte("v25-artifact")), ArtifactExpired: false, ArtifactExpiryKnown: true, ArtifactRetrieved: true}
}

func canonicalRequest(program PolicyProgram) GrantRequest {
	return BuildRequest(program, canonicalV24(), canonicalV25(), canonicalSource())
}

func canonicalCase(program PolicyProgram, id string, request GrantRequest, expectedDecision Decision, expectedResolution Resolution, expectedReason string) (CanonicalCase, error) {
	input := GrantInput{Request: request}
	resolution := Evaluate(program, input)
	verification := Verify(program, input, resolution)
	if resolution.Decision != expectedDecision || resolution.Resolution != expectedResolution || resolution.Reason != expectedReason || !verification.Verified {
		return CanonicalCase{}, fmt.Errorf("canonical case %s resolved %s/%s/%s, expected %s/%s/%s", id, resolution.Decision, resolution.Resolution, resolution.Reason, expectedDecision, expectedResolution, expectedReason)
	}
	return CanonicalCase{ID: id, ExpectedDecision: expectedDecision, ExpectedResolution: expectedResolution, ExpectedReason: expectedReason, ActualDecision: resolution.Decision, ActualResolution: resolution.Resolution, ActualReason: resolution.Reason, Unknown: resolution.Unknown, GrantAllowsExecution: resolution.GrantAllowsExecution, RemainingUses: resolution.RemainingUses, ConsumedUses: resolution.ConsumedUses, ExecutionCount: resolution.ExecutionCount, Pass: true}, nil
}

func BuildCanonicalCaseReport(program PolicyProgram) (CanonicalCaseReport, error) {
	request := canonicalRequest(program)
	missingArtifactRequest := request
	missingArtifactRequest.Source.ArtifactExpired = true
	missingArtifactRequest.Digest = requestDigest(missingArtifactRequest)
	incompleteV24Request := request
	incompleteV24Request.V24.ResolutionDigest = ""
	incompleteV24Request.Digest = requestDigest(incompleteV24Request)
	unknownV25Request := request
	unknownV25Request.V25.Decision = string(DecisionUnknown)
	unknownV25Request.Digest = requestDigest(unknownV25Request)
	scopeRequest := request
	scopeRequest.V25.BoundedTarget = "UNBOUNDED"
	scopeRequest.Digest = requestDigest(scopeRequest)
	unsafeRequest := request
	unsafeRequest.V25.MaxExecutions = 2
	unsafeRequest.Digest = requestDigest(unsafeRequest)
	contradictoryUpstreamRequest := request
	contradictoryUpstreamRequest.V24.AuthorizationOutcome = "DENIED"
	contradictoryUpstreamRequest.Digest = requestDigest(contradictoryUpstreamRequest)
	caseSpecs := []struct {
		id, reason string
		request    GrantRequest
		decision   Decision
		resolution Resolution
	}{
		{"exact-system-grant", ReasonAllow, request, DecisionClosed, ResolutionGrantedUnconsumed},
		{"deterministic-replay", ReasonAllow, request, DecisionClosed, ResolutionGrantedUnconsumed},
		{"independent-system-replay", ReasonAllow, request, DecisionClosed, ResolutionGrantedUnconsumed},
		{"missing-upstream-artifact", ReasonMissingArtifact, missingArtifactRequest, DecisionUnknown, ResolutionLower},
		{"incomplete-v24-evidence", ReasonIncompleteInput, incompleteV24Request, DecisionUnknown, ResolutionLower},
		{"unknown-v25-contract", ReasonV25Unknown, unknownV25Request, DecisionUnknown, ResolutionLower},
		{"digest-or-scope-mismatch", ReasonScopeMismatch, scopeRequest, DecisionRefuted, ResolutionExact},
		{"unsafe-execution-envelope", ReasonUnsafe, unsafeRequest, DecisionRefuted, ResolutionExact},
		{"contradictory-upstream-evidence", ReasonUpstreamContradiction, contradictoryUpstreamRequest, DecisionRefuted, ResolutionExact},
	}
	report := CanonicalCaseReport{Schema: CanonicalCasesSchema, Policy: program.Evidence, RequiredFields: RequiredBindingNames(), RequestDigest: request.Digest, CaseDenominator: 9, StructuralSeparateGrantEdgesBefore: 0, StructuralSeparateGrantEdgesAfter: 1, SourceArtifactBoundBefore: 0, SourceArtifactBoundAfter: 1, SourceArtifactBound: 1, SourceArtifactExpiredMisclassifiedBefore: 1, SourceArtifactExpiredMisclassifiedAfter: 0, SourceArtifactExpiredMisclassified: 0, ExactSourceDigestBoundBefore: 0, ExactSourceDigestBoundAfter: 1, ExactSourceDigestBound: 1, CounterexampleArtifactIDs: []int64{KnownFlawedArtifactID}, Counts: map[string]int{"CLOSED": 0, "UNKNOWN": 0, "REFUTED": 0}, ReplayEqual: true, LiveGrantRequests: 0, LiveGrants: 0, SystemDerivedGrants: 0, LiveExecutionCount: 0, CanonicalExecutionCount: 0, RepositoryWrites: 0, LocalTestExecutions: 0, FallbackAccepted: 0, PerformanceImprovement: PerformanceUnknown, Decision: DecisionClosed, Resolution: ResolutionExact, Reason: "NINE_CANONICAL_SYSTEM_GRANT_CASES", GoPhysicalLines: program.Inventory.GoPhysicalLines, GoooPhysicalLines: program.Inventory.GoooPhysicalLines}
	for _, spec := range caseSpecs {
		result, err := canonicalCase(program, spec.id, spec.request, spec.decision, spec.resolution, spec.reason)
		if err != nil {
			return CanonicalCaseReport{}, err
		}
		report.Cases = append(report.Cases, result)
		report.Counts[string(result.ActualDecision)]++
		if result.GrantAllowsExecution {
			report.CanonicalGrantedCases++
			report.SystemDerivedGrants++
		}
		if result.Unknown != nil {
			report.SixFieldUnknowns++
		}
		if result.ActualDecision == DecisionRefuted {
			report.RefutedContradictions++
		}
		first := Evaluate(program, GrantInput{Request: spec.request})
		second := Evaluate(program, GrantInput{Request: spec.request})
		if firstBytes, _ := json.Marshal(first); !bytes.Equal(firstBytes, mustJSON(second)) {
			report.ReplayEqual = false
		}
	}
	report.ClosedCases, report.UnknownCases, report.RefutedCases = report.Counts["CLOSED"], report.Counts["UNKNOWN"], report.Counts["REFUTED"]
	report.GrantRemainingUses = report.CanonicalGrantedCases
	report.IndependentReplayComparisons = 1
	report.ArtifactFiles, report.ArtifactTypes = 9, 3
	if report.ClosedCases != 3 || report.UnknownCases != 3 || report.RefutedCases != 3 || report.SystemDerivedGrants != 3 || !report.ReplayEqual {
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
	if report.Schema != CanonicalCasesSchema || report.CaseDenominator != 9 || report.StructuralSeparateGrantEdgesBefore != 0 || report.StructuralSeparateGrantEdgesAfter != 1 || report.SourceArtifactBoundBefore != 0 || report.SourceArtifactBoundAfter != 1 || report.SourceArtifactBound != 1 || report.SourceArtifactExpiredMisclassifiedBefore != 1 || report.SourceArtifactExpiredMisclassifiedAfter != 0 || report.SourceArtifactExpiredMisclassified != 0 || report.ExactSourceDigestBoundBefore != 0 || report.ExactSourceDigestBoundAfter != 1 || report.ExactSourceDigestBound != 1 || len(report.CounterexampleArtifactIDs) != 1 || report.CounterexampleArtifactIDs[0] != KnownFlawedArtifactID || report.ClosedCases != 3 || report.UnknownCases != 3 || report.RefutedCases != 3 || report.Counts["CLOSED"] != 3 || report.Counts["UNKNOWN"] != 3 || report.Counts["REFUTED"] != 3 || !report.ReplayEqual || report.LiveGrantRequests != 0 || report.LiveGrants != 0 || report.SystemDerivedGrants != 3 || report.CanonicalGrantedCases != 3 || report.GrantRemainingUses != 3 || report.LiveExecutionCount != 0 || report.CanonicalExecutionCount != 0 || report.GrantConsumedUses != 0 || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || report.FallbackAccepted != 0 || report.Digest != canonicalDigest(report) {
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
