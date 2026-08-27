package interventionconsumer

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/intervention"
)

const consumerTestHead = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

const consumerTestSource = `package invarianttransformation
namespace meta
entity Transformation id "gooo://invariant-transformation/value/transformation"
activity PreservedTranslation() -> Transformation computes "case=preserved-translation;input=2;candidate=add:1;expected=3;invariant=candidate-output-equals-expected;replay=present;effect=none"
`

func TestConsumerReconstructsBothInterventions(t *testing.T) {
	report, err := intervention.Build([]byte(consumerTestSource), consumerTestHead)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	audit, err := VerifyReport(raw, []byte(consumerTestSource), consumerTestHead, DependencyBoundary{})
	if err != nil {
		t.Fatal(err)
	}
	if audit.ReconstructedCases != 2 || audit.ExpectedCases != 2 || audit.Decision != "PASS" || audit.RepositoryWrites != 0 || audit.MutationAuthority {
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
	audit, err := VerifyReport(raw, []byte(consumerTestSource), consumerTestHead, DependencyBoundary{})
	if err != nil {
		t.Fatal(err)
	}
	if audit.CoherentTamperRejected != 1 || audit.ExpectedCoherentTamperRejections != 1 {
		t.Fatalf("tamper regression=%+v", audit)
	}
}
