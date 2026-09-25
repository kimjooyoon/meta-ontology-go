package toolchaincli

import (
	"testing"

	cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

func TestReplayIgnoresRunnerScopedPeakRSS(t *testing.T) {
	first := cliruntime.Observation{Arguments: []string{"version"}, PeakRSSKiB: 11}
	replay := first
	replay.PeakRSSKiB = 17
	if !deterministicReplayEqual(first, replay) {
		t.Fatal("runner-scoped peak RSS changed deterministic replay identity")
	}
}
