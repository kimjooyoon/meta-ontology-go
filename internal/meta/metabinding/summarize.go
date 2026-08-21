package metabinding

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func summarize(indicators []sourcepolicy.Indicator, index map[string]Binding) (Summary, []Witness) {
	var summary Summary
	witnesses := make(map[string]Witness)
	for _, indicator := range indicators {
		operation := string(indicator.Operation)
		binding, exists := index[operation]
		witness := witnesses[operation]
		if witness.Operation == "" {
			witness = Witness{Binding: binding, Bound: exists}
			if !exists {
				witness.Operation, witness.Registry = operation, "unregistered"
			}
		}
		reasons := bindingReasons(indicator, index)
		witness.IndicatorCount++
		if len(reasons) != 0 {
			witness.Bound = false
			witness.Reasons = mergeReasons(witness.Reasons, reasons)
			summary.UnboundIndicators++
		} else {
			summary.BoundIndicators++
		}
		witnesses[operation] = witness
	}
	summary.UsedOperations = len(witnesses)
	total := summary.BoundIndicators + summary.UnboundIndicators
	if total != 0 {
		summary.CoverageBasisPoints = summary.BoundIndicators * 10000 / total
	}
	result := make([]Witness, 0, len(witnesses))
	for _, witness := range witnesses {
		result = append(result, witness)
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].Operation < result[right].Operation
	})
	return summary, result
}

func mergeReasons(left, right []string) []string {
	seen := make(map[string]bool, len(left)+len(right))
	for _, reason := range append(append([]string(nil), left...), right...) {
		seen[reason] = true
	}
	result := make([]string, 0, len(seen))
	for reason := range seen {
		result = append(result, reason)
	}
	sort.Strings(result)
	return result
}

func selfIndicator(unbound int) sourcepolicy.Indicator {
	return sourcepolicy.Indicator{
		MetricID: sourcepolicy.Dimension(MetricID), Family: sourcepolicy.Family("meta"),
		Subject: ".", SubjectKind: sourcepolicy.SubjectKind("PROJECT_ROOT"),
		Value: unbound, Limit: 0, Relation: sourcepolicy.Relation("less_or_equal"),
		Applicability: sourcepolicy.Applicability("APPLICABLE"),
		ApplicabilityRule: "gooo.catalog.source-policy.default-applicability.v1",
		ApplicabilityReason: sourcepolicy.ApplicabilityReason("CATALOG_APPLICABLE"),
		Blocking: true, Satisfied: unbound == 0, Proof: sourcepolicy.ProofChoice("coherence"),
		Producer: "metabinding.Build", Consumer: "self-improvement-cycle",
		Operation: sourcepolicy.Operation("bind-indicator-meta-program"),
		Detail: "ontology=" + OntologyPath,
	}
}
