package main

const (
	rootREADMEMetric   = "gooo.metric.documentation.root-readme-presence.v1"
	rootREADMEOntology = "examples/root-readme-indicator/main.gooo"
)

func rootREADMEValue(files []fileMetric) int {
	for _, file := range files {
		if file.Path == "README.md" {
			return 1
		}
	}
	return 0
}
