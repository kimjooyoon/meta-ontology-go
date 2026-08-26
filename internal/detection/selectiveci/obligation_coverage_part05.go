package selectiveci

import (
	"math"
)

func coverageCommandRecords(required []string, registry Registry) (uint64, uint64, CoverageReason) {
	commands := make(map[string]struct{}, len(registry.Commands))
	for _, command := range registry.Commands {
		commands[command.ID] = struct{}{}
	}
	bindings := bindingIndex(registry.Obligations)
	bound := make(map[string]struct{})
	var scanned uint64
	for _, obligation := range required {
		binding, ok := bindings[obligation]
		if !ok {
			return uint64(len(bound)), scanned, CoverageReasonMissingObligation
		}
		if len(binding.CommandIDs) == 0 {
			return uint64(len(bound)), scanned, CoverageReasonMissingCommand
		}
		for _, command := range binding.CommandIDs {
			var ok bool
			scanned, ok = coverageAdd(scanned, 1)
			if !ok {
				return uint64(len(bound)), 0, CoverageReasonWorkOverflow
			}
			if _, registered := commands[command]; !registered {
				return uint64(len(bound)), scanned, CoverageReasonDanglingCommand
			}
			bound[command] = struct{}{}
		}
	}
	return uint64(len(bound)), scanned, ""
}
func coverageWorkUnits(roots, obligations, commandBindings uint64) (uint64, bool) {
	total, ok := coverageAdd(roots, obligations)
	if !ok {
		return 0, false
	}
	return coverageAdd(total, commandBindings)
}
func coverageAdd(left, right uint64) (uint64, bool) {
	if math.MaxUint64-left < right {
		return 0, false
	}
	return left + right, true
}
