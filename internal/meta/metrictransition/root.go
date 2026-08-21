package metrictransition

import (
	"encoding/json"
	"fmt"
)

type metricMetaEnvelope struct {
	Meta struct {
		Indicators []rootIndicator `json:"indicators"`
	} `json:"meta"`
}

type rootIndicator struct {
	MetricID           string `json:"metric_id"`
	Subject            string `json:"subject"`
	Value              int    `json:"value"`
	Applicability      string `json:"applicability"`
	ApplicabilityReason string `json:"applicability_reason"`
	Blocking           bool   `json:"blocking"`
	Satisfied          bool   `json:"satisfied"`
}

func rootPolicy(raw []byte, counts Counts) (RootPolicy, error) {
	var envelope metricMetaEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return RootPolicy{}, fmt.Errorf("decode root metric evidence: %w", err)
	}
	kinds := 0
	if counts.DirectFiles > 0 {
		kinds++
	}
	if counts.DirectFolders > 0 {
		kinds++
	}
	checks := []struct {
		id, applicability, reason string
		value                     int
	}{
		{"gooo.metric.layout.direct-entries.v1", "NOT_APPLICABLE", "ROOT_TOPOLOGY_EXEMPT", counts.DirectFiles + counts.DirectFolders},
		{"gooo.metric.layout.entry-kinds.v1", "NOT_APPLICABLE", "ROOT_TOPOLOGY_EXEMPT", kinds},
		{"gooo.metric.layout.direct-files.v1", "APPLICABLE", "", counts.DirectFiles},
		{"gooo.metric.layout.direct-folders.v1", "APPLICABLE", "", counts.DirectFolders},
	}
	for _, check := range checks {
		if !hasRootIndicator(envelope.Meta.Indicators, check.id, check.applicability, check.reason, check.value) {
			return RootPolicy{}, fmt.Errorf("root metric evidence is incomplete: %s", check.id)
		}
	}
	return RootPolicy{Subject: ".", SubjectKind: "PROJECT_ROOT", CountsApplicability: "OBSERVED", TopologyApplicability: "NOT_APPLICABLE", TopologyReason: "ROOT_TOPOLOGY_EXEMPT", READMERequirement: "NOT_APPLICABLE"}, nil
}

func hasRootIndicator(indicators []rootIndicator, id, applicability, reason string, value int) bool {
	for _, indicator := range indicators {
		if indicator.Subject == "." && indicator.MetricID == id && indicator.Value == value && indicator.Applicability == applicability && indicator.Satisfied && !indicator.Blocking && (reason == "" || indicator.ApplicabilityReason == reason) {
			return true
		}
	}
	return false
}
