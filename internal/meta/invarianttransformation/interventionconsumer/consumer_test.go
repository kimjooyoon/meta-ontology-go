package interventionconsumer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/executor"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/intervention"
)

const consumerTestHead = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

const consumerTestSource = `package invarianttransformation
namespace meta
entity Transformation id "gooo://invariant-transformation/value/transformation"
activity PreservedTranslation() -> Transformation computes "case=preserved-translation;kind=PRESERVED;input=2;candidate=add:1;expected=3;invariant=candidate-output-equals-expected;invariant-id=candidate-output-equals-expected-v1;domain=bounded-fixture-input-domain-v1;replay=add:1;effect=none"
activity SemanticViolation() -> Transformation computes "case=semantic-violation;kind=VIOLATION;input=2;candidate=add:2;expected=3;invariant=candidate-output-equals-expected;invariant-id=candidate-output-equals-expected-v1;domain=bounded-fixture-input-domain-v1;replay=add:2;effect=none"
activity MissingRegressionWitness() -> Transformation computes "case=missing-regression-witness;kind=EVIDENCE_MISSING;input=2;candidate=add:1;expected=3;invariant=candidate-output-equals-expected;invariant-id=candidate-output-equals-expected-v1;domain=bounded-fixture-input-domain-v1;replay=unavailable;effect=none"
activity ApprovedArtifact() -> Transformation computes "case=approved-artifact;kind=APPROVED_ARTIFACT;input=5;candidate=add:1;expected=6;invariant=candidate-output-equals-expected;invariant-id=candidate-output-equals-expected-v1;domain=bounded-fixture-input-domain-v1;replay=add:1;effect=approved-artifact"
`

func testDependency(t *testing.T, report intervention.Report, source []byte) DependencyBoundary {
	t.Helper()
	fixture, err := parseFixtureCase(source, "approved-artifact")
	if err != nil {
		t.Fatal(err)
	}
	receipt := reconstructReceipt(fixture, source, consumerTestHead)
	judgment := reconstructJudgment(receipt)
	path := filepath.Join(os.Getenv("RUNNER_TEMP"), "observed.bin")
	effect, err := executor.Emit(receipt, judgment, consumerTestHead, path)
	if err != nil {
		t.Fatal(err)
	}
	return DependencyBoundary{ArtifactEvidence: effect.Artifact, UnknownEffectScopes: report.UnknownEffectScopes}
}

func TestConsumerReconstructsBothInterventions(t *testing.T) {
	t.Setenv("RUNNER_TEMP", t.TempDir())
	report, err := intervention.Build([]byte(consumerTestSource), consumerTestHead)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	audit, err := VerifyReport(raw, []byte(consumerTestSource), consumerTestHead, testDependency(t, report, []byte(consumerTestSource)))
	if err != nil {
		t.Fatal(err)
	}
	if audit.ReconstructedCases != 3 || audit.ExpectedCases != 3 || audit.ActualReplay != 3 || audit.ExpectedActualReplay != 3 || audit.Decision != "PASS" || !audit.ArtifactObserved {
		t.Fatalf("audit=%+v", audit)
	}
}

func TestConsumerRejectsCoherentResealedTamper(t *testing.T) {
	t.Setenv("RUNNER_TEMP", t.TempDir())
	report, err := intervention.Build([]byte(consumerTestSource), consumerTestHead)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	audit, err := VerifyReport(raw, []byte(consumerTestSource), consumerTestHead, testDependency(t, report, []byte(consumerTestSource)))
	if err != nil {
		t.Fatal(err)
	}
	if audit.CoherentTamperRejected != 1 || audit.ExpectedCoherentTamperRejections != 1 {
		t.Fatalf("tamper regression=%+v", audit)
	}
}
