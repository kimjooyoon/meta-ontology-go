package metricevidence

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func (report Report) GoSplitIndicators() []Indicator {
	indicators := make([]Indicator, 0)
	for _, indicator := range report.Meta.Indicators {
		if !indicator.Satisfied && indicator.MetricID == sourcepolicy.DimensionGoFileLines &&
			indicator.MetaOperation == sourcepolicy.OperationSplitGo {
			indicators = append(indicators, indicator)
		}
	}
	sort.Slice(indicators, func(i, j int) bool { return indicators[i].Subject < indicators[j].Subject })
	return indicators
}

func Contains(indicators []Indicator, subject string) bool {
	for _, indicator := range indicators {
		if indicator.Subject == subject {
			return true
		}
	}
	return false
}
