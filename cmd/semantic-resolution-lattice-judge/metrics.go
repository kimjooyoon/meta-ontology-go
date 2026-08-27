package main

import "errors"

func validateMetrics(metrics []metric) error {
	if len(metrics) != 5 {
		return errors.New("metric cardinality is invalid")
	}
	expected := map[string]struct {
		numerator                    int
		class, unit, relation, proof string
	}{
		"gooo.metric.meta-resolution-lattice.exact-observation.count.v1":  {1, "outcome", "cases", "greater_or_equal", "FOUNDATION"},
		"gooo.metric.meta-resolution-lattice.invariant-descent.count.v1":  {1, "driver", "cases", "greater_or_equal", "COHERENCE"},
		"gooo.metric.meta-resolution-lattice.claim-preservation.count.v1": {4, "driver", "cases", "greater_or_equal", "REGRESSION"},
		"gooo.metric.meta-resolution-lattice.replay.count.v1":             {4, "driver", "cases", "greater_or_equal", "REGRESSION"},
		"gooo.metric.meta-resolution-lattice.write-guardrail.v1":          {0, "guardrail", "repository_writes", "less_or_equal", "FOUNDATION"},
	}
	proofs, seen := map[string]bool{}, map[string]bool{}
	for _, item := range metrics {
		want, known := expected[item.ID]
		if !known || seen[item.ID] || item.Numerator != want.numerator || item.Denominator != 4 || item.Numerator < 0 || item.Numerator > 4 || item.Class != want.class || item.Unit != want.unit || item.Relation != want.relation || item.Proof != want.proof || item.Producer == "" || item.Consumer == "" || item.MetaOperation == "" {
			return errors.New("metric is not fixed and provenance-bound")
		}
		seen[item.ID], proofs[item.Proof] = true, true
	}
	if len(proofs) != 3 || len(seen) != len(expected) {
		return errors.New("metric proof trilemma is incomplete")
	}
	return nil
}
