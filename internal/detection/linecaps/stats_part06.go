package linecaps

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func directoryMetricObservations(directory DirectoryMetric) []sourcepolicy.Observation {
	kinds := 0
	if directory.DirectFiles > 0 {
		kinds++
	}
	if directory.DirectFolders > 0 {
		kinds++
	}
	return []sourcepolicy.Observation{
		metricObservation(directory.Path, sourcepolicy.DimensionDirectFiles, directory.DirectFiles),
		metricObservation(directory.Path, sourcepolicy.DimensionDirectFolders, directory.DirectFolders),
		metricObservation(directory.Path, sourcepolicy.DimensionRecursiveFiles, directory.RecursiveFiles),
		metricObservation(directory.Path, sourcepolicy.DimensionRecursiveFolders, directory.RecursiveFolders),
		metricObservation(directory.Path, sourcepolicy.DimensionDirectEntries, directory.DirectFiles+directory.DirectFolders),
		metricObservation(directory.Path, sourcepolicy.DimensionDirectoryKinds, kinds),
	}
}

func directoryDepth(path string) int {
	if path == "." {
		return 0
	}
	return strings.Count(path, "/") + 1
}
