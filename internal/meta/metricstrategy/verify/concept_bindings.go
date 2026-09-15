package metricstrategyverify

import (
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

func replayLanguageConceptBindings(repository fs.FS, source []strategy.Binding) ([]strategy.Binding, error) {
	value := languageconcept.BuildArtifact(repository)
	if err := languageconcept.ValidateArtifact(repository, value); err != nil {
		return nil, fmt.Errorf("independent concept strategy artifact: %w", err)
	}
	if !value.Ready() {
		return nil, fmt.Errorf("independent concept strategy requires PASS, got %s/%s", value.Decision, value.Reason)
	}
	indicators, err := replayConceptIndicators(value)
	if err != nil {
		return nil, err
	}
	operations, err := replayConceptOperations(value, source)
	if err != nil {
		return nil, err
	}
	result := append([]strategy.Binding(nil), source...)
	result = append(result, indicators...)
	result = append(result, operations...)
	sort.Slice(result, func(i, j int) bool { return result[i].IndicatorID < result[j].IndicatorID })
	return result, nil
}

func replayConceptIndicators(value languageconcept.Artifact) ([]strategy.Binding, error) {
	result := make([]strategy.Binding, 0, len(value.Report.Indicators))
	for _, indicator := range value.Report.Indicators {
		family := strings.ToUpper(indicator.ProofChoice)
		carrier, ok := replayConceptCarriers[family]
		if !ok {
			return nil, fmt.Errorf("independent concept indicator %q has unknown proof choice", indicator.MetricID)
		}
		expected, actual := strconv.Itoa(indicator.Target), strconv.Itoa(indicator.Value)
		status := "UNSATISFIED"
		if indicator.Satisfied && expected == actual {
			status = "SATISFIED"
		}
		digest, err := replayConceptEvidenceDigest(map[string]string{"artifact_digest": value.ArtifactDigest, "kind": "INDICATOR", "subject_id": indicator.MetricID, "class": strings.ToUpper(indicator.Class), "proof_choice": family, "carrier_operation": carrier, "expected": expected, "actual": actual, "status": status})
		if err != nil {
			return nil, err
		}
		result = append(result, strategy.Binding{IndicatorID: replayConceptIndicatorID(indicator), Family: family, Trilemma: replayConceptTrilemma(family), MetaOperation: carrier, Expected: expected, Actual: actual, Status: status, EvidenceDigest: digest})
	}
	return result, nil
}
