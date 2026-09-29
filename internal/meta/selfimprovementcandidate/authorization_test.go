package selfimprovementcandidate

import (
	"encoding/json"
	"testing"
	"testing/fstest"

	valuewitnessinput "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementvaluewitnessinput"
)

const authorizationContractPath = "examples/self-improvement/authorization.gooo"

const authorizationContract = `package selfimprovement
namespace selfimprovement

entity NonExecutingImprovementCandidate id "gooo://self-improvement/entity/non-executing-improvement-candidate"
entity CandidateAuthorizationRequest id "gooo://self-improvement/entity/candidate-authorization-request"
entity CandidateAuthorizationDecision id "gooo://self-improvement/entity/system-derived-candidate-authorization-decision"
entity CandidateAuthorizationResolution id "gooo://self-improvement/entity/candidate-authorization-resolution"

activity RequestCandidateAuthorization(NonExecutingImprovementCandidate) -> CandidateAuthorizationRequest
activity DeriveCandidateAuthorization(CandidateAuthorizationRequest) -> CandidateAuthorizationDecision
activity ResolveCandidateAuthorization(CandidateAuthorizationDecision) -> CandidateAuthorizationResolution
`

func authorizationRepository() fstest.MapFS {
	return fstest.MapFS{authorizationContractPath: &fstest.MapFile{Data: []byte(authorizationContract), Mode: 0o444}}
}

func authorizationCandidateRaw(head string, runID int64) []byte {
	report := Evaluate(validRepository(), candidateContractPath, head, runID, sourceBytes(validSource(head, runID)))
	raw, _ := json.Marshal(report)
	return raw
}

func authorizationMetadata(head string, runID int64) ArtifactMetadata {
	return ArtifactMetadata{Repository: "kimjooyoon/meta-ontology-go", SubjectSHA: head, RunID: runID, RunAttempt: 1,
		ArtifactID: 200, ArtifactName: "self-improvement-nonexecuting-candidate-test",
		ArchiveDigest: fixtureDigest(), SizeBytes: 1234}
}

func authorizationJSONRoundTrip[T any](t *testing.T, value T) T {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var roundTripped T
	if err := json.Unmarshal(raw, &roundTripped); err != nil {
		t.Fatal(err)
	}
	return roundTripped
}

func TestBuildAuthorizationRequestBindsExactCandidateAndLeavesLiveUnknown(t *testing.T) {
	head, sourceRunID, artifactRunID := fixtureSHA("e"), int64(45), int64(46)
	raw := authorizationCandidateRaw(head, sourceRunID)
	request, err := BuildAuthorizationRequest(authorizationRepository(), authorizationContractPath, raw, authorizationMetadata(head, artifactRunID))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAuthorizationRequest(request); err != nil {
		t.Fatal(err)
	}
	if request.Candidate.SubjectSHA != head || request.Candidate.SourceWorkflowRunID != sourceRunID ||
		request.Artifact.RunID != artifactRunID ||
		request.Candidate.CandidateDigest == "" || request.Candidate.CandidateReportDigest == "" ||
		request.Candidate.SourceObservationDigest == "" || request.Candidate.PolicyDigest == "" ||
		request.Candidate.ContractCanonicalDigest == "" || request.Candidate.ScopeDigest == "" ||
		request.ExecutionAllowed || request.RepositoryWrites != 0 || request.LiveAuthorized != 0 || request.LiveState != "UNKNOWN" {
		t.Fatalf("request is not an exact read-only binding: %+v", request)
	}
	if request.Metrics.StructuralUnboundEdgesBefore != 1 || request.Metrics.StructuralUnboundEdgesAfter != 0 {
		t.Fatalf("structural edge transition was not recorded: %+v", request.Metrics)
	}
	second, err := BuildAuthorizationRequest(authorizationRepository(), authorizationContractPath, raw, authorizationMetadata(head, artifactRunID))
	if err != nil {
		t.Fatal(err)
	}
	firstBytes, _ := json.Marshal(request)
	secondBytes, _ := json.Marshal(second)
	if string(firstBytes) != string(secondBytes) {
		t.Fatal("authorization request was not deterministic")
	}
}

