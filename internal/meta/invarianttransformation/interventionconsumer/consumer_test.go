package interventionconsumer

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/intervention"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
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

func testDependency(t *testing.T, report intervention.Report) DependencyBoundary {
	t.Helper()
	path := t.TempDir() + "/observed.bin"
	data := []byte("independent artifact observation")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return DependencyBoundary{ArtifactEvidence: model.ArtifactEvidence{Path: path, ContentDigest: model.DigestBytes(data), Size: len(data), CaseID: "approved-artifact", SubjectSHA: consumerTestHead, AuthorizationDigest: model.DigestBytes([]byte("authorization")), Producer: model.ProducerID, Executor: model.ExecutorID, Consumer: model.ConsumerID, EffectReceiptDigest: model.DigestBytes([]byte("effect")), RepositoryNetStatusUnchanged: true}, UnknownEffectScopes: report.UnknownEffectScopes}
}

func TestConsumerReconstructsBothInterventions(t *testing.T) {
	report, err := intervention.Build([]byte(consumerTestSource), consumerTestHead)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	audit, err := VerifyReport(raw, []byte(consumerTestSource), consumerTestHead, testDependency(t, report))
	if err != nil {
		t.Fatal(err)
	}
	if audit.ReconstructedCases != 3 || audit.ExpectedCases != 3 || audit.ActualReplay != 3 || audit.ExpectedActualReplay != 3 || audit.Decision != "PASS" || !audit.ArtifactObserved {
		t.Fatalf("audit=%+v", audit)
	}
}

func TestConsumerRejectsCoherentResealedTamper(t *testing.T) {
	report, err := intervention.Build([]byte(consumerTestSource), consumerTestHead)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	audit, err := VerifyReport(raw, []byte(consumerTestSource), consumerTestHead, testDependency(t, report))
	if err != nil {
		t.Fatal(err)
	}
	if audit.CoherentTamperRejected != 1 || audit.ExpectedCoherentTamperRejections != 1 {
		t.Fatalf("tamper regression=%+v", audit)
	}
}
