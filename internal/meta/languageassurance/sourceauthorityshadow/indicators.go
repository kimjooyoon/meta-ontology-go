package sourceauthorityshadow

func buildIndicators(receipt Receipt, headBound bool) []Indicator {
	evaluation := receipt.Evaluation
	recorded := evaluation.ReceiptDigest != ""
	backed := evaluation.Observation == "SATISFIED" && evaluation.Summary.CoverageBPS == 10000
	preserved := evaluation.Observation != "UNKNOWN" ||
		(receipt.Observation == "UNKNOWN" && receipt.Resolution == "INVARIANT_ONLY" && receipt.Enforcement == "BLOCK")
	return []Indicator{
		metric("gooo.metric.semantic.source-authority-shadow-recording.coverage-bps.v1", "OUTCOME", "COHERENCE",
			"sourceauthorityshadow.Observe", "source-authority-shadow-ci", "record-shadow-observation",
			basis(recorded), 10000, "basis_points", "GREATER_OR_EQUAL", receipt.Resolution, recorded),
		metric(evaluation.MetricID, "DRIVER", evaluation.ProofChoice, "sourceauthorityeval.Observe",
			"sourceauthorityshadow.Observe", evaluation.MetaOperation, evaluation.Summary.CoverageBPS,
			10000, "basis_points", "GREATER_OR_EQUAL", evaluation.Resolution, backed),
		metric("gooo.metric.evidence.source-authority-shadow-head.coverage-bps.v1", "DRIVER", "FOUNDATION",
			"sourceauthorityshadow.Observe", "source-authority-shadow-ci", "bind-exact-shadow-head",
			basis(headBound), 10000, "basis_points", "GREATER_OR_EQUAL", receipt.Resolution, headBound),
		metric("gooo.metric.epistemic.source-authority-shadow-unknown-laundering.v1", "GUARDRAIL", "REGRESSION",
			"sourceauthorityshadow.Observe", "language-assurance-gate", "preserve-shadow-unknown",
			basis(!preserved), 0, "paths", "LESS_OR_EQUAL", receipt.Resolution, preserved),
		metric("gooo.metric.semantic.source-authority-shadow-promotion-credit.v1", "GUARDRAIL", "COHERENCE",
			"sourceauthorityshadow.Observe", "language-assurance-readiness", "deny-shadow-promotion-credit",
			receipt.PromotionCreditBPS, 0, "basis_points", "LESS_OR_EQUAL", "EXACT", true),
		metric("gooo.metric.effects.source-authority-shadow-writes.v1", "GUARDRAIL", "FOUNDATION",
			"sourceauthorityshadow.Observe", "source-authority-shadow-ci", "preserve-shadow-read-only",
			receipt.RepositoryWrites, 0, "repository_writes", "LESS_OR_EQUAL", "EXACT", true),
	}
}

func metric(id, class, proof, producer, consumer, operation string, value, target int,
	unit, relation, resolution string, satisfied bool) Indicator {
	return Indicator{MetricID: id, Class: class, ProofChoice: proof, Producer: producer,
		Consumer: consumer, MetaOperation: operation, Value: value, Target: target,
		Unit: unit, Relation: relation, Resolution: resolution, Satisfied: satisfied}
}

func basis(satisfied bool) int {
	if satisfied {
		return 10000
	}
	return 0
}
