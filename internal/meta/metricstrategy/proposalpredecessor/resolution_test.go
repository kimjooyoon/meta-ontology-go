package proposalpredecessor

import (
	"strings"
	"testing"
)

func testResolutionEvidence() ObservationEvidence {
	return ObservationEvidence{
		Schema:           ObservationSchema,
		CachePath:        "/tmp/proposal-observation.json",
		CacheBytes:       1,
		CacheDigest:      "sha256:" + strings.Repeat("0", 64),
		ResponseTotal:    1,
		ResponseConsumed: 1,
	}
}

func TestResealedResolutionContextMismatchFailsClosed(t *testing.T) {
	const (
		repository  = "owner/repository"
		current     = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		predecessor = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	)

	mutations := []struct {
		name   string
		mutate func(*ResolutionReceipt)
	}{
		{"repository", func(receipt *ResolutionReceipt) { receipt.Repository = "other/repository" }},
		{"current head", func(receipt *ResolutionReceipt) { receipt.CurrentHeadSHA = strings.Repeat("c", 40) }},
		{"predecessor", func(receipt *ResolutionReceipt) { receipt.PredecessorSHA = strings.Repeat("d", 40) }},
	}
	for _, test := range mutations {
		t.Run(test.name, func(t *testing.T) {
			receipt, err := BuildResolution(
				repository, current, predecessor, ReasonAPIUnavailable, nil, testResolutionEvidence(),
			)
			if err != nil {
				t.Fatal(err)
			}
			test.mutate(&receipt)
			receipt, err = sealResolution(receipt)
			if err != nil {
				t.Fatal(err)
			}
			if err := ValidateResolution(receipt, repository, current, predecessor); err == nil ||
				err.Error() != "FAIL_CLOSED: proposal predecessor resolution context mismatch" {
				t.Fatalf("context mutation accepted or misclassified: %v", err)
			}
		})
	}
}
