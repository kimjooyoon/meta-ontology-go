package semanticresolution

func canonicalLatticeMetrics(counterfactuals []LatticeCounterfactual) []LatticeMetric {
	decisionChanges, claimChanges, semanticInfluence := 0, 0, 0
	for _, item := range counterfactuals {
		if item.DecisionChanged {
			decisionChanges++
		}
		if item.ClaimTransitionChanged {
			claimChanges++
		}
		if item.Kind == "SEMANTIC" && item.DecisionChanged && item.ClaimTransitionChanged {
			semanticInfluence++
		}
	}
	return []LatticeMetric{
		metric("gooo.metric.meta-resolution-lattice.exact-observation.count.v1", "outcome", 1, "cases", "greater_or_equal", "semanticresolution.ResolvePartialObservation", "semantic-resolution-lattice-judge", "observe-exact-or-partial", ProofLevelFoundation),
		metric("gooo.metric.meta-resolution-lattice.invariant-descent.count.v1", "driver", 1, "cases", "greater_or_equal", "semanticresolution.ResolvePartialObservation", "semantic-resolution-lattice-judge", "lower-to-invariant-only", ProofLevelCoherence),
		metric("gooo.metric.meta-resolution-lattice.claim-preservation.count.v1", "driver", 4, "cases", "greater_or_equal", "semanticresolution.ReplayPartialObservation", "semantic-resolution-lattice-judge", "preserve-claim-state", ProofLevelRegression),
		metric("gooo.metric.meta-resolution-lattice.replay.count.v1", "driver", 4, "cases", "greater_or_equal", "semanticresolution.ReplayPartialObservation", "semantic-resolution-lattice-judge", "replay-resolution-descent", ProofLevelRegression),
		metric("gooo.metric.meta-resolution-lattice.write-guardrail.v1", "guardrail", 0, "repository_writes", "less_or_equal", "semanticresolution.ResolvePartialObservation", "semantic-resolution-lattice-judge", "preserve-read-only-resolution", ProofLevelFoundation),
		metricWithDenominator("gooo.metric.meta-resolution-lattice.semantic-decision-influence.v1", "outcome", decisionChanges, LatticeCounterfactualDenominator, "counterfactuals", "greater_or_equal", "semanticresolution.CasesFromGoooSource", "semantic-resolution-lattice-judge", "reconstruct-decision-from-gooo-value", ProofLevelCoherence),
		metricWithDenominator("gooo.metric.meta-resolution-lattice.semantic-claim-transition-influence.v1", "driver", claimChanges, LatticeCounterfactualDenominator, "counterfactuals", "greater_or_equal", "semanticresolution.ClaimsFromGoooSource", "semantic-resolution-lattice-judge", "reconstruct-claim-transition-from-gooo-value", ProofLevelRegression),
		metricWithDenominator("gooo.metric.meta-resolution-lattice.semantic-influence.v1", "outcome", semanticInfluence, LatticeCounterfactualDenominator, "counterfactuals", "greater_or_equal", "semanticresolution.CasesFromGoooSource", "semantic-resolution-lattice-judge", "reconstruct-decision-and-claim-transition", ProofLevelCoherence),
	}
}
