package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type metricSpec struct {
	Schema        string `json:"schema"`
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Activity      string `json:"activity"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
}

func readMetricSpec(path string) (metricSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return metricSpec{}, err
	}
	var spec metricSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return metricSpec{}, err
	}
	if spec.Schema != "gooo/language-assurance-metric-spec/v1" || spec.MetricID == "" || spec.MetaOperation == "" {
		return metricSpec{}, fmt.Errorf("incomplete metric spec")
	}
	return spec, nil
}
