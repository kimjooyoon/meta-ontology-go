package metacircularboundary

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundaryconsumer"
)

func TestBoundarySuite(t *testing.T) {
	input, report := baselineTestReport(t)
	if err := metacircularboundaryconsumer.Judge(report, input); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionPass || report.Summary.CasesTotal != len(contractCases()) || report.Summary.CasesPassed != len(contractCases()) || report.Summary.ExplicitAuthorizations != 1 || report.Summary.AllowedExecutions != 1 || report.Summary.ReplayMatches != len(contractCases()) || report.Summary.ReceiptSelfSealValid != len(contractCases()) || report.Summary.ExecutionArtifactsObserved != 1 {
		t.Fatalf("unexpected boundary evidence: %+v", report)
	}
}

func TestDescriptionCannotMintAuthority(t *testing.T) {
	input, report := baselineTestReport(t)
	caseResult := report.Cases[0]
	if caseResult.Observation.Authorization != AuthorizationDenied || caseResult.Observation.Execution != ExecutionBlocked || caseResult.Observation.Reason != ReasonDescriptionOnly || caseResult.Receipt.GrantDigest == "" {
		t.Fatalf("description gained authority: %+v", caseResult.Observation)
	}
	_ = input
}

func TestForgedAndOutOfScopeCapabilitiesAreBlocked(t *testing.T) {
	_, report := baselineTestReport(t)
	for _, index := range []int{2, 3} {
		if report.Cases[index].Observation.Authorization != AuthorizationDenied || report.Cases[index].Observation.Execution != ExecutionBlocked || report.Cases[index].Observation.ExecutionArtifactPresent {
			t.Fatalf("case %d crossed boundary: %+v", index, report.Cases[index].Observation)
		}
	}
}

func TestConsumerRejectsSharedWrongCaseFactCounterexample(t *testing.T) {
	input, report := baselineTestReport(t)
	tampered := report
	tampered.Cases = append([]CaseResult(nil), report.Cases...)
	tampered.Cases[0].Attempt.RequestExecution = false
	tampered.ReportDigest = sealReport(tampered).ReportDigest
	if err := metacircularboundaryconsumer.Judge(tampered, input); err == nil {
		t.Fatal("consumer accepted a report with a shared wrong case fact")
	}
}

func TestExpectedDecisionCannotMintAuthorization(t *testing.T) {
	input, report := baselineTestReport(t)
	definition := report.Cases[1].Definition
	definition.ExpectedDecision = DecisionPass
	observation := report.Cases[1].Observation
	observation.Decision = DecisionFailClosed
	observation.Authorization = AuthorizationDenied
	observation.Execution = ExecutionBlocked
	observation.Reason = ReasonForgedCapability
	receipt := buildReceipt(report.Source, definition, report.Cases[1].Attempt, report.Cases[1].Grant, observation, ExecutionArtifact{}, false)
	if receipt.Decision == definition.ExpectedDecision || receipt.Decision != DecisionFailClosed || receipt.Authorization != AuthorizationDenied {
		t.Fatalf("ExpectedDecision minted authorization: definition=%+v receipt=%+v", definition, receipt)
	}
	_ = input
}

func TestUnknownComputationDataRemainsOpen(t *testing.T) {
	source := testSource(t)
	mutated := bytes.Replace(source, []byte("request=NONE|request_execution=true"), []byte("request=UNKNOWN|request_execution=true"), 1)
	input := Input{Path: ExpectedSourcePath, HeadSHA: strings.Repeat("f", 40), Source: mutated, GrantEvidence: testGrant(t, mutated), EffectEvidence: testEffect(t)}
	report := Evaluate(input)
	if report.Decision != DecisionOpen || report.Resolution != ResolutionLower || report.Reason != ReasonCaseDataUnknown || report.Coordinate != (Coordinate{Stage: "PARSE_COMPUTES", Step: "read-case-facts", Reason: ReasonCaseDataUnknown}) {
		t.Fatalf("unknown computation was not lower-resolution open: %+v", report)
	}
	if report.Cases[0].Observation.Decision != DecisionOpen || report.Cases[0].Observation.Reason != ReasonCaseDataUnknown {
		t.Fatalf("unknown case fact was not open: %+v", report.Cases[0])
	}
}

func TestContradictoryCapabilityEvidenceIsRefuted(t *testing.T) {
	source := testSource(t)
	mutated := bytes.Replace(source, []byte("request=READ_ONLY|request_execution=true"), []byte("request=READ_ONLY|request=NONE|request_execution=true"), 1)
	input := Input{Path: ExpectedSourcePath, HeadSHA: strings.Repeat("1", 40), Source: mutated, GrantEvidence: testGrant(t, mutated), EffectEvidence: testEffect(t)}
	report := Evaluate(input)
	caseResult := report.Cases[1]
	if report.Decision != DecisionRefuted || report.Resolution != ResolutionExact || report.Reason != ReasonContradictory || caseResult.Observation.Decision != DecisionRefuted || caseResult.Receipt.Decision != DecisionRefuted {
		t.Fatalf("contradictory evidence was not refuted: report=%+v case=%+v", report, caseResult)
	}
	if caseResult.Receipt.ClaimTransitions[1].After != "REFUTED" || caseResult.Receipt.ClaimTransitions[2].After != "UNKNOWN" {
		t.Fatalf("contradictory claim state did not persist: %+v", caseResult.Receipt.ClaimTransitions)
	}
	if err := metacircularboundaryconsumer.Judge(report, input); err != nil {
		t.Fatalf("independent consumer rejected valid refutation: %v", err)
	}
}

