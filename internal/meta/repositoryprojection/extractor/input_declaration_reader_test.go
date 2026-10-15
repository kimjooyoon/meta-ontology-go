package extractor

import "testing"

func TestInputDeclarationReaderExtractionFirst(t *testing.T) {
	targets := []string{
		"cmd/gooo#TestInputDeclarationSavedSourceBinding",
		"cmd/gooo#readInputDeclarationObservation",
		"cmd/gooo#readInputDeclarationArchive",
	}
	// Repository projection may move these declarations to helper files.
	// Observe each declaration's actual rendered capacity in either layout.
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}
