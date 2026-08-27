package languagesyntax_test

import (
	"bytes"
	"io/fs"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

const testHead = "0000000000000000000000000000000000000000"

func fixture(t *testing.T) (fs.FS, []byte) {
	t.Helper()
	repository := os.DirFS("../../../../..")
	raw, err := fs.ReadFile(repository, "examples/language-syntax-roundtrip/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	return repository, raw
}

func TestCompleteCorpusProvesSyntaxRoundTrip(t *testing.T) {
	repository, raw := fixture(t)
	report := languagesyntax.Evaluate(repository, testHead, raw, languageconcept.BuildArtifact(repository))
	if err := languagesyntax.Validate(report, testHead); err != nil {
		t.Fatal(err)
	}
	if report.Decision != languagesyntax.DecisionPass || report.Resolution != languagesyntax.ResolutionExact ||
		report.Summary.Satisfied != 31 || report.Summary.ValidCases != 28 ||
		report.Summary.InvalidCases != 3 || report.Summary.Unresolved != 0 || report.Summary.GoooLines != 457 ||
		len(report.Source.GoooFiles) != 34 || len(report.Source.PackageUnits) != 2 ||
		len(report.Source.PackageUnits[0].Members) != 2 || len(report.Source.PackageUnits[1].Members) != 3 {
		t.Fatalf("report = %#v", report)
	}
}

func TestUnknownRegistryLowersResolution(t *testing.T) {
	repository, canonical := fixture(t)
	unknownField := bytes.Replace(canonical, []byte("{"), []byte(`{"unknown":true,`), 1)
	for _, raw := range [][]byte{[]byte(`{"schema":"UNKNOWN","cases":[]}`), unknownField} {
		report := languagesyntax.Evaluate(repository, testHead, raw, languageconcept.BuildArtifact(repository))
		if err := languagesyntax.Validate(report, testHead); err != nil {
			t.Fatal(err)
		}
		if report.Decision != languagesyntax.DecisionClosed || report.Resolution != languagesyntax.ResolutionLower ||
			report.Summary.Executed != 0 || report.Summary.Unresolved != 30 {
			t.Fatalf("unknown registry was not lowered: %#v", report)
		}
	}
}
