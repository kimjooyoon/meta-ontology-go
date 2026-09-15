package metricintervention

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metrictransition"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type rootCheck struct {
	id, applicability, reason string
	value                     int
}

func rootEvidenceComplete(report sourcepolicy.Report, counts metrictransition.Counts) bool {
	kinds := 0
	if counts.DirectFiles > 0 {
		kinds++
	}
	if counts.DirectFolders > 0 {
		kinds++
	}
	checks := []rootCheck{
		{"gooo.metric.layout.direct-entries.v1", "NOT_APPLICABLE", "ROOT_TOPOLOGY_EXEMPT", counts.DirectFiles + counts.DirectFolders},
		{"gooo.metric.layout.entry-kinds.v1", "NOT_APPLICABLE", "ROOT_TOPOLOGY_EXEMPT", kinds},
		{"gooo.metric.layout.direct-files.v1", "APPLICABLE", "", counts.DirectFiles},
		{"gooo.metric.layout.direct-folders.v1", "APPLICABLE", "", counts.DirectFolders},
	}
	for _, check := range checks {
		if !hasRootEvidence(report.Indicators, check) {
			return false
		}
	}
	return true
}

func hasRootEvidence(indicators []sourcepolicy.Indicator, check rootCheck) bool {
	for _, indicator := range indicators {
		reasonMatches := check.reason == "" || string(indicator.ApplicabilityReason) == check.reason
		if indicator.Subject == "." && string(indicator.MetricID) == check.id && indicator.Value == check.value && string(indicator.Applicability) == check.applicability && indicator.Satisfied && !indicator.Blocking && reasonMatches {
			return true
		}
	}
	return false
}
