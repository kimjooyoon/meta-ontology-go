package main

import "errors"

func validateMetrics(metrics []metric, counterfactuals []counterfactual, claims []claim) error {
	if len(metrics) != 8 {
		return errors.New("metric cardinality is invalid")
	}
	decisionChanges, claimChanges, semanticInfluence, preservedClaims := 0, 0, 0, 0
	for _, item := range claims {
		if item.Preserved {
			preservedClaims++
		}
	}
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
	expected := map[string]struct {
		numerator, denominator        int
		class, unit, relation, proof  string
		producer, consumer, operation string
	}{
		"gooo.metric.meta-resolution-lattice.exact-observation.count.v1":             {1, 4, "outcome", "cases", "greater_or_equal", "FOUNDATION", "semanticresolution.ResolvePartialObservation", "semantic-resolution-lattice-judge", "observe-exact-or-partial"},
		"gooo.metric.meta-resolution-lattice.invariant-descent.count.v1":             {1, 4, "driver", "cases", "greater_or_equal", "COHERENCE", "semanticresolution.ResolvePartialObservation", "semantic-resolution-lattice-judge", "lower-to-invariant-only"},
		"gooo.metric.meta-resolution-lattice.claim-preservation.count.v1":            {preservedClaims, 4, "driver", "cases", "greater_or_equal", "REGRESSION", "semanticresolution.ReplayPartialObservation", "semantic-resolution-lattice-judge", "preserve-claim-state"},
		"gooo.metric.meta-resolution-lattice.replay.count.v1":                        {4, 4, "driver", "cases", "greater_or_equal", "REGRESSION", "semanticresolution.ReplayPartialObservation", "semantic-resolution-lattice-judge", "replay-resolution-descent"},
		"gooo.metric.meta-resolution-lattice.write-guardrail.v1":                     {0, 4, "guardrail", "repository_writes", "less_or_equal", "FOUNDATION", "semanticresolution.ResolvePartialObservation", "semantic-resolution-lattice-judge", "preserve-read-only-resolution"},
		"gooo.metric.meta-resolution-lattice.semantic-decision-influence.v1":         {decisionChanges, 2, "outcome", "counterfactuals", "greater_or_equal", "COHERENCE", "semanticresolution.CasesFromGoooSource", "semantic-resolution-lattice-judge", "reconstruct-decision-from-gooo-value"},
		"gooo.metric.meta-resolution-lattice.semantic-claim-transition-influence.v1": {claimChanges, 2, "driver", "counterfactuals", "greater_or_equal", "REGRESSION", "semanticresolution.ClaimsFromGoooSource", "semantic-resolution-lattice-judge", "reconstruct-claim-transition-from-gooo-value"},
		"gooo.metric.meta-resolution-lattice.semantic-influence.v1":                  {semanticInfluence, 2, "outcome", "counterfactuals", "greater_or_equal", "COHERENCE", "semanticresolution.CasesFromGoooSource", "semantic-resolution-lattice-judge", "reconstruct-decision-and-claim-transition"},
	}
	proofs, seen := map[string]bool{}, map[string]bool{}
	for _, item := range metrics {
		want, known := expected[item.ID]
		if !known || seen[item.ID] || item.Numerator != want.numerator || item.Denominator != want.denominator || item.Numerator < 0 || item.Numerator > item.Denominator || item.Class != want.class || item.Unit != want.unit || item.Relation != want.relation || item.Proof != want.proof || item.Producer != want.producer || item.Consumer != want.consumer || item.MetaOperation != want.operation {
			return errors.New("metric is not fixed, semantic, and provenance-bound")
		}
		seen[item.ID], proofs[item.Proof] = true, true
	}
	if len(proofs) != 3 || len(seen) != len(expected) {
		return errors.New("metric proof trilemma is incomplete")
	}
	return nil
}
