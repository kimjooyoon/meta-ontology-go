package verticalsliceclosureshadow

func buildIndicators(summary Summary) []Indicator {
	return []Indicator{
		indicator("boundary-closure", "OUTCOME", "COHERENCE",
			summary.BoundariesSatisfied, boundaryTotal,
			summary.BoundariesSatisfied == boundaryTotal),
		indicator("boundary-evidence", "DRIVER", "FOUNDATION",
			summary.EvidenceAvailable, boundaryTotal,
			summary.EvidenceAvailable == boundaryTotal),
		indicator("transitive-links", "DRIVER", "COHERENCE",
			summary.LinksSatisfied, linkTotal, summary.LinksSatisfied == linkTotal),
		indicator("unknown-top-decisions", "GUARDRAIL", "FOUNDATION",
			summary.UnknownTopDecisions, 0, summary.UnknownTopDecisions == 0),
		indicator("observed-repository-writes", "GUARDRAIL", "REGRESSION",
			summary.ObservedRepositoryWrites, 0, summary.ObservedRepositoryWrites == 0),
		indicator("shadow-promotion-applied", "GUARDRAIL", "FOUNDATION", 0, 0, true),
	}
}

func indicator(id, class, proof string, value, target int, satisfied bool) Indicator {
	return Indicator{MetricID: "gooo.metric.capability.vertical-slice-closure." + id + ".v1",
		Class: class, ProofChoice: proof, Producer: "verticalsliceclosureshadow.Evaluate",
		Consumer: "language-assurance-promotion-gate", MetaOperation: MetaOperation,
		Value: value, Target: target, Satisfied: satisfied}
}
