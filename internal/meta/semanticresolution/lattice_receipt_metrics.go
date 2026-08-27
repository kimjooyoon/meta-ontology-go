package semanticresolution

import (
	"fmt"
	"strings"
)

func validateMetrics(metrics []LatticeMetric) error {
	if len(metrics) != 5 {
		return fmt.Errorf("lattice metric cardinality is invalid")
	}
	proofs := map[ProofLevel]bool{}
	seen := map[string]bool{}
	for _, item := range metrics {
		if item.ID == "" || seen[item.ID] || item.Denominator != LatticeCaseDenominator || item.Numerator < 0 || item.Numerator > item.Denominator || item.Producer == "" || item.Consumer == "" || item.MetaOperation == "" {
			return fmt.Errorf("lattice metric binding is invalid")
		}
		if item.Proof != ProofLevelFoundation && item.Proof != ProofLevelCoherence && item.Proof != ProofLevelRegression {
			return fmt.Errorf("lattice metric proof level is invalid")
		}
		seen[item.ID], proofs[item.Proof] = true, true
	}
	if len(proofs) != 3 || !strings.Contains(metrics[0].ID, "exact-observation") {
		return fmt.Errorf("lattice proof trilemma is incomplete")
	}
	return nil
}