func TestVerifyAuthorizationResolutionAcceptsExactJSONRoundTripValue(t *testing.T) {
	head := fixtureSHA("a")
	request, err := BuildAuthorizationRequest(authorizationRepository(), authorizationContractPath,
		authorizationCandidateRaw(head, 50), authorizationMetadata(head, 500))
	if err != nil {
		t.Fatal(err)
	}
	resolution := ResolveAuthorization(request)
	if resolution.Decision != AuthorizationClosed {
		t.Fatalf("expected CLOSED resolution, got %+v", resolution)
	}
	roundTrippedRequest := authorizationJSONRoundTrip(t, request)
	roundTrippedResolution := authorizationJSONRoundTrip(t, resolution)
	if roundTrippedRequest.Candidate.ExecutionInput == roundTrippedResolution.Candidate.ExecutionInput {
		t.Fatal("round-trip regression fixture unexpectedly reused the same input pointer")
	}
	if err := VerifyAuthorizationResolution(roundTrippedRequest, roundTrippedResolution); err != nil {
		t.Fatalf("exact JSON round-trip was not independently verified: %v", err)
	}
}

func TestResolveAuthorizationRefutesNestedExecutionInputContradictions(t *testing.T) {
	head := fixtureSHA("b")
	request, err := BuildAuthorizationRequest(authorizationRepository(), authorizationContractPath, authorizationCandidateRaw(head, 51), authorizationMetadata(head, 51))
	if err != nil {
		t.Fatal(err)
	}
	mutations := []struct {
		name   string
		mutate func(*valuewitnessinput.ExecutionInput)
	}{
		{name: "source-bytes", mutate: func(input *valuewitnessinput.ExecutionInput) { input.Source.Bytes += "\n" }},
		{name: "source-digest", mutate: func(input *valuewitnessinput.ExecutionInput) {
			input.Source.Digest = digestBytes([]byte("different-source"))
		}},
		{name: "corpus", mutate: func(input *valuewitnessinput.ExecutionInput) { input.Corpus[0].ExpectedOutput++ }},
		{name: "activity", mutate: func(input *valuewitnessinput.ExecutionInput) { input.Activity.ValueProgram = "int.add:2" }},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			mutated := authorizationJSONRoundTrip(t, request)
			mutation.mutate(mutated.Candidate.ExecutionInput)
			mutated.Digest = requestDigest(mutated)
			resolution := ResolveAuthorization(mutated)
			if resolution.Decision != AuthorizationRefuted {
				t.Fatalf("nested execution-input contradiction was not REFUTED: %+v", resolution)
			}
		})
	}
}

func TestVerifyAuthorizationResolutionDistinguishesNilAndPresentExecutionInput(t *testing.T) {
	head := fixtureSHA("c")
	request, err := BuildAuthorizationRequest(authorizationRepository(), authorizationContractPath, authorizationCandidateRaw(head, 52), authorizationMetadata(head, 52))
	if err != nil {
		t.Fatal(err)
	}
	nilRequest := authorizationJSONRoundTrip(t, request)
	nilRequest.Candidate.ExecutionInput = nil
	nilRequest.Digest = requestDigest(nilRequest)
	if resolution := ResolveAuthorization(nilRequest); resolution.Decision != AuthorizationUnknown {
		t.Fatalf("nil execution input was treated as present: %+v", resolution)
	}
	resolution := ResolveAuthorization(request)
	nilResolution := authorizationJSONRoundTrip(t, resolution)
	nilResolution.Candidate.ExecutionInput = nil
	nilResolution.Digest = resolutionDigest(nilResolution)
	if err := VerifyAuthorizationResolution(request, nilResolution); err == nil {
		t.Fatal("nil and present execution inputs were treated as equal")
	}
}

func TestResolveAuthorizationDerivesEligibilityAndKeepsExecutionClosed(t *testing.T) {
	head := fixtureSHA("f")
	request, err := BuildAuthorizationRequest(authorizationRepository(), authorizationContractPath, authorizationCandidateRaw(head, 46), authorizationMetadata(head, 46))
	if err != nil {
		t.Fatal(err)
	}
	resolution := ResolveAuthorization(request)
	if resolution.Decision != AuthorizationClosed || resolution.Outcome != AuthorizationAuthorized ||
		resolution.Reason != AuthorizationClosedAllow || resolution.LiveAuthorized != 1 ||
		resolution.LiveState != "CLOSED/AUTHORIZED" || resolution.ExecutionAuthorized {
		t.Fatalf("unexpected system-derived resolution: %+v", resolution)
	}
	if err := VerifyAuthorizationResolution(request, resolution); err != nil {
		t.Fatal(err)
	}
	if resolution.Authorization == nil || !resolution.Authorization.Authorized || resolution.Authorization.ExecutionAuthorized ||
		resolution.Authorization.Schema != "gooo/semantic-self-adoption-system-decision/v1" ||
		resolution.Authorization.AuthorizationMode != "system-derived-exact-evidence" ||
		resolution.Authorization.DecisionSource != AuthorizationEvidenceSource ||
		resolution.Authorization.DecisionRule != AuthorizationEvidenceRule ||
		resolution.Authorization.RepositoryWrites != 0 || resolution.Authorization.LocalTestExecutions != 0 {
		t.Fatalf("system authorization gained forbidden authority: %+v", resolution.Authorization)
	}
}

