package changedsurfacereceiptactivation

import eligibility "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/changedsurfacereceipteligibility"

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Resolution    string `json:"resolution"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

func buildIndicators(summary Summary, applied int, resolution string) []Indicator {
	return []Indicator{
		indicator("gooo.metric.semantic.changed-surface-receipt-activation-total.v1", "OUTCOME", "COHERENCE", "apply-changed-surface-receipt-transition", "transitions", "GREATER_OR_EQUAL", applied, 1, resolution),
		indicator("gooo.metric.evidence.changed-surface-receipt-activation-capsule-bps.v1", "DRIVER", "FOUNDATION", "bind-changed-surface-receipt-activation-capsules", "basis_points", "GREATER_OR_EQUAL", summary.CapsuleCoverageBPS, 10000, resolution),
		indicator("gooo.metric.evidence.changed-surface-receipt-predecessor-bps.v1", "DRIVER", "FOUNDATION", "consume-changed-surface-receipt-predecessor", "basis_points", "GREATER_OR_EQUAL", summary.PredecessorSemanticsBPS, 10000, resolution),
		indicator("gooo.metric.epistemic.changed-surface-receipt-activation-unknown.v1", "GUARDRAIL", "REGRESSION", "preserve-changed-surface-receipt-activation-unknown", "paths", "LESS_OR_EQUAL", summary.UnknownPaths, 0, resolution),
		indicator("gooo.metric.effects.changed-surface-receipt-activation-writes.v1", "GUARDRAIL", "FOUNDATION", "preserve-changed-surface-receipt-activation-read-only", "writes", "LESS_OR_EQUAL", 0, 0, ResolutionExact),
		indicator("gooo.metric.semantic.changed-surface-receipt-activation-blocked.v1", "GUARDRAIL", "COHERENCE", "deny-changed-surface-receipt-activation-without-proof", "paths", "LESS_OR_EQUAL", summary.BlockedPaths, 0, resolution),
	}
}

func indicator(id, class, proof, operation, unit, relation string, value, target int, resolution string) Indicator {
	satisfied := value >= target
	if relation == "LESS_OR_EQUAL" {
		satisfied = value <= target
	}
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "changedsurfacereceiptactivation.Evaluate", Consumer: "language-assurance-gate",
		MetaOperation: operation, Unit: unit, Relation: relation, Resolution: resolution,
		Value: value, Target: target, Satisfied: satisfied}
}

func validEligibilityIndicators(indicators []eligibility.Indicator) bool {
	counts := map[string]int{}
	for _, item := range indicators {
		if !item.Satisfied {
			return false
		}
		counts[item.Class]++
	}
	return len(indicators) == 6 && counts["OUTCOME"] == 1 && counts["DRIVER"] == 2 && counts["GUARDRAIL"] == 3
}
