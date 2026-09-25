package toolchaincli

import (
	"reflect"

	cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

func deterministicReplayEqual(first, replay cliruntime.Observation) bool {
	first.PeakRSSKiB = 0
	replay.PeakRSSKiB = 0
	return reflect.DeepEqual(first, replay)
}
