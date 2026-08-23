package languagepackageruntime

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

func TestUnknownCorpusLowersRuntimeResolution(t *testing.T) {
	repository := os.DirFS("../../../..")
	report := Evaluate(Input{ExpectedHeadSHA: "0000000000000000000000000000000000000000",
		ConceptArtifact: languageconcept.BuildArtifact(repository),
		RegistryRaw: []byte(`{"schema":"unknown","cases":[]}`)})
	if report.Decision != DecisionClosed || report.Resolution != ResolutionLower ||
		report.Summary.Unresolved != 18 || report.Source.ObservationKnown {
		t.Fatalf("report = %#v", report)
	}
}
