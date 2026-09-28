package selfimprovementexecutiongrant

import (
	"os"
	"testing"
)

func testProgram(t *testing.T) PolicyProgram {
	t.Helper()
	program, err := CompilePolicy(os.DirFS("../../.."), PolicyPath)
	if err != nil {
		t.Fatal(err)
	}
	return program
}

func TestPolicyUsesFirstClassSemanticIR(t *testing.T) {
	program := testProgram(t)
	if program.Evidence.StateCount != 12 || program.Evidence.TransitionCount != 9 || program.Evidence.CaseCount != 9 || program.Evidence.ClosedCases != 3 || program.Evidence.UnknownCases != 3 || program.Evidence.RefutedCases != 3 {
		t.Fatalf("execution grant policy denominator drifted: %#v", program.Evidence)
	}
	if program.Evidence.SourceDigest == "" || program.Evidence.CanonicalDigest == "" || program.Evidence.SemanticIRDigest == "" || !program.Inventory.Observed {
		t.Fatal("execution grant policy did not retain parser, formatter, semantic IR, and inventory evidence")
	}
}

func TestCanonicalCasesAreNineSeparateGrantCases(t *testing.T) {
	report, err := BuildCanonicalCaseReport(testProgram(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateCanonicalCases(report); err != nil {
		t.Fatal(err)
	}
	if report.StructuralSeparateGrantEdgesBefore != 0 || report.StructuralSeparateGrantEdgesAfter != 1 || report.SourceArtifactBoundBefore != 0 || report.SourceArtifactBoundAfter != 1 || report.SourceArtifactExpiredMisclassifiedBefore != 1 || report.SourceArtifactExpiredMisclassifiedAfter != 0 || report.ExactSourceDigestBoundBefore != 0 || report.ExactSourceDigestBoundAfter != 1 || len(report.CounterexampleArtifactIDs) != 1 || report.CounterexampleArtifactIDs[0] != KnownFlawedArtifactID || report.CanonicalGrantedCases != 3 || report.SystemDerivedGrants != 3 || report.CanonicalExecutionCount != 0 || report.GrantConsumedUses != 0 || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || report.FallbackAccepted != 0 {
		t.Fatalf("canonical grant boundary drifted: %#v", report)
	}
}

func TestExactSystemEvidenceProducesUnconsumedGrant(t *testing.T) {
	program := testProgram(t)
	request := canonicalRequest(program)
	input := GrantInput{Request: request, Live: true}
	resolution := Evaluate(program, input)
	if resolution.Decision != DecisionClosed || resolution.Resolution != ResolutionGrantedUnconsumed || !resolution.GrantAllowsExecution || resolution.RemainingUses != 1 || resolution.ConsumedUses != 0 || resolution.ExecutionCount != 0 || resolution.OneUseEnforced {
		t.Fatalf("ALLOW crossed the execution or consumption boundary: %#v", resolution)
	}
	if err := ValidateGrantReceipt(*resolution.Receipt); err != nil {
		t.Fatal(err)
	}
	if resolution.SystemEvidence == nil || resolution.SystemEvidence.DecisionSource != DecisionSourceSystem ||
		resolution.SystemEvidence.DecisionRule != DecisionRule || resolution.Metrics.SystemDerivedGrants != 1 {
		t.Fatalf("grant was not derived from system evidence: %#v", resolution)
	}
	verification := Verify(program, input, resolution)
	if !verification.Verified || verification.ExecutionCount != 0 || verification.GrantConsumedUses != 0 {
		t.Fatalf("independent grant replay failed: %#v", verification)
	}
}

func TestSystemGrantDoesNotExecuteOrConsume(t *testing.T) {
	program := testProgram(t)
	request := canonicalRequest(program)
	resolution := Evaluate(program, GrantInput{Request: request})
	if resolution.Decision != DecisionClosed || resolution.Resolution != ResolutionGrantedUnconsumed || !resolution.GrantAllowsExecution || resolution.RemainingUses != 1 || resolution.ConsumedUses != 0 || resolution.ExecutionCount != 0 || resolution.OneUseEnforced {
		t.Fatalf("system grant crossed the execution or consumption boundary: %#v", resolution)
	}
	if err := VerifyGrantResolution(resolution); err != nil {
		t.Fatal(err)
	}
}

func TestLiveExactEvidenceDerivesGrant(t *testing.T) {
	program := testProgram(t)
	request := canonicalRequest(program)
	resolution := Evaluate(program, GrantInput{Request: request, Live: true})
	if resolution.Decision != DecisionClosed || resolution.Resolution != ResolutionGrantedUnconsumed || resolution.SystemEvidence == nil || !resolution.GrantAllowsExecution {
		t.Fatalf("live exact evidence did not derive eligibility: %#v", resolution)
	}
	if resolution.Metrics.LiveGrantRequests != 1 || resolution.Metrics.LiveGrants != 1 || resolution.ExecutionCount != 0 || resolution.ConsumedUses != 0 {
		t.Fatalf("live grant metrics crossed the boundary: %#v", resolution.Metrics)
	}
	if resolution.Metrics.SourceArtifactBound != 1 || resolution.Metrics.SourceArtifactBoundAfter != 1 || resolution.Metrics.ExactSourceDigestBound != 1 || resolution.Metrics.ExactSourceDigestBoundAfter != 1 || resolution.Metrics.SourceArtifactExpiredMisclassified != 0 || resolution.Metrics.SourceArtifactExpiredMisclassifiedAfter != 0 {
		t.Fatalf("retrieved source was not bound exactly: %#v", resolution.Metrics)
	}
	if verification := Verify(program, GrantInput{Request: request, Live: true}, resolution); !verification.Verified {
		t.Fatalf("independent UNKNOWN replay failed: %#v", verification)
	}
}

func TestSourceRetrievalFailureIsUnknownWithoutExpiredMisclassification(t *testing.T) {
	program := testProgram(t)
	request := canonicalRequest(program)
	request.Source.ArtifactRetrieved = false
	request.Source.ArtifactRetrievalError = ReasonSourceRetrievalFailed
	request.Digest = requestDigest(request)
	resolution := Evaluate(program, GrantInput{Request: request, Live: true})
	if resolution.Decision != DecisionUnknown || resolution.Resolution != ResolutionLower || resolution.Reason != ReasonSourceRetrievalFailed || resolution.Unknown == nil || resolution.Unknown.Stage != "FETCH" || resolution.Unknown.Step != "2" || resolution.Unknown.BlockedBy != "source_artifact_retrieval" {
		t.Fatalf("retrieval failure was not preserved as UNKNOWN: %#v", resolution)
	}
	if !contains(resolution.Obligations, "source_artifact_retrieval") || !contains(resolution.Frontier, "retry-exact-source-artifact-retrieval") {
		t.Fatalf("retrieval obligation was not preserved: %#v", resolution)
	}
	if resolution.Metrics.SourceArtifactBound != 0 || resolution.Metrics.SourceArtifactBoundAfter != 0 || resolution.Metrics.ExactSourceDigestBound != 0 || resolution.Metrics.ExactSourceDigestBoundAfter != 0 || resolution.Metrics.SourceArtifactExpiredMisclassified != 0 || resolution.Metrics.SourceArtifactExpiredMisclassifiedAfter != 0 || resolution.Metrics.LiveGrants != 0 || resolution.ExecutionCount != 0 {
		t.Fatalf("retrieval failure crossed a safety boundary: %#v", resolution.Metrics)
	}
	if verification := Verify(program, GrantInput{Request: request, Live: true}, resolution); !verification.Verified {
		t.Fatalf("independent retrieval-failure replay failed: %#v", verification)
	}
}

func TestScopeSafetyAndUpstreamContradictionsRefute(t *testing.T) {
	program := testProgram(t)
	request := canonicalRequest(program)
	scope := request
	scope.V25.BoundedTarget = "UNBOUNDED"
	scope.Digest = requestDigest(scope)
	if resolution := Evaluate(program, GrantInput{Request: scope}); resolution.Decision != DecisionRefuted || resolution.Reason != ReasonScopeMismatch {
		t.Fatalf("scope contradiction was not refuted: %#v", resolution)
	}
	unsafe := request
	unsafe.V25.MaxExecutions = 2
	unsafe.Digest = requestDigest(unsafe)
	if resolution := Evaluate(program, GrantInput{Request: unsafe}); resolution.Decision != DecisionRefuted || resolution.Reason != ReasonUnsafe {
		t.Fatalf("unsafe grant was not refuted: %#v", resolution)
	}
	contradiction := request
	contradiction.V24.AuthorizationOutcome = "DENIED"
	contradiction.Digest = requestDigest(contradiction)
	if resolution := Evaluate(program, GrantInput{Request: contradiction}); resolution.Decision != DecisionRefuted || resolution.Reason != ReasonUpstreamContradiction {
		t.Fatalf("contradictory upstream evidence was not refuted: %#v", resolution)
	}
}
