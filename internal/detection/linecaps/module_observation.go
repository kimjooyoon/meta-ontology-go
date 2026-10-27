package linecaps

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

// Bind aggregate subjects through their structured physical origins, never
// by parsing a display string or assuming the repository root owns them.
func bindModuleObservation(observation sourcepolicy.Observation, paths []string,
	files []FileMetric, byPath map[string]int) sourcepolicy.Observation {
	if len(paths) == 0 {
		observation.ModuleBoundaryReason = sourcepolicy.ApplicabilityReasonModuleOwnerUnavailable
		return observation
	}
	for index, path := range paths {
		fileIndex, ok := byPath[path]
		if !ok {
			observation.ModuleBoundaryReason = sourcepolicy.ApplicabilityReasonModuleOwnerUnavailable
			return observation
		}
		file := files[fileIndex]
		observation.GeneratorOwned = observation.GeneratorOwned || file.GeneratorOwned
		if index == 0 {
			observation.ModuleRoot = file.ModuleRoot
		} else if file.ModuleRoot != observation.ModuleRoot {
			observation.ModuleRoot = ""
			observation.ModuleBoundaryReason = sourcepolicy.ApplicabilityReasonCrossModuleSource
			return observation
		}
		if file.ModuleBoundaryReason != "" {
			observation.ModuleBoundaryReason = file.ModuleBoundaryReason
		}
	}
	return observation
}