func TestResolveAuthorizationMissingSystemEvidenceIsExactSixFieldUnknown(t *testing.T) {
	head := fixtureSHA("1")
	request, err := BuildAuthorizationRequest(authorizationRepository(), authorizationContractPath, authorizationCandidateRaw(head, 47), authorizationMetadata(head, 47))
	if err != nil {
		t.Fatal(err)
	}
	request.Candidate.PolicyDigest = ""
	request.Digest = requestDigest(request)
	resolution := ResolveAuthorization(request)
	if resolution.Decision != AuthorizationUnknown || resolution.Reason != AuthorizationFieldReason || resolution.Authorization != nil ||
		resolution.Unknown == nil || resolution.Unknown.Stage != AuthorizationUnknownStage || resolution.Unknown.Step != AuthorizationUnknownStep ||
		resolution.Unknown.UnknownClass != AuthorizationUnknownClass ||
		len(resolution.Unknown.BlockedBy) != 1 || resolution.Unknown.BlockedBy[0] != "candidate_adapter_field" ||
		resolution.Unknown.NextOperation != "PROVIDE_COMPLETE_CANDIDATE_BINDING" {
		t.Fatalf("incomplete system evidence was not causal UNKNOWN: %+v", resolution)
	}
	if err := VerifyAuthorizationResolution(request, resolution); err != nil {
		t.Fatal(err)
	}
}

func TestResolveAuthorizationRefutesCandidateContradictions(t *testing.T) {
	head := fixtureSHA("2")
	request, err := BuildAuthorizationRequest(authorizationRepository(), authorizationContractPath, authorizationCandidateRaw(head, 48), authorizationMetadata(head, 48))
	if err != nil {
		t.Fatal(err)
	}
	wrong := authorizationJSONRoundTrip(t, request)
	wrong.Candidate.ExecutionInput.CandidateDigest = digestBytes([]byte("not-the-candidate"))
	wrong.Digest = requestDigest(wrong)
	resolution := ResolveAuthorization(wrong)
	if resolution.Decision != AuthorizationRefuted || resolution.Reason != AuthorizationRefutedReason {
		t.Fatalf("candidate contradiction was not REFUTED: %+v", resolution)
	}
	if err := VerifyAuthorizationResolution(wrong, resolution); err != nil {
		t.Fatal(err)
	}
}

func TestBuildCanonicalAuthorizationCasesFixesNineCaseDenominator(t *testing.T) {
	head := fixtureSHA("3")
	request, err := BuildAuthorizationRequest(authorizationRepository(), authorizationContractPath, authorizationCandidateRaw(head, 49), authorizationMetadata(head, 49))
	if err != nil {
		t.Fatal(err)
	}
	report, err := BuildCanonicalCases(request)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateCanonicalCases(report); err != nil {
		t.Fatal(err)
	}
	if report.CaseDenominator != 9 || report.ClosedCases != 3 || report.UnknownCases != 3 || report.RefutedCases != 3 ||
		report.Metrics.StructuralUnboundEdgesBefore != 1 || report.Metrics.StructuralUnboundEdgesAfter != 0 ||
		report.LiveAuthorized != 0 || report.Roundtrip.AuthorizationRoundtripExactBefore != 0 ||
		report.Roundtrip.AuthorizationRoundtripExactAfter != 1 || report.Roundtrip.PointerIdentityDependencyBefore != 1 ||
		report.Roundtrip.PointerIdentityDependencyAfter != 0 || report.Roundtrip.CounterexampleRunID != 33926584593 {
		t.Fatalf("unexpected canonical metrics: %+v", report)
	}
}
