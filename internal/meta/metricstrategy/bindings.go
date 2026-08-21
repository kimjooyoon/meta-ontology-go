package metricstrategy

import (
	"fmt"
	"sort"
	"strings"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
)

func buildBindings(indicators []metric.Indicator) ([]Binding, error) {
	bindings := make([]Binding, 0, len(indicators))
	seen := make(map[string]bool, len(indicators))
	for _, indicator := range indicators {
		family := strings.ToUpper(indicator.Family)
		if seen[indicator.ID] || !knownChoice(family) || indicator.MetaOperation == "" || indicator.EvidenceDigest == "" {
			return nil, fmt.Errorf("metric strategy indicator %q is invalid", indicator.ID)
		}
		seen[indicator.ID] = true
		bindings = append(bindings, Binding{IndicatorID: indicator.ID, Family: family, Trilemma: indicator.Trilemma, MetaOperation: indicator.MetaOperation, Expected: indicator.Expected, Actual: indicator.Actual, Status: indicator.Status, EvidenceDigest: indicator.EvidenceDigest})
	}
	sort.Slice(bindings, func(i, j int) bool { return bindings[i].IndicatorID < bindings[j].IndicatorID })
	return bindings, nil
}

func knownChoice(value string) bool {
	for _, choice := range proofChoices() {
		if value == choice {
			return true
		}
	}
	return false
}

