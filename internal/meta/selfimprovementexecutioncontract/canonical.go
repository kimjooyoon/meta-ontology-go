package selfimprovementexecutioncontract

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
)

func canonicalInput() ContractInput {
	candidate := CandidateEvidence{
		StableID:       digestBytes([]byte("v25-canonical-candidate")),
		Digest:         digestBytes([]byte("v25-canonical-candidate-digest")),
		SubjectSHA:     "0123456789abcdef0123456789abcdef01234567",
		ExperimentKind: CandidateExperimentKind, MetaOperation: CandidateMetaOperation,
		InputDigest: digestBytes([]byte("v25-canonical-candidate-input")),
	}
	observation := ObservationEvidence{
		Schema: ObservationSchema, CandidateInputDigest: candidate.InputDigest, CandidateStableID: candidate.StableID,
		SubjectSHA: candidate.SubjectSHA, Phase: Phase(KnownPhase), OperationID: OperationID(KnownOperationID),
		BoundedTarget: KnownBoundedTarget, SourceSpans: []SourceSpan{{SourceID: "examples/self-improvement/project.gooo", StartLine: 1, EndLine: 12}},
		ObservedCount: 1, ObservedCountKnown: true,
	}
	observation.ObservationDigest = observationDigest(observation)
	candidate.ObservationDigest = observation.ObservationDigest
	return ContractInput{
		Candidate: candidate,
		Authorization: AuthorizationBinding{
			RequestSchema:  selfimprovementcandidate.AuthorizationRequestSchema,
			RequestDigest:  digestBytes([]byte("v25-canonical-authorization-request")),
			ContractID:     selfimprovementcandidate.AuthorizationContractID,
			ContractDigest: digestBytes([]byte("v24-authorization-contract")),
			Scope:          selfimprovementcandidate.AuthorizationScope, Decision: "DENY",
			ExecutionAllowed: false, RepositoryWrites: 0, LocalTestExecutions: 0,
			LiveAuthorized: 0, LiveState: "UNKNOWN",
		},
		Observation: observation, Registry: KnownRegistry(),
	}
}

func BuildCanonicalCaseReport(program PolicyProgram) CanonicalCaseReport {
	base := canonicalInput()
	cases := []struct {
		id       string
		input    func(ContractInput) ContractInput
		expected Decision
		reason   string
	}{
		{"EXACT_DECLARATION", func(input ContractInput) ContractInput { return input }, DecisionClosed, ReasonDeclared},
		{"DETERMINISTIC_REPLAY", func(input ContractInput) ContractInput { return input }, DecisionClosed, ReasonDeclared},
		{"EXPLICIT_SAFETY_LOCK", func(input ContractInput) ContractInput { input.Authorization.Decision = "ALLOW"; return input }, DecisionClosed, ReasonDeclared},
		{"MISSING_OBSERVATION_EVIDENCE", func(input ContractInput) ContractInput { input.Observation = ObservationEvidence{}; return input }, DecisionUnknown, ReasonMissingObservation},
		{"UNSUPPORTED_OPERATION_MAPPING", func(input ContractInput) ContractInput {
			input.Candidate.MetaOperation = "unregistered-operation"
			return input
		}, DecisionUnknown, ReasonUnsupportedMapping},
		{"MISSING_REGISTRY_IDENTITY", func(input ContractInput) ContractInput { input.Registry = RegistryEvidence{}; return input }, DecisionUnknown, ReasonMissingRegistry},
		{"CANDIDATE_DIGEST_MISMATCH", func(input ContractInput) ContractInput {
			input.Observation.CandidateInputDigest = digestBytes([]byte("different-input"))
			return input
		}, DecisionRefuted, ReasonCandidateConflict},
		{"SOURCE_SPAN_CONTRADICTION", func(input ContractInput) ContractInput {
			input.Observation.SourceSpans[0].EndLine = 0
			input.Observation.ObservationDigest = observationDigest(input.Observation)
			return input
		}, DecisionRefuted, ReasonSpanConflict},
		{"MAX_EXECUTIONS_OR_WRITE_AUTHORITY", func(input ContractInput) ContractInput { input.Registry.MaxExecutions = 2; return input }, DecisionRefuted, ReasonSafetyConflict},
	}
	report := CanonicalCaseReport{
		Schema: CanonicalCasesSchema, Policy: program.Evidence, RequiredFields: RequiredFieldNames(), CaseDenominator: 9,
		BoundFields: PreExecutionRequiredField, MissingFields: 0, ContradictoryFields: 0,
		StructuralCandidateToOperationBefore: 0, StructuralCandidateToOperationAfter: 1,
		Counts: map[string]int{string(DecisionClosed): 0, string(DecisionUnknown): 0, string(DecisionRefuted): 0},
		Cases:  make([]CanonicalCase, 0, len(cases)), LiveExecutionCount: 0, CanonicalExecutionCount: 0,
		ExecutionGrants: 0, RepositoryWrites: 0, LocalTestExecutions: 0, FallbackAccepted: 0,
		IndependentReplayComparisons: 1, ArtifactFiles: 9, ArtifactTypes: 3,
		PerformanceImprovement: PerformanceUnknown, Decision: DecisionClosed, Resolution: ResolutionDeclared,
		Reason: "NINE_CANONICAL_PRE_EXECUTION_CASES",
	}
	report.GoPhysicalLines = program.Inventory.GoPhysicalLines
	report.GoooPhysicalLines = program.Inventory.GoooPhysicalLines
	report.ReplayEqual = true
	for _, current := range cases {
		caseInput := current.input(base)
		resolution := Evaluate(program, caseInput)
		verification := Verify(program, caseInput, resolution)
		replay := Evaluate(program, caseInput)
		replayEqual := resolutionDigest(resolution) == resolutionDigest(replay)
		if !replayEqual {
			report.ReplayEqual = false
		}
		pass := resolution.Decision == current.expected && resolution.Reason == current.reason && replayEqual && verification.Verified
		report.Cases = append(report.Cases, CanonicalCase{ID: current.id, ExpectedDecision: current.expected, ExpectedReason: current.reason,
			ActualDecision: resolution.Decision, ActualReason: resolution.Reason, Unknown: resolution.Unknown,
			MissingFields: cloneStrings(resolution.MissingFields), ContradictoryFields: cloneStrings(resolution.ContradictoryFields), Pass: pass})
		report.Counts[string(resolution.Decision)]++
	}
	report.ClosedCases = report.Counts[string(DecisionClosed)]
	report.UnknownCases = report.Counts[string(DecisionUnknown)]
	report.RefutedCases = report.Counts[string(DecisionRefuted)]
	for _, current := range report.Cases {
		if !current.Pass {
			report.Decision, report.Resolution, report.Reason = DecisionRefuted, ResolutionExact, fmt.Sprintf("CANONICAL_CASE_FAILED:%s", current.ID)
			break
		}
	}
	report.Digest = canonicalDigest(report)
	return report
}
