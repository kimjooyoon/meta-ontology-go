package linecaps

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

func directoryMetricObservations(directory DirectoryMetric, workflowDiscoveryRoot bool, workflowDiscoveryDetail string) []sourcepolicy.Observation {
	kinds := 0
	if directory.DirectFiles > 0 {
		kinds++
	}
	if directory.DirectFolders > 0 {
		kinds++
	}
	observations := []sourcepolicy.Observation{
		metricObservation(directory.Path, sourcepolicy.DimensionDirectFiles, directory.DirectFiles),
		metricObservation(directory.Path, sourcepolicy.DimensionDirectFolders, directory.DirectFolders),
		metricObservation(directory.Path, sourcepolicy.DimensionRecursiveFiles, directory.RecursiveFiles),
		metricObservation(directory.Path, sourcepolicy.DimensionRecursiveFolders, directory.RecursiveFolders),
		metricObservation(directory.Path, sourcepolicy.DimensionDirectEntries, directory.DirectFiles+directory.DirectFolders),
		metricObservation(directory.Path, sourcepolicy.DimensionDirectoryKinds, kinds),
	}
	if workflowDiscoveryRoot {
		for index := range observations {
			switch observations[index].Dimension {
			case sourcepolicy.DimensionDirectEntries, sourcepolicy.DimensionDirectoryKinds:
				observations[index].SemanticRole = sourcepolicy.SemanticRoleWorkflowDiscoveryRoot
			}
		}
	}
	observations[4].Detail = workflowDiscoveryDetail
	return observations
}
