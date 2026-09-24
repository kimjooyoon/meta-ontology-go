package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicorchestration"
)

func testDigest() string { return strings.Repeat("a", 64) }

func testPolicy() publicorchestration.Policy {
	digest := testDigest()
	return publicorchestration.Policy{
		SourceDigest: digest, SemanticDigest: digest, EvaluatorDigest: digest,
		Operation: publicorchestration.Operation,
	}
}

func testUpstreamReport() publicorchestration.Report {
	digest := testDigest()
	return publicorchestration.Report{
		Schema: publicorchestration.ReportSchema, Decision: publicorchestration.DecisionClosed,
		CaseID: publicorchestration.CaseAuthorizedOrchestration, PolicySourceDigest: digest,
		PolicySemanticDigest: digest, PolicyEvaluatorDigest: digest, Operation: publicorchestration.Operation,
		StatePath: []string{"AUTHORIZE", "CERTIFY", "GENERATE", "VALIDATE", "REUSE", "EVIDENCE"},
		Boundary:  "AUTHORIZE", CandidateDigest: digest, CandidateID: "candidate-1",
		HandoffDigest: digest, AuthorizationDigest: digest, CertificateDigest: digest, ReceiptDigest: digest,
		CaseDenominator: 6, ArtifactDenominator: publicorchestration.ArtifactDenominator,
		RepositoryWrites: 0, LocalTestExecutions: 0, RuntimeComparable: false,
		RuntimeUnknown: &publicorchestration.UnknownState{
			Stage: "UTILITY", Step: "COMPARE_MANUAL_ORCHESTRATED_RUNTIME", Reason: "RUNTIME_MODES_NOT_EQUIVALENT",
			UnknownClass: "INCOMPARABLE", NextOperation: "PREDECLARE_EQUIVALENT_RUNTIME_MEASUREMENT_RULE",
			BlockedBy: []string{"manual-vs-orchestrated-runtime-mode"},
		},
		After: publicorchestration.Metrics{WallMS: 1, PeakRSSKib: 1},
		Comparisons: publicorchestration.Comparisons{
			GeneratedBytesEqual: true, GeneratedSemanticEqual: true, TestContractBytesEqual: true,
			ReceiptBindingEqual: true, ContinuityPreserved: true, SafetyOutcomesPreserved: true,
		},
	}
}

func writeUpstreamFixture(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/orchestration-report.json"
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadUpstreamAcceptsTypedSourceBoundReport(t *testing.T) {
	report := testUpstreamReport()
	got, data, err := readUpstream(writeUpstreamFixture(t, report), testPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if got.CandidateDigest != testDigest() || got.CandidateID != "candidate-1" || len(data) == 0 {
		t.Fatalf("typed upstream evidence was not preserved: %+v", got)
	}
	if !cache.Digest(got.HandoffDigest).Known() {
		t.Fatal("handoff digest was not retained")
	}
}

func TestReadUpstreamRejectsLabelsOnlyReport(t *testing.T) {
	path := writeUpstreamFixture(t, map[string]string{
		"schema": publicorchestration.ReportSchema, "decision": publicorchestration.DecisionClosed,
		"operation": publicorchestration.Operation,
	})
	if _, _, err := readUpstream(path, testPolicy()); err == nil {
		t.Fatal("labels-only upstream report was admitted")
	}
}

func TestReadUpstreamRejectsWriteAuthorityContradiction(t *testing.T) {
	report := testUpstreamReport()
	report.RepositoryWrites = 1
	if _, _, err := readUpstream(writeUpstreamFixture(t, report), testPolicy()); err == nil {
		t.Fatal("contradictory repository-write evidence was admitted")
	}
}
