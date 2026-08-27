package verify

import (
	"fmt"
	"strings"
)

func metricFrom(fields map[string]string) (cMetric, error) {
	keys := []string{"id", "family", "prior_claim", "producer", "consumer", "meta_operation", "evidence_path", "depends_on"}
	for _, key := range keys {
		if _, ok := fields[key]; !ok {
			return cMetric{}, fmt.Errorf("consumer metric is missing %s", key)
		}
	}
	metric := cMetric{id: fields["id"], family: fields["family"], claim: fields["prior_claim"], producer: fields["producer"], consumer: fields["consumer"], operation: fields["meta_operation"], evidence: fields["evidence_path"]}
	if fields["depends_on"] != "" {
		metric.dependsOn = strings.Split(fields["depends_on"], ",")
	}
	return metric, nil
}

func scenarioFrom(fields map[string]string) (cScenario, error) {
	for _, key := range []string{"id", "remove_relation", "dependency", "reason"} {
		if _, ok := fields[key]; !ok {
			return cScenario{}, fmt.Errorf("consumer scenario is missing %s", key)
		}
	}
	return cScenario{id: fields["id"], removeRelation: fields["remove_relation"], dependency: fields["dependency"], reason: fields["reason"]}, nil
}
