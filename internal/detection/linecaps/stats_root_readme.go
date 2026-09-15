package linecaps

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

const rootREADMEOntology = "examples/root-readme-indicator/main.gooo"

func rootREADMEObservation(files []FileMetric) sourcepolicy.Observation {
	value := 0
	for _, file := range files {
		if file.Path == "README.md" {
			value = 1
			break
		}
	}
	observation := metricObservation(".", sourcepolicy.DimensionRootREADME, value)
	observation.Detail = "ontology=" + rootREADMEOntology
	return observation
}
