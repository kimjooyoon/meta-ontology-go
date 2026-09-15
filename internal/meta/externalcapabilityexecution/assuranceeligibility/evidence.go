package assuranceeligibility

import "encoding/json"

func decode(input Input) (evidence, error) {
	var value evidence
	items := []struct {
		name string
		to   any
	}{
		{AssuranceName, &value.Assurance},
		{ParentReportName, &value.ParentReport},
		{ParentObservationName, &value.ParentObservation},
		{ParentSuiteName, &value.ParentSuite},
		{CapabilityReportName, &value.CapabilityReport},
		{CapabilityObservationName, &value.CapabilityObservation},
		{CapabilitySuiteName, &value.CapabilitySuite},
	}
	for _, item := range items {
		if err := json.Unmarshal(input.Payloads[item.name], item.to); err != nil {
			return evidence{}, err
		}
	}
	return value, nil
}

func bindings(input Input) []ArtifactBinding {
	result := make([]ArtifactBinding, 0, len(artifactNames))
	for _, name := range artifactNames {
		payload := input.Payloads[name]
		digest := ""
		if len(payload) != 0 {
			digest = digestBytes(payload)
		}
		result = append(result, ArtifactBinding{Name: name, ObservedDigest: digest})
	}
	return result
}

func markDecoded(report *Report) {
	for index := range report.Artifacts {
		binding := &report.Artifacts[index]
		binding.Exact = binding.ObservedDigest != ""
		if binding.Name == AssuranceName {
			binding.Exact = binding.ObservedDigest == AssuranceDigest
		}
		if binding.Exact {
			report.Summary.EvidenceExact++
		}
	}
}
