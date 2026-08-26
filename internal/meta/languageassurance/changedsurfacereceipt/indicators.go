package changedsurfacereceipt

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

func buildIndicators(summary Summary, resolution string) []Indicator {
	ambiguous := summary.OrphanReceipts + summary.ChangedDuplicates + summary.ReceiptDuplicates + summary.MalformedPaths
	unknown := summary.MissingReceipts + summary.UnknownReceipts
	return []Indicator{
		indicator(MetricID, "OUTCOME", "COHERENCE", MetaOperation, "basis_points", "GREATER_OR_EQUAL", summary.TotalityBPS, 10000, resolution),
		indicator("gooo.metric.evidence.changed-surface-set-bps.v1", "DRIVER", "FOUNDATION", "observe-changed-surfaces", "basis_points", "GREATER_OR_EQUAL", summary.ChangedSetBPS, 10000, resolution),
		indicator("gooo.metric.evidence.unique-surface-receipt-bps.v1", "DRIVER", "FOUNDATION", "bind-surface-receipts", "basis_points", "GREATER_OR_EQUAL", summary.UniqueBindingBPS, 10000, resolution),
		indicator("gooo.metric.epistemic.surface-receipt-unknown.v1", "GUARDRAIL", "REGRESSION", "preserve-receipt-unknown", "paths", "LESS_OR_EQUAL", unknown, 0, resolution),
		indicator("gooo.metric.semantic.surface-receipt-ambiguity.v1", "GUARDRAIL", "COHERENCE", "reject-ambiguous-receipts", "paths", "LESS_OR_EQUAL", ambiguous, 0, resolution),
		indicator("gooo.metric.effects.surface-receipt-shadow-writes.v1", "GUARDRAIL", "FOUNDATION", "observe-shadow-writes", "writes", "LESS_OR_EQUAL", 0, 0, ResolutionExact),
	}
}

func indicator(id, class, proof, operation, unit, relation string, value, target int, resolution string) Indicator {
	satisfied := value >= target
	if relation == "LESS_OR_EQUAL" {
		satisfied = value <= target
	}
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "changedsurfacereceipt.Evaluate", Consumer: "changed-surface-receipt-shadow-gate",
		MetaOperation: operation, Unit: unit, Relation: relation, Resolution: resolution,
		Value: value, Target: target, Satisfied: satisfied}
}
