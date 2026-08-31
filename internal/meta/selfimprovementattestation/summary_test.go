package selfimprovementattestation

import (
	"strings"
	"testing"
)

func TestSummaryProjectsReaderResolutionsWithoutCollapsingThem(t *testing.T) {
	receipt, err := Resolve(validRequest())
	if err != nil {
		t.Fatal(err)
	}
	summary := Summary(receipt)
	wants := []string{
		"| NON_ATTESTING_READER | LOWER_RESOLUTION | 7 | 8 | 8750 |",
		"| ATTESTATION_READER | EXACT | 8 | 8 | 10000 |",
		"producer-attestation OPEN -> DISCHARGED",
		"Open / unknown / false: `0 / 0 / 0`",
		"Repository writes: `0`",
		"whole-language-transport-complete",
	}
	for _, want := range wants {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary is missing %q", want)
		}
	}
}
