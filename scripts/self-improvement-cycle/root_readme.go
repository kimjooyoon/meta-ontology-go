package main

const (
	rootREADMEMetric   = "gooo.metric.documentation.root-readme-presence.v1"
	rootREADMEOntology = "examples/root-readme-indicator/main.gooo"
)

func rootREADMEValue(files []metricsFile) int {
	for _, file := range files {
		if file.Path == "README.md" {
			return 1
		}
	}
	return 0
}

func rootREADMEException(indicator metricsIndicator, binding MetricsBinding) bool {
	return indicator.MetricID == rootREADMEMetric && indicator.Value == binding.RootREADMEValue &&
		indicator.Detail == "ontology="+binding.RootREADMEOntology && indicator.Applicability == "NOT_APPLICABLE" &&
		indicator.ApplicabilityReason == "ROOT_README_EXEMPT" && !indicator.Blocking &&
		indicator.Decision == "NOT_APPLICABLE"
}

func rootTopologyException(indicator metricsIndicator, root MetricsSnapshot) bool {
	expected := root.DirectFolders + root.DirectFiles
	if indicator.MetricID == "gooo.metric.layout.entry-kinds.v1" {
		expected = 0
		if root.DirectFolders > 0 {
			expected++
		}
		if root.DirectFiles > 0 {
			expected++
		}
	} else if indicator.MetricID != "gooo.metric.layout.direct-entries.v1" {
		return false
	}
	return indicator.Value == expected && indicator.Applicability == "NOT_APPLICABLE" &&
		indicator.ApplicabilityReason == "ROOT_TOPOLOGY_EXEMPT" && !indicator.Blocking &&
		indicator.Decision == "NOT_APPLICABLE"
}
