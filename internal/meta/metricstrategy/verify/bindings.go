package metricstrategyverify

import (
	"fmt"
	"sort"
	"strings"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

func replayBindings(indicators []metric.Indicator) ([]strategy.Binding, error) {
	bindings := make([]strategy.Binding, 0, len(indicators))
	seen := make(map[string]bool, len(indicators))
	for _, indicator := range indicators {
		family := strings.ToUpper(indicator.Family)
		if seen[indicator.ID] || !replayChoice(family) || indicator.MetaOperation == "" || indicator.EvidenceDigest == "" {
			return nil, fmt.Errorf("independent indicator %q is invalid", indicator.ID)
		}
		seen[indicator.ID] = true
		bindings = append(bindings, strategy.Binding{IndicatorID: indicator.ID, Family: family, Trilemma: indicator.Trilemma, MetaOperation: indicator.MetaOperation, Expected: indicator.Expected, Actual: indicator.Actual, Status: indicator.Status, EvidenceDigest: indicator.EvidenceDigest})
	}
	sort.Slice(bindings, func(i, j int) bool { return bindings[i].IndicatorID < bindings[j].IndicatorID })
	return bindings, nil
}

func replayChoices() []string { return []string{"FOUNDATION", "COHERENCE", "REGRESSION"} }

func replayChoice(value string) bool {
	for _, choice := range replayChoices() {
		if value == choice {
			return true
		}
	}
	return false
}

