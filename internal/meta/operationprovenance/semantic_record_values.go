package operationprovenance

import (
	"fmt"
	"strings"
)

func metricFromFields(fields map[string]string) (metricSpec, error) {
	keys := []string{"id", "family", "prior_claim", "producer", "consumer", "meta_operation", "evidence_path", "depends_on"}
	if len(fields) != len(keys) {
		return metricSpec{}, fmt.Errorf("metric record has unexpected field cardinality %d", len(fields))
	}
	for _, key := range keys {
		if _, ok := fields[key]; !ok {
			return metricSpec{}, fmt.Errorf("metric record is missing %s", key)
		}
	}
	metric := metricSpec{ID: fields["id"], Family: fields["family"], PriorClaim: fields["prior_claim"], Producer: fields["producer"], Consumer: fields["consumer"], MetaOperation: fields["meta_operation"], EvidencePath: fields["evidence_path"]}
	if fields["depends_on"] != "" {
		metric.DependsOn = strings.Split(fields["depends_on"], ",")
	}
	if metric.ID == "" || metric.Family == "" || metric.PriorClaim == "" {
		return metricSpec{}, fmt.Errorf("metric identity, family, and prior claim are required")
	}
	return metric, nil
}

func scenarioFromFields(fields map[string]string) (scenarioSpec, error) {
	if len(fields) != 4 {
		return scenarioSpec{}, fmt.Errorf("scenario record has unexpected field cardinality %d", len(fields))
	}
	for _, key := range []string{"id", "remove_relation", "dependency", "reason"} {
		if _, ok := fields[key]; !ok {
			return scenarioSpec{}, fmt.Errorf("scenario record is missing %s", key)
		}
	}
	if fields["id"] == "" || fields["reason"] == "" {
		return scenarioSpec{}, fmt.Errorf("scenario identity and reason are required")
	}
	return scenarioSpec{ID: fields["id"], RemoveRelation: fields["remove_relation"], Dependency: fields["dependency"], Reason: fields["reason"]}, nil
}
