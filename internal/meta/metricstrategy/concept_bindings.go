package metricstrategy

import (
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

func bindLanguageConcepts(repository fs.FS, source []Binding) ([]Binding, error) {
	value := languageconcept.BuildArtifact(repository)
	if err := languageconcept.ValidateArtifact(repository, value); err != nil {
		return nil, fmt.Errorf("concept-governed strategy artifact: %w", err)
	}
	if !value.Ready() {
		return nil, fmt.Errorf("concept-governed strategy requires PASS, got %s/%s", value.Decision, value.Reason)
	}
	indicators, err := conceptIndicatorBindings(value)
	if err != nil {
		return nil, err
	}
	operations, err := conceptOperationBindings(value, source)
	if err != nil {
		return nil, err
	}
	result := append([]Binding(nil), source...)
	result = append(result, indicators...)
	result = append(result, operations...)
	seen := make(map[string]bool, len(result))
	for _, binding := range result {
		if seen[binding.IndicatorID] {
			return nil, fmt.Errorf("duplicate concept-governed indicator %q", binding.IndicatorID)
		}
		seen[binding.IndicatorID] = true
	}
	sort.Slice(result, func(i, j int) bool { return result[i].IndicatorID < result[j].IndicatorID })
	return result, nil
}

func conceptIndicatorBindings(value languageconcept.Artifact) ([]Binding, error) {
	result := make([]Binding, 0, len(value.Report.Indicators))
	for _, indicator := range value.Report.Indicators {
		family := strings.ToUpper(indicator.ProofChoice)
		carrier, ok := conceptCarriers[family]
		if !ok {
			return nil, fmt.Errorf("concept indicator %q has unknown proof choice", indicator.MetricID)
		}
		expected, actual := strconv.Itoa(indicator.Target), strconv.Itoa(indicator.Value)
		status := "UNSATISFIED"
		if indicator.Satisfied && expected == actual {
			status = "SATISFIED"
		}
		digest, err := conceptEvidenceDigest(map[string]string{"artifact_digest": value.ArtifactDigest, "kind": "INDICATOR", "subject_id": indicator.MetricID, "class": strings.ToUpper(indicator.Class), "proof_choice": family, "carrier_operation": carrier, "expected": expected, "actual": actual, "status": status})
		if err != nil {
			return nil, err
		}
		result = append(result, Binding{IndicatorID: conceptIndicatorID(indicator), Family: family, Trilemma: conceptTrilemma(family), MetaOperation: carrier, Expected: expected, Actual: actual, Status: status, EvidenceDigest: digest})
	}
	return result, nil
}