func TestUnknownPermissionEvidenceRemainsLowerResolution(t *testing.T) {
	source := testSource(t)
	input := Input{Path: ExpectedSourcePath, HeadSHA: strings.Repeat("2", 40), Source: source, GrantEvidence: testGrant(t, source), EffectEvidence: testEffectWithAuthority(t, AuthorityUnknown)}
	report := Evaluate(input)
	want := Coordinate{Stage: "OBSERVE_EFFECT", Step: "resolve-workspace-permission-and-output", Reason: ReasonEffectUnknown}
	if report.Decision != DecisionOpen || report.Resolution != ResolutionLower || report.Reason != ReasonEffectUnknown || report.Coordinate != want {
		t.Fatalf("unknown permission was not lower-resolution open: %+v", report)
	}
}

func baselineTestReport(t *testing.T) (Input, Report) {
	t.Helper()
	source := testSource(t)
	grant := testGrant(t, source)
	effect := testEffect(t)
	input := Input{Path: ExpectedSourcePath, HeadSHA: strings.Repeat("a", 40), Source: source, GrantEvidence: grant, EffectEvidence: effect}
	initial := Evaluate(input)
	artifact, err := ExecuteReadOnlyMetaOperation(initial.Source, initial.Cases[1].Grant, initial.Cases[1].Definition.ID)
	if err != nil {
		t.Fatal(err)
	}
	input.ExecutionArtifacts = []ExecutionArtifact{artifact}
	run := Evaluate(input)
	replay := testReplay(t, run)
	input.ReplayEvidence = replay
	return input, Evaluate(input)
}

func testSource(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../../examples/meta-circular-boundary/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func testGrant(t *testing.T, source []byte) []byte {
	t.Helper()
	observed, err := observeSource(ExpectedSourcePath, source)
	if err != nil {
		t.Fatal(err)
	}
	artifact := ExternalGrantArtifact{Schema: GrantSchema, Producer: GrantProducer, Policy: "READ_ONLY_EXTERNAL_GRANT_POLICY_V1", Grants: []ExternalGrant{
		{CaseID: "description-only", Decision: GrantDeny, Reason: ReasonDescriptionOnly},
		{CaseID: "explicit-read-only-capability", Decision: GrantDecision, Reason: ReasonExplicitCapability, Issuer: "external-authority", SubjectDigest: observed.SemanticDigest, Operation: MetaOperationID, Scope: ScopeReadOnly, Handle: "grant://external-authority/read-only/explicit"},
		{CaseID: "forged-capability", Decision: GrantDecision, Reason: ReasonForgedCapability, Issuer: "forged-authority", SubjectDigest: observed.SemanticDigest, Operation: MetaOperationID, Scope: ScopeReadOnly, Handle: "grant://forged/handle"},
		{CaseID: "write-capability-out-of-scope", Decision: GrantDecision, Reason: ReasonOutOfScopeCapability, Issuer: "external-authority", SubjectDigest: observed.SemanticDigest, Operation: MetaOperationID, Scope: ScopeWrite, Handle: "grant://external-authority/write"},
	}}
	encoded, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func testEffect(t *testing.T) []byte {
	return testEffectWithAuthority(t, AuthorityDenied)
}

func testEffectWithAuthority(t *testing.T, authority string) []byte {
	t.Helper()
	digest := "sha256:" + strings.Repeat("a", 64)
	encoded, err := json.Marshal(EffectEvidence{Schema: EffectSchema, Producer: EffectProducer, OutputPath: "runner-temp/meta-circular-boundary", TrackedBeforeDigest: digest, TrackedAfterDigest: digest, UntrackedBeforeDigest: digest, UntrackedAfterDigest: digest, OutputOutsideRepository: true, PermissionEvidence: "workflow-contents-read-only", MutationAuthority: authority})
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func testReplay(t *testing.T, report Report) []byte {
	t.Helper()
	evidence := ReplayEvidence{Schema: ReplaySchema, Producer: ReplayProducer, Equal: true}
	for _, item := range report.Receipts {
		evidence.ReceiptDigestsA = append(evidence.ReceiptDigestsA, item.ReceiptDigest)
		evidence.ReceiptDigestsB = append(evidence.ReceiptDigestsB, item.ReceiptDigest)
		if item.ExecutionArtifact == nil {
			evidence.ExecutionDigestsA = append(evidence.ExecutionDigestsA, "")
			evidence.ExecutionDigestsB = append(evidence.ExecutionDigestsB, "")
		} else {
			evidence.ExecutionDigestsA = append(evidence.ExecutionDigestsA, item.ExecutionArtifact.OutputDigest)
			evidence.ExecutionDigestsB = append(evidence.ExecutionDigestsB, item.ExecutionArtifact.OutputDigest)
		}
	}
	encoded, err := json.Marshal(evidence)
	if err != nil {
		t.Fatal(err)
	}
	evidence.EvidenceDigest = digestBytes(encoded)
	encoded, err = json.Marshal(evidence)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
