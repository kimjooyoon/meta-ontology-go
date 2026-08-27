package semanticresolution

func canonicalLatticeMetrics() []LatticeMetric {
	return []LatticeMetric{
		metric("gooo.metric.meta-resolution-lattice.exact-observation.count.v1", "outcome", 1, "cases", "greater_or_equal", "semanticresolution.ResolvePartialObservation", "semantic-resolution-lattice-judge", "observe-exact-or-partial", ProofLevelFoundation),
		metric("gooo.metric.meta-resolution-lattice.invariant-descent.count.v1", "driver", 1, "cases", "greater_or_equal", "semanticresolution.ResolvePartialObservation", "semantic-resolution-lattice-judge", "lower-to-invariant-only", ProofLevelCoherence),
		metric("gooo.metric.meta-resolution-lattice.claim-preservation.count.v1", "driver", 4, "cases", "greater_or_equal", "semanticresolution.ReplayPartialObservation", "semantic-resolution-lattice-judge", "preserve-claim-state", ProofLevelRegression),
		metric("gooo.metric.meta-resolution-lattice.replay.count.v1", "driver", 4, "cases", "greater_or_equal", "semanticresolution.ReplayPartialObservation", "semantic-resolution-lattice-judge", "replay-resolution-descent", ProofLevelRegression),
		metric("gooo.metric.meta-resolution-lattice.write-guardrail.v1", "guardrail", 0, "repository_writes", "less_or_equal", "semanticresolution.ResolvePartialObservation", "semantic-resolution-lattice-judge", "preserve-read-only-resolution", ProofLevelFoundation),
	}
}
