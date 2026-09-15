package toolchainusecases

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

const testHead = "0000000000000000000000000000000000000000"

func fixture(t *testing.T) (Registry, []byte) {
	t.Helper()
	raw, err := os.ReadFile("../../../../examples/toolchain-executable-use-cases/usecases.json")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := decodeRegistry(raw)
	if err != nil {
		t.Fatal(err)
	}
	return registry, raw
}

func TestEvaluateExecutesCanonicalAndFailClosedCases(t *testing.T) {
	repository := os.DirFS("../../../..")
	_, raw := fixture(t)
	report := Evaluate(repository, testHead, raw, languageconcept.BuildArtifact(repository))
	if err := Validate(report, testHead); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionPass || report.Resolution != ResolutionExact ||
		report.Summary.Satisfied != 3 || report.Summary.PassPaths != 1 ||
		report.Summary.FailClosedPaths != 2 || report.Summary.Unresolved != 0 {
		t.Fatalf("report = %#v", report)
	}
}
