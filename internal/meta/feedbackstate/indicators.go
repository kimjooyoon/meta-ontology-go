package feedbackstate

func indicators(observed observation, ready bool) []Indicator {
	return []Indicator{
		metric("semantic-snapshot.readiness-bps", "outcome", 10000, "basis_points", "greater_or_equal", "coherence", "Evaluate", "select-predecessor-semantic-state", boolBPS(ready), ready),
		metric("semantic-snapshot.identity-bps", "driver", 10000, "basis_points", "greater_or_equal", "foundation", "bindReceipt", "bind-predecessor-semantic-evidence", boolBPS(observed.identity), observed.identity),
		metric("semantic-snapshot.payload-bps", "driver", 10000, "basis_points", "greater_or_equal", "foundation", "bindReceipt", "bind-predecessor-payload", boolBPS(observed.payload), observed.payload),
		metric("semantic-snapshot.replay-bps", "driver", 10000, "basis_points", "greater_or_equal", "regression", "bindReceipt", "replay-predecessor-semantic-receipt", boolBPS(observed.replay), observed.replay),
		metric("semantic-snapshot.transition-bps", "driver", 10000, "basis_points", "greater_or_equal", "coherence", "resolve", "validate-predecessor-semantic-transition", boolBPS(observed.semantic), observed.semantic),
		metric("semantic-snapshot.false-fixed-point", "guardrail", 0, "decisions", "less_or_equal", "coherence", "resolve", "reject-false-fixed-point", observed.falseFixed, observed.falseFixed == 0),
		metric("semantic-snapshot.descents", "guardrail", 2, "descents", "less_or_equal", "foundation", "resolve", "bound-semantic-resolution", observed.descents, observed.descents <= 2),
		metric("semantic-snapshot.writes", "guardrail", 0, "repository_writes", "less_or_equal", "foundation", "Evaluate", "preserve-read-only-semantic-state", observed.writes, observed.writes == 0),
	}
}

func metric(id, class string, target int, unit, relation, choice, producer, operation string, value int, satisfied bool) Indicator {
	return Indicator{
		MetricID: "gooo.metric.meta." + id + ".v1", Class: class, Target: target,
		Unit: unit, Relation: relation, ProofChoice: choice,
		Producer: "feedbackstate." + producer, Consumer: "self-improvement-cycle",
		MetaOperation: operation, Value: value, Satisfied: satisfied,
	}
}
