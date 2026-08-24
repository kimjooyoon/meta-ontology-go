package main

import (
	"fmt"
	"os"
)

const rawMetricID = "gooo.metric.evidence.raw-reconstruction.v1"

func transformRegistry(path string, spec metricSpec) ([]byte, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	files, file, err := parseSource(path, source)
	if err != nil {
		return nil, err
	}
	if !hasString(file, spec.MetricID) {
		return nil, fmt.Errorf("write-set denominator entry not found")
	}
	if err := addOperatingOperation(file, spec); err != nil {
		return nil, err
	}
	if err := addMetaOperation(file, spec); err != nil {
		return nil, err
	}
	return formatSource(files, file)
}
