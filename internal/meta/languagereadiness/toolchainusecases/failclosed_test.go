package toolchainusecases

import (
	"bytes"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

func TestUnknownRegistryLowersResolution(t *testing.T) {
	repository := os.DirFS("../../../..")
	_, canonical := fixture(t)
	unknownField := bytes.Replace(canonical, []byte("{"), []byte(`{"unknown":true,`), 1)
	registries := [][]byte{[]byte(`{"schema":"unknown","cases":[]}`), unknownField}
	for _, raw := range registries {
		report := Evaluate(repository, testHead, raw, languageconcept.BuildArtifact(repository))
		if err := Validate(report, testHead); err != nil {
			t.Fatal(err)
		}
		if report.Decision != DecisionClosed || report.Resolution != ResolutionLower ||
			report.Summary.Executed != 0 || report.Summary.Unresolved != totalCases ||
			report.Source.RegistryDigest == registryDigest() || report.Indicators[7].Value != 1 {
			t.Fatalf("report = %#v", report)
		}
		for _, indicator := range report.Indicators {
			if indicator.Resolution != ResolutionLower {
				t.Fatalf("indicator = %#v", indicator)
			}
		}
	}
}

func TestInvalidCanonicalArtifactFailsClosedExactly(t *testing.T) {
	repository := os.DirFS("../../../..")
	_, raw := fixture(t)
	artifact := languageconcept.BuildArtifact(repository)
	artifact.RepositoryWrites = 1
	report := Evaluate(repository, testHead, raw, artifact)
	if err := Validate(report, testHead); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionClosed || report.Resolution != ResolutionExact ||
		report.Summary.Executed != 3 || report.Summary.Satisfied != 2 ||
		report.Summary.NotSatisfied != 1 {
		t.Fatalf("report = %#v", report)
	}
}
