package toolchaincli

import "reflect"

// DeterministicReplayEqual compares the process contract while excluding the
// runner-scoped peak RSS. The original observations keep their resource values.
func DeterministicReplayEqual(first, replay Observation) bool {
	first.PeakRSSKiB = 0
	replay.PeakRSSKiB = 0
	return reflect.DeepEqual(first, replay)
}
