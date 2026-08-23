package languagepackageruntime

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

func TestFixedCorpusProducesExactRuntimeEvidence(t *testing.T) {
	repository := os.DirFS("../../../..")
	raw, err := os.ReadFile("../../../../examples/language-package-runtime/corpus.json")
	if err != nil { t.Fatal(err) }
	report := Evaluate(Input{ExpectedHeadSHA: "0000000000000000000000000000000000000000",
		ConceptArtifact: languageconcept.BuildArtifact(repository), RegistryRaw: raw})
	if err := Validate(report, report.Source.ExpectedHeadSHA); err != nil { t.Fatal(err) }
	if report.Decision != DecisionPass || report.Summary.Satisfied != 18 ||
		report.Summary.PositivePaths != 10 || report.Summary.GuardrailRejections != 8 ||
		len(report.Indicators) != 18 || len(report.Proofs) != 3 {
		t.Fatalf("report = %#v", report)
	}
}
